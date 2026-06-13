// Package groq provides the Groq LLM and TTS providers.
package groq

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the Groq LLM descriptor.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures LLM.
type LLMOptions struct {
	// APIKey overrides the GROQ_API_KEY environment variable.
	APIKey          string
	Model           string   // default "llama-3.3-70b-versatile"
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
	l := &LLM{Model: zrt.StrOr(opts.Model, "llama-3.3-70b-versatile"), Temperature: temp, MaxOutputTokens: mot}
	l.Init("groq", zrt.APIKeyOr(opts.APIKey, "GROQ_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "groq", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
