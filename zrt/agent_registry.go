package zrt

import (
	"cmp"
	"context"
	"strings"
	"sync"
	"time"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
	"google.golang.org/grpc"
)

// registeredSessionState tracks one runtime-dispatched session.
type registeredSessionState struct {
	sessionID          string
	room               map[string]string
	dispatchMetadata   map[string]string
	agentOverride      *pb.AgentConfigOverride
	jobCtx             *JobContext
	agentSession       *AgentSession
	agent              Agent
	cancelEntry        context.CancelFunc
	participantPresent bool
	customSTTPump      *customSTTPump
	customTTSPump      *customTTSPump
	mu                 sync.Mutex
}

// agentRegistry implements registered-agent mode (RegisterAgent bidi stream).
type agentRegistry struct {
	addr              string
	entrypoint        EntrypointFunc
	jobctxFactory     func() *JobContext
	agentTemplate     Agent
	pipeline          *Pipeline
	agentKind         string
	maxConcurrent     int
	labels            map[string]string
	authToken         string
	defaultRecording  *RecordingConfig
	initializeTimeout time.Duration

	conn     *grpc.ClientConn
	client   pb.AgentRuntimeClient
	stream   grpc.BidiStreamingClient[pb.AgentStreamIn, pb.AgentStreamOut]
	outbound chan *pb.AgentStreamIn

	mu             sync.Mutex
	sessions       map[string]*registeredSessionState
	running        bool
	draining       bool
	agentID        string
	registeredCh   chan struct{}
	registeredOnce sync.Once
}

func newAgentRegistry(addr string, entrypoint EntrypointFunc, jobctxFactory func() *JobContext, agentTemplate Agent, pipeline *Pipeline, agentKind string, maxConcurrent int, labels map[string]string, authToken string, defaultRecording *RecordingConfig, initializeTimeout time.Duration) *agentRegistry {
	return &agentRegistry{
		addr:              addr,
		entrypoint:        entrypoint,
		jobctxFactory:     jobctxFactory,
		agentTemplate:     agentTemplate,
		pipeline:          pipeline,
		agentKind:         agentKind,
		maxConcurrent:     maxConcurrent,
		labels:            labels,
		authToken:         authToken,
		defaultRecording:  defaultRecording,
		initializeTimeout: initializeTimeout,
		outbound:          make(chan *pb.AgentStreamIn, 256),
		sessions:          map[string]*registeredSessionState{},
		registeredCh:      make(chan struct{}),
	}
}

func (r *agentRegistry) bindSession(sid string, s *AgentSession) {
	r.mu.Lock()
	state := r.sessions[sid]
	r.mu.Unlock()
	if state != nil {
		state.mu.Lock()
		state.agentSession = s
		state.agent = s.agent
		state.mu.Unlock()
	}
}

func (r *agentRegistry) setSessionAgent(sid string, agent Agent) {
	r.mu.Lock()
	state := r.sessions[sid]
	r.mu.Unlock()
	if state != nil {
		state.mu.Lock()
		state.agent = agent
		state.mu.Unlock()
	}
}

func (r *agentRegistry) enqueue(ev *pb.AgentStreamIn) {
	t := time.NewTimer(2 * time.Second)
	defer t.Stop()
	select {
	case r.outbound <- ev:
	case <-t.C:
		logger.Warnf("registry outbound queue full — dropping event")
	}
}


func (r *agentRegistry) enqueueResult(ev *pb.AgentStreamIn) {
	t := time.NewTimer(28 * time.Second)
	defer t.Stop()
	select {
	case r.outbound <- ev:
	case <-t.C:
		logger.Warnf("registry outbound full for 28s — tool result dropped (turn will time out)")
	}
}

func (r *agentRegistry) run(ctx context.Context) error {
	conn, err := openGRPCChannel(r.addr, r.authToken)
	if err != nil {
		return err
	}
	r.conn = conn
	r.client = pb.NewAgentRuntimeClient(conn)
	r.mu.Lock()
	r.running = true
	r.mu.Unlock()
	defer func() {
		r.closeAllSessions()
		if r.conn != nil {
			r.conn.Close()
		}
	}()
	return r.runStream(ctx)
}

func (r *agentRegistry) runStream(ctx context.Context) error {
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := r.client.RegisterAgent(streamCtx)
	if err != nil {
		return err
	}
	r.stream = stream

	sendDone := make(chan struct{})
	go func() {
		defer close(sendDone)
		reg := buildAgentRegistration(r.agentTemplate, r.pipeline, r.agentKind, r.maxConcurrent, r.authToken, r.labels, r.defaultRecording, nil)
		if err := stream.Send(&pb.AgentStreamIn{Payload: &pb.AgentStreamIn_Register{Register: reg}}); err != nil {
			return
		}
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case ev := <-r.outbound:
				if err := stream.Send(ev); err != nil {
					return
				}
			case <-ticker.C:
				if err := stream.Send(&pb.AgentStreamIn{Payload: &pb.AgentStreamIn_Keepalive{Keepalive: &pb.Keepalive{}}}); err != nil {
					return
				}
			case <-streamCtx.Done():
				return
			}
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			if ctx.Err() == nil {
				logger.Errorf("RegisterAgent stream error: %v", err)
			}
			break
		}
		r.dispatchInbound(ctx, msg)
	}
	cancel()
	<-sendDone
	return nil
}

func (r *agentRegistry) dispatchInbound(ctx context.Context, msg *pb.AgentStreamOut) {
	sid := msg.GetSessionId()
	switch p := msg.GetPayload().(type) {
	case *pb.AgentStreamOut_Registered:
		r.mu.Lock()
		r.agentID = p.Registered.GetAgentId()
		r.mu.Unlock()
		r.registeredOnce.Do(func() { close(r.registeredCh) })
		logger.Infof("Agent registered: agent_id=%s kind=%s capacity=%d runtime=%s", p.Registered.GetAgentId(), r.agentKind, r.maxConcurrent, p.Registered.GetRuntimeVersion())
		return
	case *pb.AgentStreamOut_SessionStarted:
		r.handleSessionStarted(ctx, p.SessionStarted)
		return
	case *pb.AgentStreamOut_SessionEnded:
		r.handleSessionEnded(p.SessionEnded)
		return
	}
	r.mu.Lock()
	state := r.sessions[sid]
	r.mu.Unlock()
	if state == nil {
		logger.Warnf("Received event for unknown session_id=%s; dropping", sid)
		return
	}
	r.routeSessionEvent(ctx, state, msg)
}

func (r *agentRegistry) handleSessionStarted(ctx context.Context, started *pb.SessionStarted) {
	sid := started.GetSessionId()
	r.mu.Lock()
	draining := r.draining
	atCapacity := len(r.sessions) >= r.maxConcurrent
	r.mu.Unlock()
	if draining {
		r.rejectSession(sid, "draining")
		return
	}
	if atCapacity {
		r.rejectSession(sid, "at_capacity")
		return
	}
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_SessionAck{SessionAck: &pb.SessionAck{SessionId: sid, Accepted: true}}})

	room := map[string]string{"room_id": started.GetRoom().GetRoomId(), "auth_token": started.GetRoom().GetAuthToken(), "agent_name": started.GetRoom().GetAgentName()}
	dispatchMeta := copyAnyMap(started.GetDispatchMetadata())
	state := &registeredSessionState{sessionID: sid, room: room, dispatchMetadata: dispatchMeta, agentOverride: started.GetAgentOverride()}
	r.mu.Lock()
	r.sessions[sid] = state
	r.mu.Unlock()

	var jobCtx *JobContext
	if r.jobctxFactory != nil {
		jobCtx = r.jobctxFactory()
	} else {
		jobCtx = NewJobContext(&RoomOptions{RoomID: room["room_id"], Name: room["agent_name"]}, nil)
	}
	if jobCtx.RoomOptions == nil {
		jobCtx.RoomOptions = &RoomOptions{}
	}
	if room["room_id"] != "" {
		jobCtx.RoomOptions.RoomID = room["room_id"]
	}
	if room["auth_token"] != "" {
		jobCtx.RoomOptions.AuthToken = room["auth_token"]
	}
	if room["agent_name"] != "" {
		jobCtx.RoomOptions.Name = room["agent_name"]
	}
	jobCtx.registeredMode = true
	jobCtx.registeredSessionID = sid
	jobCtx.registeredRegistry = r
	jobCtx.Metadata = stringMapToAny(dispatchMeta)
	state.jobCtx = jobCtx

	entryCtx, cancel := context.WithCancel(ctx)
	state.cancelEntry = cancel
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Errorf("session %s entrypoint panicked: %v", sid, rec)
			}
			jobCtx.shutdown(context.Background())
			r.mu.Lock()
			delete(r.sessions, sid)
			r.mu.Unlock()
		}()
		if err := r.entrypoint(entryCtx, jobCtx); err != nil {
			logger.Errorf("session %s entrypoint error: %v", sid, err)
		}
	}()
}

func (r *agentRegistry) handleSessionEnded(ended *pb.SessionEnded) {
	sid := ended.GetSessionId()
	r.mu.Lock()
	state := r.sessions[sid]
	r.mu.Unlock()
	if state == nil {
		return
	}
	signaling := ended.GetSignalingSessionId()
	logger.Infof("Session ended: %s reason=%s", sid, ended.GetReason())
	state.mu.Lock()
	sess := state.agentSession
	state.mu.Unlock()
	if sess != nil {
		if signaling != "" {
			sess.setSignalingSessionID(signaling)
		}
		reason := cmp.Or(ended.GetReason(), "runtime_session_ended")
		sess.Close(context.Background(), reason)
	}
	if state.customSTTPump != nil {
		state.customSTTPump.close()
	}
	if state.customTTSPump != nil {
		state.customTTSPump.close()
	}
	if state.cancelEntry != nil {
		state.cancelEntry()
	}
}

func (r *agentRegistry) rejectSession(sid, reason string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_SessionAck{SessionAck: &pb.SessionAck{SessionId: sid, Accepted: false, RejectReason: reason}}})
}

func (r *agentRegistry) routeSessionEvent(ctx context.Context, state *registeredSessionState, msg *pb.AgentStreamOut) {
	state.mu.Lock()
	sess := state.agentSession
	state.mu.Unlock()
	if sess != nil && sess.AgentState() == AgentStateClosing {
		return
	}
	switch p := msg.GetPayload().(type) {
	case *pb.AgentStreamOut_ToolCall:
		tc := p.ToolCall
		var onResult func(any) any
		agent := state.agent
		if sess != nil {
			agent = sess.ActiveAgent()
			onResult = sess.interceptToolResult
		}
		if agent == nil {
			agent = r.agentTemplate
		}
		tools := agent.base().tools
		sid := state.sessionID
		go executeToolWithFiller(ctx, tools, tc.GetCallId(), tc.GetToolName(), tc.GetArgumentsJson(),
			toolFiller(tools, tc.GetToolName()),
			func(text string) { r.sendSay(sid, text, false, "", true) },
			func(callID, resultJSON string, isErr bool) { r.sendToolResult(sid, callID, resultJSON, isErr) },
			onResult)
		return
	case *pb.AgentStreamOut_BeforeLlm:
		if sess == nil {
			return
		}
		go func() {
			modified, skip := runBeforeLLM(sess.pipeline, p.BeforeLlm)
			r.sendBeforeLLMResult(state.sessionID, modified, skip, msg.GetRequestId())
		}()
		return
	case *pb.AgentStreamOut_LlmTokenForReview:
		if sess == nil {
			return
		}
		repl, drop := runLLMTokenReview(sess.pipeline, p.LlmTokenForReview.GetText(), p.LlmTokenForReview.GetTokenId())
		r.sendModifyLLMToken(state.sessionID, p.LlmTokenForReview.GetTokenId(), repl, drop)
		return
	case *pb.AgentStreamOut_CustomSttAudio:
		r.handleCustomSTTAudio(state, p.CustomSttAudio)
		return
	case *pb.AgentStreamOut_CustomTtsSynthesize:
		r.handleCustomTTSSynthesize(state, p.CustomTtsSynthesize)
		return
	case *pb.AgentStreamOut_Error:
		r.handleError(state, p.Error)
		return
	case *pb.AgentStreamOut_Participant:
		t := p.Participant.GetType()
		if strings.Contains(strings.ToLower(t), "join") {
			state.mu.Lock()
			state.participantPresent = true
			state.mu.Unlock()
		}
	}
	if sess == nil {
		logger.Debugf("Dropping event for %s: AgentSession not bound yet", state.sessionID)
		return
	}
	r.handleObserveEvent(sess, msg)
}

func (r *agentRegistry) handleObserveEvent(s *AgentSession, msg *pb.AgentStreamOut) {
	switch p := msg.GetPayload().(type) {
	case *pb.AgentStreamOut_Transcript:
		transcriptRegistered(s, p.Transcript)
	case *pb.AgentStreamOut_AgentSpeech:
		agentSpeechRegistered(s, p.AgentSpeech)
	case *pb.AgentStreamOut_TurnComplete:
		turnCompleteRegistered(s, p.TurnComplete)
	case *pb.AgentStreamOut_SayComplete:
		evSayComplete(s)
	case *pb.AgentStreamOut_Interrupt:
		evInterrupt(s, p.Interrupt)
	case *pb.AgentStreamOut_StateChange:
		evStateChange(s, p.StateChange)
	case *pb.AgentStreamOut_RecordingStatus:
		evRecordingStatus(s, p.RecordingStatus)
	case *pb.AgentStreamOut_Warning:
		evWarning(s, p.Warning)
	case *pb.AgentStreamOut_Metrics:
		evMetrics(s, p.Metrics)
	case *pb.AgentStreamOut_Dtmf:
		evDTMF(s, p.Dtmf)
	case *pb.AgentStreamOut_VisionFrame:
		evVisionFrame(s, p.VisionFrame)
	case *pb.AgentStreamOut_AudioFrame:
		evAudioFrame(s, p.AudioFrame)
	case *pb.AgentStreamOut_StreamEvent:
		evStreamEvent(s, p.StreamEvent)
	case *pb.AgentStreamOut_SignalingSessionAssigned:
		evSignaling(s, p.SignalingSessionAssigned.GetSessionId())
	case *pb.AgentStreamOut_VoicemailDetected:
		evVoicemail(s, p.VoicemailDetected)
	case *pb.AgentStreamOut_A2AMessage:
		evA2A(s, p.A2AMessage)
	case *pb.AgentStreamOut_AgentSwitched:
		evAgentSwitched(s, p.AgentSwitched)
	case *pb.AgentStreamOut_KbHits:
		evKBHits(s, p.KbHits)
	case *pb.AgentStreamOut_Participant:
		evParticipant(s, p.Participant)
	case *pb.AgentStreamOut_GenerationStarted:
		evGenerationStarted(s, p.GenerationStarted)
	case *pb.AgentStreamOut_GenerationComplete:
		evGenerationComplete(s, p.GenerationComplete)
	case *pb.AgentStreamOut_GenerationChunk:
		evGenerationChunk(s, p.GenerationChunk)
	case *pb.AgentStreamOut_SynthesisStarted:
		evSynthesisStarted(s, p.SynthesisStarted)
	case *pb.AgentStreamOut_SynthesisInterrupted:
		evSynthesisInterrupted(s, p.SynthesisInterrupted)
	case *pb.AgentStreamOut_FirstAudioByte:
		evFirstAudioByte(s, p.FirstAudioByte)
	case *pb.AgentStreamOut_LastAudioByte:
		evLastAudioByte(s, p.LastAudioByte)
	case *pb.AgentStreamOut_WordTiming:
		evWordTiming(s, p.WordTiming)
	case *pb.AgentStreamOut_TtsCapabilities:
		evTTSCapabilities(s, p.TtsCapabilities)
	case *pb.AgentStreamOut_SttStreamStarted:
		evSTTStreamStarted(s, p.SttStreamStarted)
	case *pb.AgentStreamOut_SttStreamEnded:
		evSTTStreamEnded(s, p.SttStreamEnded)
	case *pb.AgentStreamOut_VadEvent:
		evVADEvent(s, p.VadEvent)
	case *pb.AgentStreamOut_EouDetected:
		evEOUDetected(s, p.EouDetected)
	case *pb.AgentStreamOut_TranscriptPreflight:
		evTranscriptPreflight(s, p.TranscriptPreflight)
	case *pb.AgentStreamOut_LlmCompleted:
		evLLMCompleted(s, p.LlmCompleted)
	case *pb.AgentStreamOut_AgentStateChanged:
		evAgentStateChanged(s, p.AgentStateChanged)
	case *pb.AgentStreamOut_UserStateChanged:
		evUserStateChanged(s, p.UserStateChanged)
	default:
		logger.Debugf("registry handleObserveEvent: unhandled event %T", msg.GetPayload())
	}
}

func (r *agentRegistry) handleError(state *registeredSessionState, err *pb.ErrorEvent) {
	state.mu.Lock()
	sess := state.agentSession
	state.mu.Unlock()
	if sess == nil {
		return
	}
	logger.Errorf("[%s][%s] %s (fatal=%v)", state.sessionID, err.GetComponent(), err.GetMessage(), err.GetIsFatal())
	sess.Emit("runtime_error", map[string]any{
		"component": err.GetComponent(), "provider": err.GetProvider(), "severity": err.GetSeverity(),
		"error_code": err.GetErrorCode(), "message": err.GetMessage(), "is_fatal": err.GetIsFatal(), "retry_attempts": err.GetRetryAttempts(),
	})
	if err.GetIsFatal() {
		go sess.Close(context.Background(), "fatal_error")
	}
}

func (r *agentRegistry) handleCustomSTTAudio(state *registeredSessionState, msg *pb.CustomSttAudioChunk) {
	state.mu.Lock()
	sess := state.agentSession
	state.mu.Unlock()
	if sess == nil {
		return
	}
	hook := sess.pipeline.Hooks.customSTT
	if hook == nil {
		return
	}
	if state.customSTTPump == nil {
		sid := state.sessionID
		state.customSTTPump = newCustomSTTPump(hook, func(res CustomSTTResult) { r.sendCustomSTTResult(sid, res) })
	}
	state.customSTTPump.push(CustomSTTAudioChunk{PCM: msg.GetPcm(), SampleRate: msg.GetSampleRate(), UtteranceID: msg.GetUtteranceId(), EndOfUtterance: msg.GetEndOfUtterance()})
}

func (r *agentRegistry) handleCustomTTSSynthesize(state *registeredSessionState, msg *pb.CustomTtsSynthesize) {
	state.mu.Lock()
	sess := state.agentSession
	state.mu.Unlock()
	if sess == nil {
		return
	}
	hook := sess.pipeline.Hooks.customTTS
	if hook == nil {
		return
	}
	if state.customTTSPump == nil {
		sid := state.sessionID
		state.customTTSPump = newCustomTTSPump(hook, func(c CustomTTSAudioChunk) { r.sendCustomTTSAudio(sid, c) })
	}
	state.customTTSPump.push(CustomTTSSynthesize{Text: msg.GetText(), UtteranceID: msg.GetUtteranceId(), Voice: msg.GetVoice()})
}

func (r *agentRegistry) beginDrain(reason string) {
	r.mu.Lock()
	if r.draining {
		r.mu.Unlock()
		return
	}
	r.draining = true
	n := len(r.sessions)
	r.mu.Unlock()
	logger.Infof("Draining agent registry: %s; active_sessions=%d", reason, n)
	r.enqueue(&pb.AgentStreamIn{Payload: &pb.AgentStreamIn_Draining{Draining: &pb.AgentDraining{Reason: reason}}})
	if n == 0 {
		r.mu.Lock()
		r.running = false
		r.mu.Unlock()
	}
}

func (r *agentRegistry) stop() {
	r.mu.Lock()
	r.running = false
	r.mu.Unlock()
	r.closeAllSessions()
}

func (r *agentRegistry) closeAllSessions() {
	r.mu.Lock()
	states := make([]*registeredSessionState, 0, len(r.sessions))
	for _, s := range r.sessions {
		states = append(states, s)
	}
	r.sessions = map[string]*registeredSessionState{}
	r.mu.Unlock()
	for _, st := range states {
		if st.cancelEntry != nil {
			st.cancelEntry()
		}
	}
}

func (r *agentRegistry) waitForRegistered(timeout time.Duration) bool {
	if timeout == 0 {
		timeout = r.initializeTimeout
	}
	select {
	case <-r.registeredCh:
		return true
	case <-time.After(timeout):
		return false
	}
}

// ---- outbound send helpers (used by registryTransport) ----

func (r *agentRegistry) sendSay(sid, text string, interruptCurrent bool, voice string, interruptible bool) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_Say{Say: &pb.SayCommand{Text: text, InterruptCurrent: interruptCurrent, Voice: voice, NonInterruptible: !interruptible}}})
}
func (r *agentRegistry) sendToolResult(sid, callID, resultJSON string, isErr bool) {
	r.enqueueResult(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_ToolResult{ToolResult: &pb.ToolCallResponse{CallId: callID, ResultJson: resultJSON, IsError: isErr}}})
}
func (r *agentRegistry) sendUpdateInstructions(sid, s string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_UpdateInstructions{UpdateInstructions: &pb.UpdateInstructionsCmd{Instructions: s}}})
}
func (r *agentRegistry) sendGenerate(sid, text string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_Generate{Generate: &pb.GenerateCmd{Text: text}}})
}
func (r *agentRegistry) sendCancelGeneration(sid string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_CancelGeneration{CancelGeneration: &pb.CancelGenerationCmd{}}})
}
func (r *agentRegistry) sendUpdateTools(sid string, tools []*FunctionTool) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_UpdateTools{UpdateTools: &pb.UpdateToolsCmd{Tools: buildToolSchemas(tools)}}})
}
func (r *agentRegistry) sendPlayBackgroundAudio(sid, url string, volume float64, looping, playbackMode bool) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_PlayBackgroundAudio{PlayBackgroundAudio: &pb.PlayBackgroundAudioCmd{FileUrl: url, Volume: float32(volume), Looping: looping, PlaybackMode: playbackMode}}})
}
func (r *agentRegistry) sendPreloadBackgroundAudio(sid, url string, volume float64) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_PreloadBackgroundAudio{PreloadBackgroundAudio: &pb.PreloadBackgroundAudioCmd{FileUrl: url, Volume: float32(volume)}}})
}
func (r *agentRegistry) sendStopBackgroundAudio(sid string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_StopBackgroundAudio{StopBackgroundAudio: &pb.StopBackgroundAudioCmd{}}})
}
func (r *agentRegistry) sendPushAudioFrame(sid string, pcm []byte, sampleRate int) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_PushAudioFrame{PushAudioFrame: &pb.PushAudioFrameCmd{Pcm: pcm, SampleRate: uint32(sampleRate)}}})
}
func (r *agentRegistry) sendSendImage(sid, mimeType string, data []byte) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_SendImage{SendImage: &pb.SendImageCmd{MimeType: mimeType, Data: data}}})
}
func (r *agentRegistry) sendSendMessageWithFrames(sid, text string, frames []frameData, numLatestFrames uint32) {
	var mf []*pb.MessageFrame
	for _, f := range frames {
		mf = append(mf, &pb.MessageFrame{MimeType: f.mimeType, Data: f.data})
	}
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_SendMessageWithFrames{SendMessageWithFrames: &pb.SendMessageWithFramesCmd{Text: text, Frames: mf, NumLatestFrames: numLatestFrames}}})
}
func (r *agentRegistry) sendRecordingStart(sid string, cfg *RecordingConfig) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_RecordingStart{RecordingStart: &pb.RecordingStartCmd{Config: buildRecordingConfig(cfg)}}})
}
func (r *agentRegistry) sendRecordingStop(sid string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_RecordingStop{RecordingStop: &pb.RecordingStopCmd{}}})
}
func (r *agentRegistry) sendCallTransfer(sid, transferTo, token string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_CallTransfer{CallTransfer: &pb.CallTransferCmd{TransferTo: transferTo, Token: token}}})
}
func (r *agentRegistry) sendPublishMessage(sid, topic, message, optionsJSON, payloadJSON string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_PublishMessage{PublishMessage: &pb.PublishMessageCmd{Topic: topic, Message: message, OptionsJson: optionsJSON, PayloadJson: payloadJSON}}})
}
func (r *agentRegistry) sendSubscribePubSub(sid, topic string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_SubscribePubsub{SubscribePubsub: &pb.SubscribePubSubCmd{Topic: topic}}})
}
func (r *agentRegistry) sendUpdateProvider(sid, component, provider string, params map[string]string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_UpdateProvider{UpdateProvider: &pb.UpdateProviderCmd{Component: component, Provider: provider, Params: copyAnyMap(params)}}})
}
func (r *agentRegistry) sendModifyLLMToken(sid string, tokenID uint64, replacement string, drop bool) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_ModifyLlmToken{ModifyLlmToken: &pb.ModifyLLMTokenCmd{TokenId: tokenID, Replacement: replacement, Drop: drop}}})
}
func (r *agentRegistry) sendBeforeLLMResult(sid, modified string, skip bool, requestID string) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, RequestId: requestID, Payload: &pb.AgentStreamIn_BeforeLlmResult{BeforeLlmResult: &pb.BeforeLLMResponse{ModifiedMessagesJson: modified, SkipTurn: skip}}})
}
func (r *agentRegistry) sendCustomSTTResult(sid string, res CustomSTTResult) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_CustomSttResult{CustomSttResult: &pb.CustomSttResult{UtteranceId: res.UtteranceID, Text: res.Text, IsFinal: res.IsFinal, Confidence: float32(res.Confidence), Language: res.Language, StartTimeUs: res.StartTimeUS, EndTimeUs: res.EndTimeUS}}})
}
func (r *agentRegistry) sendCustomTTSAudio(sid string, c CustomTTSAudioChunk) {
	r.enqueue(&pb.AgentStreamIn{SessionId: sid, Payload: &pb.AgentStreamIn_CustomTtsAudio{CustomTtsAudio: &pb.CustomTtsAudioChunk{UtteranceId: c.UtteranceID, Pcm: c.PCM, SampleRate: c.SampleRate, EndOfSynthesis: c.EndOfSynthesis}}})
}

// ---- registryTransport adapts the registry to sessionTransport ----

type registryTransport struct {
	registry *agentRegistry
	sid      string
}

func (t *registryTransport) sendSay(text string, ic bool, voice string, interruptible bool) error {
	t.registry.sendSay(t.sid, text, ic, voice, interruptible)
	return nil
}
func (t *registryTransport) sendCancelGeneration() error {
	t.registry.sendCancelGeneration(t.sid)
	return nil
}
func (t *registryTransport) sendGenerate(text string) error {
	t.registry.sendGenerate(t.sid, text)
	return nil
}
func (t *registryTransport) sendUpdateInstructions(s string) error {
	t.registry.sendUpdateInstructions(t.sid, s)
	return nil
}
func (t *registryTransport) sendUpdateTools(tools []*FunctionTool) error {
	t.registry.sendUpdateTools(t.sid, tools)
	return nil
}
func (t *registryTransport) sendUpdateProvider(component, provider string, params map[string]string) error {
	t.registry.sendUpdateProvider(t.sid, component, provider, params)
	return nil
}
func (t *registryTransport) sendCallTransfer(transferTo, token string) error {
	t.registry.sendCallTransfer(t.sid, transferTo, token)
	return nil
}
func (t *registryTransport) sendPlayBackgroundAudio(url string, volume float64, looping, playbackMode bool) error {
	t.registry.sendPlayBackgroundAudio(t.sid, url, volume, looping, playbackMode)
	return nil
}
func (t *registryTransport) sendPreloadBackgroundAudio(url string, volume float64) error {
	t.registry.sendPreloadBackgroundAudio(t.sid, url, volume)
	return nil
}
func (t *registryTransport) sendStopBackgroundAudio() error {
	t.registry.sendStopBackgroundAudio(t.sid)
	return nil
}
func (t *registryTransport) sendRecordingStart(cfg *RecordingConfig) error {
	t.registry.sendRecordingStart(t.sid, cfg)
	return nil
}
func (t *registryTransport) sendRecordingStop() error {
	t.registry.sendRecordingStop(t.sid)
	return nil
}
func (t *registryTransport) sendPushAudioFrame(pcm []byte, sampleRate int) error {
	t.registry.sendPushAudioFrame(t.sid, pcm, sampleRate)
	return nil
}
func (t *registryTransport) sendSendImage(mimeType string, data []byte) error {
	t.registry.sendSendImage(t.sid, mimeType, data)
	return nil
}
func (t *registryTransport) sendSendMessageWithFrames(text string, frames []frameData, numLatestFrames uint32) error {
	t.registry.sendSendMessageWithFrames(t.sid, text, frames, numLatestFrames)
	return nil
}
func (t *registryTransport) sendPublishMessage(topic, message, optionsJSON, payloadJSON string) error {
	t.registry.sendPublishMessage(t.sid, topic, message, optionsJSON, payloadJSON)
	return nil
}
func (t *registryTransport) sendSubscribePubSub(topic string) error {
	t.registry.sendSubscribePubSub(t.sid, topic)
	return nil
}
func (t *registryTransport) stub() pb.AgentRuntimeClient { return t.registry.client }
func (t *registryTransport) sessionID() string           { return t.sid }

func stringMapToAny(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
