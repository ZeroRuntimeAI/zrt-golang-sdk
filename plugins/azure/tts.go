package azure

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Azure text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice        string
	SpeechRegion string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	SpeechKey    string
	SpeechRegion string // default from AZURE_REGION, else "eastus"
	Voice        string // default "en-US-JennyNeural"
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Voice:        zrt.StrOr(opts.Voice, "en-US-JennyNeural"),
		SpeechRegion: zrt.StrOr(opts.SpeechRegion, zrt.EnvOr("AZURE_REGION", "eastus")),
	}
	t.InitTTS("azure", zrt.APIKeyOr(opts.SpeechKey, "AZURE_SPEECH_KEY"), 24000)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "azure", Voice: t.Voice}
}
