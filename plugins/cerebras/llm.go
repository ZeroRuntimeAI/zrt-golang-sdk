// Package cerebras provides the Cerebras LLM provider.
package cerebras

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the Cerebras LLM descriptor.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures LLM.
type LLMOptions struct {
	// APIKey overrides the CEREBRAS_API_KEY environment variable.
	APIKey              string
	Model               string   // default "llama3.3-70b"
	Temperature         *float64 // nil uses the default (0.7).
	MaxCompletionTokens *int     // default 1024
}

// NewLLM builds an LLM.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{Model: zrt.StrOr(opts.Model, "llama3.3-70b"), Temperature: temp, MaxOutputTokens: zrt.IntOr(opts.MaxCompletionTokens, 1024)}
	l.Init("cerebras", zrt.APIKeyOr(opts.APIKey, "CEREBRAS_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "cerebras", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
