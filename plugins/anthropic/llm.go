package anthropic

import (
	"strings"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// LLM is the Anthropic Claude chat-completion provider, wrapping the
// Messages API with optional extended-thinking support.
type LLM struct {
	zrt.BaseLLM
	// Model is the Anthropic model id.
	Model string
	// Temperature is the sampling temperature in [0.0, 1.0]; lower values are more deterministic.
	Temperature float64
	// MaxOutputTokens is the maximum number of tokens the model may generate.
	MaxOutputTokens int
	// TopP is the nucleus sampling probability mass in (0.0, 1.0]; nil = provider default.
	TopP *float64
	// TopK restricts sampling to the top-K most likely tokens; nil = provider default.
	TopK *int
	// StopSequences are strings at which the model stops generating.
	StopSequences []string
	// ThinkingBudget is the token budget for Claude's extended thinking; nil disables thinking.
	ThinkingBudget *int
}

// LLMOptions configures NewLLM.
type LLMOptions struct {
	// APIKey is the Anthropic API key; falls back to ANTHROPIC_API_KEY.
	APIKey string
	// Model is the Anthropic model id. Defaults to "claude-sonnet-4-20250514".
	Model string
	// Temperature is the sampling temperature in [0.0, 1.0]; nil applies 0.7.
	Temperature *float64
	// MaxOutputTokens is the maximum number of tokens the model may generate; 0 applies 1024.
	MaxOutputTokens int
	// TopP is the nucleus sampling probability mass in (0.0, 1.0]; nil = provider default.
	TopP *float64
	// TopK restricts sampling to the top-K most likely tokens; nil = provider default.
	TopK *int
	// StopSequences are strings at which the model stops generating.
	StopSequences []string
	// ThinkingBudget is the token budget for Claude's extended thinking; nil disables thinking.
	ThinkingBudget *int
}

// NewLLM builds an LLM from opts, applying defaults.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	mot := opts.MaxOutputTokens
	if mot == 0 {
		mot = 1024
	}
	l := &LLM{
		Model:           zrt.StrOr(opts.Model, "claude-sonnet-4-20250514"),
		Temperature:     temp,
		MaxOutputTokens: mot,
		TopP:            opts.TopP,
		TopK:            opts.TopK,
		StopSequences:   opts.StopSequences,
		ThinkingBudget:  opts.ThinkingBudget,
	}
	l.Init("anthropic", zrt.APIKeyOr(opts.APIKey, "ANTHROPIC_API_KEY"))
	return l
}

// LLMConfig returns the runtime configuration for this provider.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "anthropic", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the provider-specific tuning parameters that are set.
func (l *LLM) Knobs() map[string]any {
	k := map[string]any{}
	if l.TopP != nil {
		k["top_p"] = *l.TopP
	}
	if l.TopK != nil {
		k["top_k"] = *l.TopK
	}
	if len(l.StopSequences) > 0 {
		k["stop_sequences"] = strings.Join(l.StopSequences, ",")
	}
	if l.ThinkingBudget != nil {
		k["thinking_budget"] = *l.ThinkingBudget
	}
	return k
}
