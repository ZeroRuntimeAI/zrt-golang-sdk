package cerebras

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

type LLM struct {
	zrt.BaseLLM
	Model           string
	Temperature     float64
	MaxOutputTokens int
	ToolChoice      string
	TopP            *float64
	Seed            *int
	Stop            string
	User            string
}

type LLMOptions struct {
	APIKey              string
	Model               string
	Temperature         *float64
	MaxCompletionTokens *int
	ToolChoice          string
	TopP                *float64
	Seed                *int
	Stop                string
	User                string
}

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

func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "cerebras", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

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
