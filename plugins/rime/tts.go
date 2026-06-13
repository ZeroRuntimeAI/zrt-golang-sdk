// Package rime provides the Rime text-to-speech provider.
package rime

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Rime text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the RIME_API_KEY environment variable.
	APIKey       string
	Speaker      string // default "river"
	ModelID      string // default "mist"
	SamplingRate int    // default 24000
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	sr := opts.SamplingRate
	if sr == 0 {
		sr = 24000
	}
	t := &TTS{Voice: zrt.StrOr(opts.Speaker, "river"), Model: zrt.StrOr(opts.ModelID, "mist")}
	t.InitTTS("rime", zrt.APIKeyOr(opts.APIKey, "RIME_API_KEY"), sr)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "rime", Voice: t.Voice}
}
