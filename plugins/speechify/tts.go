// Package speechify provides the Speechify text-to-speech provider.
package speechify

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Speechify text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures a Speechify TTS engine.
type TTSOptions struct {
	// APIKey is the Speechify API key. If empty, the SPEECHIFY_API_KEY environment variable is used.
	APIKey string
	// VoiceID selects the voice. Defaults to "kristy".
	VoiceID string
	// Model selects the model. Defaults to "simba-english".
	Model string
}

// NewTTS creates a Speechify TTS engine from the given options.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.VoiceID, "kristy"), Model: zrt.StrOr(opts.Model, "simba-english")}
	t.InitTTS("speechify", zrt.APIKeyOr(opts.APIKey, "SPEECHIFY_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "speechify", Voice: t.Voice}
}
