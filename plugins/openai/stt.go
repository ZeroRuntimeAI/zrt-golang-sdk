package openai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is the OpenAI speech-to-text provider.
type STT struct {
	zrt.BaseSTT
	// Model is the transcription model id. Defaults to "gpt-4o-transcribe".
	Model string
	// Language is the input language code. Defaults to "en".
	Language string
	// Stream enables streamed transcription. Defaults to true.
	Stream bool
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the output audio sample rate in Hz. Defaults to 24000.
	OutputSampleRate int
	// Prompt biases transcription toward expected vocabulary; empty = none.
	Prompt string
	// TurnDetection selects the turn-detection mode. Defaults to "server_vad".
	TurnDetection string
	// VADThreshold is the voice-activity detection threshold; nil = provider default.
	VADThreshold *float64
	// VADPrefixPaddingMs is the audio padding before detected speech, in ms; nil = provider default.
	VADPrefixPaddingMs *int
	// VADSilenceDurationMs is the trailing silence that ends a turn, in ms; nil = provider default.
	VADSilenceDurationMs *int
	// NoiseReduction selects the noise-reduction mode. Defaults to "near_field".
	NoiseReduction string
	// ResponseFormat is the transcription response format. Defaults to "json".
	ResponseFormat string
	// BaseURL overrides the API base URL; empty = default.
	BaseURL string
}

// STTOptions configures NewSTT.
type STTOptions struct {
	// APIKey is the OpenAI API key; falls back to OPENAI_API_KEY.
	APIKey string
	// Model is the transcription model id. Defaults to "gpt-4o-transcribe".
	Model string
	// Language is the input language code. Defaults to "en".
	Language string
	// Stream enables streamed transcription; nil applies true.
	Stream *bool
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the output audio sample rate in Hz. Defaults to 24000.
	OutputSampleRate int
	// Prompt biases transcription toward expected vocabulary; empty = none.
	Prompt string
	// TurnDetection selects the turn-detection mode. Defaults to "server_vad".
	TurnDetection string
	// VADThreshold is the voice-activity detection threshold; nil = provider default.
	VADThreshold *float64
	// VADPrefixPaddingMs is the audio padding before detected speech, in ms; nil = provider default.
	VADPrefixPaddingMs *int
	// VADSilenceDurationMs is the trailing silence that ends a turn, in ms; nil = provider default.
	VADSilenceDurationMs *int
	// NoiseReduction selects the noise-reduction mode. Defaults to "near_field".
	NoiseReduction string
	// ResponseFormat is the transcription response format. Defaults to "json".
	ResponseFormat string
	// BaseURL overrides the API base URL; empty = default.
	BaseURL string
}

// NewSTT builds an STT from opts, applying defaults and resolving the API key.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                zrt.StrOr(opts.Model, "gpt-4o-transcribe"),
		Language:             zrt.StrOr(opts.Language, "en"),
		Stream:               zrt.BoolOr(opts.Stream, true),
		InputSampleRate:      zrt.IntZeroOr(opts.InputSampleRate, 48000),
		OutputSampleRate:     zrt.IntZeroOr(opts.OutputSampleRate, 24000),
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

// STTConfig returns the runtime configuration for this provider.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "openai_stt", Model: s.Model, Language: s.Language}
}

// Knobs returns the set of provider parameters to pass to the runtime.
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
