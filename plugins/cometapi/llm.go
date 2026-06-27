// Package cometapi provides the CometAPI LLM provider.
package cometapi

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is a CometAPI language model configured for use as an agent's LLM.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures a CometAPI LLM.
type LLMOptions struct {
	// APIKey is the CometAPI API key. If empty, the COMETAPI_API_KEY
	// environment variable is used.
	APIKey string
	// Model is the CometAPI model to use. Defaults to "gpt-4o-mini".
	Model string
	// Temperature controls randomness. nil uses the default (0.7).
	Temperature *float64
	// MaxCompletionTokens caps the response length. nil uses the default (1024).
	MaxCompletionTokens *int
}

// NewLLM creates a CometAPI LLM from opts, applying defaults for any unset
// fields.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{Model: zrt.StrOr(opts.Model, "gpt-4o-mini"), Temperature: temp, MaxOutputTokens: zrt.IntOr(opts.MaxCompletionTokens, 1024)}
	l.Init("cometapi", zrt.APIKeyOr(opts.APIKey, "COMETAPI_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM and reports the model configuration.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "cometapi", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}
