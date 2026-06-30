package deepgram

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type TTS struct {
	zrt.BaseTTS
	Model    string
	Voice    string
	Stream   bool
	Encoding string
	BaseURL  string
}

type TTSOptions struct {
	APIKey   string
	Model    string
	Voice    string
	Language string
	Stream   *bool
	Encoding string
	BaseURL  string
}

func NewTTS(opts TTSOptions) *TTS {
	model := zrt.StrOr(opts.Model, "aura-2-andromeda-en")
	t := &TTS{
		Model:    model,
		Voice:    zrt.StrOr(opts.Voice, model),
		Stream:   zrt.BoolOr(opts.Stream, true),
		Encoding: zrt.StrOr(opts.Encoding, "linear16"),
		BaseURL:  zrt.StrOr(opts.BaseURL, "wss://api.deepgram.com/v1/speak"),
	}
	t.InitTTS("deepgram", zrt.APIKeyOr(opts.APIKey, "DEEPGRAM_API_KEY"), 24000)
	return t
}

func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "deepgram", Model: t.Model, Voice: t.Voice}
}

func (t *TTS) Knobs() map[string]any {
	return map[string]any{
		"tts_stream":   t.Stream,
		"tts_encoding": t.Encoding,
		"base_url":     t.BaseURL,
	}
}
