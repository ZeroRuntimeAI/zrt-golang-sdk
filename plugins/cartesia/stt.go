package cartesia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type STT struct {
	zrt.BaseSTT
	Model                  string
	Language               string
	InputSampleRate        int
	OutputSampleRate       int
	Encoding               string
	CartesiaVersion        string
	MinVolume              *float64
	MaxSilenceDurationSecs *float64
	BaseURL                string
}

type STTOptions struct {
	APIKey                 string
	Model                  string
	Language               string
	InputSampleRate        int
	OutputSampleRate       int
	Encoding               string
	CartesiaVersion        string
	MinVolume              *float64
	MaxSilenceDurationSecs *float64
	BaseURL                string
}

func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                  zrt.StrOr(opts.Model, "ink-whisper"),
		Language:               zrt.StrOr(opts.Language, "en"),
		InputSampleRate:        orInt(opts.InputSampleRate, 48000),
		OutputSampleRate:       orInt(opts.OutputSampleRate, 16000),
		Encoding:               zrt.StrOr(opts.Encoding, "pcm_s16le"),
		CartesiaVersion:        zrt.StrOr(opts.CartesiaVersion, "2026-03-01"),
		MinVolume:              opts.MinVolume,
		MaxSilenceDurationSecs: opts.MaxSilenceDurationSecs,
		BaseURL:                opts.BaseURL,
	}
	s.Init("cartesia_stt", zrt.APIKeyOr(opts.APIKey, "CARTESIA_API_KEY"))
	return s
}

func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "cartesia", Model: s.Model, Language: s.Language}
}

func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"input_sample_rate":  s.InputSampleRate,
		"output_sample_rate": s.OutputSampleRate,
		"encoding":           s.Encoding,
		"cartesia_version":   s.CartesiaVersion,
	}
	if s.MinVolume != nil {
		k["min_volume"] = *s.MinVolume
	}
	if s.MaxSilenceDurationSecs != nil {
		k["max_silence_duration_secs"] = *s.MaxSilenceDurationSecs
	}
	if s.BaseURL != "" {
		k["base_url"] = s.BaseURL
	}
	return k
}

func orInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
