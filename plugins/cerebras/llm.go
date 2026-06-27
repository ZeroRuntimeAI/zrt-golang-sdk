// Package cerebras provides the Cerebras LLM provider.
package cerebras

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is a Cerebras language model configured for use as an agent's LLM.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures a Cerebras LLM.
type LLMOptions struct {
	// APIKey is the Cerebras API key. If empty, the CEREBRAS_API_KEY
	// environment variable is used.
	APIKey string
	// Model is the Cerebras model to use. Defaults to "llama3.3-70b".
	Model string
	// Temperature controls randomness. nil uses the default (0.7).
	Temperature *float64
	// MaxCompletionTokens caps the response length. nil uses the default (1024).
	MaxCompletionTokens *int
}

// NewLLM creates a Cerebras LLM from opts, applying defaults for any unset
// fields.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{Model: zrt.StrOr(opts.Model, "llama3.3-70b"), Temperature: temp, MaxOutputTokens: zrt.IntOr(opts.MaxCompletionTokens, 1024)}
	l.Init("cerebras", zrt.APIKeyOr(opts.APIKey, "CEREBRAS_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM and reports the model configuration.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "cerebras", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
