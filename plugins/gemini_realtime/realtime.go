// Package gemini_realtime provides the Gemini Live speech-to-speech model.
package gemini_realtime

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// RealtimeOptions configures a Gemini Live Realtime model. The zero value is
// valid: empty fields fall back to the defaults noted below, and nil pointer
// fields leave the corresponding setting unset.
type RealtimeOptions struct {
	// APIKey authenticates with Gemini. Overrides the GOOGLE_API_KEY
	// environment variable.
	APIKey             string
	Model              string   // default "gemini-3.1-flash-live-preview"
	Voice              string   // default "Puck"
	ResponseModalities []string // default ["AUDIO"]
	// TopP is the nucleus sampling probability. nil = provider default.
	TopP *float64
	// TopK limits sampling to the K most likely tokens. nil = provider default.
	TopK *int
	// MaxOutputTokens caps the number of tokens generated per response. nil = provider default.
	MaxOutputTokens *int
	// PresencePenalty penalizes tokens that have already appeared. nil = provider default.
	PresencePenalty *float64
	// FrequencyPenalty penalizes tokens by their frequency so far. nil = provider default.
	FrequencyPenalty *float64
	// CandidateCount is the number of response candidates to generate. nil = provider default.
	CandidateCount *int
	// LanguageCode is the BCP-47 code of the output language. nil = provider default.
	LanguageCode *string
	// ThinkingBudget is the token budget allotted to model reasoning. nil = provider default.
	ThinkingBudget *int
	// IncludeThoughts includes the model's thinking traces in output. Defaults to false.
	IncludeThoughts bool
	// VADStartSensitivity tunes how readily VAD detects speech onset. nil = provider default.
	VADStartSensitivity *string
	// VADEndSensitivity tunes how readily VAD detects speech end. nil = provider default.
	VADEndSensitivity *string
	// VADPrefixPaddingMS is the audio retained before detected speech, in ms. nil = provider default.
	VADPrefixPaddingMS *int
	// VADSilenceDurationMS is the trailing silence that ends a turn, in ms. nil = provider default.
	VADSilenceDurationMS *int
	// ContextCompressionTriggerTokens is the context size that triggers compression. nil = provider default.
	ContextCompressionTriggerTokens *int
	// SessionResumptionHandle resumes a prior session from its handle. nil = start a new session.
	SessionResumptionHandle *string
	// EnableInputTranscription transcribes the user's speech. nil = enabled (default true).
	EnableInputTranscription *bool
	// EnableOutputTranscription transcribes the model's speech. nil = enabled (default true).
	EnableOutputTranscription *bool

	// VertexAI routes the session through Vertex AI instead of the Gemini API. Defaults to false.
	VertexAI bool
	// VertexProjectID is the Google Cloud project id used when VertexAI is set.
	VertexProjectID string
	VertexLocation  string // default "us-central1"
	// VertexServiceAccountJSON holds service-account credentials as a JSON string or map.
	VertexServiceAccountJSON any // string or map
	// VertexServiceAccountPath is the path to a service-account JSON file. Falls back to GOOGLE_APPLICATION_CREDENTIALS.
	VertexServiceAccountPath string
}

// Realtime is a configured Gemini Live speech-to-speech model.
type Realtime struct {
	zrt.BaseRealtime
	// Model is the resolved Gemini Live model id.
	Model string
	// Voice is the resolved output voice name.
	Voice string
	// Modalities is the resolved list of response modalities.
	Modalities []string
	params     map[string]string
	vertex     *zrt.VertexInfo
}

// NewRealtime returns a Gemini Live Realtime model configured from opts.
func NewRealtime(opts RealtimeOptions) *Realtime {
	modalities := opts.ResponseModalities
	if len(modalities) == 0 {
		modalities = []string{"AUDIO"}
	}
	params := map[string]string{}
	putFloat := func(k string, p *float64) {
		if p != nil {
			params[k] = zrt.FloatStr(*p)
		}
	}
	putInt := func(k string, p *int) {
		if p != nil {
			params[k] = strconv.Itoa(*p)
		}
	}
	putStr := func(k string, p *string) {
		if p != nil {
			params[k] = *p
		}
	}
	putFloat("top_p", opts.TopP)
	putInt("top_k", opts.TopK)
	putInt("max_output_tokens", opts.MaxOutputTokens)
	putInt("candidate_count", opts.CandidateCount)
	putFloat("presence_penalty", opts.PresencePenalty)
	putFloat("frequency_penalty", opts.FrequencyPenalty)
	putStr("language_code", opts.LanguageCode)
	putInt("thinking_budget", opts.ThinkingBudget)
	if opts.IncludeThoughts {
		params["include_thoughts"] = "true"
	} else {
		params["include_thoughts"] = "false"
	}
	putStr("vad_start_sensitivity", opts.VADStartSensitivity)
	putStr("vad_end_sensitivity", opts.VADEndSensitivity)
	putInt("vad_prefix_padding_ms", opts.VADPrefixPaddingMS)
	putInt("vad_silence_duration_ms", opts.VADSilenceDurationMS)
	putInt("context_compression_trigger_tokens", opts.ContextCompressionTriggerTokens)
	params["enable_input_transcription"] = zrt.BoolStr(zrt.BoolOr(opts.EnableInputTranscription, true))
	params["enable_output_transcription"] = zrt.BoolStr(zrt.BoolOr(opts.EnableOutputTranscription, true))
	if opts.SessionResumptionHandle != nil {
		params["session_resumption_handle"] = *opts.SessionResumptionHandle
	}
	var vertex *zrt.VertexInfo
	if opts.VertexAI && opts.VertexProjectID != "" {
		params["vertex_project_id"] = opts.VertexProjectID
		loc := zrt.StrOr(opts.VertexLocation, "us-central1")
		params["vertex_location"] = loc
		saJSON := ""
		switch v := opts.VertexServiceAccountJSON.(type) {
		case string:
			saJSON = v
		case map[string]any:
			b, _ := json.Marshal(v)
			saJSON = string(b)
		}
		saPath := opts.VertexServiceAccountPath
		if saPath == "" {
			saPath = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		}
		if saJSON != "" {
			params["vertex_service_account_json"] = saJSON
		} else if saPath != "" {
			params["vertex_service_account_path"] = saPath
		}
		vertex = &zrt.VertexInfo{ProjectID: opts.VertexProjectID, Location: loc, ServiceAccountJSON: saJSON, ServiceAccountPath: saPath}
	}
	r := &Realtime{Model: zrt.StrOr(opts.Model, "gemini-3.1-flash-live-preview"), Voice: zrt.StrOr(opts.Voice, "Puck"), Modalities: modalities, params: params, vertex: vertex}
	r.Init("gemini_live", zrt.APIKeyOr(opts.APIKey, "GOOGLE_API_KEY"))
	return r
}

// RealtimeInfo implements zrt.RealtimeModel.
func (r *Realtime) RealtimeInfo() zrt.RealtimeInfo {
	return zrt.RealtimeInfo{Model: r.Model, Voice: r.Voice, Params: r.params, ResponseModalities: r.Modalities, Vertex: r.vertex}
}

// boolOr returns *p when set, else def. Used for tri-state options whose
// default is true (so a caller must pass &false to disable).
func boolOr(p *bool, def bool) bool {
	if p != nil {
		return *p
	}
	return def
}
