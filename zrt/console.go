package zrt

import (
	"cmp"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// isConsoleMode reports whether the agent was launched in console mode, either
// via the --console flag or ZRT_MODE=console. 
func isConsoleMode() bool {
	for _, a := range os.Args {
		if a == "--console" {
			return true
		}
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ZRT_MODE")), "console")
}

// consoleStage is one STT/LLM/TTS entry in the pipeline description handed to the
// engine via ZRT_CONSOLE_PIPELINE.
type consoleStage struct {
	Kind     string `json:"kind"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// pipelineJSON encodes the pipeline's STT/LLM/TTS provider+model into the JSON
// shape the zrt-console engine expects. Returns "" if the pipeline is nil.
func pipelineJSON(p *Pipeline) string {
	if p == nil {
		return ""
	}
	stages := []consoleStage{}
	if p.STT != nil {
		c := p.STT.STTConfig()
		stages = append(stages, consoleStage{Kind: "STT", Provider: c.Provider, Model: c.Model})
	}
	// The LLM slot may hold a realtime (speech-to-speech) model that does not
	// implement LLMConfig(); include it only when it is a text LLM.
	if llm, ok := p.LLM.(LLM); ok && llm != nil {
		c := llm.LLMConfig()
		stages = append(stages, consoleStage{Kind: "LLM", Provider: c.Provider, Model: c.Model})
	}
	if p.TTS != nil {
		c := p.TTS.TTSConfig()
		model := c.Model
		if model == "" {
			model = c.Voice
		}
		stages = append(stages, consoleStage{Kind: "TTS", Provider: c.Provider, Model: model})
	}
	b, err := json.Marshal(map[string]any{"stages": stages})
	if err != nil {
		return ""
	}
	return string(b)
}

// spawnConsoleEngine resolves and launches the local zrt-console engine wired to
// the current terminal. The returned command has already been started.
func spawnConsoleEngine(room roomConfigData, pipelineCfg string) (*exec.Cmd, error) {
	binary, err := resolveEngineBinary(requiredEngineVersion)
	if err != nil {
		return nil, err
	}
	name := cmp.Or(room.AgentName, "terminal")
	cmd := exec.Command(binary, "--room", room.RoomID, "--token", room.AuthToken, "--name", name)
	env := append(os.Environ(), "ZRT_CONSOLE_DRIVER=sdk")
	if pipelineCfg != "" {
		env = append(env, "ZRT_CONSOLE_PIPELINE="+pipelineCfg)
	}
	cmd.Env = env
	// Wire the engine TUI directly to the real terminal.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func driveConsoleEngine(ctx context.Context, s *AgentSession, room roomConfigData) error {
	if room.RoomID == "" {
		return errConsoleRoomRequired
	}

	cmd, err := spawnConsoleEngine(room, pipelineJSON(s.pipeline))
	if err != nil {
		return err
	}

	logPath := filepath.Join(os.TempDir(), "zrt-console-host.log")
	prevLogger := logger
	var logFile *os.File
	if f, ferr := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644); ferr == nil {
		logFile = f
		SetLogger(NewSlogLogger(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug}))))
	}
	restoreLogging := func() {
		SetLogger(prevLogger)
		if logFile != nil {
			_ = logFile.Close()
		}
	}

	if err := s.agent.OnEnter(s.bindBus(ctx)); err != nil {
		logger.Errorf("on_enter error: %v", err)
	}
	waitErr := cmd.Wait()

	restoreLogging()
	if err := s.Close(ctx, "console_exit"); err != nil {
		logger.Errorf("console session close error: %v", err)
	}
	if logFile != nil {
		logger.Infof("[zrt-console] SDK logs during the session were written to %s", logPath)
	}
	return waitErr
}
