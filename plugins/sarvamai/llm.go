package sarvamai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the Sarvam AI LLM descriptor.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures LLM.
type LLMOptions struct {
	// APIKey overrides the SARVAM_API_KEY environment variable.
	APIKey              string
	Model               string   // default "sarvam-30b"
	Temperature         *float64 // nil uses the default (0.7).
	MaxCompletionTokens *int     // default 1024
}

// NewLLM builds an LLM.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{Model: zrt.StrOr(opts.Model, "sarvam-30b"), Temperature: temp, MaxOutputTokens: zrt.IntOr(opts.MaxCompletionTokens, 1024)}
	l.Init("sarvamai", zrt.APIKeyOr(opts.APIKey, "SARVAM_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "sarvamai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs implements the credential knob source.
func (l *LLM) Knobs() map[string]any {
	return map[string]any{"model": l.Model}
}
