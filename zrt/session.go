package zrt

import (
	"context"
	"strings"
	"sync"
	"time"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
)

// roomConfigData carries the resolved room settings into config building.
type roomConfigData struct {
	RoomID                      string
	AuthToken                   string
	AgentName                   string
	AutoEndSession              bool
	SessionTimeoutSeconds       int
	NoParticipantTimeoutSeconds int
	AudioListenerEnabled        bool
	AgentParticipantID          string
	Playground                  bool
	Vision                      bool
	RecordingEnabled            bool
	BackgroundAudioEnabled      bool
	SendLogsToDashboard         bool
}

type frameData struct {
	mimeType string
	data     []byte
}

// sessionTransport abstracts the outbound channel (direct gRPC bridge or the
// registered-agent registry), so AgentSession command methods don't branch.
type sessionTransport interface {
	sendSay(text string, interruptCurrent bool, voice string, interruptible bool) error
	sendCancelGeneration() error
	sendGenerate(text string) error
	sendUpdateInstructions(s string) error
	sendUpdateTools(tools []*FunctionTool) error
	sendUpdateProvider(component, provider string, params map[string]string) error
	sendCallTransfer(transferTo, token string) error
	sendPlayBackgroundAudio(url string, volume float64, looping, playbackMode bool) error
	sendPreloadBackgroundAudio(url string, volume float64) error
	sendStopBackgroundAudio() error
	sendRecordingStart(cfg *RecordingConfig) error
	sendRecordingStop() error
	sendPushAudioFrame(pcm []byte, sampleRate int) error
	sendSendImage(mimeType string, data []byte) error
	sendSendMessageWithFrames(text string, frames []frameData) error
	stub() pb.AgentRuntimeClient
	sessionID() string
}

// AgentSessionOptions configures an AgentSession.
type AgentSessionOptions struct {
	WakeUp            int
	WakeUpMaxAttempts int
	OnWakeUp          func()
	BackgroundAudio   any
	DTMFHandler       *DTMFHandler
	VoiceMailDetector *VoiceMailDetector
	Recording         *RecordingConfig
}

// AgentSession runs an agent against the ZRT runtime.
type AgentSession struct {
	EventEmitter

	agent    Agent
	pipeline *Pipeline

	wakeUp            int
	wakeUpMaxAttempts int
	wakeUpCount       int
	onWakeUp          func()
	recording         *RecordingConfig
	backgroundAudio   any
	dtmfHandler       *DTMFHandler
	voicemailDetector *VoiceMailDetector

	mu                 sync.Mutex
	userState          UserState
	agentState         AgentState
	sessionID          string
	signalingSessionID string

	transport sessionTransport
	bridge    *grpcBridge

	currentUtterance *UtteranceHandle
	recordingState   map[string]any
	lastEOU          map[string]any
	lastAgentSpeech  map[string]any
	ttsCapabilities  map[string]any

	isSynthesizing      bool
	synthesisDone       chan struct{}
	chatHistoryCache    []map[string]any
	transcriptMirror    []map[string]any
	thinkingAudioActive bool

	sttObservationQueue chan STTResponse

	shutdownOnce  sync.Once
	shutdownCh    chan struct{}
	eventLoopDone chan struct{}

	// start options
	waitForParticipant         bool
	waitForParticipantTimeout  int
	sipHangupOnShutdown        bool
	waitForAudioStream         bool
	waitForAudioStreamExplicit bool
	runtimeOptions             map[string]string
	callerSessionID            string

	participantPresent bool
	participantCh      chan struct{}
	audioStreamActive  bool
	audioStreamCh      chan struct{}
	audioObserved      bool
	audioObservedCh    chan struct{}

	wakeUpReset chan struct{}

	shutdownCallbacks    []func()
	boundRegistry        *agentRegistry
	registeredSession    string
	meetingJoinedEmitted bool
	jobCtx               *JobContext
	audioTrackCache      *AudioTrack
}

// NewAgentSession creates a session for agent + pipeline.
func NewAgentSession(agent Agent, pipeline *Pipeline, opts AgentSessionOptions) *AgentSession {
	s := &AgentSession{
		agent:             agent,
		pipeline:          pipeline,
		wakeUp:            opts.WakeUp,
		wakeUpMaxAttempts: opts.WakeUpMaxAttempts,
		onWakeUp:          opts.OnWakeUp,
		recording:         opts.Recording,
		backgroundAudio:   opts.BackgroundAudio,
		dtmfHandler:       opts.DTMFHandler,
		voicemailDetector: opts.VoiceMailDetector,
		userState:         UserStateIdle,
		agentState:        AgentStateStarting,
		synthesisDone:     make(chan struct{}),
		shutdownCh:        make(chan struct{}),
		participantCh:     make(chan struct{}),
		audioStreamCh:     make(chan struct{}),
		audioObservedCh:   make(chan struct{}),
		wakeUpReset:       make(chan struct{}, 1),
	}
	close(s.synthesisDone) // starts "done"
	agent.base().session = s
	pipeline.setAgent(agent)
	for _, comp := range []any{pipeline.STT, pipeline.LLM, pipeline.TTS, pipeline.VAD, pipeline.TurnDetector} {
		if ss, ok := comp.(sessionSettable); ok && comp != nil {
			ss.setSession(s)
		}
	}
	if opts.VoiceMailDetector != nil {
		if pipeline.VoiceMailDetector == nil {
			pipeline.VoiceMailDetector = opts.VoiceMailDetector
		}
		s.On("voicemail_detected", func(payload any) {
			if m, ok := payload.(map[string]any); ok {
				opts.VoiceMailDetector.onRuntimeEvent(m)
			}
		})
	}
	if opts.DTMFHandler != nil {
		s.On("dtmf_received", func(payload any) {
			if m, ok := payload.(map[string]any); ok {
				if d, _ := m["digit"].(string); d != "" {
					opts.DTMFHandler.dispatch(d)
				}
			}
		})
	}
	s.On("synthesis_started", func(any) { s.markSynthesisStarted() })
	s.On("last_audio_byte", func(any) { s.markSynthesisDone() })
	s.On("synthesis_interrupted", func(any) { s.markSynthesisDone() })
	s.On("participant_joined", func(any) { s.markParticipantArrived() })
	s.On("stream_enabled", func(payload any) { s.markAudioStreamActive(payload) })
	// transcript mirror (for GetContextHistory fallback).
	s.On("transcript_preflight", func(p any) { s.mirrorUserTranscript(p) })
	s.On("user_turn_end", func(p any) { s.mirrorUserTranscript(p) })
	s.On("generation_complete", func(p any) { s.mirrorAgentGeneration(p) })
	s.On("generation_started", func(any) { s.autoStartThinkingAudio() })
	s.On("first_audio_byte", func(any) { s.autoStopThinkingAudio() })
	s.On("synthesis_interrupted", func(any) { s.autoStopThinkingAudio() })
	s.On("agent_turn_end", func(any) { s.autoStopThinkingAudio() })
	return s
}

func (s *AgentSession) mirrorUserTranscript(payload any) {
	text := extractMirrorText(payload)
	if text == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if n := len(s.transcriptMirror); n > 0 {
		last := s.transcriptMirror[n-1]
		if last["role"] == "user" && last["content"] == text {
			return
		}
	}
	s.transcriptMirror = append(s.transcriptMirror, map[string]any{"role": "user", "content": text})
}

func (s *AgentSession) mirrorAgentGeneration(payload any) {
	text := extractMirrorText(payload)
	if text == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if n := len(s.transcriptMirror); n > 0 {
		last := s.transcriptMirror[n-1]
		if last["role"] == "assistant" && last["content"] == text {
			return
		}
	}
	s.transcriptMirror = append(s.transcriptMirror, map[string]any{"role": "assistant", "content": text})
}

func extractMirrorText(payload any) string {
	switch p := payload.(type) {
	case string:
		return strings.TrimSpace(p)
	case map[string]any:
		for _, key := range []string{"text", "transcript", "content"} {
			if v, ok := p[key].(string); ok && strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

func (s *AgentSession) autoStartThinkingAudio() {
	s.mu.Lock()
	if s.thinkingAudioActive {
		s.mu.Unlock()
		return
	}
	cfg := s.agent.base().thinkingBackgroundConfig
	if cfg == nil {
		s.mu.Unlock()
		return
	}
	s.thinkingAudioActive = true
	s.mu.Unlock()
	if err := s.PlayBackgroundAudio(context.Background(), cfg); err != nil {
		s.mu.Lock()
		s.thinkingAudioActive = false
		s.mu.Unlock()
		logger.Errorf("auto thinking-audio start failed: %v", err)
	}
}

func (s *AgentSession) autoStopThinkingAudio() {
	s.mu.Lock()
	if !s.thinkingAudioActive {
		s.mu.Unlock()
		return
	}
	s.thinkingAudioActive = false
	s.mu.Unlock()
	if err := s.StopBackgroundAudio(context.Background()); err != nil {
		logger.Errorf("auto thinking-audio stop failed: %v", err)
	}
}

// UserState returns the current user state.
func (s *AgentSession) UserState() UserState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.userState
}

// AgentState returns the current agent state.
func (s *AgentSession) AgentState() AgentState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.agentState
}

// SessionID returns the runtime session id (empty before start).
func (s *AgentSession) SessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionID
}

// SignalingSessionID returns the signaling session id, if assigned.
func (s *AgentSession) SignalingSessionID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.signalingSessionID
}

// Pipeline returns the session's pipeline.
func (s *AgentSession) Pipeline() *Pipeline { return s.pipeline }

// Agent returns the session's agent.
func (s *AgentSession) Agent() Agent { return s.agent }

// CurrentUtterance returns the in-flight utterance handle (may be nil).
func (s *AgentSession) CurrentUtterance() *UtteranceHandle {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentUtterance
}

// IsSynthesizing reports whether the agent is currently speaking.
func (s *AgentSession) IsSynthesizing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isSynthesizing
}

// TTSCapabilities returns the runtime-reported TTS capabilities (may be nil).
func (s *AgentSession) TTSCapabilities() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ttsCapabilities
}

// RecordingState returns the latest recording status (may be nil).
func (s *AgentSession) RecordingState() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recordingState
}

// IsVoicemailDetected reports whether voicemail was detected.
func (s *AgentSession) IsVoicemailDetected() bool {
	if s.voicemailDetector == nil {
		return false
	}
	return s.voicemailDetector.IsDetected()
}

func (s *AgentSession) markSynthesisStarted() {
	s.mu.Lock()
	if !s.isSynthesizing {
		s.isSynthesizing = true
		s.synthesisDone = make(chan struct{})
	}
	s.mu.Unlock()
}

func (s *AgentSession) markSynthesisDone() {
	s.mu.Lock()
	if s.isSynthesizing {
		s.isSynthesizing = false
		close(s.synthesisDone)
	}
	s.mu.Unlock()
}

// WaitForSynthesisDone blocks until synthesis completes or the timeout elapses.
// A zero timeout waits indefinitely. Returns true if synthesis finished.
func (s *AgentSession) WaitForSynthesisDone(timeout time.Duration) bool {
	s.mu.Lock()
	if !s.isSynthesizing {
		s.mu.Unlock()
		return true
	}
	done := s.synthesisDone
	s.mu.Unlock()
	if timeout <= 0 {
		<-done
		return true
	}
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// ---- state updates ----

func (s *AgentSession) updateUserState(state UserState) {
	s.mu.Lock()
	s.userState = state
	if !s.audioObserved && (state == UserStateSpeaking || state == UserStateListening) {
		s.audioObserved = true
		closeOnce(&s.audioObservedCh)
	}
	s.mu.Unlock()
	if state == UserStateSpeaking {
		s.resetWakeUpTimer()
		s.Emit("speech_in", map[string]any{"participant_id": ""})
	}
	s.Emit("user_state_changed", map[string]any{"state": state})
}

func (s *AgentSession) updateAgentState(state AgentState) {
	s.mu.Lock()
	prev := s.agentState
	s.agentState = state
	changed := prev != state
	if changed && (state == AgentStateListening || state == AgentStateIdle || state == AgentStateSpeaking || state == AgentStateThinking) {
		if !s.audioObserved {
			s.audioObserved = true
			closeOnce(&s.audioObservedCh)
		}
	}
	s.mu.Unlock()
	if changed && (state == AgentStateListening || state == AgentStateIdle || state == AgentStateSpeaking || state == AgentStateThinking) {
		s.resetWakeUpTimer()
	}
	s.Emit("agent_state_changed", map[string]any{"state": state})
}

func (s *AgentSession) updateRecordingState(status map[string]any) {
	s.mu.Lock()
	s.recordingState = status
	s.mu.Unlock()
	s.Emit("recording_status_changed", status)
}

func (s *AgentSession) setSignalingSessionID(id string) {
	s.mu.Lock()
	s.signalingSessionID = id
	s.mu.Unlock()
}

func (s *AgentSession) markParticipantArrived() {
	s.mu.Lock()
	s.participantPresent = true
	closeOnce(&s.participantCh)
	s.mu.Unlock()
}

func (s *AgentSession) markAudioStreamActive(payload any) {
	m, ok := payload.(map[string]any)
	if !ok {
		return
	}
	kind, _ := m["kind"].(string)
	enabled, _ := m["enabled"].(bool)
	if !enabled || !strings.HasPrefix(strings.ToLower(kind), "audio") {
		return
	}
	s.mu.Lock()
	if !s.audioStreamActive {
		s.audioStreamActive = true
		closeOnce(&s.audioStreamCh)
	}
	s.mu.Unlock()
}

func (s *AgentSession) resetWakeUpTimer() {
	s.mu.Lock()
	s.wakeUpCount = 0
	s.mu.Unlock()
	select {
	case s.wakeUpReset <- struct{}{}:
	default:
	}
}

func closeOnce(ch *chan struct{}) {
	select {
	case <-*ch:
		// already closed
	default:
		close(*ch)
	}
}

// AddShutdownCallback registers a callback run during Close.
func (s *AgentSession) AddShutdownCallback(cb func()) {
	s.mu.Lock()
	s.shutdownCallbacks = append(s.shutdownCallbacks, cb)
	s.mu.Unlock()
}

func (s *AgentSession) emitMeetingJoinedOnce() {
	s.mu.Lock()
	if s.meetingJoinedEmitted {
		s.mu.Unlock()
		return
	}
	s.meetingJoinedEmitted = true
	id := s.sessionID
	s.mu.Unlock()
	payload := map[string]any{"session_id": id}
	s.Emit("meeting_joined", payload)
	s.Emit("agent_joined", payload)
}
