// Package lmnt provides the LMNT text-to-speech provider.
package lmnt

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the LMNT text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the LMNT_API_KEY environment variable.
	APIKey string
	Voice  string // default "ava"
	Model  string // default "blizzard"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "ava"), Model: zrt.StrOr(opts.Model, "blizzard")}
	t.InitTTS("lmnt", zrt.APIKeyOr(opts.APIKey, "LMNT_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "lmnt", Voice: t.Voice}
}
