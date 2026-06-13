// Package navana provides the Navana Namo turn detector.
package navana

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// NamoTurnDetector is the Navana Namo turn detector (provider "namo").
type NamoTurnDetector struct {
	zrt.BaseEOU
	Language string
}

// NamoTurnDetectorOptions configures NamoTurnDetector.
type NamoTurnDetectorOptions struct {
	Threshold float64 // default 0.7
	Language  string  // default "en"
}

// NewNamoTurnDetector builds a NamoTurnDetector.
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
