// Package navana provides the Navana Namo turn detector.
package navana

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// NamoTurnDetector is the Navana Namo end-of-utterance turn detector.
type NamoTurnDetector struct {
	zrt.BaseEOU
	Language string
}

// NamoTurnDetectorOptions configures a NamoTurnDetector.
type NamoTurnDetectorOptions struct {
	// Threshold is the end-of-utterance confidence cutoff. Defaults to 0.7.
	Threshold float64
	// Language is the spoken language code. Defaults to "en".
	Language string
}

// NewNamoTurnDetector creates a NamoTurnDetector from the given options.
func NewNamoTurnDetector(opts NamoTurnDetectorOptions) *NamoTurnDetector {
	threshold := opts.Threshold
	if threshold == 0 {
		threshold = 0.7
	}
	d := &NamoTurnDetector{Language: zrt.StrOr(opts.Language, "en")}
	d.InitEOU("namo", threshold)
	return d
}

// TurnConfig implements zrt.EOU.
func (d *NamoTurnDetector) TurnConfig() zrt.TurnRuntimeConfig {
	return zrt.TurnRuntimeConfig{Threshold: float32(d.Threshold()), HasThreshold: true, Language: d.Language}
}
