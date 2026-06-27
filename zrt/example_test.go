package zrt_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/cartesia"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/deepgram"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/google"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
)

// assistant is a minimal agent that greets the caller on entry.
type assistant struct{ zrt.BaseAgent }

func (a *assistant) OnEnter(ctx context.Context) error {
	_, err := a.Session(ctx).Say(ctx, "Hi! How can I help?")
	return err
}
func (a *assistant) OnExit(ctx context.Context) error { return nil }

// Build a speech-to-speech pipeline from provider plugins and wrap it in an
// AgentSession. Provider API keys are read from the environment when the
// Options APIKey field is left empty.
func Example() {
	pipeline := zrt.NewPipeline(zrt.PipelineOptions{
		STT: deepgram.NewSTT(deepgram.STTOptions{}),
		LLM: google.NewLLM(google.LLMOptions{Model: "gemini-2.5-flash"}),
		TTS: cartesia.NewTTS(cartesia.TTSOptions{}),
	})

	agent := &assistant{BaseAgent: zrt.NewBaseAgent(zrt.AgentOptions{
		Instructions: "You are a friendly voice assistant. Keep replies short.",
	})}

	session := zrt.NewAgentSession(agent, pipeline, zrt.AgentSessionOptions{})
	_ = session // session.Start(ctx, jobCtx, zrt.StartOptions{}) runs the agent.
}

// Register a Go function the LLM can call during the conversation.
func ExampleNewFunctionTool() {
	weather := zrt.NewFunctionTool(
		"get_weather",
		"Get the weather for a city.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"city": map[string]any{"type": "string", "description": "City name"},
			},
			"required": []any{"city"},
		},
		func(ctx context.Context, args map[string]any) (any, error) {
			city, _ := args["city"].(string)
			return map[string]any{"city": city, "temp_c": 22}, nil
		},
	)
	fmt.Println(weather.Info.Name)
	// Output: get_weather
}

// Branch on failure modes with errors.Is against the exported sentinels.
func ExampleErrNoCredentials() {
	err := fmt.Errorf("starting session: %w", zrt.ErrNoCredentials)
	fmt.Println(errors.Is(err, zrt.ErrNoCredentials))
	// Output: true
}

// Float64 reads more clearly than Ptr for optional float fields.
func ExampleFloat64() {
	threshold := zrt.Float64(0.4)
	fmt.Println(*threshold)
	// Output: 0.4
}

// A tool whose result takes a moment to compute can speak a short "filler" line
// to cover the latency. WithFiller attaches the line and an optional grace period
// (ms): the filler is spoken only if the tool is still running after the grace,
// so a fast tool stays silent.
func ExampleWithFiller() {
	checkAvailability := zrt.NewFunctionTool(
		"check_availability",
		"Check open appointment slots for a given day.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"day": map[string]any{"type": "string", "description": "The day to check."},
			},
			"required": []any{"day"},
		},
		func(ctx context.Context, args map[string]any) (any, error) {
			day, _ := args["day"].(string)
			// ...a slow availability lookup; the filler covers this latency...
			return "We have 10:00 AM and 2:30 PM open on " + day, nil
		},
		// Speak the filler only if the lookup runs longer than 500ms.
		zrt.WithFiller("Let me check the schedule...", 500),
	)

	info := checkAvailability.ToolInfo()
	fmt.Printf("%s filler=%q grace=%dms\n", info.Name, info.Filler, info.FillerGracePeriod)
	// Output: check_availability filler="Let me check the schedule..." grace=500ms
}
