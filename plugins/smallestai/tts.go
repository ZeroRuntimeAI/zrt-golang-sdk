// Package smallestai provides the Smallest AI text-to-speech provider.
package smallestai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Smallest AI text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the SMALLESTAI_API_KEY environment variable.
	APIKey  string
	VoiceID string // default "emily"
	Model   string // default "lightning"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.VoiceID, "emily"), Model: zrt.StrOr(opts.Model, "lightning")}
	t.InitTTS("smallestai", zrt.APIKeyOr(opts.APIKey, "SMALLESTAI_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "smallestai", Voice: t.Voice}
}
