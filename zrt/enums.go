package zrt

// Version is the SDK version.
const Version = "0.0.1-beta.1"

// UserState describes the user's conversational state.
type UserState string

const (
	// UserStateIdle means the user is neither speaking nor being listened to.
	UserStateIdle UserState = "idle"
	// UserStateSpeaking means the user is currently speaking.
	UserStateSpeaking UserState = "speaking"
	// UserStateListening means the user is listening to the agent.
	UserStateListening UserState = "listening"
)

// AgentState describes the agent's conversational state.
type AgentState string

const (
	// AgentStateStarting means the agent is initializing its pipeline.
	AgentStateStarting AgentState = "starting"
	// AgentStateIdle means the agent is active but neither speaking nor listening.
	AgentStateIdle AgentState = "idle"
	// AgentStateSpeaking means the agent is currently producing speech.
	AgentStateSpeaking AgentState = "speaking"
	// AgentStateListening means the agent is listening for user input.
	AgentStateListening AgentState = "listening"
	// AgentStateThinking means the agent is processing input or generating a response.
	AgentStateThinking AgentState = "thinking"
	// AgentStateClosing means the agent is shutting down its pipeline.
	AgentStateClosing AgentState = "closing"
)

// PipelineMode describes the resolved pipeline topology.
type PipelineMode string

const (
	// PipelineModeRealtime is a speech-to-speech realtime-model pipeline.
	PipelineModeRealtime PipelineMode = "realtime"
	// PipelineModeFullCascading is a full STT -> LLM -> TTS cascading pipeline.
	PipelineModeFullCascading PipelineMode = "full_cascading"
	// PipelineModeLLMTTSOnly runs only LLM and TTS.
	PipelineModeLLMTTSOnly PipelineMode = "llm_tts_only"
	// PipelineModeSTTLLMOnly runs only STT and LLM.
	PipelineModeSTTLLMOnly PipelineMode = "stt_llm_only"
	// PipelineModeLLMOnly runs only the LLM.
	PipelineModeLLMOnly PipelineMode = "llm_only"
	// PipelineModeSTTOnly runs only STT.
	PipelineModeSTTOnly PipelineMode = "stt_only"
	// PipelineModeTTSOnly runs only TTS.
	PipelineModeTTSOnly PipelineMode = "tts_only"
	// PipelineModeSTTTTSOnly runs only STT and TTS.
	PipelineModeSTTTTSOnly PipelineMode = "stt_tts_only"
	// PipelineModeHybrid mixes a realtime model with cascading components.
	PipelineModeHybrid PipelineMode = "hybrid"
	// PipelineModePartialCascading is a partial cascading pipeline.
	PipelineModePartialCascading PipelineMode = "partial_cascading"
)

// RealtimeMode describes the realtime speech-to-speech sub-mode.
type RealtimeMode string

const (
	// RealtimeModeFullS2S runs the realtime model as full speech-to-speech.
	RealtimeModeFullS2S RealtimeMode = "full_s2s"
	// RealtimeModeHybridSTT uses an external STT with the realtime model.
	RealtimeModeHybridSTT RealtimeMode = "hybrid_stt"
	// RealtimeModeHybridTTS uses an external TTS with the realtime model.
	RealtimeModeHybridTTS RealtimeMode = "hybrid_tts"
	// RealtimeModeLLMOnly uses the realtime model for text generation only.
	RealtimeModeLLMOnly RealtimeMode = "llm_only"
)

// PipelineComponent identifies a pipeline slot.
type PipelineComponent string

const (
	// ComponentSTT is the speech-to-text slot.
	ComponentSTT PipelineComponent = "stt"
	// ComponentLLM is the language-model slot.
	ComponentLLM PipelineComponent = "llm"
	// ComponentTTS is the text-to-speech slot.
	ComponentTTS PipelineComponent = "tts"
	// ComponentVAD is the voice-activity-detection slot.
	ComponentVAD PipelineComponent = "vad"
	// ComponentTurnDetector is the turn-detection slot.
	ComponentTurnDetector PipelineComponent = "turn_detector"
	// ComponentAvatar is the avatar slot.
	ComponentAvatar PipelineComponent = "avatar"
	// ComponentDenoise is the noise-cancellation slot.
	ComponentDenoise PipelineComponent = "denoise"
	// ComponentRealtimeModel is the realtime speech-to-speech model slot.
	ComponentRealtimeModel PipelineComponent = "realtime_model"
)

// SpeechEventType identifies an STT speech event type.
type SpeechEventType string

const (
	// SpeechEventStart signals the start of detected speech.
	SpeechEventStart SpeechEventType = "start_of_speech"
	// SpeechEventInterim carries an interim (non-final) transcript.
	SpeechEventInterim SpeechEventType = "interim_transcript"
	// SpeechEventPreflight carries a preflight transcript emitted before finalization.
	SpeechEventPreflight SpeechEventType = "preflight_transcript"
	// SpeechEventFinal carries a final transcript.
	SpeechEventFinal SpeechEventType = "final_transcript"
	// SpeechEventEnd signals the end of detected speech.
	SpeechEventEnd SpeechEventType = "end_of_speech"
)

// VADEventType identifies a VAD edge event.
type VADEventType string

const (
	// VADEventStartOfSpeech signals the VAD detected the start of speech.
	VADEventStartOfSpeech VADEventType = "start_of_speech"
	// VADEventEndOfSpeech signals the VAD detected the end of speech.
	VADEventEndOfSpeech VADEventType = "end_of_speech"
)

// ChatRole identifies a chat message role.
type ChatRole string

const (
	// ChatRoleSystem is the system-instruction role.
	ChatRoleSystem ChatRole = "system"
	// ChatRoleDeveloper is the developer-instruction role.
	ChatRoleDeveloper ChatRole = "developer"
	// ChatRoleUser is the end-user role.
	ChatRoleUser ChatRole = "user"
	// ChatRoleAssistant is the assistant (model) role.
	ChatRoleAssistant ChatRole = "assistant"
)

// ToolChoice identifies the tool-choice policy.
type ToolChoice string

const (
	// ToolChoiceAuto lets the model decide whether to call a tool.
	ToolChoiceAuto ToolChoice = "auto"
	// ToolChoiceNone forbids the model from calling tools.
	ToolChoiceNone ToolChoice = "none"
	// ToolChoiceRequired forces the model to call a tool.
	ToolChoiceRequired ToolChoice = "required"
)

// RecordingFormat identifies the recording container/codec.
type RecordingFormat string

const (
	// RecordingFormatWAV records to a WAV container.
	RecordingFormatWAV RecordingFormat = "wav"
	// RecordingFormatOGGOpus records to an Ogg container with Opus codec.
	RecordingFormatOGGOpus RecordingFormat = "ogg_opus"
	// RecordingFormatMP3 records to an MP3 file.
	RecordingFormatMP3 RecordingFormat = "mp3"
	// RecordingFormatFLAC records to a FLAC file.
	RecordingFormatFLAC RecordingFormat = "flac"
)

// RecordingChannelMode identifies the recording channel layout.
type RecordingChannelMode string

const (
	// RecordingChannelMixed mixes all participants into a single channel.
	RecordingChannelMixed RecordingChannelMode = "mixed"
	// RecordingChannelDual records participants on separate channels.
	RecordingChannelDual RecordingChannelMode = "dual_channel"
)

// RecordingTranscriptFormat identifies the recording transcript format.
type RecordingTranscriptFormat string

const (
	// RecordingTranscriptJSON writes the transcript as JSON.
	RecordingTranscriptJSON RecordingTranscriptFormat = "json"
	// RecordingTranscriptSRT writes the transcript as SubRip (SRT) subtitles.
	RecordingTranscriptSRT RecordingTranscriptFormat = "srt"
	// RecordingTranscriptVTT writes the transcript as WebVTT subtitles.
	RecordingTranscriptVTT RecordingTranscriptFormat = "vtt"
)

// RecordingState identifies a recording lifecycle state.
type RecordingState string

const (
	// RecordingStateIdle means no recording is in progress.
	RecordingStateIdle RecordingState = "idle"
	// RecordingStateRecording means a recording is actively capturing.
	RecordingStateRecording RecordingState = "recording"
	// RecordingStateFinalizing means the recording is being finalized.
	RecordingStateFinalizing RecordingState = "finalizing"
	// RecordingStateUploading means the recording is being uploaded.
	RecordingStateUploading RecordingState = "uploading"
	// RecordingStateCompleted means the recording finished successfully.
	RecordingStateCompleted RecordingState = "completed"
	// RecordingStateFailed means the recording failed.
	RecordingStateFailed RecordingState = "failed"
)
