// Package nvidia provides the NVIDIA Riva STT and TTS providers.
package nvidia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is an NVIDIA Riva speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	Model    string
	Language string
}

// STTOptions configures an NVIDIA STT.
type STTOptions struct {
	// APIKey overrides the NVIDIA_API_KEY environment variable.
	APIKey string
	// Model selects the recognition model.
	// Defaults to "parakeet-1.1b-en-US-asr-streaming-silero-vad-sortformer".
	Model string
	// LanguageCode is the recognition language. Defaults to "en-US".
	LanguageCode string
}

// NewSTT returns an NVIDIA STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:    zrt.StrOr(opts.Model, "parakeet-1.1b-en-US-asr-streaming-silero-vad-sortformer"),
		Language: zrt.StrOr(opts.LanguageCode, "en-US"),
	}
	s.Init("nvidia", zrt.APIKeyOr(opts.APIKey, "NVIDIA_API_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "nvidia", Model: s.Model, Language: s.Language}
}
