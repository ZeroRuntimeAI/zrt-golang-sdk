// Package google provides the Google Gemini LLM and Google STT/TTS providers.
package google

import (
	"encoding/json"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// LLM is a Google Gemini language model configured for use as an agent's LLM.
type LLM struct {
	zrt.BaseLLM
	// Model is the Gemini model to use.
	Model string
	// Temperature controls randomness.
	Temperature float64
	// MaxOutputTokens caps the response length.
	MaxOutputTokens int
	// ThinkingBudget is the token budget for the model's reasoning.
	ThinkingBudget *int
	// IncludeThoughts includes the model's reasoning in the response.
	IncludeThoughts bool
	// SafetySettings configures content-safety thresholds.
	SafetySettings []zrt.SafetySetting
	// TopP is the nucleus-sampling probability mass. nil uses the provider default.
	TopP *float64
	// TopK limits sampling to the K most likely tokens. nil uses the provider default.
	TopK *int
	// PresencePenalty penalizes tokens that have already appeared. nil uses the provider default.
	PresencePenalty *float64
	// FrequencyPenalty penalizes tokens in proportion to their frequency. nil uses the provider default.
	FrequencyPenalty *float64
	// Seed makes sampling deterministic for a given prompt. nil uses the provider default.
	Seed *int

	vertexProjectID          string
	vertexLocation           string
	vertexServiceAccountJSON string
	vertexServiceAccountPath string
}

// LLMOptions configures a Google Gemini LLM.
type LLMOptions struct {
	// APIKey is the Google API key. If empty, the GOOGLE_API_KEY environment
	// variable is used.
	APIKey string
	// Model is the Gemini model to use. Defaults to "gemini-2.5-flash-lite".
	Model string
	// Temperature controls randomness. nil uses the default (0.7).
	Temperature *float64
	// MaxOutputTokens caps the response length. Defaults to 8192.
	MaxOutputTokens int
	// ThinkingBudget is the token budget for the model's reasoning. Defaults to 0.
	ThinkingBudget *int
	// IncludeThoughts includes the model's reasoning in the response.
	IncludeThoughts bool
	// SafetySettings configures content-safety thresholds.
	SafetySettings []zrt.SafetySetting
	// TopP is the nucleus-sampling probability mass.
	TopP *float64
	// TopK limits sampling to the K most likely tokens.
	TopK *int
	// PresencePenalty penalizes tokens that have already appeared.
	PresencePenalty *float64
	// FrequencyPenalty penalizes tokens in proportion to their frequency.
	FrequencyPenalty *float64
	// Seed makes sampling deterministic for a given prompt.
	Seed *int

	// VertexAI routes requests through Vertex AI instead of the Gemini API.
	VertexAI bool
	// ProjectID is the Google Cloud project ID for Vertex AI.
	ProjectID string
	// Location is the Vertex AI region. Defaults to "us-central1".
	Location string
	// ServiceAccountJSON holds Vertex AI service-account credentials as either a
	// JSON string or a map[string]any.
	ServiceAccountJSON any
	// ServiceAccountPath is the path to a Vertex AI service-account JSON file.
	ServiceAccountPath string
}

// NewLLM creates a Google Gemini LLM from opts, applying defaults for any unset
// fields. When VertexAI is set, requests are routed through Vertex AI using the
// supplied project, location, and service-account credentials.
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
		MaxOutputTokens:  zrt.IntZeroOr(opts.MaxOutputTokens, 8192),
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

// LLMConfig implements zrt.LLM and reports the model configuration.
func (l *LLM) LLMConfig() zrt.LLMRuntimeConfig {
	return zrt.LLMRuntimeConfig{Provider: "gemini", Model: l.Model, Temperature: float32(l.Temperature), MaxOutputTokens: uint32(l.MaxOutputTokens)}
}

// Knobs returns the provider-specific tuning options that are set.
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

// GeminiLLMExtras implements zrt.GeminiExtrasProvider and reports Gemini-specific
// settings, including Vertex AI credentials when configured.
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
