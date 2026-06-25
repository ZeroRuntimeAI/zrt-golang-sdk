package zrt

import "cmp"

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

// STTConfig returns the primary provider's config (endpointing fixed at 50) with
// the remaining providers attached as an ordered fallback chain for the runtime.
func (f *FallbackSTT) STTConfig() STTRuntimeConfig {
	if len(f.Providers) == 0 {
		return STTRuntimeConfig{EndpointingMs: 50}
	}
	primary := f.Providers[0].STTConfig()
	out := STTRuntimeConfig{Provider: primary.Provider, Model: primary.Model, Language: primary.Language, EndpointingMs: 50}
	for _, p := range f.Providers[1:] {
		c := p.STTConfig()
		out.Fallbacks = append(out.Fallbacks, STTRuntimeConfig{
			Provider:      c.Provider,
			Model:         c.Model,
			Language:      c.Language,
			EndpointingMs: cmp.Or(c.EndpointingMs, 50),
		})
	}
	return out
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

// LLMConfig returns the primary provider's config with the remaining providers
// attached as an ordered fallback chain for the runtime.
func (f *FallbackLLM) LLMConfig() LLMRuntimeConfig {
	if len(f.Providers) == 0 {
		return LLMRuntimeConfig{Temperature: 0.7, MaxOutputTokens: 1024}
	}
	primary := f.Providers[0].LLMConfig()
	out := LLMRuntimeConfig{
		Provider:        primary.Provider,
		Model:           primary.Model,
		Temperature:     primary.Temperature,
		MaxOutputTokens: primary.MaxOutputTokens,
	}
	for _, p := range f.Providers[1:] {
		c := p.LLMConfig()
		out.Fallbacks = append(out.Fallbacks, LLMRuntimeConfig{
			Provider:        c.Provider,
			Model:           c.Model,
			Temperature:     c.Temperature,
			MaxOutputTokens: c.MaxOutputTokens,
		})
	}
	return out
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

// TTSConfig returns the primary provider's config (voice only) with the remaining
// providers attached as an ordered fallback chain for the runtime.
func (f *FallbackTTS) TTSConfig() TTSRuntimeConfig {
	if len(f.Providers) == 0 {
		return TTSRuntimeConfig{}
	}
	primary := f.Providers[0].TTSConfig()
	out := TTSRuntimeConfig{Provider: primary.Provider, Voice: primary.Voice}
	for _, p := range f.Providers[1:] {
		c := p.TTSConfig()
		out.Fallbacks = append(out.Fallbacks, TTSRuntimeConfig{Provider: c.Provider, Voice: c.Voice})
	}
	return out
}
