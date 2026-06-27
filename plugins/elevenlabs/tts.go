// Package elevenlabs provides the ElevenLabs text-to-speech provider.
package elevenlabs

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the ElevenLabs text-to-speech provider.
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

// TTSOptions configures an ElevenLabs TTS instance. Nil pointer fields fall back
// to their default values.
type TTSOptions struct {
	// APIKey overrides the ELEVENLABS_API_KEY environment variable.
	APIKey string
	// Voice is the ElevenLabs voice id. Defaults to "21m00Tcm4TlvDq8ikWAM".
	Voice string
	// Model is the ElevenLabs model. Defaults to "eleven_turbo_v2".
	Model string
	// Stability controls voice consistency. Defaults to 0.5.
	Stability *float64
	// SimilarityBoost controls adherence to the original voice. Defaults to 0.75.
	SimilarityBoost *float64
	// Style controls stylistic exaggeration. Defaults to 0.0.
	Style *float64
	// UseSpeakerBoost enhances similarity to the speaker. Defaults to true.
	UseSpeakerBoost *bool
	// ApplyTextNormalization controls text normalization mode.
	ApplyTextNormalization string
	// EnableWordTimestamps requests per-word timing information.
	EnableWordTimestamps bool
}

// NewTTS returns an ElevenLabs TTS configured from opts.
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

// Knobs returns the provider-specific TTS settings.
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
