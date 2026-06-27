// Package assemblyai provides the AssemblyAI speech-to-text provider.
package assemblyai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is an AssemblyAI speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	Model    string
	Language string
	Region   string
}

// STTOptions configures an AssemblyAI STT.
type STTOptions struct {
	// APIKey overrides the ASSEMBLYAI_API_KEY environment variable.
	APIKey string
	// SpeechModel selects the recognition model. Defaults to "universal-streaming-english".
	SpeechModel string
	// Region selects the service region. Defaults to "US".
	Region string
}

// NewSTT returns an AssemblyAI STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:    zrt.StrOr(opts.SpeechModel, "universal-streaming-english"),
		Language: "en-US",
		Region:   zrt.StrOr(opts.Region, "US"),
	}
	s.Init("assemblyai", zrt.APIKeyOr(opts.APIKey, "ASSEMBLYAI_API_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "assemblyai", Model: s.Model, Language: s.Language}
}
