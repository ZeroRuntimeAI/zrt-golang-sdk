package zrt

import "testing"

type testInferenceSTT struct{ BaseSTT }

func (s *testInferenceSTT) STTConfig() STTRuntimeConfig {
	return STTRuntimeConfig{Provider: "deepgram", Model: "nova-2", Language: "en-US", EndpointingMs: 50}
}

func TestInferenceContract(t *testing.T) {
	t.Setenv("ZRT_AUTH_TOKEN", "test-token")

	stt := &testInferenceSTT{}
	stt.Init("deepgram", "")
	stt.SetInference("", "")
	stt.SetInferenceConfig(map[string]any{"model": "nova-2", "language": "en-US"})

	cfg := buildSTTConfig(stt)
	if cfg.Inference == nil {
		t.Fatal("expected STTProviderConfig.Inference to be non-nil for an inference provider")
	}

	creds := buildCredentials(&Pipeline{STT: stt}, nil, nil)
	if creds["zrt_auth_token"] != "test-token" {
		t.Errorf("expected zrt_auth_token credential, got %q", creds["zrt_auth_token"])
	}
	if creds["stt_inference_config"] == "" {
		t.Error("expected stt_inference_config credential to be set")
	}
}
