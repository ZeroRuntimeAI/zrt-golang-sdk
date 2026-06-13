package zrt

// Version is the SDK version.
const Version = "0.0.1-beta.1"

// UserState describes the user's conversational state.
type UserState string

const (
	UserStateIdle      UserState = "idle"
	UserStateSpeaking  UserState = "speaking"
	UserStateListening UserState = "listening"
)

// AgentState describes the agent's conversational state.
type AgentState string

const (
	AgentStateStarting  AgentState = "starting"
	AgentStateIdle      AgentState = "idle"
	AgentStateSpeaking  AgentState = "speaking"
	AgentStateListening AgentState = "listening"
	AgentStateThinking  AgentState = "thinking"
	AgentStateClosing   AgentState = "closing"
)

// PipelineMode describes the resolved pipeline topology.
type PipelineMode string

const (
	PipelineModeRealtime         PipelineMode = "realtime"
	PipelineModeFullCascading    PipelineMode = "full_cascading"
	PipelineModeLLMTTSOnly       PipelineMode = "llm_tts_only"
	PipelineModeSTTLLMOnly       PipelineMode = "stt_llm_only"
	PipelineModeLLMOnly          PipelineMode = "llm_only"
	PipelineModeSTTOnly          PipelineMode = "stt_only"
	PipelineModeTTSOnly          PipelineMode = "tts_only"
	PipelineModeSTTTTSOnly       PipelineMode = "stt_tts_only"
	PipelineModeHybrid           PipelineMode = "hybrid"
	PipelineModePartialCascading PipelineMode = "partial_cascading"
)

// RealtimeMode describes the realtime speech-to-speech sub-mode.
type RealtimeMode string

const (
	RealtimeModeFullS2S   RealtimeMode = "full_s2s"
	RealtimeModeHybridSTT RealtimeMode = "hybrid_stt"
	RealtimeModeHybridTTS RealtimeMode = "hybrid_tts"
	RealtimeModeLLMOnly   RealtimeMode = "llm_only"
)

// PipelineComponent identifies a pipeline slot.
type PipelineComponent string

const (
	ComponentSTT           PipelineComponent = "stt"
	ComponentLLM           PipelineComponent = "llm"
	ComponentTTS           PipelineComponent = "tts"
	ComponentVAD           PipelineComponent = "vad"
	ComponentTurnDetector  PipelineComponent = "turn_detector"
	ComponentAvatar        PipelineComponent = "avatar"
	ComponentDenoise       PipelineComponent = "denoise"
	ComponentRealtimeModel PipelineComponent = "realtime_model"
)

// SpeechEventType identifies an STT speech event type.
type SpeechEventType string

const (
	SpeechEventStart     SpeechEventType = "start_of_speech"
	SpeechEventInterim   SpeechEventType = "interim_transcript"
	SpeechEventPreflight SpeechEventType = "preflight_transcript"
	SpeechEventFinal     SpeechEventType = "final_transcript"
	SpeechEventEnd       SpeechEventType = "end_of_speech"
)

// VADEventType identifies a VAD edge event.
type VADEventType string

const (
	VADEventStartOfSpeech VADEventType = "start_of_speech"
	VADEventEndOfSpeech   VADEventType = "end_of_speech"
)

// ChatRole identifies a chat message role.
type ChatRole string

const (
	ChatRoleSystem    ChatRole = "system"
	ChatRoleUser      ChatRole = "user"
	ChatRoleAssistant ChatRole = "assistant"
)

// ToolChoice identifies the tool-choice policy.
type ToolChoice string

const (
	ToolChoiceAuto     ToolChoice = "auto"
	ToolChoiceNone     ToolChoice = "none"
	ToolChoiceRequired ToolChoice = "required"
)

// RecordingFormat identifies the recording container/codec.
type RecordingFormat string

const (
	RecordingFormatWAV     RecordingFormat = "wav"
	RecordingFormatOGGOpus RecordingFormat = "ogg_opus"
	RecordingFormatMP3     RecordingFormat = "mp3"
	RecordingFormatFLAC    RecordingFormat = "flac"
)

// RecordingChannelMode identifies the recording channel layout.
type RecordingChannelMode string

const (
	RecordingChannelMixed RecordingChannelMode = "mixed"
	RecordingChannelDual  RecordingChannelMode = "dual_channel"
)

// RecordingTranscriptFormat identifies the recording transcript format.
type RecordingTranscriptFormat string

const (
	RecordingTranscriptJSON RecordingTranscriptFormat = "json"
	RecordingTranscriptSRT  RecordingTranscriptFormat = "srt"
	RecordingTranscriptVTT  RecordingTranscriptFormat = "vtt"
)

// RecordingState identifies a recording lifecycle state.
type RecordingState string

const (
	RecordingStateIdle       RecordingState = "idle"
	RecordingStateRecording  RecordingState = "recording"
	RecordingStateFinalizing RecordingState = "finalizing"
	RecordingStateUploading  RecordingState = "uploading"
	RecordingStateCompleted  RecordingState = "completed"
	RecordingStateFailed     RecordingState = "failed"
)
