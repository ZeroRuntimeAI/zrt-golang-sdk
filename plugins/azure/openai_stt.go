package azure

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// OpenAISTT is the Azure OpenAI speech-to-text provider.
type OpenAISTT struct {
	zrt.BaseSTT
	// Model is the Azure OpenAI transcription model name.
	Model string
	// Deployment is the Azure OpenAI deployment name; used in preference to Model.
	Deployment string
	// Endpoint is the Azure OpenAI endpoint URL.
	Endpoint string
	// APIVersion is the Azure OpenAI REST API version.
	APIVersion string
	// Language is the recognition language.
	Language string
	// Stream enables streaming transcription.
	Stream bool
	// InputSampleRate is the input audio sample rate in Hz.
	InputSampleRate int
	// OutputSampleRate is the output audio sample rate in Hz.
	OutputSampleRate int
	// Prompt is an optional transcription prompt for biasing recognition.
	Prompt string
	// Temperature is the sampling temperature. nil omits the knob.
	Temperature *float64
	// ResponseFormat is the transcription response format (e.g. "json").
	ResponseFormat string
	// TurnDetection is the turn-detection mode (e.g. "server_vad").
	TurnDetection string
}

// OpenAISTTOptions configures NewOpenAISTT.
type OpenAISTTOptions struct {
	// APIKey authenticates with Azure OpenAI. Falls back to AZURE_API_KEY, then AZURE_OPENAI_API_KEY.
	APIKey string
	// AzureEndpoint is the Azure OpenAI endpoint. Falls back to AZURE_OPENAI_ENDPOINT.
	AzureEndpoint string
	// Deployment is the Azure OpenAI deployment name. Falls back to AZURE_OPENAI_STT_DEPLOYMENT.
	Deployment string
	// APIVersion is the Azure OpenAI REST API version. Defaults to "2025-03-01-preview".
	APIVersion string
	// Model is the Azure OpenAI transcription model name.
	Model string
	// Language is the recognition language. Defaults to "en".
	Language string
	// Stream enables streaming transcription. Defaults to false.
	Stream *bool
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the output audio sample rate in Hz. Defaults to 24000.
	OutputSampleRate int
	// Prompt is an optional transcription prompt for biasing recognition.
	Prompt string
	// Temperature is the sampling temperature. nil = provider default.
	Temperature *float64
	// ResponseFormat is the transcription response format. Defaults to "json".
	ResponseFormat string
	// TurnDetection is the turn-detection mode. Defaults to "server_vad".
	TurnDetection string
}

// NewOpenAISTT returns an Azure OpenAI STT configured from opts, applying defaults.
func NewOpenAISTT(opts OpenAISTTOptions) *OpenAISTT {
	key := opts.APIKey
	if key == "" {
		key = zrt.StrOr(os.Getenv("AZURE_API_KEY"), os.Getenv("AZURE_OPENAI_API_KEY"))
	}
	s := &OpenAISTT{
		Model:            opts.Model,
		Deployment:       zrt.StrOr(opts.Deployment, os.Getenv("AZURE_OPENAI_STT_DEPLOYMENT")),
		Endpoint:         zrt.StrOr(opts.AzureEndpoint, os.Getenv("AZURE_OPENAI_ENDPOINT")),
		APIVersion:       zrt.StrOr(opts.APIVersion, "2025-03-01-preview"),
		Language:         zrt.StrOr(opts.Language, "en"),
		Stream:           zrt.BoolOr(opts.Stream, false),
		InputSampleRate:  zrt.IntZeroOr(opts.InputSampleRate, 48000),
		OutputSampleRate: zrt.IntZeroOr(opts.OutputSampleRate, 24000),
		Prompt:           opts.Prompt,
		Temperature:      opts.Temperature,
		ResponseFormat:   zrt.StrOr(opts.ResponseFormat, "json"),
		TurnDetection:    zrt.StrOr(opts.TurnDetection, "server_vad"),
	}
	s.Init("azure_openai_stt", key)
	return s
}

func (s *OpenAISTT) modelOrDeployment() string {
	if s.Deployment != "" {
		return s.Deployment
	}
	return s.Model
}

// STTConfig returns the runtime provider, model (or deployment), and language.
func (s *OpenAISTT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "azure_openai_stt", Model: s.modelOrDeployment(), Language: s.Language}
}

// Knobs returns the extra provider parameters, omitting unset optional fields.
func (s *OpenAISTT) Knobs() map[string]any {
	k := map[string]any{
		"api_version":        s.APIVersion,
		"stream":             s.Stream,
		"input_sample_rate":  s.InputSampleRate,
		"output_sample_rate": s.OutputSampleRate,
		"response_format":    s.ResponseFormat,
		"turn_detection":     s.TurnDetection,
	}
	if s.Endpoint != "" {
		k["endpoint"] = s.Endpoint
	}
	if s.Deployment != "" {
		k["deployment"] = s.Deployment
	}
	if s.Prompt != "" {
		k["prompt"] = s.Prompt
	}
	if s.Temperature != nil {
		k["temperature"] = *s.Temperature
	}
	return k
}
