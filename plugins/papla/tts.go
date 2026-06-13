// Package papla provides the Papla text-to-speech provider.
package papla

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Papla text-to-speech descriptor.
// Papla has no separate voice; the model id is used as the voice.
type TTS struct {
	zrt.BaseTTS
	Voice string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the PAPLA_API_KEY environment variable.
	APIKey  string
	ModelID string // default "papla_p1"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.ModelID, "papla_p1")}
	t.InitTTS("papla", zrt.APIKeyOr(opts.APIKey, "PAPLA_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "papla", Voice: t.Voice}
}
