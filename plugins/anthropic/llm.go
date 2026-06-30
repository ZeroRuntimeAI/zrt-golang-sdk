package anthropic

import (
	"strings"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
	TopP            *float64
	TopK            *int
	StopSequences   []string
	ThinkingBudget  *int
}

type LLMOptions struct {
	APIKey          string
	Model           string
	Temperature     *float64
	MaxOutputTokens int
	TopP            *float64
	TopK            *int
	StopSequences   []string
	ThinkingBudget  *int
}

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

func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "anthropic", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

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
