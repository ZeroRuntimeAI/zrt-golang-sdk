package sarvamai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Sarvam AI text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice         string
	Model         string
	Language      string
	Streaming     bool
	Pitch         float64
	Pace          float64
	Loudness      float64
	Temperature   float64
	Preprocessing bool
	Bitrate       string
}

// TTSOptions configures TTS. Nil pointers use default values.
type TTSOptions struct {
	// APIKey overrides the SARVAM_API_KEY environment variable.
	APIKey        string
	Model         string   // default "bulbul:v3"
	Language      string   // default "en-IN"
	Speaker       string   // default "shubh"
	Streaming     *bool    // default true
	Pitch         *float64 // default 0.0
	Pace          *float64 // default 1.0
	Loudness      *float64 // default 1.0
	Temperature   *float64 // default 0.6
	Preprocessing *bool    // default false
	Bitrate       string   // default "128k"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Voice:         zrt.StrOr(opts.Speaker, "shubh"),
		Model:         zrt.StrOr(opts.Model, "bulbul:v3"),
		Language:      zrt.StrOr(opts.Language, "en-IN"),
		Streaming:     zrt.BoolOr(opts.Streaming, true),
		Pitch:         zrt.FloatOr(opts.Pitch, 0.0),
		Pace:          zrt.FloatOr(opts.Pace, 1.0),
		Loudness:      zrt.FloatOr(opts.Loudness, 1.0),
		Temperature:   zrt.FloatOr(opts.Temperature, 0.6),
		Preprocessing: zrt.BoolOr(opts.Preprocessing, false),
		Bitrate:       zrt.StrOr(opts.Bitrate, "128k"),
	}
	t.InitTTS("sarvamai", zrt.APIKeyOr(opts.APIKey, "SARVAM_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "sarvamai", Voice: t.Voice}
}

// Knobs implements the credential knob source.
func (t *TTS) Knobs() map[string]any {
	return map[string]any{
		"model":         t.Model,
		"language":      t.Language,
		"streaming":     t.Streaming,
		"pitch":         t.Pitch,
		"pace":          t.Pace,
		"loudness":      t.Loudness,
		"temperature":   t.Temperature,
		"preprocessing": t.Preprocessing,
		"bitrate":       t.Bitrate,
	}
}
