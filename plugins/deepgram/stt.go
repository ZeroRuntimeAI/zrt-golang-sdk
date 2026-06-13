// Package deepgram provides the Deepgram speech-to-text provider.
package deepgram

import (
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// STT is the Deepgram speech-to-text descriptor.
type STT struct {
	zrt.BaseSTT
	Model           string
	Language        string
	SampleRate      int
	Endpointing     int
	InterimResults  bool
	Punctuate       bool
	SmartFormat     bool
	FillerWords     bool
	Keywords        []string
	Keyterm         []string
	ProfanityFilter bool
	Numerals        bool
	Tag             []string
	Diarize         bool
	Redact          []string
	BaseURL         string
}

// STTOptions configures STT. Nil pointers use default values.
type STTOptions struct {
	// APIKey overrides the DEEPGRAM_API_KEY environment variable.
	APIKey          string
	Model           string // default "nova-2"
	Language        string // default "en-US"
	SampleRate      int    // default 48000
	Endpointing     *int   // default 50
	InterimResults  *bool  // default true
	Punctuate       *bool  // default true
	SmartFormat     *bool  // default true
	FillerWords     *bool  // default true
	Keywords        []string
	Keyterm         []string
	ProfanityFilter *bool // default false
	Numerals        *bool // default false
	Tag             []string
	Diarize         *bool // default false
	Redact          []string
	BaseURL         string // default "wss://api.deepgram.com/v1/listen"
}

// NewSTT builds an STT.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:           zrt.StrOr(opts.Model, "nova-2"),
		Language:        zrt.StrOr(opts.Language, "en-US"),
		SampleRate:      orInt(opts.SampleRate, 48000),
		Endpointing:     zrt.IntOr(opts.Endpointing, 50),
		InterimResults:  zrt.BoolOr(opts.InterimResults, true),
		Punctuate:       zrt.BoolOr(opts.Punctuate, true),
		SmartFormat:     zrt.BoolOr(opts.SmartFormat, true),
		FillerWords:     zrt.BoolOr(opts.FillerWords, true),
		Keywords:        opts.Keywords,
		Keyterm:         opts.Keyterm,
		ProfanityFilter: zrt.BoolOr(opts.ProfanityFilter, false),
		Numerals:        zrt.BoolOr(opts.Numerals, false),
		Tag:             opts.Tag,
		Diarize:         zrt.BoolOr(opts.Diarize, false),
		Redact:          opts.Redact,
		BaseURL:         zrt.StrOr(opts.BaseURL, "wss://api.deepgram.com/v1/listen"),
	}
	s.Init("deepgram", zrt.APIKeyOr(opts.APIKey, "DEEPGRAM_API_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "deepgram", Model: s.Model, Language: s.Language, EndpointingMs: uint32(s.Endpointing)}
}

// Knobs implements the credential knob source.
func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"smart_format":     s.SmartFormat,
		"punctuate":        s.Punctuate,
		"filler_words":     s.FillerWords,
		"profanity_filter": s.ProfanityFilter,
		"numerals":         s.Numerals,
		"interim_results":  s.InterimResults,
		"diarize":          s.Diarize,
		"base_url":         s.BaseURL,
	}
	if len(s.Keywords) > 0 {
		k["keywords"] = s.Keywords
	}
	if len(s.Keyterm) > 0 {
		k["keyterm"] = s.Keyterm
	}
	if len(s.Tag) > 0 {
		k["tag"] = s.Tag
	}
	if len(s.Redact) > 0 {
		k["redact"] = s.Redact
	}
	return k
}

func orInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
