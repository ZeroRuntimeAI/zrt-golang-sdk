package cometapi

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

type LLM struct {
	zrt.BaseLLM
	Model             string
	Temperature       float64
	MaxOutputTokens   int
	BaseURL           string
	ToolChoice        string
	TopP              *float64
	FrequencyPenalty  *float64
	PresencePenalty   *float64
	Seed              *int
	Stop              string
	User              string
	ParallelToolCalls *bool
	ResponseFormat    string
}

type LLMOptions struct {
	APIKey              string
	Model               string
	Temperature         *float64
	MaxCompletionTokens *int
	BaseURL             string
	ToolChoice          string
	TopP                *float64
	FrequencyPenalty    *float64
	PresencePenalty     *float64
	Seed                *int
	Stop                string
	User                string
	ParallelToolCalls   *bool
	ResponseFormat      string
}

func NewLLM(opts LLMOptions) *LLM {
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:             zrt.StrOr(opts.Model, "gpt-4o-mini"),
		Temperature:       temp,
		MaxOutputTokens:   zrt.IntOr(opts.MaxCompletionTokens, 1024),
		BaseURL:           zrt.StrOr(opts.BaseURL, os.Getenv("COMETAPI_BASE_URL")),
		ToolChoice:        zrt.StrOr(opts.ToolChoice, "auto"),
		TopP:              opts.TopP,
		FrequencyPenalty:  opts.FrequencyPenalty,
		PresencePenalty:   opts.PresencePenalty,
		Seed:              opts.Seed,
		Stop:              opts.Stop,
		User:              opts.User,
		ParallelToolCalls: opts.ParallelToolCalls,
		ResponseFormat:    opts.ResponseFormat,
	}
	l.Init("cometapi", zrt.APIKeyOr(opts.APIKey, "COMETAPI_API_KEY"))
	return l
}

func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "cometapi", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

func (l *LLM) Knobs() map[string]any {
	k := map[string]any{}
	if l.BaseURL != "" {
		k["base_url"] = l.BaseURL
	}
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
	return k
}
