package zrt

// Fallback default timing knobs.
const (
	DefaultTemporaryDisableSec           = 60
	DefaultPermanentDisableAfterAttempts = 3
	DefaultRecoveryCheckIntervalSec      = 300
)

// FallbackSTT wraps an ordered list of STT providers; the runtime falls back on
// failure.
//
// Only the primary provider's provider/model/language reach the runtime config
// (endpointing fixed at 50); per-provider API keys passed inline are not
// forwarded — set them via environment variables.
type FallbackSTT struct {
	BaseSTT
	Providers []STT
}

// NewFallbackSTT builds a FallbackSTT from an ordered provider list.
func NewFallbackSTT(providers ...STT) *FallbackSTT {
	f := &FallbackSTT{Providers: providers}
	f.Init("fallback_stt", "")
	return f
}

// STTConfig returns the primary provider's config (endpointing fixed at 50).
func (f *FallbackSTT) STTConfig() STTRuntimeConfig {
	if len(f.Providers) == 0 {
		return STTRuntimeConfig{EndpointingMs: 50}
	}
	c := f.Providers[0].STTConfig()
	return STTRuntimeConfig{Provider: c.Provider, Model: c.Model, Language: c.Language, EndpointingMs: 50}
}

// FallbackLLM wraps an ordered list of LLM providers.
type FallbackLLM struct {
	BaseLLM
	Providers []LLM
}

// NewFallbackLLM builds a FallbackLLM from an ordered provider list.
func NewFallbackLLM(providers ...LLM) *FallbackLLM {
	f := &FallbackLLM{Providers: providers}
	f.Init("fallback_llm", "")
	return f
}

// LLMConfig returns the primary provider's config.
func (f *FallbackLLM) LLMConfig() LLMRuntimeConfig {
	if len(f.Providers) == 0 {
		return LLMRuntimeConfig{Temperature: 0.7, MaxOutputTokens: 1024}
	}
	return f.Providers[0].LLMConfig()
}

// FallbackTTS wraps an ordered list of TTS providers.
type FallbackTTS struct {
	BaseTTS
	Providers []TTS
}

// NewFallbackTTS builds a FallbackTTS from an ordered provider list.
func NewFallbackTTS(providers ...TTS) *FallbackTTS {
	f := &FallbackTTS{Providers: providers}
	f.InitTTS("fallback_tts", "", 24000)
	return f
}

// TTSConfig returns the primary provider's config (voice only).
func (f *FallbackTTS) TTSConfig() TTSRuntimeConfig {
	if len(f.Providers) == 0 {
		return TTSRuntimeConfig{}
	}
	c := f.Providers[0].TTSConfig()
	return TTSRuntimeConfig{Provider: c.Provider, Voice: c.Voice}
}
