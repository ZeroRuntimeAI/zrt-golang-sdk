package zrt

import (
	"cmp"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

const beforeLLMHookTimeout = 4500 * time.Millisecond

// openGRPCChannel dials the runtime and returns a client connection.
func openGRPCChannel(addr, authToken string) (*grpc.ClientConn, error) {
	insecureEnv := strings.ToLower(os.Getenv("ZRT_RUNTIME_INSECURE"))
	useInsecure := insecureEnv == "1" || insecureEnv == "true" || insecureEnv == "yes" || insecureEnv == "on"

	mdInjector := func(ctx context.Context) context.Context {
		return metadata.NewOutgoingContext(ctx, clientMetadata(authToken))
	}
	unary := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(mdInjector(ctx), method, req, reply, cc, opts...)
	}
	stream := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(mdInjector(ctx), desc, cc, method, opts...)
	}

	kaOpt := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	})
	dialOpts := []grpc.DialOption{
		kaOpt,
		grpc.WithUnaryInterceptor(unary),
		grpc.WithStreamInterceptor(stream),
	}
	if useInsecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}
	return grpc.NewClient(addr, dialOpts...)
}

// grpcBridge owns the direct (non-registered) runtime connection for a session.
type grpcBridge struct {
	addr      string
	agent     Agent
	pipeline  *Pipeline
	session   *AgentSession
	room      roomConfigData
	recording *RecordingConfig

	conn     *grpc.ClientConn
	client   pb.AgentRuntimeClient
	sid      string
	outbound chan *pb.ClientEvent

	mu      sync.Mutex
	running bool
	stream  grpc.BidiStreamingClient[pb.ClientEvent, pb.RuntimeEvent]

	customSTTPump *customSTTPump
	customTTSPump *customTTSPump
}

func newGrpcBridge(addr string, agent Agent, pipeline *Pipeline, session *AgentSession, room roomConfigData, recording *RecordingConfig) *grpcBridge {
	return &grpcBridge{
		addr:      addr,
		agent:     agent,
		pipeline:  pipeline,
		session:   session,
		room:      room,
		recording: recording,
		outbound:  make(chan *pb.ClientEvent, 256),
	}
}

func (b *grpcBridge) sessionID() string           { return b.sid }
func (b *grpcBridge) stub() pb.AgentRuntimeClient { return b.client }

func (b *grpcBridge) createSession(ctx context.Context) (string, error) {
	conn, err := openGRPCChannel(b.addr, b.room.AuthToken)
	if err != nil {
		return "", fmt.Errorf("%w: opening gRPC channel: %w", ErrConnection, err)
	}
	b.conn = conn
	b.client = pb.NewAgentRuntimeClient(conn)

	if b.room.SendLogsToDashboard && b.room.AuthToken == "" {
		logger.Warnf("send_logs_to_dashboard=true but auth_token is empty — dashboard analytics will be skipped. Set room auth_token (or ZRT_AUTH_TOKEN).")
	}
	cfg := b.buildSessionConfig()
	resp, err := b.client.CreateSession(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("CreateSession failed: %w", err)
	}
	if rej := resp.GetRejected(); rej != nil {
		return "", fmt.Errorf("%w: %s — %s", ErrSessionRejected, rej.GetReason(), rej.GetMessage())
	}
	b.sid = resp.GetSession().GetSessionId()
	return b.sid, nil
}

func (b *grpcBridge) buildSessionConfig() *pb.SessionConfig {
	sessionOptions := map[string]string{}
	s := b.session
	if s != nil {
		s.mu.Lock()
		if s.waitForParticipant {
			sessionOptions["wait_for_participant"] = "true"
		}
		if s.waitForParticipantTimeout != 0 {
			sessionOptions["wait_for_participant_timeout_ms"] = fmt.Sprintf("%d", s.waitForParticipantTimeout)
		}
		if s.sipHangupOnShutdown {
			sessionOptions["sip_hangup_on_shutdown"] = "true"
		}
		if s.waitForAudioStream {
			sessionOptions["wait_for_audio_stream"] = "true"
		}
		for k, v := range s.runtimeOptions {
			if v != "" {
				sessionOptions[k] = v
			}
		}
		caller := s.callerSessionID
		s.mu.Unlock()
		return buildSessionConfig(b.pipeline, b.agent, b.room, b.recording, sessionOptions, caller)
	}
	return buildSessionConfig(b.pipeline, b.agent, b.room, b.recording, sessionOptions, "")
}

func (b *grpcBridge) destroySession() {
	if b.client != nil && b.sid != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := b.client.DestroySession(ctx, &pb.DestroyRequest{SessionId: b.sid, Reason: "sdk_shutdown"})
		cancel()
		if err != nil {
			logger.Warnf("Error destroying session: %v", err)
		}
	}
	b.mu.Lock()
	b.running = false
	b.mu.Unlock()
	if b.customSTTPump != nil {
		b.customSTTPump.close()
	}
	if b.customTTSPump != nil {
		b.customTTSPump.close()
	}
	if b.conn != nil {
		b.conn.Close()
	}
}

// enqueue pushes an outbound event (non-blocking with a small wait).
func (b *grpcBridge) enqueue(ev *pb.ClientEvent) error {
	t := time.NewTimer(2 * time.Second)
	defer t.Stop()
	select {
	case b.outbound <- ev:
		return nil
	case <-t.C:
		logger.Warnf("outbound queue full — dropping event")
		return fmt.Errorf("outbound queue full")
	}
}

// ---- sessionTransport implementation ----

func (b *grpcBridge) sendSay(text string, interruptCurrent bool, voice string, interruptible bool) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_Say{Say: &pb.SayCommand{Text: text, InterruptCurrent: interruptCurrent, Voice: voice, NonInterruptible: !interruptible}}})
}
func (b *grpcBridge) sendCancelGeneration() error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_CancelGeneration{CancelGeneration: &pb.CancelGenerationCmd{}}})
}
func (b *grpcBridge) sendGenerate(text string) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_Generate{Generate: &pb.GenerateCmd{Text: text}}})
}
func (b *grpcBridge) sendUpdateInstructions(s string) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_UpdateInstructions{UpdateInstructions: &pb.UpdateInstructionsCmd{Instructions: s}}})
}
func (b *grpcBridge) sendUpdateTools(tools []*FunctionTool) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_UpdateTools{UpdateTools: &pb.UpdateToolsCmd{Tools: buildToolSchemas(tools)}}})
}
func (b *grpcBridge) sendUpdateProvider(component, provider string, params map[string]string) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_UpdateProvider{UpdateProvider: &pb.UpdateProviderCmd{Component: component, Provider: provider, Params: copyAnyMap(params)}}})
}
func (b *grpcBridge) sendCallTransfer(transferTo, token string) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_CallTransfer{CallTransfer: &pb.CallTransferCmd{TransferTo: transferTo, Token: token}}})
}
func (b *grpcBridge) sendPublishMessage(topic, message, optionsJSON, payloadJSON string) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_PublishMessage{PublishMessage: &pb.PublishMessageCmd{Topic: topic, Message: message, OptionsJson: optionsJSON, PayloadJson: payloadJSON}}})
}
func (b *grpcBridge) sendSubscribePubSub(topic string) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_SubscribePubsub{SubscribePubsub: &pb.SubscribePubSubCmd{Topic: topic}}})
}
func (b *grpcBridge) sendPlayBackgroundAudio(url string, volume float64, looping, playbackMode bool) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_PlayBackgroundAudio{PlayBackgroundAudio: &pb.PlayBackgroundAudioCmd{FileUrl: url, Volume: float32(volume), Looping: looping, PlaybackMode: playbackMode}}})
}
func (b *grpcBridge) sendPreloadBackgroundAudio(url string, volume float64) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_PreloadBackgroundAudio{PreloadBackgroundAudio: &pb.PreloadBackgroundAudioCmd{FileUrl: url, Volume: float32(volume)}}})
}
func (b *grpcBridge) sendStopBackgroundAudio() error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_StopBackgroundAudio{StopBackgroundAudio: &pb.StopBackgroundAudioCmd{}}})
}
func (b *grpcBridge) sendRecordingStart(cfg *RecordingConfig) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_RecordingStart{RecordingStart: &pb.RecordingStartCmd{Config: buildRecordingConfig(cfg)}}})
}
func (b *grpcBridge) sendRecordingStop() error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_RecordingStop{RecordingStop: &pb.RecordingStopCmd{}}})
}
func (b *grpcBridge) sendPushAudioFrame(pcm []byte, sampleRate int) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_PushAudioFrame{PushAudioFrame: &pb.PushAudioFrameCmd{Pcm: pcm, SampleRate: uint32(sampleRate)}}})
}
func (b *grpcBridge) sendSendImage(mimeType string, data []byte) error {
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_SendImage{SendImage: &pb.SendImageCmd{MimeType: mimeType, Data: data}}})
}
func (b *grpcBridge) sendSendMessageWithFrames(text string, frames []frameData) error {
	var mf []*pb.MessageFrame
	for _, f := range frames {
		mf = append(mf, &pb.MessageFrame{MimeType: f.mimeType, Data: f.data})
	}
	return b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_SendMessageWithFrames{SendMessageWithFrames: &pb.SendMessageWithFramesCmd{Text: text, Frames: mf}}})
}

func (b *grpcBridge) sendToolResult(callID, resultJSON string, isErr bool) {
	_ = b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_ToolResult{ToolResult: &pb.ToolCallResponse{CallId: callID, ResultJson: resultJSON, IsError: isErr}}})
}
func (b *grpcBridge) sendModifyLLMToken(tokenID uint64, replacement string, drop bool) {
	_ = b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_ModifyLlmToken{ModifyLlmToken: &pb.ModifyLLMTokenCmd{TokenId: tokenID, Replacement: replacement, Drop: drop}}})
}
func (b *grpcBridge) sendBeforeLLMResult(requestID, modifiedMessages string, skip bool) {
	_ = b.enqueue(&pb.ClientEvent{SessionId: b.sid, RequestId: requestID, Event: &pb.ClientEvent_BeforeLlmResult{BeforeLlmResult: &pb.BeforeLLMResponse{ModifiedMessagesJson: modifiedMessages, SkipTurn: skip}}})
}
func (b *grpcBridge) sendCustomSTTResult(r CustomSTTResult) {
	_ = b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_CustomSttResult{CustomSttResult: &pb.CustomSttResult{UtteranceId: r.UtteranceID, Text: r.Text, IsFinal: r.IsFinal, Confidence: float32(r.Confidence), Language: r.Language, StartTimeUs: r.StartTimeUS, EndTimeUs: r.EndTimeUS}}})
}
func (b *grpcBridge) sendCustomTTSAudio(c CustomTTSAudioChunk) {
	_ = b.enqueue(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_CustomTtsAudio{CustomTtsAudio: &pb.CustomTtsAudioChunk{UtteranceId: c.UtteranceID, Pcm: c.PCM, SampleRate: c.SampleRate, EndOfSynthesis: c.EndOfSynthesis}}})
}

// ---- event loop ----

func (b *grpcBridge) runEventLoop(ctx context.Context) {
	b.mu.Lock()
	b.running = true
	b.mu.Unlock()

	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := b.client.EventStream(streamCtx)
	if err != nil {
		logger.Errorf("EventStream open failed: %v", err)
		go b.session.Close(context.Background(), "event_loop_error")
		return
	}
	b.stream = stream

	sendDone := make(chan struct{})
	go func() {
		defer close(sendDone)
		_ = stream.Send(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_Keepalive{Keepalive: &pb.Keepalive{}}})
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case ev := <-b.outbound:
				if err := stream.Send(ev); err != nil {
					return
				}
			case <-ticker.C:
				if err := stream.Send(&pb.ClientEvent{SessionId: b.sid, Event: &pb.ClientEvent_Keepalive{Keepalive: &pb.Keepalive{}}}); err != nil {
					return
				}
			case <-streamCtx.Done():
				return
			}
		}
	}()

	shouldClose := true
	closeReason := "stream_closed_by_runtime"
	for {
		ev, err := stream.Recv()
		if err != nil {
			if ctx.Err() == nil {
				logger.Errorf("gRPC stream error: %v", err)
				closeReason = "grpc_error"
			} else {
				shouldClose = false
			}
			break
		}
		b.handleRuntimeEvent(ctx, ev)
	}
	cancel()
	<-sendDone
	b.mu.Lock()
	b.running = false
	b.mu.Unlock()
	if shouldClose {
		go func() {
			if err := b.session.Close(context.Background(), closeReason); err != nil {
				logger.Errorf("session.Close(reason=%s) failed after stream end: %v", closeReason, err)
			}
		}()
	}
}

func (b *grpcBridge) handleRuntimeEvent(ctx context.Context, event *pb.RuntimeEvent) {
	if b.session.AgentState() == AgentStateClosing {
		return
	}
	s := b.session
	switch ev := event.GetEvent().(type) {
	case *pb.RuntimeEvent_ToolCall:
		tc := ev.ToolCall
		go executeTool(ctx, b.agent.base().tools, tc.GetCallId(), tc.GetToolName(), tc.GetArgumentsJson(), b.sendToolResult)
	case *pb.RuntimeEvent_BeforeLlm:
		go b.handleBeforeLLM(event.GetRequestId(), ev.BeforeLlm)
	case *pb.RuntimeEvent_LlmTokenForReview:
		b.handleLLMTokenForReview(ev.LlmTokenForReview)
	case *pb.RuntimeEvent_CustomSttAudio:
		b.handleCustomSTTAudio(ev.CustomSttAudio)
	case *pb.RuntimeEvent_CustomTtsSynthesize:
		b.handleCustomTTSSynthesize(ev.CustomTtsSynthesize)
	case *pb.RuntimeEvent_Error:
		b.handleError(ev.Error)
	case *pb.RuntimeEvent_Transcript:
		transcriptDirect(s, ev.Transcript)
	case *pb.RuntimeEvent_AgentSpeech:
		agentSpeechDirect(s, ev.AgentSpeech)
	case *pb.RuntimeEvent_TurnComplete:
		turnCompleteDirect(s, ev.TurnComplete)
	case *pb.RuntimeEvent_SayComplete:
		evSayComplete(s)
	case *pb.RuntimeEvent_Interrupt:
		evInterrupt(s, ev.Interrupt)
	case *pb.RuntimeEvent_StateChange:
		evStateChange(s, ev.StateChange)
	case *pb.RuntimeEvent_RecordingStatus:
		evRecordingStatus(s, ev.RecordingStatus)
	case *pb.RuntimeEvent_Warning:
		evWarning(s, ev.Warning)
	case *pb.RuntimeEvent_Metrics:
		evMetrics(s, ev.Metrics)
	case *pb.RuntimeEvent_Dtmf:
		evDTMF(s, ev.Dtmf)
	case *pb.RuntimeEvent_VisionFrame:
		evVisionFrame(s, ev.VisionFrame)
	case *pb.RuntimeEvent_AudioFrame:
		evAudioFrame(s, ev.AudioFrame)
	case *pb.RuntimeEvent_StreamEvent:
		evStreamEvent(s, ev.StreamEvent)
	case *pb.RuntimeEvent_SignalingSessionAssigned:
		evSignaling(s, ev.SignalingSessionAssigned.GetSessionId())
	case *pb.RuntimeEvent_VoicemailDetected:
		evVoicemail(s, ev.VoicemailDetected)
	case *pb.RuntimeEvent_A2AMessage:
		evA2A(s, ev.A2AMessage)
	case *pb.RuntimeEvent_PubsubMessage:
		evPubSubMessage(s, ev.PubsubMessage)
	case *pb.RuntimeEvent_AgentSwitched:
		evAgentSwitched(s, ev.AgentSwitched)
	case *pb.RuntimeEvent_KbHits:
		evKBHits(s, ev.KbHits)
	case *pb.RuntimeEvent_Participant:
		evParticipant(s, ev.Participant)
	case *pb.RuntimeEvent_AgentStateChanged:
		evAgentStateChanged(s, ev.AgentStateChanged)
	case *pb.RuntimeEvent_UserStateChanged:
		evUserStateChanged(s, ev.UserStateChanged)
	case *pb.RuntimeEvent_GenerationStarted:
		evGenerationStarted(s, ev.GenerationStarted)
	case *pb.RuntimeEvent_GenerationComplete:
		evGenerationComplete(s, ev.GenerationComplete)
	case *pb.RuntimeEvent_GenerationChunk:
		evGenerationChunk(s, ev.GenerationChunk)
	case *pb.RuntimeEvent_SynthesisStarted:
		evSynthesisStarted(s, ev.SynthesisStarted)
	case *pb.RuntimeEvent_SynthesisInterrupted:
		evSynthesisInterrupted(s, ev.SynthesisInterrupted)
	case *pb.RuntimeEvent_FirstAudioByte:
		evFirstAudioByte(s, ev.FirstAudioByte)
	case *pb.RuntimeEvent_LastAudioByte:
		evLastAudioByte(s, ev.LastAudioByte)
	case *pb.RuntimeEvent_WordTiming:
		evWordTiming(s, ev.WordTiming)
	case *pb.RuntimeEvent_TtsCapabilities:
		evTTSCapabilities(s, ev.TtsCapabilities)
	case *pb.RuntimeEvent_SttStreamStarted:
		evSTTStreamStarted(s, ev.SttStreamStarted)
	case *pb.RuntimeEvent_SttStreamEnded:
		evSTTStreamEnded(s, ev.SttStreamEnded)
	case *pb.RuntimeEvent_VadEvent:
		evVADEvent(s, ev.VadEvent)
	case *pb.RuntimeEvent_TranscriptPreflight:
		evTranscriptPreflight(s, ev.TranscriptPreflight)
	case *pb.RuntimeEvent_EouDetected:
		evEOUDetected(s, ev.EouDetected)
	case *pb.RuntimeEvent_UserTurnStart:
		evUserTurnStart(s, ev.UserTurnStart)
	case *pb.RuntimeEvent_UserTurnEnd:
		evUserTurnEnd(s, ev.UserTurnEnd)
	case *pb.RuntimeEvent_AgentTurnStart:
		evAgentTurnStart(s, ev.AgentTurnStart)
	case *pb.RuntimeEvent_AgentTurnEnd:
		evAgentTurnEnd(s, ev.AgentTurnEnd)
	case *pb.RuntimeEvent_LlmCompleted:
		evLLMCompleted(s, ev.LlmCompleted)
	default:
		logger.Debugf("handleRuntimeEvent: unhandled event %T", event.GetEvent())
	}
}

func (b *grpcBridge) handleError(err *pb.ErrorEvent) {
	logger.Errorf("[%s] %s (fatal=%v)", err.GetComponent(), err.GetMessage(), err.GetIsFatal())
	payload := map[string]any{
		"component": err.GetComponent(), "provider": err.GetProvider(), "severity": err.GetSeverity(),
		"error_code": err.GetErrorCode(), "message": err.GetMessage(), "is_fatal": err.GetIsFatal(),
		"retry_attempts": err.GetRetryAttempts(),
	}
	b.session.Emit("runtime_error", payload)
	for _, h := range b.pipeline.Hooks.errorHooks {
		safeHook("error", func() { h(payload) })
	}
	effectivelyFatal := err.GetIsFatal() || isUnrecoverableAuthError(err.GetComponent(), err.GetMessage())
	if !err.GetIsFatal() && effectivelyFatal {
		logger.Errorf("[%s] treating as fatal (provider-auth error): %s", err.GetComponent(), err.GetMessage())
	}
	if effectivelyFatal {
		b.mu.Lock()
		b.running = false
		b.mu.Unlock()
		comp := cmp.Or(err.GetComponent(), "runtime")
		go b.session.Close(context.Background(), "fatal_error:"+comp)
	}
}

func (b *grpcBridge) handleBeforeLLM(requestID string, blh *pb.BeforeLLMHook) {
	modified, skip := runBeforeLLM(b.pipeline, blh)
	b.sendBeforeLLMResult(requestID, modified, skip)
}

func (b *grpcBridge) handleLLMTokenForReview(t *pb.LLMTokenForReviewEvent) {
	replacement, drop := runLLMTokenReview(b.pipeline, t.GetText(), t.GetTokenId())
	b.sendModifyLLMToken(t.GetTokenId(), replacement, drop)
}

func (b *grpcBridge) handleCustomSTTAudio(msg *pb.CustomSttAudioChunk) {
	hook := b.pipeline.Hooks.customSTT
	if hook == nil {
		return
	}
	if b.customSTTPump == nil {
		b.customSTTPump = newCustomSTTPump(hook, b.sendCustomSTTResult)
	}
	b.customSTTPump.push(CustomSTTAudioChunk{PCM: msg.GetPcm(), SampleRate: msg.GetSampleRate(), UtteranceID: msg.GetUtteranceId(), EndOfUtterance: msg.GetEndOfUtterance()})
}

func (b *grpcBridge) handleCustomTTSSynthesize(msg *pb.CustomTtsSynthesize) {
	hook := b.pipeline.Hooks.customTTS
	if hook == nil {
		return
	}
	if b.customTTSPump == nil {
		b.customTTSPump = newCustomTTSPump(hook, b.sendCustomTTSAudio)
	}
	b.customTTSPump.push(CustomTTSSynthesize{Text: msg.GetText(), UtteranceID: msg.GetUtteranceId(), Voice: msg.GetVoice()})
}

// ---- shared hook runners ----

func runBeforeLLM(p *Pipeline, blh *pb.BeforeLLMHook) (string, bool) {
	hook := p.Hooks.beforeLLM
	if hook == nil {
		return "", false
	}
	var messages []any
	if blh.GetMessagesJson() != "" {
		if err := json.Unmarshal([]byte(blh.GetMessagesJson()), &messages); err != nil {
			logger.Warnf("before_llm: failed to decode messages_json: %v", err)
		}
	}
	data := BeforeLLMData{Messages: messages, TokenCount: blh.GetTokenCount(), TurnNumber: blh.GetTurnNumber()}
	resultCh := make(chan *BeforeLLMResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("before_llm hook panicked: %v", r)
				resultCh <- nil
			}
		}()
		resultCh <- hook(data)
	}()
	select {
	case res := <-resultCh:
		if res == nil {
			return "", false
		}
		if res.SkipTurn {
			return "", true
		}
		if res.Messages != nil {
			b, err := json.Marshal(res.Messages)
			if err != nil {
				logger.Errorf("before_llm hook returned non-serializable messages: %v", err)
				return "", false
			}
			return string(b), false
		}
		return "", false
	case <-time.After(beforeLLMHookTimeout):
		logger.Warnf("before_llm hook exceeded %.1fs SDK timeout; proceeding with original messages", beforeLLMHookTimeout.Seconds())
		return "", false
	}
}

func runLLMTokenReview(p *Pipeline, text string, tokenID uint64) (string, bool) {
	if hook := p.Hooks.llmStream; hook != nil {
		repl, drop := hook(text, tokenID)
		if repl == text {
			repl = ""
		}
		return repl, drop
	}
	cur := text
	dropped := false
	for _, hook := range p.Hooks.llmTokenForReview {
		repl, d := hook(cur, tokenID)
		dropped = dropped || d
		if repl != "" && repl != cur {
			cur = repl
		}
	}
	if cur == text {
		return "", dropped
	}
	return cur, dropped
}
