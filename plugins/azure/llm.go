package azure

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// LLM is the Azure OpenAI large-language-model provider.
type LLM struct {
	zrt.BaseLLM
	// Model is the Azure OpenAI deployment name used for completions.
	Model string
	// Endpoint is the Azure OpenAI endpoint URL.
	Endpoint string
	// APIVersion is the Azure OpenAI REST API version.
	APIVersion string
	// Temperature is the sampling temperature.
	Temperature float64
	// MaxOutputTokens caps the number of tokens generated per response.
	MaxOutputTokens int
	// TopP is the nucleus-sampling probability mass. nil omits the knob.
	TopP *float64
	// FrequencyPenalty penalizes tokens by frequency. nil omits the knob.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens by prior presence. nil omits the knob.
	PresencePenalty *float64
	// Seed makes sampling deterministic. nil omits the knob.
	Seed *int
	// Stop is the stop sequence that ends generation.
	Stop string
	// User is an end-user identifier forwarded for abuse monitoring.
	User string
	// ToolChoice controls how the model selects tools.
	ToolChoice string
	// ParallelToolCalls enables parallel tool calls. nil omits the knob.
	ParallelToolCalls *bool
	// ResponseFormat requests a specific response format (e.g. "json_object").
	ResponseFormat string
	// ReasoningEffort sets the reasoning effort for reasoning-capable models.
	ReasoningEffort string
}

// LLMOptions configures NewLLM.
type LLMOptions struct {
	// APIKey authenticates with Azure OpenAI. Falls back to AZURE_OPENAI_API_KEY.
	APIKey string
	// AzureEndpoint is the Azure OpenAI endpoint. Falls back to AZURE_OPENAI_ENDPOINT.
	AzureEndpoint string
	// Deployment is the Azure OpenAI deployment name.
	Deployment string
	// APIVersion is the Azure OpenAI REST API version. Defaults to "2024-10-21".
	APIVersion string
	// Temperature is the sampling temperature. Defaults to 0.7.
	Temperature *float64
	// MaxOutputTokens caps tokens generated per response. Defaults to 1024.
	MaxOutputTokens int
	// TopP is the nucleus-sampling probability mass. nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes tokens by frequency. nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens by prior presence. nil = provider default.
	PresencePenalty *float64
	// Seed makes sampling deterministic. nil = provider default.
	Seed *int
	// Stop is the stop sequence that ends generation.
	Stop string
	// User is an end-user identifier forwarded for abuse monitoring.
	User string
	// ToolChoice controls how the model selects tools.
	ToolChoice string
	// ParallelToolCalls enables parallel tool calls. nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat requests a specific response format (e.g. "json_object").
	ResponseFormat string
	// ReasoningEffort sets the reasoning effort for reasoning-capable models.
	ReasoningEffort string
}

// NewLLM returns an Azure OpenAI LLM configured from opts, applying defaults.
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

// LLMConfig returns the runtime provider, model, and generation limits.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "azure_openai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the extra provider parameters, omitting unset optional fields.
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
