package turn_detector

import "github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"

// NamoTurnDetectorV1 is the Namo v1 turn detector (provider "namo").
type NamoTurnDetectorV1 struct {
	zrt.BaseEOU
	Language string
}

// NamoTurnDetectorV1Options configures NamoTurnDetectorV1.
type NamoTurnDetectorV1Options struct {
	Threshold float64 // default 0.7
	Language  string  // default "en"
}

// NewNamoTurnDetectorV1 builds a NamoTurnDetectorV1.
func NewNamoTurnDetectorV1(opts NamoTurnDetectorV1Options) *NamoTurnDetectorV1 {
	threshold := opts.Threshold
	if threshold == 0 {
		threshold = 0.7
	}
	d := &NamoTurnDetectorV1{Language: zrt.StrOr(opts.Language, "en")}
	d.InitEOU("namo", threshold)
	return d
}

// TurnConfig implements zrt.EOU.
func (d *NamoTurnDetectorV1) TurnConfig() zrt.TurnRuntimeConfig {
	return zrt.TurnRuntimeConfig{Threshold: float32(d.Threshold()), HasThreshold: true, Language: d.Language}
}
