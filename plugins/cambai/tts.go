// Package cambai provides the CambAI text-to-speech provider.
package cambai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the CambAI text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures a CambAI TTS instance.
type TTSOptions struct {
	// APIKey overrides the CAMBAI_API_KEY environment variable.
	APIKey string
	// Voice is the CambAI voice. Defaults to "147320".
	Voice string
	// Model is the CambAI model. Defaults to "mars-pro".
	Model      string
	SampleRate int
}

// NewTTS returns a CambAI TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	sr := opts.SampleRate
	if sr == 0 {
		sr = 24000
	}
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "147320"), Model: zrt.StrOr(opts.Model, "mars-pro")}
	t.InitTTS("cambai", zrt.APIKeyOr(opts.APIKey, "CAMBAI_API_KEY"), sr)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "cambai", Model: t.Model, Voice: t.Voice}
}
