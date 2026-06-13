// Package neuphonic provides the Neuphonic text-to-speech provider.
package neuphonic

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Neuphonic text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the NEUPHONIC_API_KEY environment variable.
	APIKey       string
	VoiceID      string // default "" (none)
	LangCode     string // default "en" (not serialized)
	SamplingRate int    // default 22050
}

// NewTTS builds a TTS.
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
