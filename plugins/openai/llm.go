// Package openai provides the OpenAI LLM provider.
package openai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the OpenAI LLM descriptor.
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

// LLMOptions configures LLM.
type LLMOptions struct {
	// APIKey overrides the OPENAI_API_KEY environment variable.
	APIKey            string
	Model             string   // default "gpt-4o"
	Temperature       *float64 // nil uses the default (0.7).
	MaxOutputTokens   int      // default 1024
	TopP              *float64
	FrequencyPenalty  *float64
	PresencePenalty   *float64
	Seed              *int
	ResponseFormat    string
	ToolChoice        string
	ParallelToolCalls *bool
}

// NewLLM builds an LLM.
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

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "openai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs implements the credential knob source.
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
