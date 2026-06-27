// Package turn_detector provides turn-detection (end-of-utterance) providers.
package turn_detector

import (
	"log"
	"strings"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// inferenceModels maps each hosted TurnDetector model to its model identifier.
// A model absent from this map is the on-device "namo" detector.
var inferenceModels = map[string]string{
	ModelNamoInference: "roberta",
	ModelEchoSmall:     "echo_small",
	ModelEchoLarge:     "echo_large",
}

// TurnDetector is the end-of-turn detector — it decides when the caller has
// finished speaking. A single type selected by TurnDetectorOptions.Model:
// "namo" (the default), "namo-inference", "echo-small", or "echo-large". Which
// extra options apply depends on the model (see TurnDetectorOptions).
type TurnDetector struct {
	zrt.BaseEOU
	model       string
	modelID     string
	host        string
	authToken   string
	baseURL     string
	isInference bool
}

// TurnDetectorOptions configures a TurnDetector. The zero value selects the
// default "namo" detector at threshold 0.7.
type TurnDetectorOptions struct {
	// Model selects the detector: "namo" (the default), "namo-inference",
	// "echo-small", or "echo-large". Empty is treated as "namo".
	Model TurnDetectorModel
	// Threshold is the end-of-turn probability cutoff, 0.0–1.0. Applies to every
	// model. Defaults to 0.7 when 0.
	Threshold float64
	// Language is a BCP-47 language hint. Only used by the "namo" model.
	Language string
	// Host is the hosted-inference service address. Hosted models only.
	Host string
	// AuthToken authenticates with the hosted-inference service. Hosted models
	// only.
	AuthToken string
	// BaseURL is the hosted-inference service base URL. Hosted models only.
	BaseURL string
}

// NewTurnDetector builds a TurnDetector from opts. A zero-value
// TurnDetectorOptions yields the default "namo" detector at threshold 0.7.
// Options that do not apply to the selected model are ignored (with a warning).
func NewTurnDetector(opts TurnDetectorOptions) *TurnDetector {
	model := zrt.StrOr(opts.Model, ModelNamo)
	threshold := opts.Threshold
	if threshold == 0 {
		threshold = 0.7
	}

	modelID, isInference := inferenceModels[model]
	d := &TurnDetector{model: model, isInference: isInference, baseURL: opts.BaseURL}

	if isInference && opts.Language != "" {
		log.Printf("turn_detector: TurnDetector: Language is ignored for model %q — it only applies to model %q.", model, ModelNamo)
	}
	if !isInference && (opts.Host != "" || opts.AuthToken != "" || opts.BaseURL != "") {
		log.Printf("turn_detector: TurnDetector: Host/AuthToken/BaseURL only apply to inference models (%q, %q, %q) — ignored for model %q.",
			ModelNamoInference, ModelEchoSmall, ModelEchoLarge, model)
	}

	if isInference {
		d.modelID = modelID
		d.host = zrt.StrOr(opts.Host, defaultNamoHost())
		d.authToken = opts.AuthToken
		d.InitEOU("namo_v2", threshold)
	} else {
		d.modelID = strings.ToLower(opts.Language)
		d.InitEOU("namo", threshold)
	}
	return d
}

// Model returns the selected model name.
func (d *TurnDetector) Model() string { return d.model }

// TurnConfig implements zrt.EOU.
func (d *TurnDetector) TurnConfig() zrt.TurnRuntimeConfig {
	cfg := zrt.TurnRuntimeConfig{
		Threshold:    float32(d.Threshold()),
		HasThreshold: true,
		ModelID:      d.modelID,
	}
	if d.isInference {
		cfg.Host = d.host
		cfg.AuthToken = d.authToken
	}
	return cfg
}

// TurnDetectorV2 is the TurnSense v2 detector.
type TurnDetectorV2 struct {
	zrt.BaseEOU
}

// NewTurnDetectorV2 builds a TurnDetectorV2.
func NewTurnDetectorV2(threshold float64) *TurnDetectorV2 {
	if threshold == 0 {
		threshold = 0.7
	}
	d := &TurnDetectorV2{}
	d.InitEOU("turnsense_v2", threshold)
	return d
}

// TurnConfig implements zrt.EOU.
func (d *TurnDetectorV2) TurnConfig() zrt.TurnRuntimeConfig {
	return zrt.TurnRuntimeConfig{Threshold: float32(d.Threshold()), HasThreshold: true}
}

// NamoTurnDetectorV1 is the Namo v1 turn detector (model from language).
type NamoTurnDetectorV1 struct {
	zrt.BaseEOU
	Language string
}

// NewNamoTurnDetectorV1 builds a NamoTurnDetectorV1.
func NewNamoTurnDetectorV1(language string, threshold float64) *NamoTurnDetectorV1 {
	if threshold == 0 {
		threshold = 0.7
	}
	d := &NamoTurnDetectorV1{Language: language}
	d.InitEOU("namo", threshold)
	return d
}

// TurnConfig implements zrt.EOU.
func (d *NamoTurnDetectorV1) TurnConfig() zrt.TurnRuntimeConfig {
	return zrt.TurnRuntimeConfig{Threshold: float32(d.Threshold()), HasThreshold: true, ModelID: strings.ToLower(d.Language)}
}

// ZeroRuntimeTurnDetector is an alias of NamoTurnDetectorV1.
type ZeroRuntimeTurnDetector = NamoTurnDetectorV1

// NewZeroRuntimeTurnDetector builds a ZeroRuntimeTurnDetector.
func NewZeroRuntimeTurnDetector(language string, threshold float64) *ZeroRuntimeTurnDetector {
	return NewNamoTurnDetectorV1(language, threshold)
}
