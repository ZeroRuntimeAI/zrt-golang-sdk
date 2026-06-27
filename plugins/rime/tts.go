// Package rime provides the Rime text-to-speech provider.
package rime

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Rime text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures a Rime TTS engine.
type TTSOptions struct {
	// APIKey is the Rime API key. If empty, the RIME_API_KEY environment variable is used.
	APIKey string
	// Speaker selects the voice. Defaults to "river".
	Speaker string
	// ModelID selects the model. Defaults to "mist".
	ModelID string
	// SamplingRate is the output sample rate in Hz. Defaults to 24000.
	SamplingRate int
}

// NewTTS creates a Rime TTS engine from the given options.
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
