package google

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is the Google speech-to-text descriptor (provider name "google_stt").
type STT struct {
	zrt.BaseSTT
	Model           string
	Language        string
	Punctuate       bool
	ProfanityFilter bool
}

// STTOptions configures STT.
type STTOptions struct {
	// APIKey overrides the GOOGLE_API_KEY environment variable.
	APIKey          string
	Languages       []string // default ["en-US"]
	Language        string   // default "en-US" (used if Languages empty)
	Model           string   // default "latest_long"
	Punctuate       *bool    // default true
	ProfanityFilter *bool    // default false
}

// NewSTT builds an STT.
func NewSTT(opts STTOptions) *STT {
	lang := opts.Language
	if len(opts.Languages) > 0 {
		lang = opts.Languages[0]
	}
	lang = zrt.StrOr(lang, "en-US")
	s := &STT{
		Model:           zrt.StrOr(opts.Model, "latest_long"),
		Language:        lang,
		Punctuate:       zrt.BoolOr(opts.Punctuate, true),
		ProfanityFilter: zrt.BoolOr(opts.ProfanityFilter, false),
	}
	s.Init("google_stt", zrt.APIKeyOr(opts.APIKey, "GOOGLE_API_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "google_stt", Model: s.Model, Language: s.Language}
}

// Knobs implements the credential knob source.
func (s *STT) Knobs() map[string]any {
	return map[string]any{"punctuate": s.Punctuate, "profanity_filter": s.ProfanityFilter}
}
