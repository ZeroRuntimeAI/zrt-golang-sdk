// Package turn_detector provides turn-detection (end-of-utterance) providers.
package turn_detector

import (
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// TurnDetector is the Namo turn detector (provider "namo").
type TurnDetector struct {
	zrt.BaseEOU
}

// NewTurnDetector builds a TurnDetector. threshold defaults to 0.7 when 0.
func NewTurnDetector(threshold float64) *TurnDetector {
	if threshold == 0 {
		threshold = 0.7
	}
	d := &TurnDetector{}
	d.InitEOU("namo", threshold)
	return d
}

// TurnConfig implements zrt.EOU.
func (d *TurnDetector) TurnConfig() zrt.TurnRuntimeConfig {
	return zrt.TurnRuntimeConfig{Threshold: float32(d.Threshold()), HasThreshold: true}
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

// ZeroRuntimeTurnDetector is an alias of NamoTurnDetectorV1.
type ZeroRuntimeTurnDetector = NamoTurnDetectorV1

// NewZeroRuntimeTurnDetector builds a ZeroRuntimeTurnDetector.
func NewZeroRuntimeTurnDetector(opts NamoTurnDetectorV1Options) *ZeroRuntimeTurnDetector {
	return NewNamoTurnDetectorV1(opts)
}
