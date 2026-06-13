// Package openai_realtime provides the OpenAI Realtime speech-to-speech model.
package openai_realtime

import (
	"strconv"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// TurnDetectionConfig configures turn detection.
type TurnDetectionConfig struct {
	Type              string // default "server_vad"
	Threshold         float64
	PrefixPaddingMS   int
	SilenceDurationMS int
	CreateResponse    bool
	InterruptResponse bool
}

// DefaultTurnDetectionConfig returns the default turn detection config.
func DefaultTurnDetectionConfig() *TurnDetectionConfig {
	return &TurnDetectionConfig{Type: "server_vad", Threshold: 0.5, PrefixPaddingMS: 300, SilenceDurationMS: 200, CreateResponse: true, InterruptResponse: true}
}

// InputAudioTranscriptionConfig configures input transcription.
type InputAudioTranscriptionConfig struct {
	Model string // default "gpt-4o-mini-transcribe"
}

// RealtimeOptions configures Realtime.
type RealtimeOptions struct {
	// APIKey overrides the OPENAI_API_KEY environment variable.
	APIKey                  string
	Model                   string   // default "gpt-4o-realtime-preview"
	Voice                   string   // default "alloy"
	Modalities              []string // default ["text","audio"]
	Temperature             *float64 // default 0.8
	MaxResponseOutputTokens string   // default "inf"
	TurnDetection           *TurnDetectionConfig
	InputAudioTranscription *InputAudioTranscriptionConfig
	ToolChoice              string // default "auto"
}

// Realtime is the OpenAI Realtime model descriptor.
type Realtime struct {
	zrt.BaseRealtime
	Model      string
	Voice      string
	Modalities []string
	params     map[string]string
}

// NewRealtime builds a Realtime from opts.
func NewRealtime(opts RealtimeOptions) *Realtime {
	model := zrt.StrOr(opts.Model, "gpt-4o-realtime-preview")
	voice := zrt.StrOr(opts.Voice, "alloy")
	modalities := opts.Modalities
	if len(modalities) == 0 {
		modalities = []string{"text", "audio"}
	}
	temp := zrt.FloatOr(opts.Temperature, 0.8)
	maxTokens := zrt.StrOr(opts.MaxResponseOutputTokens, "inf")
	toolChoice := zrt.StrOr(opts.ToolChoice, "auto")
	td := opts.TurnDetection
	if td == nil {
		td = DefaultTurnDetectionConfig()
	}
	iat := opts.InputAudioTranscription
	if iat == nil {
		iat = &InputAudioTranscriptionConfig{Model: "gpt-4o-mini-transcribe"}
	}
	params := map[string]string{
		"temperature":                     zrt.FloatStr(temp),
		"tool_choice":                     toolChoice,
		"max_response_output_tokens":      maxTokens,
		"input_audio_transcription_model": iat.Model,
	}
	if td != nil {
		params["turn_detection_type"] = td.Type
		params["turn_detection_threshold"] = zrt.FloatStr(td.Threshold)
		params["turn_detection_prefix_padding_ms"] = strconv.Itoa(td.PrefixPaddingMS)
		params["turn_detection_silence_duration_ms"] = strconv.Itoa(td.SilenceDurationMS)
		params["turn_detection_create_response"] = boolStr(td.CreateResponse)
		params["turn_detection_interrupt_response"] = boolStr(td.InterruptResponse)
	}
	r := &Realtime{Model: model, Voice: voice, Modalities: modalities, params: params}
	r.Init("openai_realtime", zrt.APIKeyOr(opts.APIKey, "OPENAI_API_KEY"))
	return r
}

// RealtimeInfo implements zrt.RealtimeModel.
func (r *Realtime) RealtimeInfo() zrt.RealtimeInfo {
	return zrt.RealtimeInfo{Model: r.Model, Voice: r.Voice, Params: r.params, ResponseModalities: r.Modalities}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
