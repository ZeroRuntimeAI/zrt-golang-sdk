// Package gladia provides the Gladia speech-to-text provider.
package gladia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is a Gladia speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	Model    string
	Language string
}

// STTOptions configures a Gladia STT.
type STTOptions struct {
	// APIKey overrides the GLADIA_API_KEY environment variable.
	APIKey string
	// Model selects the recognition model. Defaults to "solaria-1".
	Model string
	// Languages lists the recognition languages; the first entry is used.
	// Defaults to "english".
	Languages []string
}

// NewSTT returns a Gladia STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	lang := "english"
	if len(opts.Languages) > 0 {
		lang = opts.Languages[0]
	}
	s := &STT{Model: zrt.StrOr(opts.Model, "solaria-1"), Language: lang}
	s.Init("gladia", zrt.APIKeyOr(opts.APIKey, "GLADIA_API_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "gladia", Model: s.Model, Language: s.Language}
}
