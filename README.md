# ZRT — Zero Runtime Go SDK

**Build real-time AI voice agents in Go — without running the infrastructure.**
You write the agent (instructions, tools, logic); **Zero Runtime** runs the live
speech-to-speech pipeline — speech-to-text → LLM → text-to-speech, with turn
detection, denoising, and interruptions — at low latency in the cloud.

> **Write the agent. We run the runtime.**

## Requirements

- Go **1.26+**
- A ZRT runtime endpoint + auth token (from your Zero Runtime account)
- API key(s) for the providers you use (e.g. Deepgram, Google, Cartesia)

## Install

```bash
go get github.com/ZeroRuntimeAI/zrt-golang-sdk@latest
```

## Quickstart

**1. Set your environment**

```bash
export ZRT_RUNTIME_ADDRESS=us1.rt.zeroruntime.ai:443   # your ZRT runtime
export ZRT_AUTH_TOKEN=<your-token>

export DEEPGRAM_API_KEY=<key>    # speech-to-text
export GOOGLE_API_KEY=<key>      # the LLM (Gemini)
export CARTESIA_API_KEY=<key>    # text-to-speech
```

**2. Write your agent** — `main.go`

```go
package main

import (
	"context"

	"github.com/ZeroRuntimeAI/zrt-golang-sdk/zrt"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/cartesia"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/deepgram"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/google"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/rnnoise"
	"github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/silero"
	td "github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/turn_detector"
)

// Assistant is your agent. Embed zrt.BaseAgent and implement OnEnter/OnExit.
type Assistant struct{ zrt.BaseAgent }

func (a *Assistant) OnEnter(ctx context.Context) error {
	_, err := a.Session().Say(ctx, "Hi! How can I help?")
	return err
}
func (a *Assistant) OnExit(ctx context.Context) error { return nil }

func entrypoint(ctx context.Context, jobCtx *zrt.JobContext) error {
	agent := &Assistant{BaseAgent: zrt.NewBaseAgent(zrt.AgentOptions{
		Instructions: "You are a friendly voice assistant. Keep replies short.",
	})}

	pipeline := zrt.NewPipeline(zrt.PipelineOptions{
		STT: deepgram.NewSTT(deepgram.STTOptions{}),
		LLM: google.NewLLM(google.LLMOptions{
			Model: "gemini-2.5-flash", MaxOutputTokens: 8192,
		}),
		TTS:          cartesia.NewTTS(cartesia.TTSOptions{}),
		VAD:          silero.NewVAD(silero.VADOptions{Threshold: zrt.Float64(0.4)}),
		TurnDetector: td.NewNamoTurnDetectorV1("en", 0.8),
		Denoise:      rnnoise.New(),
		EOUConfig:    &zrt.EOUConfig{Mode: "ADAPTIVE", MinMaxSpeechWaitTimeout: []float64{0.1, 0.3}},
		InterruptConfig: &zrt.InterruptConfig{
			InterruptMinDuration: 0.5, InterruptMinWords: 2, ResumeOnFalseInterrupt: true,
		},
	})

	session := zrt.NewAgentSession(agent, pipeline, zrt.AgentSessionOptions{})
	return session.Start(ctx, jobCtx, zrt.StartOptions{
		WaitForParticipant: true,
		RunUntilShutdown:   true,
	})
}

func main() {
	jobctx := func() *zrt.JobContext {
		return zrt.NewJobContext(&zrt.RoomOptions{Name: "Assistant", Playground: true}, nil)
	}
	if err := zrt.NewWorkerJob(entrypoint, jobctx, nil).Start(); err != nil {
		panic(err)
	}
}
```

**3. Run it**

```bash
go run .
```

That's it — speech in → your agent → speech out, in real time.

## Examples

Runnable examples live in [`examples/`](examples) — quickstart, a tool with a
filler line (`booking`), provider fallback chains (`fallback`), multi-agent
handoff (`handoff`), the ChatContext API + runtime context management
(`chat_context`, `summary_llm`), and two background-audio walkthroughs
(`background_audio`, `background_audio_hold_music`). Run any of them with
`go run ./examples/<name>`. See [`examples/README.md`](examples/README.md).

More live in a dedicated repo:
**[ZeroRuntimeAI/zrt-golang-sdk-examples](https://github.com/ZeroRuntimeAI/zrt-golang-sdk-examples)**

## How it works

| Piece | What it is |
|---|---|
| **`zrt.Agent`** | Your behavior — instructions, tools, what it says on enter/exit. Embed `zrt.BaseAgent`. |
| **`zrt.Pipeline`** | The voice stack: STT → LLM → TTS, plus VAD, turn detection, and denoising. |
| **`zrt.WorkerJob`** | Runs your agent and connects it to Zero Runtime. |

## Give your agent tools

Let the LLM call your Go functions — register a `FunctionTool` with a JSON schema
and a handler:

```go
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

agent := &Assistant{BaseAgent: zrt.NewBaseAgent(zrt.AgentOptions{
	Instructions: "...",
	Tools:        []*zrt.FunctionTool{weather},
})}
```

Your tool runs in your worker; the runtime calls it when the LLM decides to.

## Providers

Mix and match — bring the best model for each stage, swap any one in a line.
Each provider lives in `github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/<name>`:

- **Speech-to-text (STT):** `deepgram`, `assemblyai`, `google`, `azure`, `gladia`, `nvidia`, `sarvamai`
- **LLM:** `openai`, `google` (Gemini), `anthropic` (Claude), `groq`, `cerebras`, `xai` (Grok), `sarvamai`, `cometapi`
- **Text-to-speech (TTS):** `cartesia`, `elevenlabs`, `google`, `aws`, `azure`, `rime`, `lmnt`, `neuphonic`, `humeai`, `inworldai`, `murfai`, `resemble`, `smallestai`, `speechify`, `cambai`, `papla`, `nvidia`, `sarvamai`, `groq`
- **Realtime speech-to-speech:** `openai_realtime`, `gemini_realtime` (Gemini Live), `ultravox`, `xai` (realtime), `azure` (Voice Live)
- **Turn detection:** `turn_detector` / `navana` (Namo) · **VAD:** `silero` · **Denoise:** `rnnoise`

```go
import "github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/elevenlabs"  // different TTS
import "github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/anthropic"   // different LLM
```

## Realtime (speech-to-speech)

Pass a realtime model in the LLM slot:

```go
import oair "github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/openai_realtime"

pipeline := zrt.NewPipeline(zrt.PipelineOptions{
	LLM: oair.NewRealtime(oair.RealtimeOptions{Voice: "alloy"}),
})
```

## Events

`AgentSession` and `Pipeline` are event emitters. Subscribe to session lifecycle
events, or register typed pipeline hooks:

```go
session.On("user_state_changed", func(p any) { /* ... */ })
session.On("metrics_collected", func(p any) { /* ... */ })

pipeline.OnEOUDetected(func(prob float64, waitMS uint32, text string) { /* ... */ })
pipeline.OnFirstAudioByte(func(ttfbMS, byteCount uint32) { /* ... */ })
```

## Serve & invoke

For multi-session workers, give your agent an `AgentID` and a pipeline, then
`Serve` it. `Serve` registers the agent with the ZRT registry over a WebSocket
connection and listens — it does **not** start a session on its own. Call `Invoke`
to start one (from anywhere: a script, a CLI, a web handler). Scale out by running
more `Serve` workers.

```go
// The agent embeds BaseAgent built with AgentOptions{AgentID: "my-agent", Pipeline: pipeline}.
agent := NewAssistant(pipeline)

// Serve registers and blocks. OnReady fires once registration is confirmed
// call Invoke to kick off a session for local testing.
zrt.Serve(agent, zrt.ServeOptions{
    OnReady: func() {
        res, err := zrt.Invoke("my-agent", zrt.InvokeOptions{})
        if err == nil && res.PlaygroundURL != "" {
            fmt.Println("Join the playground:", res.PlaygroundURL)
        }
    },
})
```

Or invoke from a separate process once the worker is serving:

```go
res, _ := zrt.Invoke("my-agent", zrt.InvokeOptions{Room: &zrt.Room{RoomID: "existing-room"}})
fmt.Println("session:", res.SessionID, "worker:", res.WorkerID)
```

## Environment variables

| Var | Purpose |
|---|---|
| `ZRT_RUNTIME_ADDRESS` | Runtime gRPC address |
| `ZRT_AUTH_TOKEN` | Pre-minted auth token |
| `ZRT_API_KEY` + `ZRT_SECRET_KEY` | Mint a JWT if no token is set |
| `ZRT_RUNTIME_INSECURE` | `1` to use an insecure (non-TLS) channel |
| `ZRT_SIGNALING_URL` | Signaling/API base (default `api.videosdk.live`) |
| `<PROVIDER>_API_KEY` | Provider keys (e.g. `DEEPGRAM_API_KEY`, `CARTESIA_API_KEY`) |

## Contact

support@videosdk.live

Copyright © 2026 Zujo Tech Pvt Ltd. All rights reserved.
