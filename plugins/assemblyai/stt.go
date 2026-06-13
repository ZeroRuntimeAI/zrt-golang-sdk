// Package assemblyai provides the AssemblyAI speech-to-text provider.
package assemblyai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is the AssemblyAI speech-to-text descriptor.
type STT struct {
	zrt.BaseSTT
	Model    string
	Language string
	Region   string
}

// STTOptions configures STT.
type STTOptions struct {
	// APIKey overrides the ASSEMBLYAI_API_KEY environment variable.
	APIKey      string
	SpeechModel string // default "universal-streaming-english"
	Region      string // default "US"
}

// NewSTT builds an STT.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:    zrt.StrOr(opts.SpeechModel, "universal-streaming-english"),
		Language: "en-US",
		Region:   zrt.StrOr(opts.Region, "US"),
	}
	s.Init("assemblyai", zrt.APIKeyOr(opts.APIKey, "ASSEMBLYAI_API_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "assemblyai", Model: s.Model, Language: s.Language}
}
