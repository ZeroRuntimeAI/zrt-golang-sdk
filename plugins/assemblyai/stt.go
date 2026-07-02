// Package assemblyai provides the AssemblyAI speech-to-text provider.
package assemblyai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is an AssemblyAI speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	// Model is the streaming model id. Defaults to "universal-streaming-english".
	Model string
	// Language is the BCP-47 transcription language code; only used when LanguageDetection is false. Defaults to "en-US".
	Language string
	// InputSampleRate is the sample rate (Hz) of the incoming audio. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the sample rate (Hz) audio is resampled to before being sent to AssemblyAI. Defaults to 16000.
	OutputSampleRate int
	// Encoding is the PCM encoding of the audio stream ("pcm_s16le" or "pcm_mulaw"). Defaults to "pcm_s16le".
	Encoding string
	// FormatTurns, when true, returns punctuated and cased transcript segments aligned to speaker turns. Defaults to true.
	FormatTurns bool
	// EndOfTurnConfidenceThreshold is the confidence score in [0.0, 1.0] above which a pause is treated as an end of turn. Defaults to 0.4.
	EndOfTurnConfidenceThreshold float64
	// MinEndOfTurnSilenceWhenConfident is the minimum silence (ms) required to confirm an end of turn when confidence exceeds the threshold. Defaults to 560.
	MinEndOfTurnSilenceWhenConfident int
	// MaxTurnSilence is the maximum silence (ms) before a turn is force-ended regardless of confidence. Defaults to 2400.
	MaxTurnSilence int
	// KeytermsPrompt is a list of domain-specific words or phrases to boost recognition accuracy.
	KeytermsPrompt []string
	// LanguageDetection, when true, auto-detects the spoken language per utterance and overrides Language. Defaults to false.
	LanguageDetection bool
	// BaseURL overrides the AssemblyAI WebSocket endpoint URL.
	BaseURL string
}

// STTOptions configures an AssemblyAI STT.
type STTOptions struct {
	// APIKey overrides the ASSEMBLYAI_API_KEY environment variable.
	APIKey string
	// Model is the streaming model id. Defaults to "universal-streaming-english".
	Model string
	// Language is the BCP-47 transcription language code; only used when LanguageDetection is false. Defaults to "en-US".
	Language string
	// InputSampleRate is the sample rate (Hz) of the incoming audio. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the sample rate (Hz) audio is resampled to before being sent to AssemblyAI. Defaults to 16000.
	OutputSampleRate int
	// Encoding is the PCM encoding of the audio stream ("pcm_s16le" or "pcm_mulaw"). Defaults to "pcm_s16le".
	Encoding string
	// FormatTurns, when true, returns punctuated and cased transcript segments aligned to speaker turns. nil = provider default (true).
	FormatTurns *bool
	// EndOfTurnConfidenceThreshold is the confidence score in [0.0, 1.0] above which a pause is treated as an end of turn. nil = provider default (0.4).
	EndOfTurnConfidenceThreshold *float64
	// MinEndOfTurnSilenceWhenConfident is the minimum silence (ms) required to confirm an end of turn when confidence exceeds the threshold. nil = provider default (560).
	MinEndOfTurnSilenceWhenConfident *int
	// MaxTurnSilence is the maximum silence (ms) before a turn is force-ended regardless of confidence. nil = provider default (2400).
	MaxTurnSilence *int
	// KeytermsPrompt is a list of domain-specific words or phrases to boost recognition accuracy.
	KeytermsPrompt []string
	// LanguageDetection, when true, auto-detects the spoken language per utterance and overrides Language. nil = provider default (false).
	LanguageDetection *bool
	// BaseURL overrides the AssemblyAI WebSocket endpoint URL.
	BaseURL string
}

// NewSTT returns an AssemblyAI STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                            zrt.StrOr(opts.Model, "universal-streaming-english"),
		Language:                         zrt.StrOr(opts.Language, "en-US"),
		InputSampleRate:                  zrt.IntZeroOr(opts.InputSampleRate, 48000),
		OutputSampleRate:                 zrt.IntZeroOr(opts.OutputSampleRate, 16000),
		Encoding:                         zrt.StrOr(opts.Encoding, "pcm_s16le"),
		FormatTurns:                      zrt.BoolOr(opts.FormatTurns, true),
		EndOfTurnConfidenceThreshold:     zrt.FloatOr(opts.EndOfTurnConfidenceThreshold, 0.4),
		MinEndOfTurnSilenceWhenConfident: zrt.IntOr(opts.MinEndOfTurnSilenceWhenConfident, 560),
		MaxTurnSilence:                   zrt.IntOr(opts.MaxTurnSilence, 2400),
		KeytermsPrompt:                   opts.KeytermsPrompt,
		LanguageDetection:                zrt.BoolOr(opts.LanguageDetection, false),
		BaseURL:                          opts.BaseURL,
	}
	s.Init("assemblyai", zrt.APIKeyOr(opts.APIKey, "ASSEMBLYAI_API_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "assemblyai", Model: s.Model, Language: s.Language}
}

// Knobs returns the provider-specific configuration knobs sent to the runtime.
func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"output_sample_rate":                     s.OutputSampleRate,
		"encoding":                               s.Encoding,
		"format_turns":                           s.FormatTurns,
		"end_of_turn_confidence_threshold":       s.EndOfTurnConfidenceThreshold,
		"min_end_of_turn_silence_when_confident": s.MinEndOfTurnSilenceWhenConfident,
		"max_turn_silence":                       s.MaxTurnSilence,
		"language_detection":                     s.LanguageDetection,
	}
	if len(s.KeytermsPrompt) > 0 {
		k["keyterms_prompt"] = s.KeytermsPrompt
	}
	if s.BaseURL != "" {
		k["base_url"] = s.BaseURL
	}
	return k
}
