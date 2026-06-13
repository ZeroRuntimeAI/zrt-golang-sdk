// Package nvidia provides the NVIDIA Riva STT and TTS providers.
package nvidia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is the NVIDIA speech-to-text descriptor.
type STT struct {
	zrt.BaseSTT
	Model    string
	Language string
}

// STTOptions configures STT.
type STTOptions struct {
	// APIKey overrides the NVIDIA_API_KEY environment variable.
	APIKey       string
	Model        string // default "parakeet-1.1b-en-US-asr-streaming-silero-vad-sortformer"
	LanguageCode string // default "en-US"
}

// NewSTT builds an STT.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:    zrt.StrOr(opts.Model, "parakeet-1.1b-en-US-asr-streaming-silero-vad-sortformer"),
		Language: zrt.StrOr(opts.LanguageCode, "en-US"),
	}
	s.Init("nvidia", zrt.APIKeyOr(opts.APIKey, "NVIDIA_API_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "nvidia", Model: s.Model, Language: s.Language}
}
