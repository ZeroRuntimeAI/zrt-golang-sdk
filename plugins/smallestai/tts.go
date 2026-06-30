// Package smallestai provides the Smallest AI text-to-speech provider.
package smallestai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Smallest AI text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	Voice    string
	Model    string
	Language string
	Speed    *float64
	Stream   bool
}

// TTSOptions configures a Smallest AI TTS engine.
type TTSOptions struct {
	// APIKey is the Smallest AI API key. If empty, the SMALLESTAI_API_KEY environment variable is used.
	APIKey string
	// VoiceID selects the voice. Defaults to "magnus".
	VoiceID string
	Voice   string
	// Model selects the model. Defaults to "lightning_v3.1".
	Model    string
	Language string
	Speed    *float64
	Stream   *bool
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
