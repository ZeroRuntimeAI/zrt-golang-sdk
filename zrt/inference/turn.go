package inference

import (
	"os"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/turn_detector"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

const defaultHTTPTurnBase = "https://inference-gateway.videosdk.live"

func resolveHTTPTurnBase(baseURL string) string {
	if baseURL != "" {
		return baseURL
	}
	if env := os.Getenv("ZRT_INFERENCE_BASE_URL"); env != "" {
		return env
	}
	return defaultHTTPTurnBase
}

type httpTurn struct {
	zrt.BaseEOU
	ModelID   string
	BaseURL   string
	AuthToken string
}

func (t *httpTurn) TurnConfig() zrt.TurnRuntimeConfig {
	return zrt.TurnRuntimeConfig{
		Threshold:    float32(t.Threshold()),
		HasThreshold: true,
		ModelID:      t.ModelID,
		Host:         t.BaseURL,
		AuthToken:    t.AuthToken,
	}
}

type TurnOptions struct {
	Threshold float64
	BaseURL   string
	AuthToken string
	Language  string
}

func newHTTPTurn(modelID string, o TurnOptions) *httpTurn {
	threshold := o.Threshold
	if threshold == 0 {
		threshold = 0.7
	}
	authToken := o.AuthToken
	if authToken == "" {
		authToken = os.Getenv("ZRT_AUTH_TOKEN")
	}
	t := &httpTurn{
		ModelID:   modelID,
		BaseURL:   resolveHTTPTurnBase(o.BaseURL),
		AuthToken: authToken,
	}
	t.InitEOU("turn", threshold)
	return t
}

func NamoTurnDetectorV1(o TurnOptions) zrt.EOU { return newHTTPTurn("namo-turn-detector-v1", o) }

func TurnDetector(o TurnOptions) zrt.EOU { return newHTTPTurn("turnsense", o) }

type TurnV2Options struct {
	Threshold float64
	Host      string
	AuthToken string
}

func newTurnV2(modelID string, o TurnV2Options) zrt.EOU {
	authToken := o.AuthToken
	if authToken == "" {
		authToken = os.Getenv("ZRT_AUTH_TOKEN")
	}
	d, _ := turn_detector.NewNamoTurnDetectorV2(turn_detector.NamoTurnDetectorV2Options{
		Threshold: o.Threshold,
		ModelID:   modelID,
		Host:      o.Host,
		AuthToken: authToken,
	})
	return d
}

type turnV2NS struct{}

var TurnV2 turnV2NS

func (turnV2NS) EchoSmall(o TurnV2Options) zrt.EOU { return newTurnV2("echo_small", o) }

func (turnV2NS) EchoLarge(o TurnV2Options) zrt.EOU { return newTurnV2("echo_large", o) }
