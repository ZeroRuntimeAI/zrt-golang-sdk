// Package resemble provides the Resemble AI text-to-speech provider.
package resemble

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Resemble AI text-to-speech engine.
type TTS struct {
	zrt.BaseTTS
	Voice string
}

// TTSOptions configures a Resemble AI TTS engine.
type TTSOptions struct {
	// APIKey is the Resemble AI API key. If empty, the RESEMBLE_API_KEY environment variable is used.
	APIKey string
	// VoiceUUID selects the voice. Defaults to "55592656".
	VoiceUUID string
	// SampleRate is the output sample rate in Hz. Defaults to 22050.
	SampleRate int
}

// NewTTS creates a Resemble AI TTS engine from the given options.
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
