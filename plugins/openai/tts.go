package openai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type TTS struct {
	zrt.BaseTTS
	Voice  string
	Model  string
	Speed  *float64
	Stream bool
}

type TTSOptions struct {
	APIKey     string
	Voice      string
	Model      string
	SampleRate int
	Speed      *float64
	Stream     *bool
}

func NewTTS(opts TTSOptions) *TTS {
	t := &TTS{
		Voice:  zrt.StrOr(opts.Voice, "ash"),
		Model:  zrt.StrOr(opts.Model, "gpt-4o-mini-tts"),
		Speed:  opts.Speed,
		Stream: zrt.BoolOr(opts.Stream, true),
	}
	t.InitTTS("openai", zrt.APIKeyOr(opts.APIKey, "OPENAI_API_KEY"), orInt(opts.SampleRate, 24000))
	return t
}

func (t *TTS) TTSConfig() zrt.TTSRuntimeConfig {
	return zrt.TTSRuntimeConfig{Provider: "openai", Model: t.Model, Voice: t.Voice}
}

func (t *TTS) Knobs() map[string]any {
	k := map[string]any{"tts_stream": t.Stream}
	if t.Speed != nil {
		k["speed"] = *t.Speed
	}
	return k
}
