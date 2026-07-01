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
	APIKey                          string
	Model                           string   // default "gemini-3.1-flash-live-preview"
	Voice                           string   // default "Puck"
	ResponseModalities              []string // default ["AUDIO"]
	TopP                            *float64
	TopK                            *int
	MaxOutputTokens                 *int
	PresencePenalty                 *float64
	FrequencyPenalty                *float64
	CandidateCount                  *int
	LanguageCode                    *string
	ThinkingBudget                  *int
	IncludeThoughts                 bool
	VADStartSensitivity             *string
	VADEndSensitivity               *string
	VADPrefixPaddingMS              *int
	VADSilenceDurationMS            *int
	ContextCompressionTriggerTokens *int
	SessionResumptionHandle         *string
	EnableInputTranscription        *bool
	EnableOutputTranscription       *bool

	VertexAI                 bool
	VertexProjectID          string
	VertexLocation           string // default "us-central1"
	VertexServiceAccountJSON any    // string or map
	VertexServiceAccountPath string
}

// Realtime is a configured Gemini Live speech-to-speech model.
type Realtime struct {
	zrt.BaseRealtime
	Model      string
	Voice      string
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
