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
	APIKey                           string
	Model                            string // default "fixie-ai/ultravox"
	Voice                            string
	LanguageHint                     string // default "en"
	Temperature                      *float64
	MaxDuration                      string
	TimeExceededMessage              string
	InputSampleRate                  int  // default 48000
	OutputSampleRate                 int  // default 24000
	ClientBufferSizeMS               int  // default 30000
	VADTurnEndpointDelayMS           *int // default 800
	VADMinimumTurnDurationMS         *int // default 600
	VADMinimumInterruptionDurationMS *int
	VADFrameActivationThreshold      *float64 // default 0.4
	FirstSpeaker                     string   // default "FIRST_SPEAKER_USER"
	EnableGreetingPrompt             bool
	BaseURL                          string
}

// Realtime is a configured Ultravox speech-to-speech model.
type Realtime struct {
	zrt.BaseRealtime
	Model  string
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
		"enable_greeting_prompt": boolStr(opts.EnableGreetingPrompt),
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

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
