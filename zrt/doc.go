// Package zrt is the Go SDK for Zero Runtime (ZRT).
//
// You author voice agents — instructions, tools, and logic — and the ZRT cloud
// runtime executes the real-time speech-to-speech pipeline (media, VAD, turn
// detection, STT, LLM, TTS) for you. The agent-facing API lives in
// this package; provider descriptors live under
// github.com/ZeroRuntimeAI/zrt-golang-sdk/plugins/<name>.
//
// # Quickstart
//
// Build a pipeline from provider plugins, wrap it in an AgentSession, and run it
// inside a WorkerJob:
//
//	pipeline := zrt.NewPipeline(zrt.PipelineOptions{
//		STT: deepgram.NewSTT(deepgram.STTOptions{}),
//		LLM: google.NewLLM(google.LLMOptions{Model: "gemini-2.5-flash"}),
//		TTS: cartesia.NewTTS(cartesia.TTSOptions{}),
//	})
//	agent := &Assistant{BaseAgent: zrt.NewBaseAgent(zrt.AgentOptions{
//		Instructions: "You are a friendly voice assistant.",
//	})}
//	session := zrt.NewAgentSession(agent, pipeline, zrt.AgentSessionOptions{})
//	// session.Start(ctx, jobCtx, zrt.StartOptions{})
//
// # Configuration
//
// Provider plugins read their API key from the environment when the Options
// APIKey field is left empty (for example deepgram reads DEEPGRAM_API_KEY). The
// runtime endpoint and credentials are read from ZRT_RUNTIME_ADDRESS and
// ZRT_AUTH_TOKEN (or ZRT_API_KEY + ZRT_SECRET_KEY). See each plugin's Options
// type for the exact variable it uses.
//
// # Errors
//
// Failures wrap exported sentinel errors so callers can branch with errors.Is —
// for example errors.Is(err, ErrNoCredentials) or errors.Is(err, ErrSessionRejected).
//
// # Logging
//
// By default the SDK logs through its built-in logger at WorkerOptions.LogLevel.
// To route logs through the standard library log/slog, set WorkerOptions.Logger
// to a *slog.Logger.
package zrt
