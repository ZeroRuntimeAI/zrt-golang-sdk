package azure

import (
	"os"
	"strconv"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// VoiceLiveTurnDetectionConfig configures Azure Voice Live turn detection.
type VoiceLiveTurnDetectionConfig struct {
	Type              string // default "server_vad"
	Threshold         float64
	PrefixPaddingMS   int
	SilenceDurationMS int // default 500
	CreateResponse    bool
	InterruptResponse bool
}

// DefaultVoiceLiveTurnDetectionConfig returns the default config.
func DefaultVoiceLiveTurnDetectionConfig() *VoiceLiveTurnDetectionConfig {
	return &VoiceLiveTurnDetectionConfig{Type: "server_vad", Threshold: 0.5, PrefixPaddingMS: 300, SilenceDurationMS: 500, CreateResponse: true, InterruptResponse: true}
}

// VoiceLiveInputAudioTranscriptionConfig configures input transcription.
type VoiceLiveInputAudioTranscriptionConfig struct {
	Model string // default "gpt-4o-mini-transcribe"
}

// VoiceLiveOptions configures VoiceLive.
type VoiceLiveOptions struct {
	// APIKey overrides the AZURE_VOICE_LIVE_API_KEY environment variable.
	APIKey                  string
	Model                   string // default "gpt-4o-realtime-preview"
	Voice                   string // default "en-US-AvaNeural"
	Endpoint                string
	Modalities              []string // default ["text","audio"]
	Temperature             *float64
	MaxResponseOutputTokens string
	TurnDetection           *VoiceLiveTurnDetectionConfig
	InputAudioTranscription *VoiceLiveInputAudioTranscriptionConfig
	ToolChoice              string // default "auto"
}

// VoiceLive is the Azure Voice Live model descriptor.
type VoiceLive struct {
	zrt.BaseRealtime
	Model      string
	Voice      string
	Endpoint   string
	Modalities []string
	params     map[string]string
}

// NewVoiceLive builds a VoiceLive from opts.
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
	params["turn_detection_create_response"] = boolStr(td.CreateResponse)
	params["turn_detection_interrupt_response"] = boolStr(td.InterruptResponse)

	r := &VoiceLive{Model: zrt.StrOr(opts.Model, "gpt-4o-realtime-preview"), Voice: zrt.StrOr(opts.Voice, "en-US-AvaNeural"), Endpoint: resolvedEndpoint, Modalities: modalities, params: params}
	r.Init("azure_voice_live", zrt.APIKeyOr(opts.APIKey, "AZURE_VOICE_LIVE_API_KEY"))
	return r
}

// RealtimeInfo implements zrt.RealtimeModel.
func (r *VoiceLive) RealtimeInfo() zrt.RealtimeInfo {
	return zrt.RealtimeInfo{Model: r.Model, Voice: r.Voice, Params: r.params, ResponseModalities: r.Modalities}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
