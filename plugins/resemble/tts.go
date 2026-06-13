// Package resemble provides the Resemble AI text-to-speech provider.
package resemble

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is the Resemble text-to-speech descriptor.
type TTS struct {
	zrt.BaseTTS
	Voice string
}

// TTSOptions configures TTS.
type TTSOptions struct {
	// APIKey overrides the RESEMBLE_API_KEY environment variable.
	APIKey     string
	VoiceUUID  string // default "55592656"
	SampleRate int    // default 22050
}

// NewTTS builds a TTS.
func NewTTS(opts TTSOptions) *TTS {
	sr := opts.SampleRate
	if sr == 0 {
		sr = 22050
	}
	t := &TTS{Voice: zrt.StrOr(opts.VoiceUUID, "55592656")}
	t.InitTTS("resemble", zrt.APIKeyOr(opts.APIKey, "RESEMBLE_API_KEY"), sr)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "resemble", Voice: t.Voice}
}
