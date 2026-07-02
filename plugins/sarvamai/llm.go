package sarvamai

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// LLM is a Sarvam AI large-language-model engine.
type LLM struct {
	zrt.BaseLLM
	// Model is the model id. Defaults to "sarvam-30b".
	Model string
	// Temperature controls sampling randomness. Defaults to 0.7.
	Temperature float64
	// MaxOutputTokens caps the number of generated tokens. Defaults to 1024.
	MaxOutputTokens int
	// ToolChoice controls tool selection. Defaults to "auto".
	ToolChoice string
	// TopP is the nucleus-sampling probability mass. nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes repeated tokens. nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens already present. nil = provider default.
	PresencePenalty *float64
	// Seed makes sampling deterministic. nil = provider default.
	Seed *int
	// Stop is the stop sequence that ends generation.
	Stop string
	// User is an opaque end-user identifier.
	User string
	// ParallelToolCalls enables concurrent tool calls. nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat requests a structured response format.
	ResponseFormat string
	// ReasoningEffort sets the reasoning effort level.
	ReasoningEffort string
	// WikiGrounding grounds responses in wiki knowledge. Defaults to false.
	WikiGrounding bool
}

// LLMOptions configures NewLLM.
type LLMOptions struct {
	// APIKey is the Sarvam AI API key. If empty, the SARVAM_API_KEY environment variable is used.
	APIKey string
	// Model is the model id. Defaults to "sarvam-30b".
	Model string
	// Temperature controls sampling randomness. Defaults to 0.7.
	Temperature *float64
	// MaxCompletionTokens caps the number of generated tokens. Defaults to 1024.
	MaxCompletionTokens *int
	// ToolChoice controls tool selection. Defaults to "auto".
	ToolChoice string
	// TopP is the nucleus-sampling probability mass. nil = provider default.
	TopP *float64
	// FrequencyPenalty penalizes repeated tokens. nil = provider default.
	FrequencyPenalty *float64
	// PresencePenalty penalizes tokens already present. nil = provider default.
	PresencePenalty *float64
	// Seed makes sampling deterministic. nil = provider default.
	Seed *int
	// Stop is the stop sequence that ends generation.
	Stop string
	// User is an opaque end-user identifier.
	User string
	// ParallelToolCalls enables concurrent tool calls. nil = provider default.
	ParallelToolCalls *bool
	// ResponseFormat requests a structured response format.
	ResponseFormat string
	// ReasoningEffort sets the reasoning effort level.
	ReasoningEffort string
	// WikiGrounding grounds responses in wiki knowledge. Defaults to false.
	WikiGrounding *bool
}

// NewLLM returns a Sarvam AI LLM configured from opts.
func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "sarvam-30b"),
		Temperature:       temp,
		MaxOutputTokens:   zrt.IntOr(opts.MaxCompletionTokens, 1024),
		ToolChoice:        zrt.StrOr(opts.ToolChoice, "auto"),
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		Stop:              opts.Stop,
		User:              opts.User,
		ParallelToolCalls: opts.ParallelToolCalls,
		ResponseFormat:    opts.ResponseFormat,
		ReasoningEffort:   opts.ReasoningEffort,
		WikiGrounding:     zrt.BoolOr(opts.WikiGrounding, false),
	}
	l.Init("sarvamai", zrt.APIKeyOr(opts.APIKey, "SARVAM_API_KEY"))
	return l
}

// LLMConfig returns the provider, model, temperature, and max output tokens for this engine.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "sarvamai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the Sarvam AI-specific options as a key/value map.
func (l *LLM) Knobs() map[string]any {
	k := map[string]any{"model": l.Model}
	k["tool_choice"] = l.ToolChoice
	if l.TopP != nil {
		k["top_p"] = *l.TopP
	}
	if l.FrequencyPenalty != nil {
		k["frequency_penalty"] = *l.FrequencyPenalty
	}
	if l.PresencePenalty != nil {
		k["presence_penalty"] = *l.PresencePenalty
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
	if l.ParallelToolCalls != nil {
		k["parallel_tool_calls"] = *l.ParallelToolCalls
	}
	if l.ResponseFormat != "" {
		k["response_format"] = l.ResponseFormat
	}
	if l.ReasoningEffort != "" {
		k["reasoning_effort"] = l.ReasoningEffort
	}
	k["wiki_grounding"] = l.WikiGrounding
	return k
}
