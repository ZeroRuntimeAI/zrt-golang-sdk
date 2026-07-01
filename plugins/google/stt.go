package google

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

type STT struct {
	zrt.BaseSTT
	Model                     string
	Language                  string
	ServiceAccountJSON        string
	ProjectID                 string
	Stream                    bool
	Location                  string
	AudioChannelCount         int
	InterimResults            bool
	Punctuate                 bool
	ProfanityFilter           bool
	EnableSpokenPunctuation   bool
	EnableSpokenEmojis        bool
	EnableWordTimeOffsets     bool
	EnableWordConfidence      bool
	MaxAlternatives           int
	EnableVoiceActivityEvents bool
	SpeechStartTimeout        *float64
	SpeechEndTimeout          *float64
	MinSpeakerCount           *int
	MaxSpeakerCount           *int
	MinConfidenceThreshold    float64
	SampleRate                int
}

type STTOptions struct {
	APIKey                    string
	CredentialsJSON           string
	ServiceAccountPath        string
	ProjectID                 string
	Languages                 []string
	Language                  string
	Model                     string
	Stream                    *bool
	Location                  string
	AudioChannelCount         int
	InterimResults            *bool
	Punctuate                 *bool
	ProfanityFilter           *bool
	EnableSpokenPunctuation   *bool
	EnableSpokenEmojis        *bool
	EnableWordTimeOffsets     *bool
	EnableWordConfidence      *bool
	MaxAlternatives           int
	EnableVoiceActivityEvents *bool
	SpeechStartTimeout        *float64
	SpeechEndTimeout          *float64
	MinSpeakerCount           *int
	MaxSpeakerCount           *int
	MinConfidenceThreshold    *float64
	SampleRate                int
}

func NewSTT(opts STTOptions) *STT {
	lang := opts.Language
	if len(opts.Languages) > 0 {
		lang = strings.Join(opts.Languages, ",")
	}
	lang = zrt.StrOr(lang, "en-US")

	saJSON := resolveServiceAccountJSON(opts.CredentialsJSON, opts.ServiceAccountPath, opts.APIKey)
	projectID := zrt.StrOr(opts.ProjectID, projectIDFromJSON(saJSON))

	s := &STT{
		Model:                     zrt.StrOr(opts.Model, "latest_long"),
		Language:                  lang,
		ServiceAccountJSON:        saJSON,
		ProjectID:                 projectID,
		Stream:                    zrt.BoolOr(opts.Stream, true),
		Location:                  zrt.StrOr(opts.Location, "us"),
		AudioChannelCount:         zrt.IntZeroOr(opts.AudioChannelCount, 1),
		InterimResults:            zrt.BoolOr(opts.InterimResults, true),
		Punctuate:                 zrt.BoolOr(opts.Punctuate, true),
		ProfanityFilter:           zrt.BoolOr(opts.ProfanityFilter, false),
		EnableSpokenPunctuation:   zrt.BoolOr(opts.EnableSpokenPunctuation, false),
		EnableSpokenEmojis:        zrt.BoolOr(opts.EnableSpokenEmojis, false),
		EnableWordTimeOffsets:     zrt.BoolOr(opts.EnableWordTimeOffsets, false),
		EnableWordConfidence:      zrt.BoolOr(opts.EnableWordConfidence, false),
		MaxAlternatives:           zrt.IntZeroOr(opts.MaxAlternatives, 1),
		EnableVoiceActivityEvents: zrt.BoolOr(opts.EnableVoiceActivityEvents, false),
		SpeechStartTimeout:        opts.SpeechStartTimeout,
		SpeechEndTimeout:          opts.SpeechEndTimeout,
		MinSpeakerCount:           opts.MinSpeakerCount,
		MaxSpeakerCount:           opts.MaxSpeakerCount,
		MinConfidenceThreshold:    zrt.FloatOr(opts.MinConfidenceThreshold, 0.0),
		SampleRate:                zrt.IntZeroOr(opts.SampleRate, 48000),
	}
	s.Init("google_stt", zrt.APIKeyOr(opts.APIKey, "GOOGLE_API_KEY"))
	return s
}

func (s *STT) STTConfig() zrt.STTRuntimeConfig {
	return zrt.STTRuntimeConfig{Provider: "google_stt", Model: s.Model, Language: s.Language}
}

func (s *STT) Knobs() map[string]any {
	k := map[string]any{
		"location":                     s.Location,
		"stream":                       s.Stream,
		"audio_channel_count":          s.AudioChannelCount,
		"interim_results":              s.InterimResults,
		"punctuate":                    s.Punctuate,
		"profanity_filter":             s.ProfanityFilter,
		"enable_spoken_punctuation":    s.EnableSpokenPunctuation,
		"enable_spoken_emojis":         s.EnableSpokenEmojis,
		"enable_word_time_offsets":     s.EnableWordTimeOffsets,
		"enable_word_confidence":       s.EnableWordConfidence,
		"max_alternatives":             s.MaxAlternatives,
		"enable_voice_activity_events": s.EnableVoiceActivityEvents,
		"min_confidence_threshold":     s.MinConfidenceThreshold,
	}
	if s.ServiceAccountJSON != "" {
		k["service_account_json"] = s.ServiceAccountJSON
	}
	if s.ProjectID != "" {
		k["project_id"] = s.ProjectID
	}
	if s.SpeechStartTimeout != nil {
		k["speech_start_timeout"] = *s.SpeechStartTimeout
	}
	if s.SpeechEndTimeout != nil {
		k["speech_end_timeout"] = *s.SpeechEndTimeout
	}
	if s.MinSpeakerCount != nil {
		k["min_speaker_count"] = *s.MinSpeakerCount
	}
	if s.MaxSpeakerCount != nil {
		k["max_speaker_count"] = *s.MaxSpeakerCount
	}
	return k
}

func coerceToJSONContent(value string) string {
	if value == "" {
		return ""
	}
	if strings.HasPrefix(strings.TrimSpace(value), "{") {
		return value
	}
	b, err := os.ReadFile(value)
	if err != nil {
		return ""
	}
	return string(b)
}

func resolveServiceAccountJSON(credentialsJSON, serviceAccountPath, apiKey string) string {
	for _, candidate := range []string{credentialsJSON, serviceAccountPath, apiKey, os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"), "credentials.json"} {
		if content := coerceToJSONContent(candidate); content != "" {
			return content
		}
	}
	return ""
}

func projectIDFromJSON(saJSON string) string {
	if saJSON == "" {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(saJSON), &m); err != nil {
		return ""
	}
	if v, ok := m["project_id"].(string); ok {
		return v
	}
	return ""
}
