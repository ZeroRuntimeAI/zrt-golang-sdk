// Package neuphonic provides the Neuphonic text-to-speech provider.
package neuphonic

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Neuphonic text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	// Voice is the Neuphonic voice.
	Voice string
}

// TTSOptions configures a Neuphonic TTS engine.
type TTSOptions struct {
	// APIKey is the Neuphonic API key. If empty, the NEUPHONIC_API_KEY environment variable is used.
	APIKey string
	// VoiceID selects the voice. Defaults to none.
	VoiceID string
	// LangCode is the language code. Defaults to "en". Accepted but ignored.
	LangCode string
	// SamplingRate is the output sample rate in Hz. Defaults to 22050.
	SamplingRate int
}

// NewTTS creates a Neuphonic TTS engine from the given options.
func NewTTS(opts TTSOptions) *TTS {
	sr := opts.SamplingRate
	if sr == 0 {
		sr = 22050
	}
	t := &TTS{Voice: opts.VoiceID}
	t.InitTTS("neuphonic", zrt.APIKeyOr(opts.APIKey, "NEUPHONIC_API_KEY"), sr)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "neuphonic", Voice: t.Voice}
}
