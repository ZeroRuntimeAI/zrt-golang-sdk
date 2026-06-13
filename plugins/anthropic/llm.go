// Package anthropic provides the Anthropic Claude LLM provider.
package anthropic

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the Anthropic Claude LLM descriptor.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures LLM.
type LLMOptions struct {
	// APIKey overrides the ANTHROPIC_API_KEY environment variable.
	APIKey          string
	Model           string   // default "claude-sonnet-4-20250514"
	Temperature     *float64 // nil uses the default (0.7).
	MaxOutputTokens int      // default 1024
}

// NewLLM builds an LLM.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	mot := opts.MaxOutputTokens
	if mot == 0 {
		mot = 1024
	}
	l := &LLM{Model: zrt.StrOr(opts.Model, "claude-sonnet-4-20250514"), Temperature: temp, MaxOutputTokens: mot}
	l.Init("anthropic", zrt.APIKeyOr(opts.APIKey, "ANTHROPIC_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "anthropic", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
