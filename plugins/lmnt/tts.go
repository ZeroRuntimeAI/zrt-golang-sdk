// Package lmnt provides the LMNT text-to-speech provider.
package lmnt

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the LMNT text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures an LMNT TTS instance.
type TTSOptions struct {
	// APIKey overrides the LMNT_API_KEY environment variable.
	APIKey string
	// Voice is the LMNT voice. Defaults to "ava".
	Voice string
	// Model is the LMNT model. Defaults to "blizzard".
	Model string
}

// NewTTS returns an LMNT TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "ava"), Model: zrt.StrOr(opts.Model, "blizzard")}
	t.InitTTS("lmnt", zrt.APIKeyOr(opts.APIKey, "LMNT_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "lmnt", Model: t.Model, Voice: t.Voice}
}
