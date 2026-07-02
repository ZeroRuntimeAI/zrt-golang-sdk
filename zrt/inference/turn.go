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

// TurnOptions configures an HTTP-gateway end-of-utterance turn detector.
type TurnOptions struct {
	// Threshold is the end-of-utterance decision threshold. Defaults to 0.7.
	Threshold float64
	// BaseURL overrides the inference gateway endpoint. Falls back to
	// $ZRT_INFERENCE_BASE_URL, then the default gateway.
	BaseURL string
	// AuthToken authenticates against the gateway. Falls back to $ZRT_AUTH_TOKEN.
	AuthToken string
	// Language is the target language hint for turn detection. Optional.
	Language string
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

// NamoTurnDetectorV1 builds an EOU turn detector using the "namo-turn-detector-v1" model.
func NamoTurnDetectorV1(o TurnOptions) zrt.EOU { return newHTTPTurn("namo-turn-detector-v1", o) }

// TurnDetector builds an EOU turn detector using the "turnsense" model.
func TurnDetector(o TurnOptions) zrt.EOU { return newHTTPTurn("turnsense", o) }

// TurnV2Options configures a Namo v2 (Echo) end-of-utterance turn detector.
type TurnV2Options struct {
	// Threshold is the end-of-utterance decision threshold.
	Threshold float64
	// Host overrides the inference gateway endpoint.
	Host string
	// AuthToken authenticates against the gateway. Falls back to $ZRT_AUTH_TOKEN.
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

// TurnV2 is the namespace for Namo v2 (Echo) turn detector constructors.
var TurnV2 turnV2NS

// EchoSmall builds a Namo v2 turn detector using the "echo_small" model.
func (turnV2NS) EchoSmall(o TurnV2Options) zrt.EOU { return newTurnV2("echo_small", o) }

// EchoLarge builds a Namo v2 turn detector using the "echo_large" model.
func (turnV2NS) EchoLarge(o TurnV2Options) zrt.EOU { return newTurnV2("echo_large", o) }
