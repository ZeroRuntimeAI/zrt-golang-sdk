package groq

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Groq text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	Voice          string
	Model          string
	Speed          float64
	ResponseFormat string
}

// TTSOptions configures a Groq TTS instance.
type TTSOptions struct {
	// APIKey overrides the GROQ_API_KEY environment variable.
	APIKey string
	// Model is the Groq model. Defaults to "canopylabs/orpheus-v1-english".
	Model string
	// Voice is the Groq voice. Defaults to "hannah".
	Voice          string
	Speed          float64
	ResponseFormat string
}

// NewTTS returns a Groq TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	speed := opts.Speed
	if speed == 0 {
		speed = 1.0
	}
	t := &TTS{
		Voice:          zrt.StrOr(opts.Voice, "hannah"),
		Model:          zrt.StrOr(opts.Model, "canopylabs/orpheus-v1-english"),
		Speed:          speed,
		ResponseFormat: zrt.StrOr(opts.ResponseFormat, "wav"),
	}
	t.InitTTS("groq", zrt.APIKeyOr(opts.APIKey, "GROQ_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "groq", Model: t.Model, Voice: t.Voice}
}

func (t *TTS) Knobs() map[string]any {
	return map[string]any{
		"response_format": t.ResponseFormat,
	}
}
