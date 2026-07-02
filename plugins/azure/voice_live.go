package azure

import (
	"os"
	"strconv"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// VoiceLiveTurnDetectionConfig controls how Azure Voice Live decides the user
// has stopped speaking.
type VoiceLiveTurnDetectionConfig struct {
	Type string // default "server_vad"
	// Threshold is the VAD activation threshold.
	Threshold float64
	// PrefixPaddingMS is the audio padding kept before detected speech, in ms.
	PrefixPaddingMS   int
	SilenceDurationMS int // default 500
	// CreateResponse triggers a model response when a turn ends.
	CreateResponse bool
	// InterruptResponse lets user speech interrupt an in-progress response.
	InterruptResponse bool
}

// DefaultVoiceLiveTurnDetectionConfig returns the default server-VAD turn
// detection settings.
func DefaultVoiceLiveTurnDetectionConfig() *VoiceLiveTurnDetectionConfig {
	return &VoiceLiveTurnDetectionConfig{Type: "server_vad", Threshold: 0.5, PrefixPaddingMS: 300, SilenceDurationMS: 500, CreateResponse: true, InterruptResponse: true}
}

// VoiceLiveInputAudioTranscriptionConfig selects the model used to transcribe
// the user's speech.
type VoiceLiveInputAudioTranscriptionConfig struct {
	// Model is the transcription model for user speech. default "gpt-4o-mini-transcribe"
	Model string // default "gpt-4o-mini-transcribe"
}

// VoiceLiveOptions configures an Azure Voice Live model. The zero value is
// valid; empty and nil fields fall back to the defaults noted below.
type VoiceLiveOptions struct {
	// APIKey authenticates with Azure Voice Live. Overrides the
	// AZURE_VOICE_LIVE_API_KEY environment variable.
	APIKey string
	Model  string // default "gpt-4o-realtime-preview"
	Voice  string // default "en-US-AvaNeural"
	// Endpoint is the Azure Voice Live endpoint. Falls back to the
	// AZURE_VOICE_LIVE_ENDPOINT environment variable.
	Endpoint   string
	Modalities []string // default ["text","audio"]
	// Temperature is the sampling temperature. nil = provider default.
	Temperature *float64
	// MaxResponseOutputTokens caps tokens per response; empty leaves it unset.
	MaxResponseOutputTokens string
	// TurnDetection configures end-of-turn detection. nil uses the server-VAD defaults.
	TurnDetection *VoiceLiveTurnDetectionConfig
	// InputAudioTranscription selects the user-speech transcription model. nil uses the default.
	InputAudioTranscription *VoiceLiveInputAudioTranscriptionConfig
	ToolChoice              string // default "auto"
}

// VoiceLive is a configured Azure Voice Live speech-to-speech model.
type VoiceLive struct {
	zrt.BaseRealtime
	// Model is the Azure Voice Live model name.
	Model string
	// Voice is the Azure voice name used for audio output.
	Voice string
	// Endpoint is the resolved Azure Voice Live endpoint URL.
	Endpoint string
	// Modalities are the response modalities the model may produce.
	Modalities []string
	params     map[string]string
}

// NewVoiceLive returns an Azure Voice Live model configured from opts.
func NewVoiceLive(opts VoiceLiveOptions) *VoiceLive {
	modalities := opts.Modalities
	if len(modalities) == 0 {
		modalities = []string{"text", "audio"}
	}
	resolvedEndpoint := opts.Endpoint
	if resolvedEndpoint == "" {
		resolvedEndpoint = os.Getenv("AZURE_VOICE_LIVE_ENDPOINT")
	}
	td := opts.TurnDetection
	if td == nil {
		td = DefaultVoiceLiveTurnDetectionConfig()
	}
	iat := opts.InputAudioTranscription
	if iat == nil {
		iat = &VoiceLiveInputAudioTranscriptionConfig{Model: "gpt-4o-mini-transcribe"}
	}
	params := map[string]string{}
	if resolvedEndpoint != "" {
		params["base_url"] = resolvedEndpoint
	}
	if opts.Temperature != nil {
		params["temperature"] = zrt.FloatStr(*opts.Temperature)
	}
	params["tool_choice"] = zrt.StrOr(opts.ToolChoice, "auto")
	if opts.MaxResponseOutputTokens != "" {
		params["max_response_output_tokens"] = opts.MaxResponseOutputTokens
	}
	params["input_audio_transcription_model"] = iat.Model
	params["turn_detection_type"] = td.Type
	params["turn_detection_threshold"] = zrt.FloatStr(td.Threshold)
	params["turn_detection_prefix_padding_ms"] = strconv.Itoa(td.PrefixPaddingMS)
	params["turn_detection_silence_duration_ms"] = strconv.Itoa(td.SilenceDurationMS)
	params["turn_detection_create_response"] = zrt.BoolStr(td.CreateResponse)
	params["turn_detection_interrupt_response"] = zrt.BoolStr(td.InterruptResponse)

	r := &VoiceLive{Model: zrt.StrOr(opts.Model, "gpt-4o-realtime-preview"), Voice: zrt.StrOr(opts.Voice, "en-US-AvaNeural"), Endpoint: resolvedEndpoint, Modalities: modalities, params: params}
	r.Init("azure_voice_live", zrt.APIKeyOr(opts.APIKey, "AZURE_VOICE_LIVE_API_KEY"))
	return r
}

// RealtimeInfo implements zrt.RealtimeModel.
func (r *VoiceLive) RealtimeInfo() zrt.RealtimeInfo {
	return zrt.RealtimeInfo{Model: r.Model, Voice: r.Voice, Params: r.params, ResponseModalities: r.Modalities}
}
