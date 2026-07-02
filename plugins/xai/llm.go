package xai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the xAI (Grok) chat-completion provider.
type LLM struct {
	zrt.BaseLLM
	// Model is the model id.
	Model string
	// Temperature is the sampling temperature.
	Temperature float64
	// MaxOutputTokens caps the number of tokens generated per response.
	MaxOutputTokens int
	// BaseURL is the xAI API endpoint.
	BaseURL string
	// ToolChoice controls tool selection ("auto", "none", "required", or a tool name).
	ToolChoice string
	// TopP is the nucleus sampling probability mass. nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes tokens by their existing frequency. nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens that already appeared. nil = provider default.
	PresencePenalty *float64
	// Seed makes sampling deterministic for a given input. nil = provider default.
	Seed *int
	// Stop is a stop sequence that ends generation.
	Stop string
	// User is an end-user identifier passed to the API.
	User string
	// ParallelToolCalls enables parallel tool calls. nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat requests a response format (e.g. "json_object").
	ResponseFormat string
	// ReasoningEffort sets the reasoning effort level for reasoning models.
	ReasoningEffort string
	// PromptCacheKey is a key used to route requests to prompt caches.
	PromptCacheKey string
	// ServiceTier selects the API service tier.
	ServiceTier string
}

// LLMOptions configures NewLLM.
type LLMOptions struct {
	// APIKey authenticates with xAI. Falls back to the XAI_API_KEY environment variable.
	APIKey string
	// Model is the model id. Defaults to "grok-4-1-fast-non-reasoning".
	Model string
	// BaseURL is the xAI API endpoint. Defaults to "https://api.x.ai/v1".
	BaseURL string
	// Temperature is the sampling temperature. nil defaults to 0.7.
	Temperature *float64
	// MaxCompletionTokens caps the number of tokens generated per response. nil defaults to 1024.
	MaxCompletionTokens *int
	// ToolChoice controls tool selection. Defaults to "auto".
	ToolChoice string
	// TopP is the nucleus sampling probability mass. nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes tokens by their existing frequency. nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens that already appeared. nil = provider default.
	PresencePenalty *float64
	// Seed makes sampling deterministic for a given input. nil = provider default.
	Seed *int
	// Stop is a stop sequence that ends generation.
	Stop string
	// User is an end-user identifier passed to the API.
	User string
	// ParallelToolCalls enables parallel tool calls. nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat requests a response format (e.g. "json_object").
	ResponseFormat string
	// ReasoningEffort sets the reasoning effort level for reasoning models.
	ReasoningEffort string
	// PromptCacheKey is a key used to route requests to prompt caches.
	PromptCacheKey string
	// ServiceTier selects the API service tier.
	ServiceTier string
}

// NewLLM builds an LLM from opts, applying defaults.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "grok-4-1-fast-non-reasoning"),
		Temperature:       temp,
		MaxOutputTokens:   zrt.IntOr(opts.MaxCompletionTokens, 1024),
		BaseURL:           zrt.StrOr(opts.BaseURL, "https://api.x.ai/v1"),
		ToolChoice:        zrt.StrOr(opts.ToolChoice, "auto"),
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		Stop:              opts.Stop,
		User:              opts.User,
		ParallelToolCalls: opts.ParallelToolCalls,
		ResponseFormat:    opts.ResponseFormat,
		ReasoningEffort:   opts.ReasoningEffort,
		PromptCacheKey:    opts.PromptCacheKey,
		ServiceTier:       opts.ServiceTier,
	}
	l.Init("xai", zrt.APIKeyOr(opts.APIKey, "XAI_API_KEY"))
	return l
}

// LLMConfig returns the runtime configuration for this LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "xai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the provider-specific tuning parameters set on this LLM.
func (l *LLM) Knobs() map[string]any {
	k := map[string]any{}
	k["tool_choice"] = l.ToolChoice
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
	if l.ParallelToolCalls != nil {
		k["parallel_tool_calls"] = *l.ParallelToolCalls
	}
	if l.ResponseFormat != "" {
		k["response_format"] = l.ResponseFormat
	}
	if l.ReasoningEffort != "" {
		k["reasoning_effort"] = l.ReasoningEffort
	}
	if l.PromptCacheKey != "" {
		k["prompt_cache_key"] = l.PromptCacheKey
	}
	if l.ServiceTier != "" {
		k["service_tier"] = l.ServiceTier
	}
	return k
}
