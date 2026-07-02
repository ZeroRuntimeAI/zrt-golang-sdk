package azure

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Azure text-to-speech provider.
type TTS struct {
	zrt.BaseTTS
	// Voice is the Azure voice name.
	Voice string
	// SpeechRegion is the Azure service region.
	SpeechRegion string
}

// TTSOptions configures an Azure TTS instance.
type TTSOptions struct {
	// SpeechKey overrides the AZURE_SPEECH_KEY environment variable.
	SpeechKey string
	// SpeechRegion is the Azure region. Defaults to AZURE_REGION, or "eastus".
	SpeechRegion string
	// Voice is the Azure voice. Defaults to "en-US-JennyNeural".
	Voice string
}

// NewTTS returns an Azure TTS configured from opts.
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
