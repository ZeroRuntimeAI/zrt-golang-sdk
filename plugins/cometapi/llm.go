// Package cometapi provides the CometAPI LLM provider.
package cometapi

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the CometAPI LLM descriptor.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures LLM.
type LLMOptions struct {
	// APIKey overrides the COMETAPI_API_KEY environment variable.
	APIKey              string
	Model               string   // default "gpt-4o-mini"
	Temperature         *float64 // nil uses the default (0.7).
	MaxCompletionTokens *int     // default 1024
}

// NewLLM builds an LLM.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{Model: zrt.StrOr(opts.Model, "gpt-4o-mini"), Temperature: temp, MaxOutputTokens: zrt.IntOr(opts.MaxCompletionTokens, 1024)}
	l.Init("cometapi", zrt.APIKeyOr(opts.APIKey, "COMETAPI_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "cometapi", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
