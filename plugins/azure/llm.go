package azure

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

type LLM struct {
	zrt.BaseLLM
	Model             string
	Endpoint          string
	APIVersion        string
	Temperature       float64
	MaxOutputTokens   int
	TopP              *float64
	FrequencyPenalty  *float64
	PresencePenalty   *float64
	Seed              *int
	Stop              string
	User              string
	ToolChoice        string
	ParallelToolCalls *bool
	ResponseFormat    string
	ReasoningEffort   string
}

type LLMOptions struct {
	APIKey            string
	AzureEndpoint     string
	Deployment        string
	APIVersion        string
	Temperature       *float64
	MaxOutputTokens   int
	TopP              *float64
	FrequencyPenalty  *float64
	PresencePenalty   *float64
	Seed              *int
	Stop              string
	User              string
	ToolChoice        string
	ParallelToolCalls *bool
	ResponseFormat    string
	ReasoningEffort   string
}

func NewLLM(opts LLMOptions) *LLM {
	l := &LLM{
		Model:             opts.Deployment,
		Endpoint:          zrt.StrOr(opts.AzureEndpoint, os.Getenv("AZURE_OPENAI_ENDPOINT")),
		APIVersion:        zrt.StrOr(opts.APIVersion, "2024-10-21"),
		Temperature:       zrt.FloatOr(opts.Temperature, 0.7),
		MaxOutputTokens:   zrt.IntZeroOr(opts.MaxOutputTokens, 1024),
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		Stop:              opts.Stop,
		User:              opts.User,
		ToolChoice:        opts.ToolChoice,
		ParallelToolCalls: opts.ParallelToolCalls,
		ResponseFormat:    opts.ResponseFormat,
		ReasoningEffort:   opts.ReasoningEffort,
	}
	l.Init("azure_openai", zrt.APIKeyOr(opts.APIKey, "AZURE_OPENAI_API_KEY"))
	return l
}

func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "azure_openai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

func (l *LLM) Knobs() map[string]any {
	k := map[string]any{"api_version": l.APIVersion}
	if l.Endpoint != "" {
		k["endpoint"] = l.Endpoint
	}
	if l.TopP != nil {
		k["top_p"] = *l.TopP
	}
	if l.FrequencyPenalty != nil {
		k["frequency_penalty"] = *l.FrequencyPenalty
	}
	if l.PresencePenalty != nil {
		k["presence_penalty"] = *l.PresencePenalty
	}
	if l.Seed != nil {
		k["seed"] = *l.Seed
	}
	if l.Stop != "" {
		k["stop"] = l.Stop
	}
	if l.User != "" {
		k["user"] = l.User
	}
	if l.ToolChoice != "" {
		k["tool_choice"] = l.ToolChoice
	}
	if l.ParallelToolCalls != nil {
		k["parallel_tool_calls"] = *l.ParallelToolCalls
	}
	if l.ResponseFormat != "" {
		k["response_format"] = l.ResponseFormat
	}
	if l.ReasoningEffort != "" {
		k["reasoning_effort"] = l.ReasoningEffort
	}
	return k
}
