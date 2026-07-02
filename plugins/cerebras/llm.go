package cerebras

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is the Cerebras large-language-model provider offering ultra-fast
// OpenAI-compatible chat completions on Cerebras wafer-scale hardware.
type LLM struct {
	zrt.BaseLLM
	// Model is the model id (e.g. "llama3.3-70b", "gpt-oss-120b", "zai-glm-4.7").
	Model string
	// Temperature is the sampling temperature (0.0-1.5).
	Temperature float64
	// MaxOutputTokens is the maximum number of tokens to generate.
	MaxOutputTokens int
	// ToolChoice is the tool selection mode ("auto", "none", or "required").
	ToolChoice string
	// TopP is the nucleus sampling probability mass; nil = provider default.
	TopP *float64
	// Seed is the seed for reproducible sampling; nil = provider default.
	Seed *int
	// Stop is the stop sequence that halts generation.
	Stop string
	// User is a stable end-user identifier used for abuse monitoring.
	User string
}

// LLMOptions configures NewLLM.
type LLMOptions struct {
	// APIKey is the Cerebras API key. Falls back to CEREBRAS_API_KEY.
	APIKey string
	// Model is the model id. Defaults to "llama3.3-70b".
	Model string
	// Temperature is the sampling temperature (0.0-1.5). Defaults to 0.7.
	Temperature *float64
	// MaxCompletionTokens is the maximum number of tokens to generate. Defaults to 1024.
	MaxCompletionTokens *int
	// ToolChoice is the tool selection mode ("auto", "none", or "required"). Defaults to "auto".
	ToolChoice string
	// TopP is the nucleus sampling probability mass; nil = provider default.
	TopP *float64
	// Seed is the seed for reproducible sampling; nil = provider default.
	Seed *int
	// Stop is the stop sequence that halts generation.
	Stop string
	// User is a stable end-user identifier used for abuse monitoring.
	User string
}

// NewLLM builds an LLM from opts, applying defaults and resolving the API key.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:           zrt.StrOr(opts.Model, "llama3.3-70b"),
		Temperature:     temp,
		MaxOutputTokens: zrt.IntOr(opts.MaxCompletionTokens, 1024),
		ToolChoice:      zrt.StrOr(opts.ToolChoice, "auto"),
		TopP:            opts.TopP,
		Seed:            opts.Seed,
		Stop:            opts.Stop,
		User:            opts.User,
	}
	l.Init("cerebras", zrt.APIKeyOr(opts.APIKey, "CEREBRAS_API_KEY"))
	return l
}

// LLMConfig returns the runtime configuration for this LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "cerebras", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the provider-specific tuning parameters as a map.
func (l *LLM) Knobs() map[string]any {
	k := map[string]any{}
	k["tool_choice"] = l.ToolChoice
	if l.TopP != nil {
		k["top_p"] = *l.TopP
	}
	if l.Seed != nil {
		k["seed"] = *l.Seed
	}
	if l.Stop != "" {
		k["stop"] = l.Stop
	}
	if l.User != "" {
		k["user"] = l.User
	}
	return k
}
