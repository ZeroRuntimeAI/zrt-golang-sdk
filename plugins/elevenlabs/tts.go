// Package elevenlabs provides the ElevenLabs text-to-speech provider.
package elevenlabs

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the ElevenLabs text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice                  string
	Model                  string
	Stability              float64
	SimilarityBoost        float64
	Style                  float64
	UseSpeakerBoost        bool
	ApplyTextNormalization string
	EnableWordTimestamps   bool
}

// TTSOptions configures TTS. Nil pointers use default values.
type TTSOptions struct {
	// APIKey overrides the ELEVENLABS_API_KEY environment variable.
	APIKey                 string
	Voice                  string   // default "21m00Tcm4TlvDq8ikWAM"
	Model                  string   // default "eleven_turbo_v2"
	Stability              *float64 // default 0.5
	SimilarityBoost        *float64 // default 0.75
	Style                  *float64 // default 0.0
	UseSpeakerBoost        *bool    // default true
	ApplyTextNormalization string
	EnableWordTimestamps   bool
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Voice:                  zrt.StrOr(opts.Voice, "21m00Tcm4TlvDq8ikWAM"),
		Model:                  zrt.StrOr(opts.Model, "eleven_turbo_v2"),
		Stability:              zrt.FloatOr(opts.Stability, 0.5),
		SimilarityBoost:        zrt.FloatOr(opts.SimilarityBoost, 0.75),
		Style:                  zrt.FloatOr(opts.Style, 0.0),
		UseSpeakerBoost:        zrt.BoolOr(opts.UseSpeakerBoost, true),
		ApplyTextNormalization: opts.ApplyTextNormalization,
		EnableWordTimestamps:   opts.EnableWordTimestamps,
	}
	t.InitTTS("elevenlabs", zrt.APIKeyOr(opts.APIKey, "ELEVENLABS_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "elevenlabs", Voice: t.Voice}
}

// Knobs implements the credential knob source.
func (t *TTS) Knobs() map[string]any {
	k := map[string]any{
		"model":                  t.Model,
		"stability":              t.Stability,
		"similarity_boost":       t.SimilarityBoost,
		"style":                  t.Style,
		"use_speaker_boost":      t.UseSpeakerBoost,
		"enable_word_timestamps": t.EnableWordTimestamps,
	}
	if t.ApplyTextNormalization != "" {
		k["apply_text_normalization"] = t.ApplyTextNormalization
	}
	return k
}
