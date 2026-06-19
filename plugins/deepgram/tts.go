package deepgram

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type TTS struct {
	zrt.BaseTTS
	Model    string
	Voice    string
	Language string
}

type TTSOptions struct {
	APIKey   string
	Model    string
	Voice    string
	Language string
}

func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Model:    zrt.StrOr(opts.Model, "aura-2"),
		Voice:    zrt.StrOr(opts.Voice, "asteria"),
		Language: zrt.StrOr(opts.Language, "en"),
	}
	t.InitTTS("deepgram", zrt.APIKeyOr(opts.APIKey, "DEEPGRAM_API_KEY"), 24000)
	return t
}

func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "deepgram", Voice: t.Voice}
}

func (t *TTS) Knobs() map[string]any {
	return map[string]any{"model": t.Model, "language": t.Language}
}
