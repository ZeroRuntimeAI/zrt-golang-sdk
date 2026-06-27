package groq

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Groq text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures a Groq TTS instance.
type TTSOptions struct {
	// APIKey overrides the GROQ_API_KEY environment variable.
	APIKey string
	// Model is the Groq model. Defaults to "playai-tts".
	Model string
	// Voice is the Groq voice. Defaults to "Fritz-PlayAI".
	Voice string
}

// NewTTS returns a Groq TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "Fritz-PlayAI"), Model: zrt.StrOr(opts.Model, "playai-tts")}
	t.InitTTS("groq", zrt.APIKeyOr(opts.APIKey, "GROQ_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "groq", Voice: t.Voice}
}
