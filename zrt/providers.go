package zrt

// ---------------------------------------------------------------------------
// Response value types
// ---------------------------------------------------------------------------

// SpeechData carries transcript data from an STT event.
type SpeechData struct {
	Text       string
	Confidence float64
	Language   string
	StartTime  float64
	EndTime    float64
	Duration   float64
}

// STTResponse is a speech-to-text event delivered to OnTranscript callbacks.
type STTResponse struct {
	EventType SpeechEventType
	Data      SpeechData
	Metadata  map[string]string
}

// LLMResponse is an LLM response chunk (the runtime drives the LLM).
type LLMResponse struct {
	Content  string
	Role     ChatRole
	Metadata map[string]any
}

// VADData carries voice-activity data from a VAD event.
type VADData struct {
	IsSpeech        bool
	Confidence      float64
	Timestamp       float64
	SpeechDuration  float64
	SilenceDuration float64
}

// VADResponse is a VAD edge event delivered to OnVADEvent callbacks.
type VADResponse struct {
	EventType VADEventType
	Data      VADData
	Metadata  map[string]any
}

// ---------------------------------------------------------------------------
// Runtime config value types
// ---------------------------------------------------------------------------

// STTRuntimeConfig is the STT slice of a pipeline config.
type STTRuntimeConfig struct {
	Provider      string
	Model         string
	Language      string
	EndpointingMs uint32
}

// LLMRuntimeConfig is the LLM slice of a pipeline config.
type LLMRuntimeConfig struct {
	Provider        string
	Model           string
	Temperature     float32
	MaxOutputTokens uint32
}

// TTSRuntimeConfig is the TTS slice of a pipeline config.
type TTSRuntimeConfig struct {
	Provider string
	Voice    string
}

// AvatarRuntimeConfig is the avatar slice of a pipeline config.
type AvatarRuntimeConfig struct {
	Provider  string
	AvatarID  string
	FaceID    string
	Model     string
	IsTrinity bool
	Params    map[string]string
}

// VADRuntimeConfig is the VAD slice of a pipeline config.
type VADRuntimeConfig struct {
	Threshold          float32
	StopThreshold      float32
	MinSpeechDuration  float32
	MinSilenceDuration float32
	PaddingDuration    float32
	MaxBufferedSpeech  float32
	ForceCPU           bool
	SmoothingFactor    float32
	InputSampleRate    uint32
	MinVolume          float32
}

// TurnRuntimeConfig is the turn-detector / EOU slice of a pipeline config.
type TurnRuntimeConfig struct {
	Threshold float32
	ModelID   string
	Host      string
	AuthToken string
	Language  string
	// HasThreshold reports whether the provider explicitly set a threshold.
	HasThreshold bool
}

// InferenceInfo marks a provider as dedicated-inference (set by inference factories).
type InferenceInfo struct {
	IsInference bool
	BaseURL     string
	Location    string
}

// RealtimeInfo describes a realtime (speech-to-speech) model.
type RealtimeInfo struct {
	Provider           string // usually empty; falls back to ProviderName()
	Model              string
	Voice              string
	Params             map[string]string
	ResponseModalities []string
	BaseURL            string
	Vertex             *VertexInfo
}

// VertexInfo holds Vertex AI configuration for Gemini providers.
type VertexInfo struct {
	ProjectID          string
	Location           string
	ServiceAccountJSON string
	ServiceAccountPath string
}

// SafetySetting is a Gemini safety setting (category + threshold).
type SafetySetting struct {
	Category  string
	Threshold string
}

// GeminiLLMExtras holds Gemini-specific LLM config (thinking, safety, vertex).
type GeminiLLMExtras struct {
	ThinkingBudget  *int
	IncludeThoughts bool
	SafetySettings  []SafetySetting
	Vertex          *VertexInfo
}

// ---------------------------------------------------------------------------
// Provider interfaces (all methods exported so plugin packages can implement)
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

// LLMLike is anything accepted in the Pipeline LLM slot (text LLM or realtime model).
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

// GeminiExtrasProvider is implemented by the Gemini LLM to expose Gemini extras.
type GeminiExtrasProvider interface {
	GeminiLLMExtras() *GeminiLLMExtras
}

// ---------------------------------------------------------------------------
// Base structs (plugins embed these and call Init/SetInference in their ctors)
// ---------------------------------------------------------------------------

// ProviderBase holds fields common to every component descriptor. Plugin
// packages embed it and call Init (and optionally SetInference / SetKnobs).
type ProviderBase struct {
	name              string
	apiKey            string
	knobs             map[string]any
	isInference       bool
	inferenceBaseURL  string
	inferenceLocation string
	session           *AgentSession
}

// Init sets the provider name and API key. Call from a plugin constructor.
func (p *ProviderBase) Init(name, apiKey string) {
	p.name = name
	p.apiKey = apiKey
}

// SetKnobs sets the provider-specific credential knobs.
func (p *ProviderBase) SetKnobs(knobs map[string]any) { p.knobs = knobs }

// SetInference marks this provider as dedicated-inference (used by the
// inference factory helpers). location may be empty.
func (p *ProviderBase) SetInference(baseURL, location string) {
	p.isInference = true
	p.inferenceBaseURL = baseURL
	p.inferenceLocation = location
}

// ProviderName returns the provider name.
func (p *ProviderBase) ProviderName() string { return p.name }

// APIKey returns the configured API key ("" if none).
func (p *ProviderBase) APIKey() string { return p.apiKey }

// Knobs returns the provider-specific credential knobs (may be nil).
func (p *ProviderBase) Knobs() map[string]any { return p.knobs }

// InferenceInfo returns the dedicated-inference markers.
func (p *ProviderBase) InferenceInfo() InferenceInfo {
	return InferenceInfo{IsInference: p.isInference, BaseURL: p.inferenceBaseURL, Location: p.inferenceLocation}
}

func (p *ProviderBase) setSession(s *AgentSession) { p.session = s }

// BaseSTT is embedded by STT plugins.
type BaseSTT struct {
	ProviderBase
	transcriptCB func(STTResponse)
}

// OnTranscript registers a callback invoked on every transcript event.
func (b *BaseSTT) OnTranscript(cb func(STTResponse))     { b.transcriptCB = cb }
func (b *BaseSTT) transcriptCallback() func(STTResponse) { return b.transcriptCB }

// VoiceEmbedding default for STT is none (kept off the STT interface).

// BaseLLM is embedded by text-LLM plugins.
type BaseLLM struct {
	ProviderBase
}

// BaseAvatar is embedded by avatar plugins. Plugins call Init in their ctor.
type BaseAvatar struct {
	ProviderBase
}

// BaseTTS is embedded by TTS plugins.
type BaseTTS struct {
	ProviderBase
	sampleRate   int
	numChannels  int
	firstAudioCB func(ttfbMS, byteCount uint32)
}

// InitTTS sets TTS audio params (default num_channels = 1).
func (b *BaseTTS) InitTTS(name, apiKey string, sampleRate int) {
	b.Init(name, apiKey)
	if sampleRate == 0 {
		sampleRate = 24000
	}
	b.sampleRate = sampleRate
	b.numChannels = 1
}

// SampleRate returns the TTS sample rate.
func (b *BaseTTS) SampleRate() int { return b.sampleRate }

// NumChannels returns the TTS channel count (always 1).
func (b *BaseTTS) NumChannels() int { return b.numChannels }

// VoiceEmbedding returns the Cartesia voice embedding (nil for other providers).
func (b *BaseTTS) VoiceEmbedding() []float64 { return nil }

// OnFirstAudioByte registers a callback for the first synthesized audio byte.
func (b *BaseTTS) OnFirstAudioByte(cb func(ttfbMS, byteCount uint32))     { b.firstAudioCB = cb }
func (b *BaseTTS) firstAudioByteCallback() func(ttfbMS, byteCount uint32) { return b.firstAudioCB }

// BaseVAD is embedded by VAD plugins.
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

// BaseRealtime is embedded by realtime (speech-to-speech) model plugins.
type BaseRealtime struct {
	ProviderBase
}

// IsRealtimeModel reports that this is a realtime model.
func (b *BaseRealtime) IsRealtimeModel() bool { return true }

// BaseEOU is embedded by turn-detector plugins.
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

// internal accessor interfaces used by session/grpc-bridge across plugin pkgs.
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
