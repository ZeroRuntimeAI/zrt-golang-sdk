// Package assemblyai provides the AssemblyAI speech-to-text provider.
package assemblyai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is an AssemblyAI speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	Model                            string
	Language                         string
	InputSampleRate                  int
	OutputSampleRate                 int
	Encoding                         string
	FormatTurns                      bool
	EndOfTurnConfidenceThreshold     float64
	MinEndOfTurnSilenceWhenConfident int
	MaxTurnSilence                   int
	KeytermsPrompt                   []string
	LanguageDetection                bool
	BaseURL                          string
}

// STTOptions configures an AssemblyAI STT.
type STTOptions struct {
	// APIKey overrides the ASSEMBLYAI_API_KEY environment variable.
	APIKey                           string
	Model                            string
	Language                         string
	InputSampleRate                  int
	OutputSampleRate                 int
	Encoding                         string
	FormatTurns                      *bool
	EndOfTurnConfidenceThreshold     *float64
	MinEndOfTurnSilenceWhenConfident *int
	MaxTurnSilence                   *int
	KeytermsPrompt                   []string
	LanguageDetection                *bool
	BaseURL                          string
}

// NewSTT returns an AssemblyAI STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                            zrt.StrOr(opts.Model, "universal-streaming-english"),
		Language:                         zrt.StrOr(opts.Language, "en-US"),
		InputSampleRate:                  orInt(opts.InputSampleRate, 48000),
		OutputSampleRate:                 orInt(opts.OutputSampleRate, 16000),
		Encoding:                         zrt.StrOr(opts.Encoding, "pcm_s16le"),
		FormatTurns:                      zrt.BoolOr(opts.FormatTurns, true),
		EndOfTurnConfidenceThreshold:     zrt.FloatOr(opts.EndOfTurnConfidenceThreshold, 0.4),
		MinEndOfTurnSilenceWhenConfident: zrt.IntOr(opts.MinEndOfTurnSilenceWhenConfident, 560),
		MaxTurnSilence:                   zrt.IntOr(opts.MaxTurnSilence, 2400),
		KeytermsPrompt:                   opts.KeytermsPrompt,
		LanguageDetection:                zrt.BoolOr(opts.LanguageDetection, false),
		BaseURL:                          opts.BaseURL,
	}
	s.Init("assemblyai", zrt.APIKeyOr(opts.APIKey, "ASSEMBLYAI_API_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "assemblyai", Model: s.Model, Language: s.Language}
}

func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"output_sample_rate":                     s.OutputSampleRate,
		"encoding":                               s.Encoding,
		"format_turns":                           s.FormatTurns,
		"end_of_turn_confidence_threshold":       s.EndOfTurnConfidenceThreshold,
		"min_end_of_turn_silence_when_confident": s.MinEndOfTurnSilenceWhenConfident,
		"max_turn_silence":                       s.MaxTurnSilence,
		"language_detection":                     s.LanguageDetection,
	}
	if len(s.KeytermsPrompt) > 0 {
		k["keyterms_prompt"] = s.KeytermsPrompt
	}
	if s.BaseURL != "" {
		k["base_url"] = s.BaseURL
	}
	return k
}

func orInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
