// Package cambai provides the CambAI text-to-speech provider.
package cambai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the CambAI text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the CAMBAI_API_KEY environment variable.
	APIKey string
	Voice  string // default "147320"
	Model  string // default "mars-pro"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "147320"), Model: zrt.StrOr(opts.Model, "mars-pro")}
	t.InitTTS("cambai", zrt.APIKeyOr(opts.APIKey, "CAMBAI_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "cambai", Voice: t.Voice}
}
