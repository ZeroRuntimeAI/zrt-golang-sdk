package nvidia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is an NVIDIA text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	// Voice is the Riva synthesis voice name.
	Voice string
	// Server is the Riva endpoint host. Accepted for compatibility; currently has no effect.
	Server string
}

// TTSOptions configures an NVIDIA TTS engine.
type TTSOptions struct {
	// APIKey is the NVIDIA API key. If empty, the NVIDIA_API_KEY environment variable is used.
	APIKey string
	// Voice selects the voice. Defaults to "English-US-Female-1".
	Voice string
	// LanguageCode is the language code. Defaults to "en-US". Accepted but ignored.
	LanguageCode string
	// SampleRate is the output sample rate in Hz. Defaults to 22050.
	SampleRate int
	// Server is the Riva endpoint host. Accepted for compatibility; currently has no effect.
	Server string
}

// NewTTS creates an NVIDIA TTS engine from the given options.
func NewTTS(opts TTSOptions) *TTS {
	sr := opts.SampleRate
	if sr == 0 {
		sr = 22050
	}
	t := &TTS{Voice: zrt.StrOr(opts.Voice, "English-US-Female-1"), Server: opts.Server}
	t.InitTTS("nvidia", zrt.APIKeyOr(opts.APIKey, "NVIDIA_API_KEY"), sr)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "nvidia", Voice: t.Voice}
}
