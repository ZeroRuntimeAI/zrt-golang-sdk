package azure

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

type OpenAISTT struct {
	zrt.BaseSTT
	Model            string
	Deployment       string
	Endpoint         string
	APIVersion       string
	Language         string
	Stream           bool
	InputSampleRate  int
	OutputSampleRate int
	Prompt           string
	Temperature      *float64
	ResponseFormat   string
	TurnDetection    string
}

type OpenAISTTOptions struct {
	APIKey           string
	AzureEndpoint    string
	Deployment       string
	APIVersion       string
	Model            string
	Language         string
	Stream           *bool
	InputSampleRate  int
	OutputSampleRate int
	Prompt           string
	Temperature      *float64
	ResponseFormat   string
	TurnDetection    string
}

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
		InputSampleRate:  orInt(opts.InputSampleRate, 48000),
		OutputSampleRate: orInt(opts.OutputSampleRate, 24000),
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

func (s *OpenAISTT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "azure_openai_stt", Model: s.modelOrDeployment(), Language: s.Language}
}

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
