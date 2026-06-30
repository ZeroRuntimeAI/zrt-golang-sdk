// Package nvidia provides the NVIDIA Riva STT and TTS providers.
package nvidia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is an NVIDIA Riva speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	Model                string
	Language             string
	Server               string
	FunctionID           string
	SampleRate           int
	UseSSL               bool
	ProfanityFilter      bool
	AutomaticPunctuation bool
}

// STTOptions configures an NVIDIA STT.
type STTOptions struct {
	// APIKey overrides the NVIDIA_API_KEY environment variable.
	APIKey string
	// Model selects the recognition model.
	// Defaults to "parakeet-1.1b-en-US-asr-streaming-silero-vad-sortformer".
	Model string
	// LanguageCode is the recognition language. Defaults to "en-US".
	LanguageCode         string
	Server               string
	FunctionID           string
	SampleRate           int
	UseSSL               *bool
	ProfanityFilter      *bool
	AutomaticPunctuation *bool
}

// NewSTT returns an NVIDIA STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                zrt.StrOr(opts.Model, "parakeet-1.1b-en-US-asr-streaming-silero-vad-sortformer"),
		Language:             zrt.StrOr(opts.LanguageCode, "en-US"),
		Server:               zrt.StrOr(opts.Server, "grpc.nvcf.nvidia.com:443"),
		FunctionID:           opts.FunctionID,
		SampleRate:           orInt(opts.SampleRate, 16000),
		UseSSL:               zrt.BoolOr(opts.UseSSL, true),
		ProfanityFilter:      zrt.BoolOr(opts.ProfanityFilter, false),
		AutomaticPunctuation: zrt.BoolOr(opts.AutomaticPunctuation, true),
	}
	s.Init("nvidia", zrt.APIKeyOr(opts.APIKey, "NVIDIA_API_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "nvidia", Model: s.Model, Language: s.Language}
}

func orInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
