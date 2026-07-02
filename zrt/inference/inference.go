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

// DeepgramSTTOptions configures a Deepgram STT routed through the ZRT
// inference gateway.
type DeepgramSTTOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the Deepgram STT model id. Defaults to "nova-2".
	Model string
	// Language is the recognition language. Defaults to "en-US".
	Language string
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// Endpointing is the silence duration in ms used to detect end of speech.
	// Defaults to 50.
	Endpointing int
	// EagerEOTThreshold is the confidence threshold for eager end-of-turn
	// detection. Optional; nil defaults to 0.6.
	EagerEOTThreshold *float64
	// EOTThreshold is the confidence threshold for end-of-turn detection.
	// Optional; nil defaults to 0.8.
	EOTThreshold *float64
	// EOTTimeoutMs is the end-of-turn timeout in ms. Optional; nil defaults to 7000.
	EOTTimeoutMs *int
	// Keyterm is a list of key terms to boost during recognition.
	Keyterm []string
}

// DeepgramSTT builds a Deepgram STT configured to run through the ZRT
// inference gateway, applying defaults for any unset options.
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

// GoogleSTTOptions configures a Google STT routed through the ZRT
// inference gateway.
type GoogleSTTOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the Google STT model id. Defaults to "chirp_3".
	Model string
	// Language is the primary recognition language. Defaults to "en-US".
	Language string
	// Languages is the set of recognition languages. Defaults to [Language].
	Languages []string
	// Location is the Google Cloud region. Defaults to "us-central1".
	Location string
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the output audio sample rate in Hz. Defaults to 16000.
	OutputSampleRate int
}

// GoogleSTT builds a Google STT configured to run through the ZRT inference
// gateway, applying defaults for any unset options.
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

// SarvamAISTTOptions configures a SarvamAI STT routed through the ZRT
// inference gateway.
type SarvamAISTTOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the SarvamAI STT model id. Defaults to "saaras:v3".
	Model string
	// Language is the recognition language. Defaults to "en-IN".
	Language string
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// OutputSampleRate is the output audio sample rate in Hz. Defaults to 16000.
	OutputSampleRate int
}

// SarvamAISTT builds a SarvamAI STT configured to run through the ZRT
// inference gateway, applying defaults for any unset options.
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

// AssemblyAISTTOptions configures an AssemblyAI STT routed through the ZRT
// inference gateway.
type AssemblyAISTTOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// SpeechModel is the AssemblyAI speech model id.
	// Defaults to "universal-streaming-english".
	SpeechModel string
	// Region is the AssemblyAI service region. Defaults to "US".
	Region string
	// InputSampleRate is the input audio sample rate in Hz. Defaults to 48000.
	InputSampleRate int
	// TargetSampleRate is the target output sample rate in Hz. Defaults to 16000.
	TargetSampleRate int
	// FormatTurns enables formatted turn output. Optional; nil defaults to true.
	FormatTurns *bool
	// KeytermsPrompt is a list of key terms to boost during recognition.
	KeytermsPrompt []string
	// EndOfTurnConfidenceThreshold is the confidence threshold for end-of-turn
	// detection. Optional; nil defaults to 0.5.
	EndOfTurnConfidenceThreshold *float64
	// MinEndOfTurnSilenceWhenConfident is the minimum end-of-turn silence in ms
	// when confident. Optional; nil defaults to 800.
	MinEndOfTurnSilenceWhenConfident *int
	// MaxTurnSilence is the maximum silence in ms within a turn. Optional; nil
	// defaults to 2000.
	MaxTurnSilence *int
	// LanguageDetection enables automatic language detection. Optional; nil
	// defaults to true.
	LanguageDetection *bool
}

// AssemblyAISTT builds an AssemblyAI STT configured to run through the ZRT
// inference gateway, applying defaults for any unset options.
func AssemblyAISTT(o AssemblyAISTTOptions) *assemblyai.STT {
	model := strOr(o.SpeechModel, "universal-streaming-english")
	region := strOr(o.Region, "US")
	inRate := intOr(o.InputSampleRate, 48000)
	outRate := intOr(o.TargetSampleRate, 16000)
	s := assemblyai.NewSTT(assemblyai.STTOptions{Model: model})
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

// CartesiaSTTOptions configures a Cartesia STT routed through the ZRT
// inference gateway.
type CartesiaSTTOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the Cartesia STT model id. Defaults to "ink-2".
	Model string
	// Language is the recognition language. Defaults to "en".
	Language string
}

// CartesiaSTT builds a Cartesia STT configured to run through the ZRT
// inference gateway, applying defaults for any unset options.
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

// GoogleLLMOptions configures a Google LLM routed through the ZRT inference
// gateway.
type GoogleLLMOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the Google LLM model id.
	Model string
	// Temperature is the sampling temperature. Optional; nil uses the model default.
	Temperature *float64
	// MaxOutputTokens caps the number of generated tokens.
	MaxOutputTokens int
}

// GoogleLLM builds a Google LLM configured to run through the ZRT inference
// gateway.
func GoogleLLM(o GoogleLLMOptions) *google.LLM {
	l := google.NewLLM(google.LLMOptions{Model: o.Model, Temperature: o.Temperature, MaxOutputTokens: o.MaxOutputTokens})
	mark(l, o.BaseURL, map[string]any{
		"model":             l.Model,
		"temperature":       l.Temperature,
		"max_output_tokens": l.MaxOutputTokens,
	})
	return l
}

// SarvamAILLMOptions configures a SarvamAI LLM routed through the ZRT
// inference gateway.
type SarvamAILLMOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the SarvamAI LLM model id.
	Model string
	// Temperature is the sampling temperature. Optional; nil uses the model default.
	Temperature *float64
	// MaxCompletionTokens caps the number of generated tokens. Optional; nil
	// uses the model default.
	MaxCompletionTokens *int
}

// SarvamAILLM builds a SarvamAI LLM configured to run through the ZRT
// inference gateway.
func SarvamAILLM(o SarvamAILLMOptions) *sarvamai.LLM {
	l := sarvamai.NewLLM(sarvamai.LLMOptions{Model: o.Model, Temperature: o.Temperature, MaxCompletionTokens: o.MaxCompletionTokens})
	mark(l, o.BaseURL, map[string]any{
		"model":             l.Model,
		"temperature":       l.Temperature,
		"max_output_tokens": l.MaxOutputTokens,
	})
	return l
}

// CartesiaTTSOptions configures a Cartesia TTS routed through the ZRT
// inference gateway.
type CartesiaTTSOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the Cartesia TTS model id. Defaults to "sonic-2".
	Model string
	// Language is the synthesis language. Defaults to "en".
	Language string
	// Voice is the voice id. Defaults to the Cartesia gateway default voice.
	Voice string
	// SampleRate is the output audio sample rate in Hz. Defaults to 24000.
	SampleRate int
}

// CartesiaTTS builds a Cartesia TTS configured to run through the ZRT
// inference gateway, applying defaults for any unset options.
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

// GoogleTTSOptions configures a Google TTS routed through the ZRT inference
// gateway.
type GoogleTTSOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the Google TTS model id. Defaults to "Chirp3-HD".
	Model string
	// Voice is the voice name. Defaults to "Achernar".
	Voice string
	// Language is the synthesis language code. Defaults to "en-US".
	Language string
	// SampleRate is the output audio sample rate in Hz. Defaults to 24000.
	SampleRate int
}

// GoogleTTS builds a Google TTS configured to run through the ZRT inference
// gateway, applying defaults for any unset options.
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

// SarvamAITTSOptions configures a SarvamAI TTS routed through the ZRT
// inference gateway.
type SarvamAITTSOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the SarvamAI TTS model id. Defaults to "bulbul:v3".
	Model string
	// Speaker is the speaker id. Defaults to "shubh".
	Speaker string
	// Language is the synthesis language. Defaults to "en-IN".
	Language string
	// SampleRate is the output audio sample rate in Hz. Defaults to 24000.
	SampleRate int
}

// SarvamAITTS builds a SarvamAI TTS configured to run through the ZRT
// inference gateway, applying defaults for any unset options.
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

// DeepgramTTSOptions configures a Deepgram TTS routed through the ZRT
// inference gateway.
type DeepgramTTSOptions struct {
	// BaseURL is the inference gateway base URL. Empty falls back to the
	// ZRT_INFERENCE_BASE_URL environment variable.
	BaseURL string
	// Model is the Deepgram TTS model id. Defaults to "aura-2".
	Model string
	// Voice is the voice id. Defaults to "asteria".
	Voice string
	// Language is the synthesis language. Defaults to "en".
	Language string
	// SampleRate is the output audio sample rate in Hz. Defaults to 24000.
	SampleRate int
}

// DeepgramTTS builds a Deepgram TTS configured to run through the ZRT
// inference gateway, applying defaults for any unset options.
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
