package groq

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the Groq large language model provider.
type LLM struct {
	zrt.BaseLLM
	// Model is the Groq model id.
	Model string
	// Temperature controls sampling randomness.
	Temperature float64
	// MaxOutputTokens caps the number of tokens generated per response.
	MaxOutputTokens int
	// TopP is the nucleus sampling probability mass; nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes tokens by frequency; nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens by presence; nil = provider default.
	PresencePenalty *float64
	// Seed sets a deterministic sampling seed; nil = provider default.
	Seed *int
	// Stop is a stop sequence that halts generation.
	Stop string
	// User is an end-user identifier passed to the provider.
	User string
	// ToolChoice controls tool selection (e.g. "auto", "none", "required").
	ToolChoice string
	// ParallelToolCalls enables parallel tool calls; nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat selects the response format (e.g. "json_object").
	ResponseFormat string
	// ReasoningEffort controls the reasoning effort level.
	ReasoningEffort string
	// ReasoningFormat controls how reasoning content is returned.
	ReasoningFormat string
	// ServiceTier selects the provider service tier.
	ServiceTier string
}

// LLMOptions configures NewLLM.
type LLMOptions struct {
	// APIKey overrides the GROQ_API_KEY environment variable.
	APIKey string
	// Model is the Groq model id. Defaults to "llama-3.3-70b-versatile".
	Model string
	// Temperature controls sampling randomness. Defaults to 0.7.
	Temperature *float64
	// MaxOutputTokens caps the number of tokens generated per response. Defaults to 1024.
	MaxOutputTokens int
	// TopP is the nucleus sampling probability mass; nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes tokens by frequency; nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens by presence; nil = provider default.
	PresencePenalty *float64
	// Seed sets a deterministic sampling seed; nil = provider default.
	Seed *int
	// Stop is a stop sequence that halts generation.
	Stop string
	// User is an end-user identifier passed to the provider.
	User string
	// ToolChoice controls tool selection (e.g. "auto", "none", "required").
	ToolChoice string
	// ParallelToolCalls enables parallel tool calls; nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat selects the response format (e.g. "json_object").
	ResponseFormat string
	// ReasoningEffort controls the reasoning effort level.
	ReasoningEffort string
	// ReasoningFormat controls how reasoning content is returned.
	ReasoningFormat string
	// ServiceTier selects the provider service tier.
	ServiceTier string
}

// NewLLM returns a Groq LLM configured from opts, applying defaults and
// falling back to the GROQ_API_KEY environment variable for the API key.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	mot := opts.MaxOutputTokens
	if mot == 0 {
		mot = 1024
	}
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "llama-3.3-70b-versatile"),
		Temperature:       temp,
		MaxOutputTokens:   mot,
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
		ReasoningFormat:   opts.ReasoningFormat,
		ServiceTier:       opts.ServiceTier,
	}
	l.Init("groq", zrt.APIKeyOr(opts.APIKey, "GROQ_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "groq", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the provider-specific parameters set on l.
func (l *LLM) Knobs() map[string]any {
	k := map[string]any{}
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
	if l.ReasoningFormat != "" {
		k["reasoning_format"] = l.ReasoningFormat
	}
	if l.ServiceTier != "" {
		k["service_tier"] = l.ServiceTier
	}
	return k
}
