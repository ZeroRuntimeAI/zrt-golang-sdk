// Package sarvamai provides the Sarvam AI STT, LLM and TTS providers.
package sarvamai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// STT is a Sarvam AI speech-to-text engine.
type STT struct {
	zrt.BaseSTT
	// Model selects the recognition model. Defaults to "saaras:v3".
	Model string
	// Language is the recognition language. Defaults to "en-IN".
	Language string
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the output audio sample rate in Hz. Defaults to 16000.
	OutputSampleRate int
	// Mode selects the transcription mode.
	Mode string
	// HighVADSensitivity raises voice-activity detection sensitivity. nil = provider default.
	HighVADSensitivity *bool
	// FlushSignals emits flush markers in the transcript stream. nil = provider default.
	FlushSignals *bool
	// Translation translates recognized speech into the target language.
	Translation bool
	// Prompt biases recognition toward the given context.
	Prompt string
	// InputAudioCodec is the input audio codec. Defaults to "pcm_s16le".
	InputAudioCodec string
	// VadSignals emits voice-activity markers in the transcript stream. Defaults to true.
	VadSignals bool
	// PositiveSpeechThreshold is the VAD probability above which a frame is speech. nil = provider default.
	PositiveSpeechThreshold *float64
	// NegativeSpeechThreshold is the VAD probability below which a frame is non-speech. nil = provider default.
	NegativeSpeechThreshold *float64
	// MinSpeechFrames is the minimum frames required to start a speech segment. nil = provider default.
	MinSpeechFrames *int
	// FirstTurnMinSpeechFrames is the minimum speech frames for the first turn. nil = provider default.
	FirstTurnMinSpeechFrames *int
	// PreSpeechPadFrames is the number of frames of audio kept before speech onset. nil = provider default.
	PreSpeechPadFrames *int
	// InterruptMinSpeechFrames is the minimum speech frames required to trigger an interruption. nil = provider default.
	InterruptMinSpeechFrames *int
}

// STTOptions configures a Sarvam AI STT.
type STTOptions struct {
	// APIKey overrides the SARVAM_API_KEY environment variable.
	APIKey string
	// Model selects the recognition model. Defaults to "saaras:v3".
	Model string
	// Language is the recognition language. Defaults to "en-IN".
	Language string
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the output audio sample rate in Hz. Defaults to 16000.
	OutputSampleRate int
	// Mode selects the transcription mode.
	Mode string
	// HighVADSensitivity raises voice-activity detection sensitivity.
	HighVADSensitivity *bool
	// FlushSignals emits flush markers in the transcript stream.
	FlushSignals *bool
	// Translation translates recognized speech into the target language.
	Translation bool
	// Prompt biases recognition toward the given context.
	Prompt string
	// InputAudioCodec is the input audio codec. Defaults to "pcm_s16le".
	InputAudioCodec string
	// VadSignals emits voice-activity markers in the transcript stream. Defaults to true.
	VadSignals *bool
	// PositiveSpeechThreshold is the VAD probability above which a frame is speech. nil = provider default.
	PositiveSpeechThreshold *float64
	// NegativeSpeechThreshold is the VAD probability below which a frame is non-speech. nil = provider default.
	NegativeSpeechThreshold *float64
	// MinSpeechFrames is the minimum frames required to start a speech segment. nil = provider default.
	MinSpeechFrames *int
	// FirstTurnMinSpeechFrames is the minimum speech frames for the first turn. nil = provider default.
	FirstTurnMinSpeechFrames *int
	// PreSpeechPadFrames is the number of frames of audio kept before speech onset. nil = provider default.
	PreSpeechPadFrames *int
	// InterruptMinSpeechFrames is the minimum speech frames required to trigger an interruption. nil = provider default.
	InterruptMinSpeechFrames *int
}

// NewSTT returns a Sarvam AI STT configured from opts.
func NewSTT(opts STTOptions) *STT {
	s := &STT{
		Model:                    zrt.StrOr(opts.Model, "saaras:v3"),
		Language:                 zrt.StrOr(opts.Language, "en-IN"),
		InputSampleRate:          zrt.IntZeroOr(opts.InputSampleRate, 48000),
		OutputSampleRate:         zrt.IntZeroOr(opts.OutputSampleRate, 16000),
		Mode:                     opts.Mode,
		HighVADSensitivity:       opts.HighVADSensitivity,
		FlushSignals:             opts.FlushSignals,
		Translation:              opts.Translation,
		Prompt:                   opts.Prompt,
		InputAudioCodec:          zrt.StrOr(opts.InputAudioCodec, "pcm_s16le"),
		VadSignals:               zrt.BoolOr(opts.VadSignals, true),
		PositiveSpeechThreshold:  opts.PositiveSpeechThreshold,
		NegativeSpeechThreshold:  opts.NegativeSpeechThreshold,
		MinSpeechFrames:          opts.MinSpeechFrames,
		FirstTurnMinSpeechFrames: opts.FirstTurnMinSpeechFrames,
		PreSpeechPadFrames:       opts.PreSpeechPadFrames,
		InterruptMinSpeechFrames: opts.InterruptMinSpeechFrames,
	}
	s.Init("sarvamai", zrt.APIKeyOr(opts.APIKey, "SARVAM_API_KEY"))
	return s
}

// STTConfig returns the provider, model, and language for this engine.
func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "sarvamai", Model: s.Model, Language: s.Language}
}

// Knobs returns the Sarvam AI-specific options as a key/value map.
func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"model":              s.Model,
		"language":           s.Language,
		"translation":        s.Translation,
		"input_sample_rate":  s.InputSampleRate,
		"output_sample_rate": s.OutputSampleRate,
		"input_audio_codec":  s.InputAudioCodec,
		"vad_signals":        s.VadSignals,
	}
	if s.Mode != "" {
		k["mode"] = s.Mode
	}
	if s.Prompt != "" {
		k["prompt"] = s.Prompt
	}
	if s.HighVADSensitivity != nil {
		k["high_vad_sensitivity"] = *s.HighVADSensitivity
	}
	if s.FlushSignals != nil {
		k["flush_signals"] = *s.FlushSignals
	}
	if s.PositiveSpeechThreshold != nil {
		k["positive_speech_threshold"] = *s.PositiveSpeechThreshold
	}
	if s.NegativeSpeechThreshold != nil {
		k["negative_speech_threshold"] = *s.NegativeSpeechThreshold
	}
	if s.MinSpeechFrames != nil {
		k["min_speech_frames"] = *s.MinSpeechFrames
	}
	if s.FirstTurnMinSpeechFrames != nil {
		k["first_turn_min_speech_frames"] = *s.FirstTurnMinSpeechFrames
	}
	if s.PreSpeechPadFrames != nil {
		k["pre_speech_pad_frames"] = *s.PreSpeechPadFrames
	}
	if s.InterruptMinSpeechFrames != nil {
		k["interrupt_min_speech_frames"] = *s.InterruptMinSpeechFrames
	}
	return k
}
