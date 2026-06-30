package openai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type STT struct {
	zrt.BaseSTT
	Model                string
	Language             string
	Stream               bool
	InputSampleRate      int
	OutputSampleRate     int
	Prompt               string
	TurnDetection        string
	VADThreshold         *float64
	VADPrefixPaddingMs   *int
	VADSilenceDurationMs *int
	NoiseReduction       string
	ResponseFormat       string
	BaseURL              string
}

type STTOptions struct {
	APIKey               string
	Model                string
	Language             string
	Stream               *bool
	InputSampleRate      int
	OutputSampleRate     int
	Prompt               string
	TurnDetection        string
	VADThreshold         *float64
	VADPrefixPaddingMs   *int
	VADSilenceDurationMs *int
	NoiseReduction       string
	ResponseFormat       string
	BaseURL              string
}

func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                zrt.StrOr(opts.Model, "gpt-4o-transcribe"),
		Language:             zrt.StrOr(opts.Language, "en"),
		Stream:               zrt.BoolOr(opts.Stream, true),
		InputSampleRate:      orInt(opts.InputSampleRate, 48000),
		OutputSampleRate:     orInt(opts.OutputSampleRate, 24000),
		Prompt:               opts.Prompt,
		TurnDetection:        zrt.StrOr(opts.TurnDetection, "server_vad"),
		VADThreshold:         opts.VADThreshold,
		VADPrefixPaddingMs:   opts.VADPrefixPaddingMs,
		VADSilenceDurationMs: opts.VADSilenceDurationMs,
		NoiseReduction:       zrt.StrOr(opts.NoiseReduction, "near_field"),
		ResponseFormat:       zrt.StrOr(opts.ResponseFormat, "json"),
		BaseURL:              opts.BaseURL,
	}
	s.Init("openai_stt", zrt.APIKeyOr(opts.APIKey, "OPENAI_API_KEY"))
	return s
}

func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "openai_stt", Model: s.Model, Language: s.Language}
}

func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"stream":             s.Stream,
		"input_sample_rate":  s.InputSampleRate,
		"output_sample_rate": s.OutputSampleRate,
		"turn_detection":     s.TurnDetection,
		"noise_reduction":    s.NoiseReduction,
		"response_format":    s.ResponseFormat,
	}
	if s.Prompt != "" {
		k["prompt"] = s.Prompt
	}
	if s.VADThreshold != nil {
		k["vad_threshold"] = *s.VADThreshold
	}
	if s.VADPrefixPaddingMs != nil {
		k["vad_prefix_padding_ms"] = *s.VADPrefixPaddingMs
	}
	if s.VADSilenceDurationMs != nil {
		k["vad_silence_duration_ms"] = *s.VADSilenceDurationMs
	}
	if s.BaseURL != "" {
		k["base_url"] = s.BaseURL
	}
	return k
}
