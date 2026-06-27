package xai

import (
	"strconv"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// TurnDetectionConfig controls how the model decides the user has stopped
// speaking.
type TurnDetectionConfig struct {
	Type              string // default "server_vad"
	Threshold         float64
	PrefixPaddingMS   int
	SilenceDurationMS int
	CreateResponse    bool
	InterruptResponse bool
}

// DefaultTurnDetectionConfig returns the default server-VAD turn detection
// settings.
func DefaultTurnDetectionConfig() *TurnDetectionConfig {
	return &TurnDetectionConfig{Type: "server_vad", Threshold: 0.5, PrefixPaddingMS: 300, SilenceDurationMS: 200, CreateResponse: true, InterruptResponse: true}
}

// InputAudioTranscriptionConfig selects the model used to transcribe the user's
// speech.
type InputAudioTranscriptionConfig struct {
	Model string // default "gpt-4o-mini-transcribe"
}

// RealtimeOptions configures an xAI Realtime model. The zero value is valid;
// empty and nil fields fall back to the defaults noted below.
type RealtimeOptions struct {
	// APIKey authenticates with xAI. Overrides the XAI_API_KEY environment
	// variable.
	APIKey                  string
	Model                   string   // default "grok-realtime"
	Voice                   string   // default "Ara"
	Modalities              []string // default ["text","audio"]
	Temperature             *float64 // default 0.8
	MaxResponseOutputTokens string   // default "inf"
	TurnDetection           *TurnDetectionConfig
	InputAudioTranscription *InputAudioTranscriptionConfig
	ToolChoice              string // default "auto"
	BaseURL                 string
}

// Realtime is a configured xAI speech-to-speech model.
type Realtime struct {
	zrt.BaseRealtime
	Model      string
	Voice      string
	Modalities []string
	params     map[string]string
}

// NewRealtime returns an xAI Realtime model configured from opts.
func NewRealtime(opts RealtimeOptions) *Realtime {
	modalities := opts.Modalities
	if len(modalities) == 0 {
		modalities = []string{"text", "audio"}
	}
	temp := zrt.FloatOr(opts.Temperature, 0.8)
	td := opts.TurnDetection
	if td == nil {
		td = DefaultTurnDetectionConfig()
	}
	iat := opts.InputAudioTranscription
	if iat == nil {
		iat = &InputAudioTranscriptionConfig{Model: "gpt-4o-mini-transcribe"}
	}
	params := map[string]string{"temperature": zrt.FloatStr(temp)}
	if opts.BaseURL != "" {
		params["base_url"] = opts.BaseURL
	}
	params["tool_choice"] = zrt.StrOr(opts.ToolChoice, "auto")
	params["max_response_output_tokens"] = zrt.StrOr(opts.MaxResponseOutputTokens, "inf")
	params["input_audio_transcription_model"] = iat.Model
	params["turn_detection_type"] = td.Type
	params["turn_detection_threshold"] = zrt.FloatStr(td.Threshold)
	params["turn_detection_prefix_padding_ms"] = strconv.Itoa(td.PrefixPaddingMS)
	params["turn_detection_silence_duration_ms"] = strconv.Itoa(td.SilenceDurationMS)
	params["turn_detection_create_response"] = boolStr(td.CreateResponse)
	params["turn_detection_interrupt_response"] = boolStr(td.InterruptResponse)

	r := &Realtime{Model: zrt.StrOr(opts.Model, "grok-realtime"), Voice: zrt.StrOr(opts.Voice, "Ara"), Modalities: modalities, params: params}
	r.Init("xai", zrt.APIKeyOr(opts.APIKey, "XAI_API_KEY"))
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
