// Package smallestai provides the Smallest AI text-to-speech provider.
package smallestai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Smallest AI text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	// Voice is the resolved voice name used for synthesis.
	Voice string
	// Model is the resolved model id.
	Model string
	// Language is the ISO 639-1 language code of the target voice.
	Language string
	// Speed scales the speaking rate; nil applies the provider default.
	Speed *float64
	// Stream reports whether audio is streamed as it is generated.
	Stream bool
}

// TTSOptions configures a Smallest AI TTS engine.
type TTSOptions struct {
	// APIKey is the Smallest AI API key. If empty, the SMALLESTAI_API_KEY environment variable is used.
	APIKey string
	// VoiceID selects the voice. Defaults to "magnus".
	VoiceID string
	// Voice selects the voice by name; it takes precedence over VoiceID when both are set.
	Voice string
	// Model selects the model. Defaults to "lightning_v3.1".
	Model string
	// Language is the ISO 639-1 language code of the target voice. Defaults to "en".
	Language string
	// Speed scales the speaking rate; nil applies the provider default.
	Speed *float64
	// Stream enables streaming synthesis. nil defaults to true.
	Stream *bool
}

// NewTTS creates a Smallest AI TTS engine from the given options.
func NewTTS(opts TTSOptions) *TTS {
	voice := zrt.StrOr(opts.Voice, zrt.StrOr(opts.VoiceID, "magnus"))
	t := &TTS{
		Voice:    voice,
		Model:    zrt.StrOr(opts.Model, "lightning_v3.1"),
		Language: zrt.StrOr(opts.Language, "en"),
		Speed:    opts.Speed,
		Stream:   zrt.BoolOr(opts.Stream, true),
	}
	t.InitTTS("smallestai", zrt.APIKeyOr(opts.APIKey, "SMALLESTAI_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "smallestai", Model: t.Model, Voice: t.Voice}
}

// Knobs returns provider-specific runtime settings for the TTS engine.
func (t *TTS) Knobs() map[string]any {
	k := map[string]any{
		"tts_stream": t.Stream,
		"model":      t.Model,
		"language":   t.Language,
	}
	if t.Speed != nil {
		k["speed"] = *t.Speed
	}
	return k
}
