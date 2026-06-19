package cartesia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type STT struct {
	zrt.BaseSTT
	Model      string
	Language   string
	SampleRate int
}

type STTOptions struct {
	APIKey     string
	Model      string
	Language   string
	SampleRate int
}

func NewSTT(opts STTOptions) *STT {
	rate := opts.SampleRate
	if rate == 0 {
		rate = 48000
	}
	s := &STT{
		Model:      zrt.StrOr(opts.Model, "ink-2"),
		Language:   zrt.StrOr(opts.Language, "en"),
		SampleRate: rate,
	}
	s.Init("cartesia_stt", zrt.APIKeyOr(opts.APIKey, "CARTESIA_API_KEY"))
	return s
}

func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "cartesia_stt", Model: s.Model, Language: s.Language}
}
