// Package murfai provides the Murf AI text-to-speech provider.
package murfai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Murf AI text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the MURFAI_API_KEY environment variable.
	APIKey string
	Voice  string // default "en-US-natalie"
	Model  string // default "Falcon"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "en-US-natalie"), Model: zrt.StrOr(opts.Model, "Falcon")}
	t.InitTTS("murfai", zrt.APIKeyOr(opts.APIKey, "MURFAI_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "murfai", Voice: t.Voice}
}
