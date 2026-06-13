package turn_detector

import (
	"fmt"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

func defaultNamoHost() string {
	return zrt.EnvOr("ZRT_INFERENCE_GATEWAY", "inference-gateway.videosdk.live:50053")
}

// NamoTurnDetectorV2 is the Namo v2 turn detector.
type NamoTurnDetectorV2 struct {
	zrt.BaseEOU
	ModelID   string
	Host      string
	AuthToken string
}

// NamoTurnDetectorV2Options configures NamoTurnDetectorV2.
type NamoTurnDetectorV2Options struct {
	Threshold float64 // default 0.7
	ModelID   string  // default "roberta"
	Host      string  // default inference gateway
	AuthToken string
}

// NewNamoTurnDetectorV2 builds a NamoTurnDetectorV2.
func NewNamoTurnDetectorV2(opts NamoTurnDetectorV2Options) (*NamoTurnDetectorV2, error) {
	threshold := opts.Threshold
	if threshold == 0 {
		threshold = 0.7
	}
	modelID := zrt.StrOr(opts.ModelID, "roberta")
	if modelID == "" {
		return nil, fmt.Errorf("NamoTurnDetectorV2: model_id is empty")
	}
	d := &NamoTurnDetectorV2{ModelID: modelID, Host: zrt.StrOr(opts.Host, defaultNamoHost()), AuthToken: opts.AuthToken}
	d.InitEOU("namo_v2", threshold)
	return d, nil
}

// TurnConfig implements zrt.EOU.
func (d *NamoTurnDetectorV2) TurnConfig() zrt.TurnRuntimeConfig {
	return zrt.TurnRuntimeConfig{Threshold: float32(d.Threshold()), HasThreshold: true, ModelID: d.ModelID, Host: d.Host, AuthToken: d.AuthToken}
}
