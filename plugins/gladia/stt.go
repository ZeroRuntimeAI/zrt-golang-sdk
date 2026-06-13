// Package gladia provides the Gladia speech-to-text provider.
package gladia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is the Gladia speech-to-text descriptor.
type STT struct {
	zrt.BaseSTT
	Model    string
	Language string
}

// STTOptions configures STT.
type STTOptions struct {
	// APIKey overrides the GLADIA_API_KEY environment variable.
	APIKey    string
	Model     string   // default "solaria-1"
	Languages []string // language = first element, default "english"
}

// NewSTT builds an STT.
func NewSTT(opts STTOptions) *STT {
	lang := "english"
	if len(opts.Languages) > 0 {
		lang = opts.Languages[0]
	}
	s := &STT{Model: zrt.StrOr(opts.Model, "solaria-1"), Language: lang}
	s.Init("gladia", zrt.APIKeyOr(opts.APIKey, "GLADIA_API_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "gladia", Model: s.Model, Language: s.Language}
}
