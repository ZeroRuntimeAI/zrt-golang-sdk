// Package xai provides the xAI Grok LLM and realtime providers.
package xai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the xAI Grok LLM descriptor.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
	BaseURL         string
}

// LLMOptions configures LLM.
type LLMOptions struct {
	// APIKey overrides the XAI_API_KEY environment variable.
	APIKey              string
	Model               string   // default "grok-4-1-fast-non-reasoning"
	BaseURL             string   // default "https://api.x.ai/v1"
	Temperature         *float64 // nil uses the default (0.7).
	MaxCompletionTokens *int     // default 1024
}

// NewLLM builds an LLM.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:           zrt.StrOr(opts.Model, "grok-4-1-fast-non-reasoning"),
		Temperature:     temp,
		MaxOutputTokens: zrt.IntOr(opts.MaxCompletionTokens, 1024),
		BaseURL:         zrt.StrOr(opts.BaseURL, "https://api.x.ai/v1"),
	}
	l.Init("xai", zrt.APIKeyOr(opts.APIKey, "XAI_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "xai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
