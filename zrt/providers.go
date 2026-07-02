package zrt

import "cmp"

// ---------------------------------------------------------------------------
// Response value types
// ---------------------------------------------------------------------------

// SpeechData carries transcript data from an STT event.
type SpeechData struct {
	// Text is the recognized transcript text.
	Text string
	// Confidence is the recognizer's confidence score for the transcript.
	Confidence float64
	// Language is the detected or configured language code.
	Language string
	// StartTime is the utterance start time, in seconds.
	StartTime float64
	// EndTime is the utterance end time, in seconds.
	EndTime float64
	// Duration is the utterance duration, in seconds.
	Duration float64
}

// STTResponse is a speech-to-text event delivered to OnTranscript callbacks.
type STTResponse struct {
	// EventType is the kind of speech event (e.g. interim or final transcript).
	EventType SpeechEventType
	// Data is the transcript payload for this event.
	Data SpeechData
	// Metadata holds provider-specific extra fields.
	Metadata map[string]string
}

// LLMResponse is a single chunk of an LLM response.
type LLMResponse struct {
	// Content is the text content of this response chunk.
	Content string
	// Role is the chat role that produced the content.
	Role ChatRole
	// Metadata holds provider-specific extra fields.
	Metadata map[string]any
}

// VADData carries voice-activity data from a VAD event.
type VADData struct {
	// IsSpeech reports whether speech is currently detected.
	IsSpeech bool
	// Confidence is the speech-probability score for this event.
	Confidence float64
	// Timestamp is the event time, in seconds.
	Timestamp float64
	// SpeechDuration is the accumulated speech duration, in seconds.
	SpeechDuration float64
	// SilenceDuration is the accumulated silence duration, in seconds.
	SilenceDuration float64
}

// VADResponse is a VAD edge event delivered to OnVADEvent callbacks.
type VADResponse struct {
	// EventType is the kind of VAD edge event (e.g. speech start or end).
	EventType VADEventType
	// Data is the voice-activity payload for this event.
	Data VADData
	// Metadata holds provider-specific extra fields.
	Metadata map[string]any
}

// ---------------------------------------------------------------------------
// Runtime config value types
// ---------------------------------------------------------------------------

// STTRuntimeConfig holds speech-to-text settings for a pipeline.
type STTRuntimeConfig struct {
	// Provider is the STT provider name.
	Provider string
	// Model is the STT model id.
	Model string
	// Language is the recognition language code.
	Language string
	// EndpointingMs is the silence duration, in milliseconds, used to finalize an utterance.
	EndpointingMs uint32
	// Fallbacks lists alternative STT configs tried in order if the primary fails.
	Fallbacks []STTRuntimeConfig
}

// LLMRuntimeConfig holds LLM settings for a pipeline.
type LLMRuntimeConfig struct {
	// Provider is the LLM provider name.
	Provider string
	// Model is the LLM model id.
	Model string
	// Temperature controls sampling randomness.
	Temperature float32
	// MaxOutputTokens caps the number of tokens generated per response.
	MaxOutputTokens uint32
	// Fallbacks lists alternative LLM configs tried in order if the primary fails.
	Fallbacks []LLMRuntimeConfig
}

// TTSRuntimeConfig holds text-to-speech settings for a pipeline.
type TTSRuntimeConfig struct {
	// Provider is the TTS provider name.
	Provider string
	// Model is the TTS model id.
	Model string
	// Language is the synthesis language code.
	Language string
	// Voice is the voice id used for synthesis.
	Voice string
	// Fallbacks lists alternative TTS configs tried in order if the primary fails.
	Fallbacks []TTSRuntimeConfig
}

// AvatarRuntimeConfig is the avatar slice of a pipeline config.
type AvatarRuntimeConfig struct {
	// Provider is the avatar provider name.
	Provider string
	// AvatarID identifies the avatar to render.
	AvatarID string
	// FaceID identifies the face used by the avatar.
	FaceID string
	// Model is the avatar model id.
	Model string
	// IsTrinity reports whether the Trinity avatar pipeline is used.
	IsTrinity bool
	// Params holds provider-specific avatar parameters.
	Params map[string]string
}

// VADRuntimeConfig holds voice-activity-detection settings for a pipeline.
type VADRuntimeConfig struct {
	// Threshold is the speech-probability threshold for detecting speech onset.
	Threshold float32
	// StopThreshold is the speech-probability threshold for detecting speech end.
	StopThreshold float32
	// MinSpeechDuration is the minimum speech duration, in seconds, to count as speech.
	MinSpeechDuration float32
	// MinSilenceDuration is the minimum silence duration, in seconds, to end an utterance.
	MinSilenceDuration float32
	// PaddingDuration is the audio padding, in seconds, added around detected speech.
	PaddingDuration float32
	// MaxBufferedSpeech is the maximum buffered speech duration, in seconds.
	MaxBufferedSpeech float32
	// ForceCPU forces the VAD model to run on CPU.
	ForceCPU bool
	// SmoothingFactor smooths the speech-probability signal.
	SmoothingFactor float32
	// InputSampleRate is the input audio sample rate, in Hz.
	InputSampleRate uint32
	// MinVolume is the minimum RMS volume threshold below which audio is treated as silence.
	MinVolume float32
}

// TurnRuntimeConfig holds turn-detection (end-of-utterance) settings for a pipeline.
type TurnRuntimeConfig struct {
	// Threshold is the end-of-utterance probability threshold.
	Threshold float32
	// ModelID is the turn-detector model id.
	ModelID string
	// Host is the turn-detector service host.
	Host string
	// AuthToken is the auth token for the turn-detector service.
	AuthToken string
	// Language is the turn-detection language code.
	Language string
	// HasThreshold reports whether a threshold was explicitly set.
	HasThreshold bool
}

// InferenceInfo describes a provider's dedicated-inference settings.
type InferenceInfo struct {
	// IsInference reports whether the provider uses dedicated inference.
	IsInference bool
	// BaseURL is the dedicated-inference endpoint base URL.
	BaseURL string
	// Location is the dedicated-inference region or location.
	Location string
}

// RealtimeInfo describes a realtime (speech-to-speech) model.
type RealtimeInfo struct {
	Provider string // defaults to ProviderName() when empty
	// Model is the realtime model id.
	Model string
	// Voice is the voice id used for speech output.
	Voice string
	// Params holds provider-specific realtime parameters.
	Params map[string]string
	// ResponseModalities lists the enabled output modalities (e.g. text, audio).
	ResponseModalities []string
	// BaseURL overrides the realtime endpoint base URL.
	BaseURL string
	// Vertex holds Vertex AI configuration for Gemini realtime models, if any.
	Vertex *VertexInfo
}

// VertexInfo holds Vertex AI configuration for Gemini providers.
type VertexInfo struct {
	// ProjectID is the Google Cloud project id.
	ProjectID string
	// Location is the Vertex AI region.
	Location string
	// ServiceAccountJSON is the service-account credentials as a JSON string.
	ServiceAccountJSON string
	// ServiceAccountPath is the filesystem path to service-account credentials.
	ServiceAccountPath string
}

// SafetySetting is a Gemini safety setting (category + threshold).
type SafetySetting struct {
	// Category is the safety category to configure.
	Category string
	// Threshold is the block threshold for the category.
	Threshold string
}

// GeminiLLMExtras holds Gemini-specific LLM config (thinking, safety, vertex).
type GeminiLLMExtras struct {
	// ThinkingBudget is the token budget for thinking; nil leaves it unset (provider default).
	ThinkingBudget *int
	// IncludeThoughts reports whether thinking output is included in responses.
	IncludeThoughts bool
	// SafetySettings lists per-category safety thresholds.
	SafetySettings []SafetySetting
	// Vertex holds Vertex AI configuration for Gemini, if any.
	Vertex *VertexInfo
}

// ---------------------------------------------------------------------------
// Provider interfaces
// ---------------------------------------------------------------------------

// Provider is implemented by every pipeline component descriptor.
type Provider interface {
	ProviderName() string
	APIKey() string
	Knobs() map[string]any
	InferenceInfo() InferenceInfo
}

// STT is a speech-to-text provider descriptor.
type STT interface {
	Provider
	STTConfig() STTRuntimeConfig
}

// LLMLike is anything accepted in the Pipeline LLM slot: a text LLM or a realtime model.
type LLMLike interface {
	ProviderName() string
}

// LLM is a text large-language-model provider descriptor.
type LLM interface {
	Provider
	LLMConfig() LLMRuntimeConfig
}

// TTS is a text-to-speech provider descriptor.
type TTS interface {
	Provider
	TTSConfig() TTSRuntimeConfig
	VoiceEmbedding() []float64
	SampleRate() int
	NumChannels() int
}

// VAD is a voice-activity-detection provider descriptor.
type VAD interface {
	Provider
	VADConfig() VADRuntimeConfig
}

// EOU is a turn-detector / end-of-utterance provider descriptor.
type EOU interface {
	Provider
	TurnConfig() TurnRuntimeConfig
}

// Avatar is a video-avatar provider descriptor.
type Avatar interface {
	Provider
	AvatarConfig() AvatarRuntimeConfig
}

// RealtimeModel is a realtime speech-to-speech model descriptor.
type RealtimeModel interface {
	LLMLike
	APIKey() string
	RealtimeInfo() RealtimeInfo
	IsRealtimeModel() bool
}

// GeminiExtrasProvider exposes Gemini-specific LLM configuration.
type GeminiExtrasProvider interface {
	GeminiLLMExtras() *GeminiLLMExtras
}

// ---------------------------------------------------------------------------
// Base structs
// ---------------------------------------------------------------------------

// ProviderBase holds fields common to every component descriptor. Embed it in a
// provider and call Init (and optionally SetInference / SetKnobs).
type ProviderBase struct {
	name              string
	apiKey            string
	knobs             map[string]any
	isInference       bool
	inferenceBaseURL  string
	inferenceLocation string
	inferenceConfig   map[string]any
	session           *AgentSession
}

// Init sets the provider name and API key.
func (p *ProviderBase) Init(name, apiKey string) {
	p.name = name
	p.apiKey = apiKey
}

// SetKnobs sets the provider-specific credential knobs.
func (p *ProviderBase) SetKnobs(knobs map[string]any) { p.knobs = knobs }

// SetInference marks this provider as dedicated-inference. location may be empty.
func (p *ProviderBase) SetInference(baseURL, location string) {
	p.isInference = true
	p.inferenceBaseURL = baseURL
	p.inferenceLocation = location
}

// SetInferenceConfig sets the dedicated-inference configuration map.
func (p *ProviderBase) SetInferenceConfig(cfg map[string]any) { p.inferenceConfig = cfg }

// InferenceConfigMap returns the dedicated-inference configuration map (may be nil).
func (p *ProviderBase) InferenceConfigMap() map[string]any { return p.inferenceConfig }

// ProviderName returns the provider name.
func (p *ProviderBase) ProviderName() string { return p.name }

// APIKey returns the configured API key ("" if none).
func (p *ProviderBase) APIKey() string { return p.apiKey }

// Knobs returns the provider-specific credential knobs (may be nil).
func (p *ProviderBase) Knobs() map[string]any { return p.knobs }

// InferenceInfo returns the dedicated-inference settings.
func (p *ProviderBase) InferenceInfo() InferenceInfo {
	return InferenceInfo{IsInference: p.isInference, BaseURL: p.inferenceBaseURL, Location: p.inferenceLocation}
}

func (p *ProviderBase) setSession(s *AgentSession) { p.session = s }

// BaseSTT is the base type for STT providers.
type BaseSTT struct {
	ProviderBase
	transcriptCB func(STTResponse)
}

// OnTranscript registers a callback invoked on every transcript event.
func (b *BaseSTT) OnTranscript(cb func(STTResponse))     { b.transcriptCB = cb }
func (b *BaseSTT) transcriptCallback() func(STTResponse) { return b.transcriptCB }

// BaseLLM is the base type for text-LLM providers.
type BaseLLM struct {
	ProviderBase
}

// BaseAvatar is embedded by avatar plugins. Plugins call Init in their ctor.
type BaseAvatar struct {
	ProviderBase
}

// BaseTTS is the base type for TTS providers.
type BaseTTS struct {
	ProviderBase
	sampleRate   int
	numChannels  int
	firstAudioCB func(ttfbMS, byteCount uint32)
}

// InitTTS sets the TTS provider name, API key, and sample rate.
func (b *BaseTTS) InitTTS(name, apiKey string, sampleRate int) {
	b.Init(name, apiKey)
	b.sampleRate = cmp.Or(sampleRate, 24000)
	b.numChannels = 1
}

// SampleRate returns the TTS sample rate.
func (b *BaseTTS) SampleRate() int { return b.sampleRate }

// NumChannels returns the TTS channel count.
func (b *BaseTTS) NumChannels() int { return b.numChannels }

// VoiceEmbedding returns the TTS voice embedding, or nil if none.
func (b *BaseTTS) VoiceEmbedding() []float64 { return nil }

// OnFirstAudioByte registers a callback for the first synthesized audio byte.
func (b *BaseTTS) OnFirstAudioByte(cb func(ttfbMS, byteCount uint32))     { b.firstAudioCB = cb }
func (b *BaseTTS) firstAudioByteCallback() func(ttfbMS, byteCount uint32) { return b.firstAudioCB }

// BaseVAD is the base type for VAD providers.
type BaseVAD struct {
	ProviderBase
	sampleRate int
	vadCB      func(VADResponse)
}

// InitVAD sets the VAD provider name and sample rate.
func (b *BaseVAD) InitVAD(name string, sampleRate int) {
	b.Init(name, "")
	b.sampleRate = sampleRate
}

// SampleRate returns the VAD sample rate.
func (b *BaseVAD) SampleRate() int { return b.sampleRate }

// OnVADEvent registers a callback for VAD edge events.
func (b *BaseVAD) OnVADEvent(cb func(VADResponse)) { b.vadCB = cb }
func (b *BaseVAD) vadCallback() func(VADResponse)  { return b.vadCB }

// BaseRealtime is the base type for realtime (speech-to-speech) models.
type BaseRealtime struct {
	ProviderBase
}

// IsRealtimeModel reports that this is a realtime model.
func (b *BaseRealtime) IsRealtimeModel() bool { return true }

// BaseEOU is the base type for turn-detector providers.
type BaseEOU struct {
	ProviderBase
	threshold float64
}

// InitEOU sets the turn-detector provider name and threshold.
func (b *BaseEOU) InitEOU(name string, threshold float64) {
	b.Init(name, "")
	b.threshold = threshold
}

// Threshold returns the configured EOU threshold.
func (b *BaseEOU) Threshold() float64 { return b.threshold }

// Internal accessor interfaces.
type sessionSettable interface{ setSession(*AgentSession) }
type transcriptObservable interface {
	transcriptCallback() func(STTResponse)
}
type firstAudioObservable interface {
	firstAudioByteCallback() func(ttfbMS, byteCount uint32)
}
type vadObservable interface {
	vadCallback() func(VADResponse)
}
