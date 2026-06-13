package groq

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Groq text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the GROQ_API_KEY environment variable.
	APIKey string
	Model  string // default "playai-tts"
	Voice  string // default "Fritz-PlayAI"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "Fritz-PlayAI"), Model: zrt.StrOr(opts.Model, "playai-tts")}
	t.InitTTS("groq", zrt.APIKeyOr(opts.APIKey, "GROQ_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "groq", Voice: t.Voice}
}
