// Package speechify provides the Speechify text-to-speech provider.
package speechify

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Speechify text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the SPEECHIFY_API_KEY environment variable.
	APIKey  string
	VoiceID string // default "kristy"
	Model   string // default "simba-english"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.VoiceID, "kristy"), Model: zrt.StrOr(opts.Model, "simba-english")}
	t.InitTTS("speechify", zrt.APIKeyOr(opts.APIKey, "SPEECHIFY_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "speechify", Voice: t.Voice}
}
