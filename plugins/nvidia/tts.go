package nvidia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the NVIDIA text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the NVIDIA_API_KEY environment variable.
	APIKey       string
	Voice        string // default "English-US-Female-1"
	LanguageCode string // default "en-US" (not serialized)
	SampleRate   int    // default 22050
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	sr := opts.SampleRate
	if sr == 0 {
		sr = 22050
	}
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "English-US-Female-1")}
	t.InitTTS("nvidia", zrt.APIKeyOr(opts.APIKey, "NVIDIA_API_KEY"), sr)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "nvidia", Voice: t.Voice}
}
