// Package inworldai provides the Inworld AI text-to-speech provider.
package inworldai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Inworld AI text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	Voice string
	Model string
}

// TTSOptions configures an Inworld AI TTS instance.
type TTSOptions struct {
	// APIKey overrides the INWORLDAI_API_KEY environment variable.
	APIKey string
	// VoiceID is the Inworld AI voice. Defaults to "Hades".
	VoiceID string
	// ModelID is the Inworld AI model. Defaults to "inworld-tts-1".
	ModelID string
}

// NewTTS returns an Inworld AI TTS configured from opts.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{Voice: zrt.StrOr(opts.VoiceID, "Hades"), Model: zrt.StrOr(opts.ModelID, "inworld-tts-1")}
	t.InitTTS("inworldai", zrt.APIKeyOr(opts.APIKey, "INWORLDAI_API_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "inworldai", Model: t.Model, Voice: t.Voice}
}
