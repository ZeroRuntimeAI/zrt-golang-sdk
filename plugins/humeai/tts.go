// Package humeai provides the Hume AI text-to-speech provider.
package humeai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Hume AI text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Speed float64
}

// TTSOptions configures a Hume AI TTS instance.
type TTSOptions struct {
	// APIKey overrides the HUMEAI_API_KEY environment variable.
	APIKey string
	// Voice is the Hume AI voice. Defaults to "Serene Assistant".
	Voice string
	// Speed scales the speaking rate. Defaults to 1.0.
	Speed float64
}

// NewTTS returns a Hume AI TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	speed := opts.Speed
	if speed == 0 {
		speed = 1.0
	}
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "Serene Assistant"), Speed: speed}
	t.InitTTS("humeai", zrt.APIKeyOr(opts.APIKey, "HUMEAI_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "humeai", Voice: t.Voice}
}
