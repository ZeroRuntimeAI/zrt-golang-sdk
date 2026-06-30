// Package deepgram provides the Deepgram speech-to-text provider.
package deepgram

import (
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// STT is a Deepgram speech-to-text engine.
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

// STTOptions configures a Deepgram STT. Nil pointer fields fall back to their defaults.
type STTOptions struct {
	// APIKey overrides the DEEPGRAM_API_KEY environment variable.
	APIKey string
	// Model selects the recognition model. Defaults to "nova-2".
	Model string
	// Language is the recognition language. Defaults to "en-US".
	Language string
	// SampleRate is the audio sample rate in Hz. Defaults to 48000.
	SampleRate int
	// Endpointing is the silence (in ms) that ends an utterance. Defaults to 50.
	Endpointing *int
	// InterimResults enables partial transcripts before an utterance is final. Defaults to true.
	InterimResults *bool
	// Punctuate adds punctuation to transcripts. Defaults to true.
	Punctuate *bool
	// SmartFormat formats entities such as dates, times, and currency. Defaults to true.
	SmartFormat *bool
	// FillerWords keeps fillers such as "uh" and "um" in transcripts. Defaults to true.
	FillerWords *bool
	// Keywords boosts recognition of the given words.
	Keywords []string
	// Keyterm boosts recognition of the given key terms.
	Keyterm []string
	// ProfanityFilter masks profanity in transcripts. Defaults to false.
	ProfanityFilter *bool
	// Numerals renders spoken numbers as digits. Defaults to false.
	Numerals *bool
	// Tag attaches arbitrary tags to the request.
	Tag []string
	// Diarize labels transcripts by speaker. Defaults to false.
	Diarize *bool
	// Redact removes the given categories of sensitive content from transcripts.
	Redact []string
	// BaseURL is the Deepgram streaming endpoint. Defaults to "wss://api.deepgram.com/v1/listen".
	BaseURL string
}

// NewSTT returns a Deepgram STT configured from opts.
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

// STTConfig returns the provider, model, language, and endpointing for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "deepgram", Model: s.Model, Language: s.Language, EndpointingMs: uint32(s.Endpointing)}
}

// Knobs returns the Deepgram-specific options as a key/value map.
func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"endpointing":      s.Endpointing,
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
