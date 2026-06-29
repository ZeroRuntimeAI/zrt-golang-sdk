package inference

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/assemblyai"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/cartesia"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/deepgram"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/google"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/sarvamai"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

const cartesiaGatewayDefaultVoice = "faf0731e-dfb9-4cfc-8119-259a79b27e12"

func resolveBaseURL(baseURL string) string {
	if baseURL != "" {
		return baseURL
	}
	return os.Getenv("ZRT_INFERENCE_BASE_URL")
}

type inferenceProvider interface {
	SetInference(baseURL, location string)
	SetInferenceConfig(cfg map[string]any)
}

func mark(p inferenceProvider, baseURL string, cfg map[string]any) {
	p.SetInference(resolveBaseURL(baseURL), "")
	p.SetInferenceConfig(clean(cfg))
}

func clean(d map[string]any) map[string]any {
	out := make(map[string]any, len(d))
	for k, v := range d {
		if v == nil {
			continue
		}
		out[k] = v
	}
	return out
}

func strOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func intOr(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
func floatPtrOr(p *float64, def float64) float64 {
	if p == nil {
		return def
	}
	return *p
}

func intPtrOr(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

func boolPtrOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

type DeepgramSTTOptions struct {
	BaseURL           string
	Model             string
	Language          string
	InputSampleRate   int
	Endpointing       int
	EagerEOTThreshold *float64
	EOTThreshold      *float64
	EOTTimeoutMs      *int
	Keyterm           []string
}

func DeepgramSTT(o DeepgramSTTOptions) *deepgram.STT {
	model := strOr(o.Model, "nova-2")
	language := strOr(o.Language, "en-US")
	rate := intOr(o.InputSampleRate, 48000)
	ep := intOr(o.Endpointing, 50)
	s := deepgram.NewSTT(deepgram.STTOptions{Model: model, Language: language, SampleRate: rate, Endpointing: &ep})
	cfg := map[string]any{
		"model":               model,
		"language":            language,
		"input_sample_rate":   rate,
		"endpointing":         ep,
		"interim_results":     true,
		"punctuate":           true,
		"smart_format":        true,
		"eager_eot_threshold": floatPtrOr(o.EagerEOTThreshold, 0.6),
		"eot_threshold":       floatPtrOr(o.EOTThreshold, 0.8),
		"eot_timeout_ms":      intPtrOr(o.EOTTimeoutMs, 7000),
	}
	if len(o.Keyterm) > 0 {
		cfg["keyterm"] = o.Keyterm
	}
	mark(s, o.BaseURL, cfg)
	return s
}

type GoogleSTTOptions struct {
	BaseURL          string
	Model            string
	Language         string
	Languages        []string
	Location         string
	InputSampleRate  int
	OutputSampleRate int
}

func GoogleSTT(o GoogleSTTOptions) *google.STT {
	model := strOr(o.Model, "chirp_3")
	language := strOr(o.Language, "en-US")
	langs := o.Languages
	if len(langs) == 0 {
		langs = []string{language}
	}
	inRate := intOr(o.InputSampleRate, 48000)
	outRate := intOr(o.OutputSampleRate, 16000)
	location := strOr(o.Location, "us-central1")
	s := google.NewSTT(google.STTOptions{Model: model, Language: language, Languages: langs})
	mark(s, o.BaseURL, map[string]any{
		"model":              model,
		"language":           language,
		"languages":          langs,
		"input_sample_rate":  inRate,
		"output_sample_rate": outRate,
		"interim_results":    true,
		"punctuate":          true,
		"location":           location,
	})
	return s
}

type SarvamAISTTOptions struct {
	BaseURL          string
	Model            string
	Language         string
	InputSampleRate  int
	OutputSampleRate int
}

func SarvamAISTT(o SarvamAISTTOptions) *sarvamai.STT {
	model := strOr(o.Model, "saaras:v3")
	language := strOr(o.Language, "en-IN")
	inRate := intOr(o.InputSampleRate, 48000)
	outRate := intOr(o.OutputSampleRate, 16000)
	s := sarvamai.NewSTT(sarvamai.STTOptions{Model: model, Language: language, InputSampleRate: inRate, OutputSampleRate: outRate})
	mark(s, o.BaseURL, map[string]any{
		"model":              model,
		"language":           language,
		"input_sample_rate":  inRate,
		"output_sample_rate": outRate,
	})
	return s
}

type AssemblyAISTTOptions struct {
	BaseURL                          string
	SpeechModel                      string
	Region                           string
	InputSampleRate                  int
	TargetSampleRate                 int
	FormatTurns                      *bool
	KeytermsPrompt                   []string
	EndOfTurnConfidenceThreshold     *float64
	MinEndOfTurnSilenceWhenConfident *int
	MaxTurnSilence                   *int
	LanguageDetection                *bool
}

func AssemblyAISTT(o AssemblyAISTTOptions) *assemblyai.STT {
	model := strOr(o.SpeechModel, "universal-streaming-english")
	region := strOr(o.Region, "US")
	inRate := intOr(o.InputSampleRate, 48000)
	outRate := intOr(o.TargetSampleRate, 16000)
	s := assemblyai.NewSTT(assemblyai.STTOptions{SpeechModel: model, Region: region})
	cfg := map[string]any{
		"model":                                  model,
		"language":                               "en-US",
		"input_sample_rate":                      inRate,
		"output_sample_rate":                     outRate,
		"format_turns":                           boolPtrOr(o.FormatTurns, true),
		"end_of_turn_confidence_threshold":       floatPtrOr(o.EndOfTurnConfidenceThreshold, 0.5),
		"min_end_of_turn_silence_when_confident": intPtrOr(o.MinEndOfTurnSilenceWhenConfident, 800),
		"max_turn_silence":                       intPtrOr(o.MaxTurnSilence, 2000),
		"language_detection":                     boolPtrOr(o.LanguageDetection, true),
		"region":                                 region,
	}
	if len(o.KeytermsPrompt) > 0 {
		cfg["keyterms_prompt"] = o.KeytermsPrompt
	}
	mark(s, o.BaseURL, cfg)
	return s
}

type CartesiaSTTOptions struct {
	BaseURL  string
	Model    string
	Language string
}

func CartesiaSTT(o CartesiaSTTOptions) *cartesia.STT {
	model := strOr(o.Model, "ink-2")
	language := strOr(o.Language, "en")
	s := cartesia.NewSTT(cartesia.STTOptions{Model: model, Language: language})
	mark(s, o.BaseURL, map[string]any{
		"model":              model,
		"language":           language,
		"input_sample_rate":  48000,
		"output_sample_rate": 16000,
	})
	return s
}

type GoogleLLMOptions struct {
	BaseURL         string
	Model           string
	Temperature     *float64
	MaxOutputTokens int
}

func GoogleLLM(o GoogleLLMOptions) *google.LLM {
	l := google.NewLLM(google.LLMOptions{Model: o.Model, Temperature: o.Temperature, MaxOutputTokens: o.MaxOutputTokens})
	mark(l, o.BaseURL, map[string]any{
		"model":             l.Model,
		"temperature":       l.Temperature,
		"max_output_tokens": l.MaxOutputTokens,
	})
	return l
}

type SarvamAILLMOptions struct {
	BaseURL             string
	Model               string
	Temperature         *float64
	MaxCompletionTokens *int
}

func SarvamAILLM(o SarvamAILLMOptions) *sarvamai.LLM {
	l := sarvamai.NewLLM(sarvamai.LLMOptions{Model: o.Model, Temperature: o.Temperature, MaxCompletionTokens: o.MaxCompletionTokens})
	mark(l, o.BaseURL, map[string]any{
		"model":             l.Model,
		"temperature":       l.Temperature,
		"max_output_tokens": l.MaxOutputTokens,
	})
	return l
}

type CartesiaTTSOptions struct {
	BaseURL    string
	Model      string
	Language   string
	Voice      string
	SampleRate int
}

func CartesiaTTS(o CartesiaTTSOptions) *cartesia.TTS {
	model := strOr(o.Model, "sonic-2")
	language := strOr(o.Language, "en")
	voice := strOr(o.Voice, cartesiaGatewayDefaultVoice)
	rate := intOr(o.SampleRate, 24000)
	t := cartesia.NewTTS(cartesia.TTSOptions{Model: model, Language: language, Voice: voice})
	mark(t, o.BaseURL, map[string]any{
		"model":       model,
		"language":    language,
		"voice":       voice,
		"sample_rate": rate,
	})
	return t
}

type GoogleTTSOptions struct {
	BaseURL    string
	Model      string
	Voice      string
	Language   string
	SampleRate int
}

func GoogleTTS(o GoogleTTSOptions) *google.TTS {
	model := strOr(o.Model, "Chirp3-HD")
	voice := strOr(o.Voice, "Achernar")
	language := strOr(o.Language, "en-US")
	rate := intOr(o.SampleRate, 24000)
	t := google.NewTTS(google.TTSOptions{Model: model, Voice: voice, LanguageCode: language})
	mark(t, o.BaseURL, map[string]any{
		"model":         model,
		"voice_name":    voice,
		"language_code": language,
		"sample_rate":   rate,
		"model_id":      model,
	})
	return t
}

type SarvamAITTSOptions struct {
	BaseURL    string
	Model      string
	Speaker    string
	Language   string
	SampleRate int
}

func SarvamAITTS(o SarvamAITTSOptions) *sarvamai.TTS {
	model := strOr(o.Model, "bulbul:v3")
	speaker := strOr(o.Speaker, "shubh")
	language := strOr(o.Language, "en-IN")
	rate := intOr(o.SampleRate, 24000)
	t := sarvamai.NewTTS(sarvamai.TTSOptions{Model: model, Speaker: speaker, Language: language})
	mark(t, o.BaseURL, map[string]any{
		"model":       model,
		"language":    language,
		"speaker":     speaker,
		"sample_rate": rate,
	})
	return t
}

type DeepgramTTSOptions struct {
	BaseURL    string
	Model      string
	Voice      string
	Language   string
	SampleRate int
}

func DeepgramTTS(o DeepgramTTSOptions) *deepgram.TTS {
	model := strOr(o.Model, "aura-2")
	voice := strOr(o.Voice, "asteria")
	language := strOr(o.Language, "en")
	rate := intOr(o.SampleRate, 24000)
	t := deepgram.NewTTS(deepgram.TTSOptions{Model: model, Voice: voice, Language: language})
	mark(t, o.BaseURL, map[string]any{
		"model":       model,
		"voice":       voice,
		"voice_id":    voice,
		"language":    language,
		"sample_rate": rate,
	})
	return t
}

var (
	_ inferenceProvider = (*deepgram.STT)(nil)
	_ inferenceProvider = (*zrt.BaseSTT)(nil)
)
