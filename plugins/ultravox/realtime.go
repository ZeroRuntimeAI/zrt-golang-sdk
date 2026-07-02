// Package ultravox provides the Ultravox speech-to-speech model.
package ultravox

import (
	"strconv"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// RealtimeOptions configures an Ultravox Realtime model. The zero value is
// valid; empty and nil fields fall back to the defaults noted below.
type RealtimeOptions struct {
	// APIKey authenticates with Ultravox. Overrides the ULTRAVOX_API_KEY
	// environment variable.
	APIKey string
	// Model is the Ultravox model id.
	Model string // default "fixie-ai/ultravox"
	// Voice is the voice used for synthesized speech.
	Voice string
	// LanguageHint is a BCP-47 hint for the expected input language.
	LanguageHint string // default "en"
	// Temperature controls sampling randomness; nil applies the provider default.
	Temperature *float64
	// MaxDuration caps the session length (e.g. "600s"); empty means no cap.
	MaxDuration string
	// TimeExceededMessage is spoken when MaxDuration is reached.
	TimeExceededMessage string
	// InputSampleRate is the input audio sample rate in Hz.
	InputSampleRate int // default 48000
	// OutputSampleRate is the output audio sample rate in Hz.
	OutputSampleRate int // default 24000
	// ClientBufferSizeMS is the client-side audio buffer size in milliseconds.
	ClientBufferSizeMS int // default 30000
	// VADTurnEndpointDelayMS is the silence delay before end-of-turn is detected, in milliseconds; nil applies the default.
	VADTurnEndpointDelayMS *int // default 800
	// VADMinimumTurnDurationMS is the minimum user turn duration in milliseconds; nil applies the default.
	VADMinimumTurnDurationMS *int // default 600
	// VADMinimumInterruptionDurationMS is the minimum speech duration required to interrupt, in milliseconds; nil omits the setting.
	VADMinimumInterruptionDurationMS *int
	// VADFrameActivationThreshold is the per-frame speech probability threshold; nil applies the default.
	VADFrameActivationThreshold *float64 // default 0.4
	// FirstSpeaker sets who speaks first in the conversation.
	FirstSpeaker string // default "FIRST_SPEAKER_USER"
	// EnableGreetingPrompt enables an initial greeting prompt.
	EnableGreetingPrompt bool
	// BaseURL overrides the Ultravox API base URL; empty uses the default.
	BaseURL string
}

// Realtime is a configured Ultravox speech-to-speech model.
type Realtime struct {
	zrt.BaseRealtime
	// Model is the resolved Ultravox model id.
	Model string
	// Voice is the resolved voice used for synthesized speech.
	Voice  string
	params map[string]string
}

// NewRealtime returns an Ultravox Realtime model configured from opts.
func NewRealtime(opts RealtimeOptions) *Realtime {
	inputSR := opts.InputSampleRate
	if inputSR == 0 {
		inputSR = 48000
	}
	outputSR := opts.OutputSampleRate
	if outputSR == 0 {
		outputSR = 24000
	}
	bufSize := opts.ClientBufferSizeMS
	if bufSize == 0 {
		bufSize = 30000
	}
	params := map[string]string{
		"language_hint":          zrt.StrOr(opts.LanguageHint, "en"),
		"input_sample_rate":      strconv.Itoa(inputSR),
		"output_sample_rate":     strconv.Itoa(outputSR),
		"client_buffer_size_ms":  strconv.Itoa(bufSize),
		"enable_greeting_prompt": zrt.BoolStr(opts.EnableGreetingPrompt),
	}
	if opts.BaseURL != "" {
		params["base_url"] = opts.BaseURL
	}
	if opts.Temperature != nil {
		params["temperature"] = zrt.FloatStr(*opts.Temperature)
	}
	if opts.MaxDuration != "" {
		params["max_duration"] = opts.MaxDuration
	}
	if opts.TimeExceededMessage != "" {
		params["time_exceeded_message"] = opts.TimeExceededMessage
	}
	params["vad_turn_endpoint_delay_ms"] = strconv.Itoa(zrt.IntOr(opts.VADTurnEndpointDelayMS, 800))
	params["vad_minimum_turn_duration_ms"] = strconv.Itoa(zrt.IntOr(opts.VADMinimumTurnDurationMS, 600))
	if opts.VADMinimumInterruptionDurationMS != nil {
		params["vad_minimum_interruption_duration_ms"] = strconv.Itoa(*opts.VADMinimumInterruptionDurationMS)
	}
	params["vad_frame_activation_threshold"] = zrt.FloatStr(zrt.FloatOr(opts.VADFrameActivationThreshold, 0.4))
	params["first_speaker"] = zrt.StrOr(opts.FirstSpeaker, "FIRST_SPEAKER_USER")

	r := &Realtime{Model: zrt.StrOr(opts.Model, "fixie-ai/ultravox"), Voice: opts.Voice, params: params}
	r.Init("ultravox", zrt.APIKeyOr(opts.APIKey, "ULTRAVOX_API_KEY"))
	return r
}

// RealtimeInfo implements zrt.RealtimeModel.
func (r *Realtime) RealtimeInfo() zrt.RealtimeInfo {
	return zrt.RealtimeInfo{Model: r.Model, Voice: r.Voice, Params: r.params, ResponseModalities: []string{"AUDIO"}}
}
