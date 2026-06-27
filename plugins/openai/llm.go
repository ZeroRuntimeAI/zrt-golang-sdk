// Package openai provides the OpenAI LLM provider.
package openai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is an OpenAI language model configured for use as an agent's LLM.
type LLM struct {
	zrt.BaseLLM
	Model             string
	Temperature       float64
	MaxOutputTokens   int
	TopP              *float64
	FrequencyPenalty  *float64
	PresencePenalty   *float64
	Seed              *int
	ResponseFormat    string
	ToolChoice        string
	ParallelToolCalls *bool
}

// LLMOptions configures an OpenAI LLM.
type LLMOptions struct {
	// APIKey is the OpenAI API key. If empty, the OPENAI_API_KEY environment
	// variable is used.
	APIKey string
	// Model is the OpenAI model to use. Defaults to "gpt-4o".
	Model string
	// Temperature controls randomness. nil uses the default (0.7).
	Temperature *float64
	// MaxOutputTokens caps the response length. Defaults to 1024.
	MaxOutputTokens int
	// TopP is the nucleus-sampling probability mass.
	TopP *float64
	// FrequencyPenalty penalizes tokens in proportion to their frequency.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens that have already appeared.
	PresencePenalty *float64
	// Seed makes sampling deterministic for a given prompt.
	Seed *int
	// ResponseFormat requests a specific output format, such as "json_object".
	ResponseFormat string
	// ToolChoice controls how the model selects tools (for example "auto" or "none").
	ToolChoice string
	// ParallelToolCalls allows the model to invoke multiple tools in one turn.
	ParallelToolCalls *bool
}

// NewLLM creates an OpenAI LLM from opts, applying defaults for any unset fields.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "gpt-4o"),
		Temperature:       temp,
		MaxOutputTokens:   orInt(opts.MaxOutputTokens, 1024),
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		ResponseFormat:    opts.ResponseFormat,
		ToolChoice:        opts.ToolChoice,
		ParallelToolCalls: opts.ParallelToolCalls,
	}
	l.Init("openai", zrt.APIKeyOr(opts.APIKey, "OPENAI_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM and reports the model configuration.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "openai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the provider-specific tuning options that are set.
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
	return k
}

func orInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
