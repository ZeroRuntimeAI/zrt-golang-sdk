// Package google provides the Google Gemini LLM and Google STT/TTS providers.
package google

import (
	"encoding/json"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// LLM is the Google Gemini LLM descriptor.
type LLM struct {
	zrt.BaseLLM
	Model            string
	Temperature      float64
	MaxOutputTokens  int
	ThinkingBudget   *int
	IncludeThoughts  bool
	SafetySettings   []zrt.SafetySetting
	TopP             *float64
	TopK             *int
	PresencePenalty  *float64
	FrequencyPenalty *float64
	Seed             *int

	vertexProjectID          string
	vertexLocation           string
	vertexServiceAccountJSON string
	vertexServiceAccountPath string
}

// LLMOptions configures LLM.
type LLMOptions struct {
	// APIKey overrides the GOOGLE_API_KEY environment variable.
	APIKey           string
	Model            string   // default "gemini-2.5-flash-lite"
	Temperature      *float64 // nil uses the default (0.7).
	MaxOutputTokens  int      // default 8192
	ThinkingBudget   *int     // default 0
	IncludeThoughts  bool
	SafetySettings   []zrt.SafetySetting
	TopP             *float64
	TopK             *int
	PresencePenalty  *float64
	FrequencyPenalty *float64
	Seed             *int

	VertexAI           bool
	ProjectID          string
	Location           string // default "us-central1"
	ServiceAccountJSON any    // string or map[string]any
	ServiceAccountPath string
}

// NewLLM builds an LLM.
func NewLLM(opts LLMOptions) *LLM {
	tb := opts.ThinkingBudget
	if tb == nil {
		zero := 0
		tb = &zero
	}
	temp := zrt.FloatOr(opts.Temperature, 0.7)
	l := &LLM{
		Model:            zrt.StrOr(opts.Model, "gemini-2.5-flash-lite"),
		Temperature:      temp,
		MaxOutputTokens:  orInt(opts.MaxOutputTokens, 8192),
		ThinkingBudget:   tb,
		IncludeThoughts:  opts.IncludeThoughts,
		SafetySettings:   opts.SafetySettings,
		TopP:             opts.TopP,
		TopK:             opts.TopK,
		PresencePenalty:  opts.PresencePenalty,
		FrequencyPenalty: opts.FrequencyPenalty,
		Seed:             opts.Seed,
	}
	if opts.VertexAI {
		l.vertexProjectID = opts.ProjectID
		l.vertexLocation = zrt.StrOr(opts.Location, "us-central1")
		switch v := opts.ServiceAccountJSON.(type) {
		case string:
			l.vertexServiceAccountJSON = v
		case map[string]any:
			b, _ := json.Marshal(v)
			l.vertexServiceAccountJSON = string(b)
		default:
			if opts.ServiceAccountPath != "" {
				l.vertexServiceAccountPath = opts.ServiceAccountPath
			}
		}
	}
	l.Init("gemini", zrt.APIKeyOr(opts.APIKey, "GOOGLE_API_KEY"))
	return l
}

// LLMConfig implements zrt.LLM.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "gemini", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs implements the credential knob source.
func (l *LLM) Knobs() map[string]any {
	k := map[string]any{"include_thoughts": l.IncludeThoughts}
	if l.ThinkingBudget != nil {
		k["thinking_budget"] = *l.ThinkingBudget
	}
	if l.TopP != nil {
		k["top_p"] = *l.TopP
	}
	if l.TopK != nil {
		k["top_k"] = *l.TopK
	}
	if l.PresencePenalty != nil {
		k["presence_penalty"] = *l.PresencePenalty
	}
	if l.FrequencyPenalty != nil {
		k["frequency_penalty"] = *l.FrequencyPenalty
	}
	if l.Seed != nil {
		k["seed"] = *l.Seed
	}
	return k
}

// GeminiLLMExtras implements zrt.GeminiExtrasProvider.
func (l *LLM) GeminiLLMExtras() *zrt.GeminiLLMExtras {
	e := &zrt.GeminiLLMExtras{
		ThinkingBudget:  l.ThinkingBudget,
		IncludeThoughts: l.IncludeThoughts,
		SafetySettings:  l.SafetySettings,
	}
	if l.vertexProjectID != "" {
		e.Vertex = &zrt.VertexInfo{
			ProjectID:          l.vertexProjectID,
			Location:           l.vertexLocation,
			ServiceAccountJSON: l.vertexServiceAccountJSON,
			ServiceAccountPath: l.vertexServiceAccountPath,
		}
	}
	return e
}

func orInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
