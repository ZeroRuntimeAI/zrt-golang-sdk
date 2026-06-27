// Package murfai provides the Murf AI text-to-speech provider.
package murfai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Murf AI text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures a Murf AI TTS engine.
type TTSOptions struct {
	// APIKey is the Murf AI API key. If empty, the MURFAI_API_KEY environment variable is used.
	APIKey string
	// Voice selects the voice. Defaults to "en-US-natalie".
	Voice string
	// Model selects the model. Defaults to "Falcon".
	Model string
}

// NewTTS creates a Murf AI TTS engine from the given options.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "en-US-natalie"), Model: zrt.StrOr(opts.Model, "Falcon")}
	t.InitTTS("murfai", zrt.APIKeyOr(opts.APIKey, "MURFAI_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "murfai", Voice: t.Voice}
}
