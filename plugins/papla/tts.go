// Package papla provides the Papla text-to-speech provider.
package papla

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// TTS is a Papla text-to-speech engine. Papla has no separate voice;
// the model id is used as the voice.
type TTS struct {
	zrt.BaseTTS
	Voice string
}

// TTSOptions configures a Papla TTS engine.
type TTSOptions struct {
	// APIKey is the Papla API key. If empty, the PAPLA_API_KEY environment variable is used.
	APIKey string
	// ModelID selects the model, which also serves as the voice. Defaults to "papla_p1".
	ModelID    string
	SampleRate int
}

// NewTTS creates a Papla TTS engine from the given options.
func NewTTS(opts TTSOptions) *TTS {
	sr := opts.SampleRate
	if sr == 0 {
		sr = 24000
	}
	t := &TTS{Voice: zrt.StrOr(opts.ModelID, "papla_p1")}
	t.InitTTS("papla", zrt.APIKeyOr(opts.APIKey, "PAPLA_API_KEY"), sr)
	return t
}

// TTSConfig implements zrt.TTS.
func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "papla", Voice: t.Voice}
}
