// Package xai provides the xAI Grok LLM and realtime providers.
package xai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is an xAI Grok large language model.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
	BaseURL         string
}

// LLMOptions configures an xAI Grok LLM.
type LLMOptions struct {
	// APIKey is the xAI API key. If empty, the XAI_API_KEY environment variable is used.
	APIKey string
	// Model is the Grok model to use. Defaults to "grok-4-1-fast-non-reasoning".
	Model string
	// BaseURL is the xAI API endpoint. Defaults to "https://api.x.ai/v1".
	BaseURL string
	// Temperature controls sampling randomness. nil uses the default (0.7).
	Temperature *float64
	// MaxCompletionTokens caps the response length. Defaults to 1024.
	MaxCompletionTokens *int
}

// NewLLM creates an xAI Grok LLM from the given options.
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
