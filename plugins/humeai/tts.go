// Package humeai provides the Hume AI text-to-speech provider.
package humeai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Hume AI text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Speed float64
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the HUMEAI_API_KEY environment variable.
	APIKey string
	Voice  string  // default "Serene Assistant"
	Speed  float64 // default 1.0
}

// NewTTS builds a TTS.
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
