package google

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is a Google speech-to-text engine (provider name "google_stt").
type STT struct {
	zrt.BaseSTT
	Model           string
	Language        string
	Punctuate       bool
	ProfanityFilter bool
}

// STTOptions configures a Google STT.
type STTOptions struct {
	// APIKey overrides the GOOGLE_API_KEY environment variable.
	APIKey string
	// Languages lists the recognition languages; the first entry is used.
	// Defaults to ["en-US"].
	Languages []string
	// Language is the recognition language used when Languages is empty.
	// Defaults to "en-US".
	Language string
	// Model selects the recognition model. Defaults to "latest_long".
	Model string
	// Punctuate adds punctuation to transcripts. Defaults to true.
	Punctuate *bool
	// ProfanityFilter masks profanity in transcripts. Defaults to false.
	ProfanityFilter *bool
}

// NewSTT returns a Google STT configured from opts.
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

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "google_stt", Model: s.Model, Language: s.Language}
}

// Knobs returns the Google-specific options as a key/value map.
func (s *STT) Knobs() map[string]any {
	return map[string]any{"punctuate": s.Punctuate, "profanity_filter": s.ProfanityFilter}
}
