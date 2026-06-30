package groq

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type LLM struct {
	zrt.BaseLLM
	Model             string
	Temperature       float64
	MaxOutputTokens   int
	TopP              *float64
	FrequencyPenalty  *float64
	PresencePenalty   *float64
	Seed              *int
	Stop              string
	User              string
	ToolChoice        string
	ParallelToolCalls *bool
	ResponseFormat    string
	ReasoningEffort   string
	ReasoningFormat   string
	ServiceTier       string
}

type LLMOptions struct {
	APIKey            string
	Model             string
	Temperature       *float64
	MaxOutputTokens   int
	TopP              *float64
	FrequencyPenalty  *float64
	PresencePenalty   *float64
	Seed              *int
	Stop              string
	User              string
	ToolChoice        string
	ParallelToolCalls *bool
	ResponseFormat    string
	ReasoningEffort   string
	ReasoningFormat   string
	ServiceTier       string
}

func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	mot := opts.MaxOutputTokens
	if mot == 0 {
		mot = 1024
	}
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "llama-3.3-70b-versatile"),
		Temperature:       temp,
		MaxOutputTokens:   mot,
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		Stop:              opts.Stop,
		User:              opts.User,
		ToolChoice:        opts.ToolChoice,
		ParallelToolCalls: opts.ParallelToolCalls,
		ResponseFormat:    opts.ResponseFormat,
		ReasoningEffort:   opts.ReasoningEffort,
		ReasoningFormat:   opts.ReasoningFormat,
		ServiceTier:       opts.ServiceTier,
	}
	l.Init("groq", zrt.APIKeyOr(opts.APIKey, "GROQ_API_KEY"))
	return l
}

func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "groq", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

func (l *LLM) Knobs() map[string]any {
	k := map[string]any{}
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
	if l.ToolChoice != "" {
		k["tool_choice"] = l.ToolChoice
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
	if l.ReasoningFormat != "" {
		k["reasoning_format"] = l.ReasoningFormat
	}
	if l.ServiceTier != "" {
		k["service_tier"] = l.ServiceTier
	}
	return k
}
