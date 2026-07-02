// Package nvidia provides the NVIDIA Riva STT and TTS providers.
package nvidia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is an NVIDIA Riva speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	// Model is the Riva/NIM ASR model identifier.
	Model string
	// Language is the recognition language as a BCP-47 tag.
	Language string
	// Server is the Riva gRPC endpoint.
	Server string
	// FunctionID is the NVCF function id for the hosted model.
	FunctionID string
	// SampleRate is the input audio sample rate in Hz.
	SampleRate int
	// UseSSL selects a TLS/SSL gRPC channel. Accepted for compatibility; currently has no effect.
	UseSSL bool
	// ProfanityFilter masks profanity in transcripts. Accepted for compatibility; currently has no effect.
	ProfanityFilter bool
	// AutomaticPunctuation adds punctuation to the transcript. Accepted for compatibility; currently has no effect.
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
	LanguageCode string
	// Server is the Riva gRPC endpoint. Defaults to "grpc.nvcf.nvidia.com:443".
	Server string
	// FunctionID is the NVCF function id for the hosted model.
	FunctionID string
	// SampleRate is the input audio sample rate in Hz. Defaults to 16000.
	SampleRate int
	// UseSSL selects a TLS/SSL gRPC channel; nil defaults to true. Accepted for compatibility; currently has no effect.
	UseSSL *bool
	// ProfanityFilter masks profanity in transcripts; nil defaults to false. Accepted for compatibility; currently has no effect.
	ProfanityFilter *bool
	// AutomaticPunctuation adds punctuation to the transcript; nil defaults to true. Accepted for compatibility; currently has no effect.
	AutomaticPunctuation *bool
}

// NewSTT returns an NVIDIA STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                zrt.StrOr(opts.Model, "parakeet-1.1b-en-US-asr-streaming-silero-vad-sortformer"),
		Language:             zrt.StrOr(opts.LanguageCode, "en-US"),
		Server:               zrt.StrOr(opts.Server, "grpc.nvcf.nvidia.com:443"),
		FunctionID:           opts.FunctionID,
		SampleRate:           zrt.IntZeroOr(opts.SampleRate, 16000),
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
