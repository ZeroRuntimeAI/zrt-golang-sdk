package openai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the OpenAI large-language-model provider.
type LLM struct {
	zrt.BaseLLM
	// Model is the model id. Defaults to "gpt-5.4-nano".
	Model string
	// Temperature is the sampling temperature. Defaults to 0.7.
	Temperature float64
	// MaxOutputTokens caps generated tokens. Defaults to 1024.
	MaxOutputTokens int
	// TopP is the nucleus sampling probability mass; nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes token frequency; nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes token presence; nil = provider default.
	PresencePenalty *float64
	// Seed forces deterministic sampling; nil = provider default.
	Seed *int
	// ResponseFormat requests a response format (e.g. "json_object"); empty = provider default.
	ResponseFormat string
	// ToolChoice controls tool selection (e.g. "auto", "none"); empty = provider default.
	ToolChoice string
	// ParallelToolCalls enables parallel tool calls; nil = provider default.
	ParallelToolCalls *bool
	// Stop is a stop sequence; empty = none.
	Stop string
	// User is an end-user identifier passed to OpenAI; empty = none.
	User string
	// ReasoningEffort sets reasoning effort for reasoning models; empty = provider default.
	ReasoningEffort string
	// Verbosity controls response verbosity; empty = provider default.
	Verbosity string
	// Streaming enables streamed responses. Defaults to false.
	Streaming bool
	// WssURL overrides the WebSocket endpoint; empty = default.
	WssURL string
	// Store enables server-side response storage. Defaults to true.
	Store bool
}

// LLMOptions configures NewLLM.
type LLMOptions struct {
	// APIKey is the OpenAI API key; falls back to OPENAI_API_KEY.
	APIKey string
	// Model is the model id. Defaults to "gpt-5.4-nano".
	Model string
	// Temperature is the sampling temperature; nil applies 0.7.
	Temperature *float64
	// MaxOutputTokens caps generated tokens. Defaults to 1024.
	MaxOutputTokens int
	// TopP is the nucleus sampling probability mass; nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes token frequency; nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes token presence; nil = provider default.
	PresencePenalty *float64
	// Seed forces deterministic sampling; nil = provider default.
	Seed *int
	// ResponseFormat requests a response format (e.g. "json_object"); empty = provider default.
	ResponseFormat string
	// ToolChoice controls tool selection (e.g. "auto", "none"); empty = provider default.
	ToolChoice string
	// ParallelToolCalls enables parallel tool calls; nil = provider default.
	ParallelToolCalls *bool
	// Stop is a stop sequence; empty = none.
	Stop string
	// User is an end-user identifier passed to OpenAI; empty = none.
	User string
	// ReasoningEffort sets reasoning effort for reasoning models; empty = provider default.
	ReasoningEffort string
	// Verbosity controls response verbosity; empty = provider default.
	Verbosity string
	// Streaming enables streamed responses; nil applies false.
	Streaming *bool
	// WssURL overrides the WebSocket endpoint; empty = default.
	WssURL string
	// Store enables server-side response storage; nil applies true.
	Store *bool
}

// NewLLM builds an LLM from opts, applying defaults and resolving the API key.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "gpt-5.4-nano"),
		Temperature:       temp,
		MaxOutputTokens:   zrt.IntZeroOr(opts.MaxOutputTokens, 1024),
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		ResponseFormat:    opts.ResponseFormat,
		ToolChoice:        opts.ToolChoice,
		ParallelToolCalls: opts.ParallelToolCalls,
		Stop:              opts.Stop,
		User:              opts.User,
		ReasoningEffort:   opts.ReasoningEffort,
		Verbosity:         opts.Verbosity,
		Streaming:         zrt.BoolOr(opts.Streaming, false),
		WssURL:            opts.WssURL,
		Store:             zrt.BoolOr(opts.Store, true),
	}
	l.Init("openai", zrt.APIKeyOr(opts.APIKey, "OPENAI_API_KEY"))
	return l
}

// LLMConfig returns the runtime configuration for this provider.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "openai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the set of non-default provider parameters to pass to the runtime.
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
	if l.ResponseFormat != "" {
		k["response_format"] = l.ResponseFormat
	}
	if l.ToolChoice != "" {
		k["tool_choice"] = l.ToolChoice
	}
	if l.ParallelToolCalls != nil {
		k["parallel_tool_calls"] = *l.ParallelToolCalls
	}
	if l.Stop != "" {
		k["stop"] = l.Stop
	}
	if l.User != "" {
		k["user"] = l.User
	}
	if l.ReasoningEffort != "" {
		k["reasoning_effort"] = l.ReasoningEffort
	}
	if l.Verbosity != "" {
		k["verbosity"] = l.Verbosity
	}
	k["streaming"] = l.Streaming
	if l.WssURL != "" {
		k["wss_url"] = l.WssURL
	}
	k["store"] = l.Store
	return k
}
