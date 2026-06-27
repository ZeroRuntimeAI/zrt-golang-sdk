package zrt

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
)

// StartOptions configures AgentSession.Start.
type StartOptions struct {
	WaitForParticipant          bool
	RunUntilShutdown            bool
	SIPHangupOnShutdown         bool
	WaitForParticipantTimeoutMS int // default 60000
	WaitForAudioStream          bool
	RuntimeOptions              map[string]string
	SessionID                   string
	// RuntimeAddress overrides ZRT_RUNTIME_ADDRESS.
	RuntimeAddress string
}

// Start connects the session to the runtime and runs the agent.
//
// jobCtx may be nil for standalone use (a minimal room config is derived). When
// RunUntilShutdown is true, Start blocks until the session closes.
func (s *AgentSession) Start(ctx context.Context, jobCtx *JobContext, opts StartOptions) error {
	logger.Infof("Starting AgentSession...")
	timeout := cmp.Or(opts.WaitForParticipantTimeoutMS, 60000)
	s.mu.Lock()
	s.agentState = AgentStateStarting
	s.waitForParticipant = opts.WaitForParticipant
	s.waitForParticipantTimeout = timeout
	s.sipHangupOnShutdown = opts.SIPHangupOnShutdown
	s.waitForAudioStream = opts.WaitForAudioStream
	s.waitForAudioStreamExplicit = opts.WaitForAudioStream
	s.runtimeOptions = opts.RuntimeOptions
	s.callerSessionID = opts.SessionID
	s.mu.Unlock()

	// SIP metadata handling from the job context.
	if jobCtx != nil {
		md := jobCtx.Metadata
		isSIP := md["sipCallTo"] != nil || md["sipCallFrom"] != nil
		if ct, ok := md["callType"].(string); ok {
			isSIP = isSIP || strings.HasPrefix(strings.ToLower(ct), "sip")
		}
		if isSIP && !s.waitForAudioStreamExplicit {
			s.mu.Lock()
			s.waitForAudioStream = true
			s.mu.Unlock()
		}
		if s.callerSessionID == "" {
			if cid, ok := md["callId"].(string); ok && cid != "" && cid != "__probe__" {
				s.callerSessionID = cid
			}
		}
		jobCtx.registerSession(s)
		s.mu.Lock()
		s.jobCtx = jobCtx
		s.mu.Unlock()
		// Bind room error / session-end callbacks.
		if jobCtx.RoomOptions != nil {
			if cb := jobCtx.RoomOptions.OnRoomError; cb != nil {
				s.On("runtime_error", func(p any) { cb(p) })
			}
			if cb := jobCtx.RoomOptions.OnSessionEnd; cb != nil {
				s.On("session_ended", func(p any) { cb(p) })
			}
		}
	}

	if err := s.maybeAutoRecording(jobCtx); err != nil {
		logger.Warnf("auto recording config skipped: %v", err)
	}

	runtimeAddr := cmp.Or(opts.RuntimeAddress, os.Getenv("ZRT_RUNTIME_ADDRESS"), "localhost:50051")

	room, err := s.resolveRoomConfig(jobCtx)
	if err != nil {
		return err
	}

	printPlaygroundURL(room)

	bridge := newGrpcBridge(runtimeAddr, s.agent, s.pipeline, s, room, s.recording)
	s.bridge = bridge
	s.transport = bridge
	sid, err := bridge.createSession(ctx)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.sessionID = sid
	s.agentState = AgentStateIdle
	s.mu.Unlock()
	logger.Infof("Session created: %s", sid)
	s.Emit("agent_state_changed", map[string]any{"state": AgentStateIdle})
	s.emitMeetingJoinedOnce()

	s.eventLoopDone = make(chan struct{})
	go func() {
		defer close(s.eventLoopDone)
		bridge.runEventLoop(ctx)
	}()

	s.awaitParticipantIfNeeded(ctx)
	if err := s.agent.OnEnter(ctx); err != nil {
		logger.Errorf("on_enter error: %v", err)
	}
	if s.wakeUp > 0 && s.onWakeUp != nil {
		go s.wakeUpWatcher()
	}

	if opts.RunUntilShutdown {
		select {
		case <-s.shutdownCh:
		case <-ctx.Done():
			logger.Infof("Session context cancelled — shutting down")
		}
		s.Close(context.Background(), "sdk_close")
	}
	return nil
}

func (s *AgentSession) resolveRoomConfig(jobCtx *JobContext) (roomConfigData, error) {
	if jobCtx == nil || jobCtx.RoomOptions == nil {
		return roomConfigData{AgentName: "Agent", SendLogsToDashboard: true}, nil
	}
	opts := jobCtx.RoomOptions
	if opts.RoomID == "" {
		if _, err := jobCtx.GetRoomID(); err != nil {
			logger.Warnf("Auto room creation skipped: %v", err)
		}
	}
	token, _ := ResolveAuthToken(opts.AuthToken)
	return roomConfigData{
		RoomID:                      opts.RoomID,
		AuthToken:                   token,
		AgentName:                   cmp.Or(opts.Name, "Agent"),
		AutoEndSession:              opts.AutoEndSession,
		SessionTimeoutSeconds:       opts.SessionTimeoutSeconds,
		NoParticipantTimeoutSeconds: opts.NoParticipantTimeoutSeconds,
		AudioListenerEnabled:        opts.AudioListenerEnabled,
		AgentParticipantID:          opts.AgentParticipantID,
		Playground:                  opts.Playground,
		Vision:                      opts.Vision,
		RecordingEnabled:            opts.Recording,
		BackgroundAudioEnabled:      opts.BackgroundAudio,
		SendLogsToDashboard:         opts.SendLogsToDashboard,
	}, nil
}

func (s *AgentSession) maybeAutoRecording(jobCtx *JobContext) error {
	if s.recording != nil || jobCtx == nil || jobCtx.RoomOptions == nil || !jobCtx.RoomOptions.Recording {
		return nil
	}
	bucket := os.Getenv("ZRT_RECORDING_S3_BUCKET")
	if bucket != "" {
		region := cmp.Or(os.Getenv("ZRT_RECORDING_S3_REGION"), "us-east-1")
		s.recording = &RecordingConfig{
			Enabled: true, AutoStart: true, Format: "ogg_opus", ChannelMode: "dual_channel",
			Storage: &S3StorageConfig{
				Bucket:          bucket,
				Region:          region,
				AccessKeyID:     os.Getenv("ZRT_RECORDING_S3_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("ZRT_RECORDING_S3_SECRET_ACCESS_KEY"),
				Prefix:          os.Getenv("ZRT_RECORDING_S3_PREFIX"),
				EndpointURL:     os.Getenv("ZRT_RECORDING_S3_ENDPOINT_URL"),
			},
		}
	} else {
		logger.Infof("[recording] RoomOptions(recording=true) → Zero Runtime cloud recording (no ZRT_RECORDING_S3_BUCKET set).")
		s.recording = &RecordingConfig{Enabled: true, AutoStart: true}
	}
	return nil
}

// ---- command methods ----

// Say speaks message via TTS and returns a handle that completes on playback.
func (s *AgentSession) Say(ctx context.Context, message string) (*UtteranceHandle, error) {
	return s.SayWith(ctx, message, SayOptions{Interruptible: true})
}

// SayOptions configures Say.
type SayOptions struct {
	Interruptible bool
	Voice         string
}

// SayWith speaks message via TTS using opts to control voice and
// interruptibility, and returns a handle that completes on playback.
func (s *AgentSession) SayWith(ctx context.Context, message string, opts SayOptions) (*UtteranceHandle, error) {
	handle := NewUtteranceHandle("", opts.Interruptible)
	s.mu.Lock()
	s.currentUtterance = handle
	t := s.transport
	s.mu.Unlock()
	if t == nil {
		return handle, ErrSessionNotStarted
	}
	return handle, t.sendSay(message, false, opts.Voice, opts.Interruptible)
}

// Reply generates a context-aware reply and (by default) waits for playback.
func (s *AgentSession) Reply(ctx context.Context, instructions string, waitForPlayback bool) (*UtteranceHandle, error) {
	handle := NewUtteranceHandle("", true)
	s.mu.Lock()
	s.currentUtterance = handle
	t := s.transport
	s.mu.Unlock()
	if t == nil {
		return handle, ErrSessionNotStarted
	}
	if err := t.sendSay(instructions, true, "", true); err != nil {
		return handle, err
	}
	if waitForPlayback {
		handle.Wait()
	}
	return handle, nil
}

// ReplyOptions configures ReplyWith.
type ReplyOptions struct {
	// WaitForPlayback blocks until playback completes before returning.
	WaitForPlayback bool
	// Frames attaches image frames to the reply.
	Frames []MessageFrame
	// LatestFrames includes this many most-recent buffered vision frames when
	// Frames is empty.
	LatestFrames int
}

// ReplyWith generates a context-aware reply following instructions, with options
// to attach image frames and wait for playback, and returns its utterance handle.
func (s *AgentSession) ReplyWith(ctx context.Context, instructions string, opts ReplyOptions) (*UtteranceHandle, error) {
	handle := NewUtteranceHandle("", true)
	s.mu.Lock()
	s.currentUtterance = handle
	t := s.transport
	s.mu.Unlock()
	if t == nil {
		return handle, ErrSessionNotStarted
	}
	if len(opts.Frames) > 0 {
		if err := s.SendMessageWithFrames(ctx, instructions, opts.Frames); err != nil {
			return handle, err
		}
	} else if opts.LatestFrames > 0 {
		if err := t.sendSendMessageWithFrames(instructions, nil, uint32(opts.LatestFrames)); err != nil {
			return handle, err
		}
	} else if err := t.sendSay(instructions, true, "", true); err != nil {
		return handle, err
	}
	if opts.WaitForPlayback {
		handle.Wait()
	}
	return handle, nil
}

// Interrupt interrupts the current utterance and cancels generation.
func (s *AgentSession) Interrupt(force bool) {
	s.mu.Lock()
	u := s.currentUtterance
	t := s.transport
	s.mu.Unlock()
	if u != nil {
		u.Interrupt(force)
	}
	if t != nil {
		_ = t.sendCancelGeneration()
	}
}

// Generate prompts the agent to generate a response to text.
func (s *AgentSession) Generate(ctx context.Context, text string) error {
	if t := s.transportRef(); t != nil {
		return t.sendGenerate(text)
	}
	return nil
}

// CancelGeneration cancels the current generation.
func (s *AgentSession) CancelGeneration(ctx context.Context) error {
	if t := s.transportRef(); t != nil {
		return t.sendCancelGeneration()
	}
	return nil
}

// UpdateInstructions updates the agent instructions on the runtime.
func (s *AgentSession) UpdateInstructions(ctx context.Context, instructions string) error {
	s.agent.base().instructions = instructions
	if t := s.transportRef(); t != nil {
		return t.sendUpdateInstructions(instructions)
	}
	return nil
}

// UpdateTools replaces the agent tools on the runtime.
func (s *AgentSession) UpdateTools(ctx context.Context, tools []*FunctionTool) error {
	s.agent.base().UpdateTools(tools)
	if t := s.transportRef(); t != nil {
		return t.sendUpdateTools(tools)
	}
	return nil
}

// UpdateProvider swaps a pipeline provider at runtime.
func (s *AgentSession) UpdateProvider(ctx context.Context, component, provider string, params map[string]string) error {
	if t := s.transportRef(); t != nil {
		return t.sendUpdateProvider(component, provider, params)
	}
	return nil
}

// CallTransfer transfers the call to transferTo.
func (s *AgentSession) CallTransfer(ctx context.Context, token, transferTo string) error {
	if transferTo == "" {
		logger.Warnf("CallTransfer: empty transferTo — ignored")
		return nil
	}
	if t := s.transportRef(); t != nil {
		return t.sendCallTransfer(transferTo, token)
	}
	return nil
}

// PlayBackgroundAudio plays a background audio file. config may be a string URL,
// a map[string]any, or a *BackgroundAudioHandlerConfig.
func (s *AgentSession) PlayBackgroundAudio(ctx context.Context, config any) error {
	url, volume, looping, playbackMode := extractBGAudioArgs(config)
	if url == "" {
		logger.Warnf("PlayBackgroundAudio: empty file_url — ignored")
		return nil
	}
	fileURL, audioData, err := resolveBGAudioPayload(url)
	if err != nil {
		return err
	}
	if t := s.transportRef(); t != nil {
		return t.sendPlayBackgroundAudio(fileURL, volume, looping, playbackMode, audioData)
	}
	return nil
}

// PreloadBackgroundAudio preloads a background audio file.
func (s *AgentSession) PreloadBackgroundAudio(ctx context.Context, config any) error {
	url, volume, _, _ := extractBGAudioArgs(config)
	if url == "" {
		logger.Warnf("PreloadBackgroundAudio: empty file_url — ignored")
		return nil
	}
	fileURL, audioData, err := resolveBGAudioPayload(url)
	if err != nil {
		return err
	}
	if t := s.transportRef(); t != nil {
		return t.sendPreloadBackgroundAudio(fileURL, volume, audioData)
	}
	return nil
}

// StopBackgroundAudio stops background audio.
func (s *AgentSession) StopBackgroundAudio(ctx context.Context) error {
	if t := s.transportRef(); t != nil {
		return t.sendStopBackgroundAudio()
	}
	return nil
}

// StartRecording begins recording (config may be nil to use the session config).
func (s *AgentSession) StartRecording(ctx context.Context, config *RecordingConfig) error {
	cfg := config
	if cfg == nil {
		cfg = s.recording
	}
	if cfg == nil {
		logger.Warnf("StartRecording: no config and no session-level recording set; ignoring")
		return nil
	}
	if t := s.transportRef(); t != nil {
		return t.sendRecordingStart(cfg)
	}
	return nil
}

// StopRecording stops recording.
func (s *AgentSession) StopRecording(ctx context.Context) error {
	if t := s.transportRef(); t != nil {
		return t.sendRecordingStop()
	}
	return nil
}

// PushAudioFrame pushes a raw PCM frame to the runtime.
func (s *AgentSession) PushAudioFrame(ctx context.Context, pcm []byte, sampleRate int) error {
	if len(pcm) == 0 {
		return nil
	}
	sampleRate = cmp.Or(sampleRate, 48000)
	if t := s.transportRef(); t != nil {
		return t.sendPushAudioFrame(pcm, sampleRate)
	}
	return nil
}

// SendImage sends an image to the runtime (data must be encoded JPEG/PNG bytes).
func (s *AgentSession) SendImage(ctx context.Context, data []byte, mimeType string) error {
	if len(data) == 0 {
		return nil
	}
	mimeType = cmp.Or(mimeType, "image/jpeg")
	if t := s.transportRef(); t != nil {
		return t.sendSendImage(mimeType, data)
	}
	return nil
}

// MessageFrame is an image frame for SendMessageWithFrames.
type MessageFrame struct {
	MimeType string
	Data     []byte
}

// MessageFramesFromCaptured converts frames returned by Agent.CaptureFrames into
// MessageFrames ready for SendMessageWithFrames or ReplyWith. Frames with no
// usable bytes are skipped, and a missing or empty MIME type defaults to image/jpeg.
func MessageFramesFromCaptured(frames []map[string]any) []MessageFrame {
	out := make([]MessageFrame, 0, len(frames))
	for _, f := range frames {
		data, _ := f["data"].([]byte)
		if len(data) == 0 {
			continue
		}
		mime, _ := f["mime_type"].(string)
		out = append(out, MessageFrame{MimeType: cmp.Or(mime, "image/jpeg"), Data: data})
	}
	return out
}

// SendMessageWithFrames sends a text message with image frames.
func (s *AgentSession) SendMessageWithFrames(ctx context.Context, message string, frames []MessageFrame) error {
	var fd []frameData
	for _, f := range frames {
		fd = append(fd, frameData{mimeType: cmp.Or(f.MimeType, "image/jpeg"), data: f.Data})
	}
	if message == "" && len(fd) == 0 {
		return nil
	}
	if t := s.transportRef(); t != nil {
		return t.sendSendMessageWithFrames(message, fd, 0)
	}
	return nil
}

// StartThinkingAudio starts the agent's configured thinking audio.
func (s *AgentSession) StartThinkingAudio(ctx context.Context) error {
	cfg := s.agent.base().thinkingBackgroundConfig
	if cfg == nil {
		logger.Warnf("StartThinkingAudio: no thinking_audio configured (call agent.SetThinkingAudio first)")
		return nil
	}
	s.mu.Lock()
	s.thinkingAudioActive = true
	s.mu.Unlock()
	return s.PlayBackgroundAudio(ctx, cfg)
}

// StopThinkingAudio stops the thinking audio.
func (s *AgentSession) StopThinkingAudio(ctx context.Context) error {
	s.mu.Lock()
	s.thinkingAudioActive = false
	s.mu.Unlock()
	return s.StopBackgroundAudio(ctx)
}

func (s *AgentSession) transportRef() sessionTransport {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transport
}

// ---- close ----

// Close ends the session.
func (s *AgentSession) Close(ctx context.Context, reason string) error {
	s.mu.Lock()
	if s.agentState == AgentStateClosing {
		s.mu.Unlock()
		return nil
	}
	s.agentState = AgentStateClosing
	bridge := s.bridge
	cbs := s.shutdownCallbacks
	s.shutdownCallbacks = nil
	s.mu.Unlock()

	logger.Infof("Closing AgentSession (reason=%s)...", reason)
	payload := map[string]any{"session_id": s.SessionID(), "reason": reason}
	s.Emit("session_ended", payload)
	s.Emit("meeting_left", payload)
	s.Emit("agent_left", payload)

	for _, cb := range cbs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("shutdown callback panicked: %v", r)
				}
			}()
			cb()
		}()
	}

	_, _ = s.FetchContextHistory(ctx, 0, false, false)
	if err := s.agent.OnExit(ctx); err != nil {
		logger.Errorf("on_exit error: %v", err)
	}
	if bridge != nil {
		bridge.destroySession()
	}
	if s.eventLoopDone != nil {
		select {
		case <-s.eventLoopDone:
		case <-time.After(2 * time.Second):
		}
	}
	s.closeEmitter()
	s.shutdownOnce.Do(func() { close(s.shutdownCh) })
	logger.Infof("Session closed")
	return nil
}

// Leave ends the session and disconnects the agent.
func (s *AgentSession) Leave(ctx context.Context) error { return s.Close(ctx, "sdk_leave") }

// Hangup ends the call and closes the session.
func (s *AgentSession) Hangup(ctx context.Context, reason string) error {
	return s.Close(ctx, "manual_hangup")
}

// ---- context history ----

// FetchContextHistory fetches conversation history from the runtime.
func (s *AgentSession) FetchContextHistory(ctx context.Context, lastN int, includeFunctionCalls, includeSystemMessages bool) ([]map[string]any, error) {
	if s.SessionID() == "" {
		return filterHistory(s.chatHistoryCache, lastN, includeFunctionCalls, includeSystemMessages), nil
	}
	t := s.transportRef()
	if t == nil || t.stub() == nil {
		return filterHistory(s.chatHistoryCache, lastN, includeFunctionCalls, includeSystemMessages), nil
	}
	resp, err := t.stub().GetContext(ctx, &pb.GetContextRequest{SessionId: s.SessionID(), LastN: 0})
	if err != nil {
		logger.Errorf("GetContext RPC failed: %v", err)
		return filterHistory(s.chatHistoryCache, lastN, includeFunctionCalls, includeSystemMessages), nil
	}
	var full []map[string]any
	for _, m := range resp.GetMessages() {
		full = append(full, map[string]any{"role": m.GetRole(), "content": m.GetContent(), "message_id": m.GetMessageId()})
	}
	s.chatHistoryCache = full
	return filterHistory(full, lastN, includeFunctionCalls, includeSystemMessages), nil
}

// RemoveMessage removes a message from the conversation context by id. Returns
// true if the message was removed.
func (s *AgentSession) RemoveMessage(ctx context.Context, messageID string) bool {
	if messageID == "" || s.SessionID() == "" {
		return false
	}
	t := s.transportRef()
	if t == nil || t.stub() == nil {
		return false
	}
	resp, err := t.stub().RemoveMessage(ctx, &pb.RemoveMessageRequest{SessionId: s.SessionID(), MessageId: messageID})
	if err != nil {
		logger.Errorf("RemoveMessage RPC failed for id=%s: %v", messageID, err)
		return false
	}
	var kept []map[string]any
	for _, m := range s.chatHistoryCache {
		if m["message_id"] != messageID {
			kept = append(kept, m)
		}
	}
	s.chatHistoryCache = kept
	return resp.GetSuccess()
}

// InjectMessage appends a message to the conversation context. Returns true on success.
func (s *AgentSession) InjectMessage(ctx context.Context, message ChatMessage) bool {
	if s.SessionID() == "" {
		logger.Warnf("InjectMessage: no active session id yet; skipping")
		return false
	}
	t := s.transportRef()
	if t == nil || t.stub() == nil {
		logger.Warnf("InjectMessage: no runtime stub available; skipping")
		return false
	}
	pm := &pb.ContextMessageProto{Role: string(message.Role), Content: message.Content, MessageId: message.MessageID}
	for _, img := range message.Images {
		pm.Images = append(pm.Images, &pb.ImageContentProto{Url: img.URL, Detail: img.Detail})
	}
	resp, err := t.stub().InjectMessage(ctx, &pb.InjectMessageRequest{SessionId: s.SessionID(), Message: pm})
	if err != nil {
		logger.Errorf("InjectMessage RPC failed: %v", err)
		return false
	}
	if !resp.GetSuccess() {
		logger.Warnf("InjectMessage rejected by runtime: %s", resp.GetError())
		return false
	}
	logger.Debugf("InjectMessage ok (role=%s, total_messages=%d)", message.Role, resp.GetTotalMessages())
	return true
}

// InjectContext appends every message of a ChatContext to the conversation
// context, in order. Returns true if every message was injected successfully.
func (s *AgentSession) InjectContext(ctx context.Context, cc *ChatContext) bool {
	if cc == nil {
		return true
	}
	ok := true
	for _, m := range cc.Messages() {
		if !s.InjectMessage(ctx, m) {
			ok = false
		}
	}
	return ok
}

// FetchChatContext fetches the latest conversation history as a typed
// *ChatContext. Pass lastN > 0 to limit the number of most-recent messages.
// Returns an empty context if the session is not started or the fetch fails.
func (s *AgentSession) FetchChatContext(ctx context.Context, lastN int) *ChatContext {
	if s.SessionID() == "" {
		logger.Warnf("FetchChatContext: no active session id yet; returning empty context")
		return &ChatContext{}
	}
	t := s.transportRef()
	if t == nil || t.stub() == nil {
		logger.Warnf("FetchChatContext: no runtime stub available; returning empty context")
		return &ChatContext{}
	}
	var ln uint32
	if lastN > 0 {
		ln = uint32(lastN)
	}
	resp, err := t.stub().GetContext(ctx, &pb.GetContextRequest{SessionId: s.SessionID(), LastN: ln})
	if err != nil {
		logger.Errorf("GetContext RPC failed: %v", err)
		return &ChatContext{}
	}
	logger.Debugf("FetchChatContext: runtime returned %d message(s) (total_messages=%d)", len(resp.GetMessages()), resp.GetTotalMessages())
	return ChatContextFromContextMessages(resp.GetMessages())
}

func filterHistory(history []map[string]any, lastN int, includeFunctionCalls, includeSystemMessages bool) []map[string]any {
	var out []map[string]any
	for _, m := range history {
		role, _ := m["role"].(string)
		if !includeSystemMessages && role == "system" {
			continue
		}
		if !includeFunctionCalls && (role == "tool" || role == "function" || role == "tool_result") {
			continue
		}
		out = append(out, m)
	}
	if lastN > 0 && len(out) > lastN {
		out = out[len(out)-lastN:]
	}
	return out
}

// ---- wake-up watcher ----

func (s *AgentSession) wakeUpWatcher() {
	s.mu.Lock()
	wakeUp := s.wakeUp
	onWakeUp := s.onWakeUp
	maxAttempts := s.wakeUpMaxAttempts
	s.mu.Unlock()
	if wakeUp <= 0 || onWakeUp == nil {
		return
	}
	timeout := time.Duration(wakeUp) * time.Second
	<-s.audioObservedCh
	for {
		if s.AgentState() == AgentStateClosing {
			return
		}
		select {
		case <-s.wakeUpReset:
			continue
		case <-s.shutdownCh:
			return
		case <-time.After(timeout):
			st := s.AgentState()
			if st == AgentStateClosing {
				return
			}
			if st == AgentStateSpeaking || st == AgentStateThinking || st == AgentStateStarting {
				continue
			}
			s.mu.Lock()
			s.wakeUpCount++
			count := s.wakeUpCount
			s.mu.Unlock()
			if maxAttempts > 0 && count > maxAttempts {
				s.Emit("wake_up_exceeded", map[string]any{"attempts": count})
				go s.Close(context.Background(), "wake_up_exceeded")
				return
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Errorf("on_wake_up callback panicked: %v", r)
					}
				}()
				onWakeUp()
			}()
		}
	}
}

// ---- await participant ----

func (s *AgentSession) awaitParticipantIfNeeded(ctx context.Context) {
	s.firePreloadAssets(ctx)
	s.mu.Lock()
	want := s.waitForParticipant
	wantAudio := s.waitForAudioStream
	present := s.participantPresent
	audioActive := s.audioStreamActive
	timeout := time.Duration(max(s.waitForParticipantTimeout, 1)) * time.Millisecond
	s.mu.Unlock()
	if !want {
		return
	}
	if present && (!wantAudio || audioActive) {
		return
	}
	logger.Infof("wait_for_participant: holding on_enter until participant joins (timeout %.1fs)", timeout.Seconds())
	deadline := time.Now().Add(timeout)
	if !present {
		select {
		case <-s.participantCh:
		case <-time.After(time.Until(deadline)):
			logger.Warnf("wait_for_participant: timed out — proceeding with on_enter anyway")
			return
		case <-ctx.Done():
			return
		}
	}
	if wantAudio && !s.audioStreamActiveFlag() {
		select {
		case <-s.audioStreamCh:
		case <-time.After(time.Until(deadline)):
			logger.Warnf("wait_for_participant: audio stream timed out — proceeding with on_enter anyway")
		case <-ctx.Done():
		}
	}
}

func (s *AgentSession) audioStreamActiveFlag() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.audioStreamActive
}

func (s *AgentSession) firePreloadAssets(ctx context.Context) {
	seen := map[string]bool{}
	queue := func(url string, volume float64) {
		if url == "" {
			return
		}
		key := fmt.Sprintf("%s|%d", url, int(volume*1000))
		if seen[key] {
			return
		}
		seen[key] = true
		go s.PreloadBackgroundAudio(ctx, map[string]any{"file_url": url, "volume": volume})
	}
	if s.backgroundAudio != nil {
		url, volume, _, _ := extractBGAudioArgs(s.backgroundAudio)
		queue(url, volume)
	}
	for _, entry := range s.agent.base().preloadBGAudio {
		url, _ := entry[0].(string)
		volume := 1.0
		if len(entry) > 1 {
			if v, ok := entry[1].(float64); ok {
				volume = v
			}
		}
		queue(url, volume)
	}
}

// ---- helpers ----

func extractBGAudioArgs(config any) (url string, volume float64, looping, playbackMode bool) {
	volume = 1.0
	switch c := config.(type) {
	case nil:
		return "", 1.0, false, false
	case string:
		return c, 1.0, false, false
	case map[string]any:
		url = firstString(c, "file_url", "file_path", "url")
		if v, ok := c["volume"].(float64); ok {
			volume = v
		}
		looping, _ = c["looping"].(bool)
		if pm, ok := c["playback_mode"].(bool); ok {
			playbackMode = pm
		} else if m, _ := c["mode"].(string); m == "playback" {
			playbackMode = true
		}
		return url, volume, looping, playbackMode
	case *BackgroundAudioHandlerConfig:
		pm := c.Mode == "playback"
		return c.File, c.Volume, c.Looping, pm
	}
	return "", 1.0, false, false
}

// maxBGAudioBytes caps inline background-audio payloads at 16 MiB.
const maxBGAudioBytes = 16 * 1024 * 1024

// resolveBGAudioPayload turns a background-audio source into a (fileURL, audioData) pair.
//   - http(s):// URLs are forwarded as-is with no bytes.
//   - A local file path is read into bytes (≤16 MiB), returned with an empty URL.
//   - Anything else is treated as an opaque/remote URL with no bytes.
func resolveBGAudioPayload(url string) (fileURL string, audioData []byte, err error) {
	if url == "" {
		return "", nil, nil
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url, nil, nil
	}
	info, statErr := os.Stat(url)
	if statErr != nil || info.IsDir() {
		// not a local file — treat as an opaque URL
		return url, nil, nil
	}
	if info.Size() > maxBGAudioBytes {
		return "", nil, fmt.Errorf("background audio file %q is %d bytes, exceeds the %d-byte (16 MiB) limit", url, info.Size(), maxBGAudioBytes)
	}
	data, readErr := os.ReadFile(url)
	if readErr != nil {
		return "", nil, readErr
	}
	return "", data, nil
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func printPlaygroundURL(room roomConfigData) {
	enabled := room.Playground
	if v := os.Getenv("ZRT_PLAYGROUND"); v == "1" || v == "true" || v == "yes" || v == "on" {
		enabled = true
	}
	if !enabled || room.RoomID == "" || room.AuthToken == "" {
		return
	}
	base := cmp.Or(os.Getenv("ZRT_PLAYGROUND_URL"), "https://playground.zeroruntime.ai//cli")
	logger.Infof("Agent started in playground mode")
	// Playground mode is an explicit, dev-facing opt-in and the URL is meant to be
	// opened in a browser, so the full token is printed (it is unusable truncated).
	logger.Infof("Interact with agent here at: %s?token=%s&meetingId=%s", base, room.AuthToken, room.RoomID)
}
