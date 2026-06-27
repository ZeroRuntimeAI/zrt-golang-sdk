// Package smallestai provides the Smallest AI text-to-speech provider.
package smallestai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Smallest AI text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures a Smallest AI TTS engine.
type TTSOptions struct {
	// APIKey is the Smallest AI API key. If empty, the SMALLESTAI_API_KEY environment variable is used.
	APIKey string
	// VoiceID selects the voice. Defaults to "emily".
	VoiceID string
	// Model selects the model. Defaults to "lightning".
	Model string
}

// NewTTS creates a Smallest AI TTS engine from the given options.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.VoiceID, "emily"), Model: zrt.StrOr(opts.Model, "lightning")}
	t.InitTTS("smallestai", zrt.APIKeyOr(opts.APIKey, "SMALLESTAI_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "smallestai", Voice: t.Voice}
}
