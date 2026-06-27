// Package anthropic provides the Anthropic Claude LLM provider.
package anthropic

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is an Anthropic Claude language model configured for use as an agent's LLM.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures an Anthropic Claude LLM.
type LLMOptions struct {
	// APIKey is the Anthropic API key. If empty, the ANTHROPIC_API_KEY
	// environment variable is used.
	APIKey string
	// Model is the Claude model to use. Defaults to "claude-sonnet-4-20250514".
	Model string
	// Temperature controls randomness. nil uses the default (0.7).
	Temperature *float64
	// MaxOutputTokens caps the response length. Defaults to 1024.
	MaxOutputTokens int
}

// NewLLM creates an Anthropic Claude LLM from opts, applying defaults for any
// unset fields.
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

// LLMConfig implements zrt.LLM and reports the model configuration.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "anthropic", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
