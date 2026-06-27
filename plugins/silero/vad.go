// Package silero provides the Silero voice-activity-detection provider.
package silero

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// VAD is a Silero voice-activity-detection engine.
type VAD struct {
	zrt.BaseVAD
	Threshold          float64
	StopThreshold      float64
	MinSpeechDuration  float64
	MinSilenceDuration float64
	PaddingDuration    float64
	MaxBufferedSpeech  float64
	InputSampleRate    int
	SmoothingFactor    float64
	ForceCPU           bool
	MinVolume          float64
}

// VADOptions configures VAD. Nil pointers use default values.
type VADOptions struct {
	Threshold           *float64 // default 0.4
	StartThreshold      *float64 // overrides Threshold
	EndThreshold        *float64 // overrides StopThreshold
	StopThreshold       *float64 // default 0.25
	MinSpeechDuration   *float64 // default 0.3
	MinSilenceDuration  *float64 // default 0.4
	PaddingDuration     *float64 // default 0.5
	MaxBufferedSpeech   *float64 // default 60.0
	SampleRate          int      // default 16000
	InputSampleRate     *int     // default 48000
	ModelSampleRate     *int     // overrides SampleRate
	SmoothingFactor     *float64 // default 0.35
	ForceCPU            bool
	MinVolume           float64 // default 0.0
	EnergyFilterEnabled *bool   // default true
}

// NewVAD builds a VAD.
func NewVAD(opts VADOptions) *VAD {
	threshold := zrt.FloatOr(opts.Threshold, 0.4)
	if opts.StartThreshold != nil {
		threshold = *opts.StartThreshold
	}
	stop := zrt.FloatOr(opts.StopThreshold, 0.25)
	if opts.EndThreshold != nil {
		stop = *opts.EndThreshold
	}
	sampleRate := opts.SampleRate
	if sampleRate == 0 {
		sampleRate = 16000
	}
	modelSR := sampleRate
	if opts.ModelSampleRate != nil {
		modelSR = *opts.ModelSampleRate
	}
	minVolume := opts.MinVolume
	if !zrt.BoolOr(opts.EnergyFilterEnabled, true) {
		minVolume = 0.0
	}
	v := &VAD{
		Threshold:          threshold,
		StopThreshold:      stop,
		MinSpeechDuration:  zrt.FloatOr(opts.MinSpeechDuration, 0.3),
		MinSilenceDuration: zrt.FloatOr(opts.MinSilenceDuration, 0.4),
		PaddingDuration:    zrt.FloatOr(opts.PaddingDuration, 0.5),
		MaxBufferedSpeech:  zrt.FloatOr(opts.MaxBufferedSpeech, 60.0),
		InputSampleRate:    zrt.IntOr(opts.InputSampleRate, 48000),
		SmoothingFactor:    zrt.FloatOr(opts.SmoothingFactor, 0.35),
		ForceCPU:           opts.ForceCPU,
		MinVolume:          minVolume,
	}
	v.InitVAD("silero", modelSR)
	return v
}

// VADConfig implements zrt.VAD.
func (v *VAD) VADConfig() zrt.VADRuntimeConfig {
	return zrt.VADRuntimeConfig{
		Threshold:          float32(v.Threshold),
		StopThreshold:      float32(v.StopThreshold),
		MinSpeechDuration:  float32(v.MinSpeechDuration),
		MinSilenceDuration: float32(v.MinSilenceDuration),
		PaddingDuration:    float32(v.PaddingDuration),
		MaxBufferedSpeech:  float32(v.MaxBufferedSpeech),
		ForceCPU:           v.ForceCPU,
		SmoothingFactor:    float32(v.SmoothingFactor),
		InputSampleRate:    uint32(v.InputSampleRate),
		MinVolume:          float32(v.MinVolume),
	}
}

// Knobs implements the credential knob source.
func (v *VAD) Knobs() map[string]any {
	return map[string]any{"smoothing_factor": v.SmoothingFactor, "min_volume": v.MinVolume}
}
