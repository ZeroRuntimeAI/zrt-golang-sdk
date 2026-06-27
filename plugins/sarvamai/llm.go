package sarvamai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is a Sarvam AI language model configured for use as an agent's LLM.
type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
}

// LLMOptions configures a Sarvam AI LLM.
type LLMOptions struct {
	// APIKey is the Sarvam AI API key. If empty, the SARVAM_API_KEY environment
	// variable is used.
	APIKey string
	// Model is the Sarvam AI model to use. Defaults to "sarvam-30b".
	Model string
	// Temperature controls randomness. nil uses the default (0.7).
	Temperature *float64
	// MaxCompletionTokens caps the response length. nil uses the default (1024).
	MaxCompletionTokens *int
}

// NewLLM creates a Sarvam AI LLM from opts, applying defaults for any unset
// fields.
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
