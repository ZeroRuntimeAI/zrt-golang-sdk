package cometapi

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// LLM is the CometAPI large-language-model provider.
type LLM struct {
	zrt.BaseLLM
	// Model is the model id. Defaults to "gpt-4o-mini".
	Model string
	// Temperature is the sampling temperature. Defaults to 0.7.
	Temperature float64
	// MaxOutputTokens caps generated tokens. Defaults to 1024.
	MaxOutputTokens int
	// BaseURL overrides the API endpoint; defaults to COMETAPI_BASE_URL.
	BaseURL string
	// ToolChoice controls tool selection (e.g. "auto", "none"). Defaults to "auto".
	ToolChoice string
	// TopP is the nucleus sampling probability mass; nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes token frequency; nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes token presence; nil = provider default.
	PresencePenalty *float64
	// Seed forces deterministic sampling; nil = provider default.
	Seed *int
	// Stop is a stop sequence; empty = none.
	Stop string
	// User is an end-user identifier passed to the provider; empty = none.
	User string
	// ParallelToolCalls enables parallel tool calls; nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat requests a response format (e.g. "json_object"); empty = provider default.
	ResponseFormat string
}

// LLMOptions configures NewLLM.
type LLMOptions struct {
	// APIKey is the CometAPI API key; falls back to COMETAPI_API_KEY.
	APIKey string
	// Model is the model id. Defaults to "gpt-4o-mini".
	Model string
	// Temperature is the sampling temperature; nil applies 0.7.
	Temperature *float64
	// MaxCompletionTokens caps generated tokens; nil applies 1024.
	MaxCompletionTokens *int
	// BaseURL overrides the API endpoint; empty applies COMETAPI_BASE_URL.
	BaseURL string
	// ToolChoice controls tool selection (e.g. "auto", "none"); empty applies "auto".
	ToolChoice string
	// TopP is the nucleus sampling probability mass; nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes token frequency; nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes token presence; nil = provider default.
	PresencePenalty *float64
	// Seed forces deterministic sampling; nil = provider default.
	Seed *int
	// Stop is a stop sequence; empty = none.
	Stop string
	// User is an end-user identifier passed to the provider; empty = none.
	User string
	// ParallelToolCalls enables parallel tool calls; nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat requests a response format (e.g. "json_object"); empty = provider default.
	ResponseFormat string
}

// NewLLM builds an LLM from opts, applying defaults and resolving the API key.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "gpt-4o-mini"),
		Temperature:       temp,
		MaxOutputTokens:   zrt.IntOr(opts.MaxCompletionTokens, 1024),
		BaseURL:           zrt.StrOr(opts.BaseURL, os.Getenv("COMETAPI_BASE_URL")),
		ToolChoice:        zrt.StrOr(opts.ToolChoice, "auto"),
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		Stop:              opts.Stop,
		User:              opts.User,
		ParallelToolCalls: opts.ParallelToolCalls,
		ResponseFormat:    opts.ResponseFormat,
	}
	l.Init("cometapi", zrt.APIKeyOr(opts.APIKey, "COMETAPI_API_KEY"))
	return l
}

// LLMConfig returns the runtime configuration for this provider.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "cometapi", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the set of non-default provider parameters to pass to the runtime.
func (l *LLM) Knobs() map[string]any {
	k := map[string]any{}
	if l.BaseURL != "" {
		k["base_url"] = l.BaseURL
	}
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
	return k
}
