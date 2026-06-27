// Package groq provides the Groq LLM and TTS providers.
package groq

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is a Groq language model configured for use as an agent's LLM.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures a Groq LLM.
type LLMOptions struct {
	// APIKey is the Groq API key. If empty, the GROQ_API_KEY environment
	// variable is used.
	APIKey string
	// Model is the Groq model to use. Defaults to "llama-3.3-70b-versatile".
	Model string
	// Temperature controls randomness. nil uses the default (0.7).
	Temperature *float64
	// MaxOutputTokens caps the response length. Defaults to 1024.
	MaxOutputTokens int
}

// NewLLM creates a Groq LLM from opts, applying defaults for any unset fields.
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

// LLMConfig implements zrt.LLM and reports the model configuration.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "groq", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
