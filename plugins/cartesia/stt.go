package cartesia

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is the Cartesia speech-to-text provider.
type STT struct {
	zrt.BaseSTT
	// Model is the Cartesia STT model id. Defaults to "ink-whisper".
	Model string
	// Language is the ISO 639-1 code of the spoken audio. Defaults to "en".
	Language string
	// InputSampleRate is the sample rate in Hz of the raw PCM audio sent to the service. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the sample rate in Hz at which Cartesia processes audio internally. Defaults to 16000.
	OutputSampleRate int
	// Encoding is the PCM encoding format of the input audio. Defaults to "pcm_s16le".
	Encoding string
	// CartesiaVersion is the Cartesia API version header in "YYYY-MM-DD" format. Defaults to "2026-03-01".
	CartesiaVersion string
	// MinVolume is the minimum RMS volume threshold (0.0-1.0) below which audio is treated as silence; nil = platform default.
	MinVolume *float64
	// MaxSilenceDurationSecs is the maximum continuous silence, in seconds, before a transcript is finalized; nil = platform default.
	MaxSilenceDurationSecs *float64
	// BaseURL overrides the Cartesia WebSocket base URL.
	BaseURL string
}

// STTOptions configures NewSTT.
type STTOptions struct {
	// APIKey overrides the CARTESIA_API_KEY environment variable.
	APIKey string
	// Model is the Cartesia STT model id. Defaults to "ink-whisper".
	Model string
	// Language is the ISO 639-1 code of the spoken audio. Defaults to "en".
	Language string
	// InputSampleRate is the sample rate in Hz of the raw PCM audio sent to the service. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the sample rate in Hz at which Cartesia processes audio internally. Defaults to 16000.
	OutputSampleRate int
	// Encoding is the PCM encoding format of the input audio. Defaults to "pcm_s16le".
	Encoding string
	// CartesiaVersion is the Cartesia API version header in "YYYY-MM-DD" format. Defaults to "2026-03-01".
	CartesiaVersion string
	// MinVolume is the minimum RMS volume threshold (0.0-1.0) below which audio is treated as silence; nil = platform default.
	MinVolume *float64
	// MaxSilenceDurationSecs is the maximum continuous silence, in seconds, before a transcript is finalized; nil = platform default.
	MaxSilenceDurationSecs *float64
	// BaseURL overrides the Cartesia WebSocket base URL.
	BaseURL string
}

// NewSTT returns a Cartesia STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                  zrt.StrOr(opts.Model, "ink-whisper"),
		Language:               zrt.StrOr(opts.Language, "en"),
		InputSampleRate:        zrt.IntZeroOr(opts.InputSampleRate, 48000),
		OutputSampleRate:       zrt.IntZeroOr(opts.OutputSampleRate, 16000),
		Encoding:               zrt.StrOr(opts.Encoding, "pcm_s16le"),
		CartesiaVersion:        zrt.StrOr(opts.CartesiaVersion, "2026-03-01"),
		MinVolume:              opts.MinVolume,
		MaxSilenceDurationSecs: opts.MaxSilenceDurationSecs,
		BaseURL:                opts.BaseURL,
	}
	s.Init("cartesia_stt", zrt.APIKeyOr(opts.APIKey, "CARTESIA_API_KEY"))
	return s
}

// STTConfig implements zrt.STT.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "cartesia", Model: s.Model, Language: s.Language}
}

// Knobs returns the provider-specific STT settings.
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
