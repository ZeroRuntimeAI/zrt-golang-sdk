// Package sarvamai provides the Sarvam AI STT, LLM and TTS providers.
package sarvamai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is the Sarvam AI speech-to-text descriptor.
type STT struct {
	zrt.BaseSTT
	Model              string
	Language           string
	InputSampleRate    int
	OutputSampleRate   int
	Mode               string
	HighVADSensitivity *bool
	FlushSignals       *bool
	Translation        bool
	Prompt             string
}

// STTOptions configures STT.
type STTOptions struct {
	// APIKey overrides the SARVAM_API_KEY environment variable.
	APIKey             string
	Model              string // default "saaras:v3"
	Language           string // default "en-IN"
	InputSampleRate    int    // default 48000
	OutputSampleRate   int    // default 16000
	Mode               string
	HighVADSensitivity *bool
	FlushSignals       *bool
	Translation        bool
	Prompt             string
}

// NewSTT builds an STT.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:              zrt.StrOr(opts.Model, "saaras:v3"),
		Language:           zrt.StrOr(opts.Language, "en-IN"),
		InputSampleRate:    orInt(opts.InputSampleRate, 48000),
		OutputSampleRate:   orInt(opts.OutputSampleRate, 16000),
		Mode:               opts.Mode,
		HighVADSensitivity: opts.HighVADSensitivity,
		FlushSignals:       opts.FlushSignals,
		Translation:        opts.Translation,
		Prompt:             opts.Prompt,
	}
	s.Init("sarvamai", zrt.APIKeyOr(opts.APIKey, "SARVAM_API_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "sarvamai", Model: s.Model, Language: s.Language}
}

// Knobs implements the credential knob source (general + STT knobs).
func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"model":              s.Model,
		"language":           s.Language,
		"translation":        s.Translation,
		"input_sample_rate":  s.InputSampleRate,
		"output_sample_rate": s.OutputSampleRate,
	}
	if s.Mode != "" {
		k["mode"] = s.Mode
	}
	if s.Prompt != "" {
		k["prompt"] = s.Prompt
	}
	if s.HighVADSensitivity != nil {
		k["high_vad_sensitivity"] = *s.HighVADSensitivity
	}
	if s.FlushSignals != nil {
		k["flush_signals"] = *s.FlushSignals
	}
	return k
}

func orInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
