package turn_detector

import (
	"errors"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

func defaultNamoHost() string {
	return zrt.EnvOr("ZRT_INFERENCE_GATEWAY", "inference-gateway.videosdk.live:50053")
}

// NamoTurnDetectorV2 is a Namo v2 end-of-utterance turn detector.
type NamoTurnDetectorV2 struct {
	zrt.BaseEOU
	// ModelID selects the model.
	ModelID string
	// Host is the hosted-inference service endpoint.
	Host string
	// AuthToken authenticates requests to Host.
	AuthToken string
}

// NamoTurnDetectorV2Options configures a NamoTurnDetectorV2.
type NamoTurnDetectorV2Options struct {
	// Threshold is the end-of-utterance confidence cutoff. Defaults to 0.7.
	Threshold float64
	// ModelID selects the model. Defaults to "roberta".
	ModelID string
	// Host is the service endpoint. Defaults to the ZRT_INFERENCE_GATEWAY value.
	Host string
	// AuthToken authenticates requests to Host.
	AuthToken string
}

// NewNamoTurnDetectorV2 creates a NamoTurnDetectorV2 from the given options.
func NewNamoTurnDetectorV2(opts NamoTurnDetectorV2Options) (*NamoTurnDetectorV2, error) {
	threshold := opts.Threshold
	if threshold == 0 {
		threshold = 0.7
	}
	modelID := zrt.StrOr(opts.ModelID, "roberta")
	if modelID == "" {
		return nil, errors.New("model_id is empty")
	}
	d := &NamoTurnDetectorV2{ModelID: modelID, Host: zrt.StrOr(opts.Host, defaultNamoHost()), AuthToken: opts.AuthToken}
	d.InitEOU("namo_v2", threshold)
	return d, nil
}

// TurnConfig implements zrt.EOU.
func (d *NamoTurnDetectorV2) TurnConfig() zrt.TurnRuntimeConfig {
	return zrt.TurnRuntimeConfig{Threshold: float32(d.Threshold()), HasThreshold: true, ModelID: d.ModelID, Host: d.Host, AuthToken: d.AuthToken}
}
