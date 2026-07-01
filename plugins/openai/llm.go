package openai

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
	ResponseFormat    string
	ToolChoice        string
	ParallelToolCalls *bool
	Stop              string
	User              string
	ReasoningEffort   string
	Verbosity         string
	Streaming         bool
	WssURL            string
	Store             bool
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
	ResponseFormat    string
	ToolChoice        string
	ParallelToolCalls *bool
	Stop              string
	User              string
	ReasoningEffort   string
	Verbosity         string
	Streaming         *bool
	WssURL            string
	Store             *bool
}

func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "gpt-5.4-nano"),
		Temperature:       temp,
		MaxOutputTokens:   zrt.IntZeroOr(opts.MaxOutputTokens, 1024),
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		ResponseFormat:    opts.ResponseFormat,
		ToolChoice:        opts.ToolChoice,
		ParallelToolCalls: opts.ParallelToolCalls,
		Stop:              opts.Stop,
		User:              opts.User,
		ReasoningEffort:   opts.ReasoningEffort,
		Verbosity:         opts.Verbosity,
		Streaming:         zrt.BoolOr(opts.Streaming, false),
		WssURL:            opts.WssURL,
		Store:             zrt.BoolOr(opts.Store, true),
	}
	l.Init("openai", zrt.APIKeyOr(opts.APIKey, "OPENAI_API_KEY"))
	return l
}

func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "openai", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
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
	if l.ResponseFormat != "" {
		k["response_format"] = l.ResponseFormat
	}
	if l.ToolChoice != "" {
		k["tool_choice"] = l.ToolChoice
	}
	if l.ParallelToolCalls != nil {
		k["parallel_tool_calls"] = *l.ParallelToolCalls
	}
	if l.Stop != "" {
		k["stop"] = l.Stop
	}
	if l.User != "" {
		k["user"] = l.User
	}
	if l.ReasoningEffort != "" {
		k["reasoning_effort"] = l.ReasoningEffort
	}
	if l.Verbosity != "" {
		k["verbosity"] = l.Verbosity
	}
	k["streaming"] = l.Streaming
	if l.WssURL != "" {
		k["wss_url"] = l.WssURL
	}
	k["store"] = l.Store
	return k
}
