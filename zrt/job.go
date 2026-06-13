package zrt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// EntrypointFunc is the agent entrypoint, invoked per job/session.
type EntrypointFunc func(ctx context.Context, jobCtx *JobContext) error

// RecordingOptions toggles per-track recording.
type RecordingOptions struct {
	Video       bool
	ScreenShare bool
}

// WebSocketConfig configures a WebSocket transport.
type WebSocketConfig struct {
	Port int    // default 8080
	Path string // default "/ws"
}

// WebRTCConfig configures a WebRTC transport.
type WebRTCConfig struct {
	SignalingURL  string
	SignalingType string // default "websocket"
	ICEServers    []any
}

// TracesOptions configures trace export.
type TracesOptions struct {
	Enabled       bool
	ExportURL     string
	ExportHeaders map[string]string
}

// MetricsOptions configures metric export.
type MetricsOptions struct {
	Enabled       bool
	ExportURL     string
	ExportHeaders map[string]string
}

// LoggingOptions configures log export.
type LoggingOptions struct {
	Enabled       bool
	Level         string // default "INFO"
	ExportURL     string
	ExportHeaders map[string]string
}

// ObservabilityOptions groups traces/metrics/logs.
type ObservabilityOptions struct {
	Traces  *TracesOptions
	Metrics *MetricsOptions
	Logs    *LoggingOptions
}

// PubSubPublishConfig configures publishing a message on a room pubsub topic.
type PubSubPublishConfig struct {
	Topic    string // e.g. "CHAT", "AGENT_EVENT"
	Message  any
	Mode     string // "sendOnly" (default) or "sendAndPersist"
	SendOnly bool
	Payload  map[string]any
}

// NewPubSubPublishConfig returns a config with default send-only mode applied.
func NewPubSubPublishConfig(topic string) *PubSubPublishConfig {
	return &PubSubPublishConfig{Topic: topic, Message: "", Mode: "sendOnly"}
}

// RoomOptions configures the room/session the agent joins.
type RoomOptions struct {
	RoomID                      string
	AuthToken                   string
	Name                        string
	AgentParticipantID          string
	Playground                  bool
	Vision                      bool
	Recording                   bool
	RecordingOptions            *RecordingOptions
	JoinMeeting                 bool
	OnRoomError                 func(any)
	OnSessionEnd                func(any)
	AutoEndSession              bool
	SessionTimeoutSeconds       int
	NoParticipantTimeoutSeconds int
	SignalingBaseURL            string
	BackgroundAudio             bool
	AudioListenerEnabled        bool
	SendLogsToDashboard         bool
	DashboardLogLevel           string
	Traces                      *TracesOptions
	Metrics                     *MetricsOptions
	Logs                        *LoggingOptions
	WebSocket                   *WebSocketConfig
	WebRTC                      *WebRTCConfig
	Observability               *ObservabilityOptions
}

// NewRoomOptions returns RoomOptions with the default values.
func NewRoomOptions() *RoomOptions {
	signaling := os.Getenv("ZRT_SIGNALING_URL")
	if signaling == "" {
		signaling = "api.videosdk.live"
	}
	return &RoomOptions{
		Name:                        "Agent",
		Playground:                  true,
		JoinMeeting:                 true,
		AutoEndSession:              true,
		SessionTimeoutSeconds:       60,
		NoParticipantTimeoutSeconds: 180,
		SignalingBaseURL:            signaling,
		SendLogsToDashboard:         true,
		DashboardLogLevel:           "INFO",
	}
}

// WorkerOptions configures the worker.
type WorkerOptions struct {
	NumIdleProcesses  int
	InitializeTimeout time.Duration
	CloseTimeout      time.Duration
	MaxProcesses      int
	AgentID           string
	AuthToken         string
	MaxRetry          int
	LoadThreshold     float64
	Register          bool
	SignalingBaseURL  string
	Host              string
	Port              int
	LogLevel          string

	// Logger routes SDK logs through a standard library *slog.Logger. When set,
	// it takes precedence over LogLevel (level filtering is handled by the slog
	// handler). When nil, the SDK uses its built-in logger at LogLevel.
	Logger *slog.Logger
}

// NewWorkerOptions returns WorkerOptions with the default values.
func NewWorkerOptions() *WorkerOptions {
	signaling := os.Getenv("ZRT_SIGNALING_URL")
	if signaling == "" {
		signaling = "api.videosdk.live"
	}
	return &WorkerOptions{
		NumIdleProcesses:  1,
		InitializeTimeout: 10 * time.Second,
		CloseTimeout:      60 * time.Second,
		MaxProcesses:      1,
		AgentID:           "ZeroRuntimeAgent",
		MaxRetry:          16,
		LoadThreshold:     0.75,
		SignalingBaseURL:  signaling,
		Host:              "0.0.0.0",
		Port:              8081,
		LogLevel:          "INFO",
	}
}

// JobContext carries per-job configuration and runtime wiring.
type JobContext struct {
	RoomOptions *RoomOptions
	Metadata    map[string]any

	mu                sync.Mutex
	shutdownCallbacks []func()
	shuttingDown      bool
	activeSession     *AgentSession
	workerJob         *WorkerJob

	registeredMode      bool
	registeredSessionID string
	registeredRegistry  *agentRegistry

	registrationProbe bool
	capturedAgent     Agent
	capturedPipeline  *Pipeline
	capturedRecording *RecordingConfig
}

// NewJobContext creates a JobContext.
func NewJobContext(roomOptions *RoomOptions, metadata map[string]any) *JobContext {
	if roomOptions == nil {
		roomOptions = NewRoomOptions()
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	return &JobContext{RoomOptions: roomOptions, Metadata: metadata}
}

// AddShutdownCallback registers a callback run during job shutdown.
func (j *JobContext) AddShutdownCallback(cb func()) {
	j.mu.Lock()
	j.shutdownCallbacks = append(j.shutdownCallbacks, cb)
	j.mu.Unlock()
}

func (j *JobContext) registerSession(s *AgentSession) {
	j.mu.Lock()
	j.activeSession = s
	wj := j.workerJob
	j.mu.Unlock()
	if wj != nil {
		wj.registerRunner(s, j.RoomOptions)
	}
}

func (j *JobContext) shutdown(ctx context.Context) {
	j.mu.Lock()
	if j.shuttingDown {
		j.mu.Unlock()
		return
	}
	j.shuttingDown = true
	sess := j.activeSession
	cbs := j.shutdownCallbacks
	j.shutdownCallbacks = nil
	wj := j.workerJob
	j.mu.Unlock()

	if sess != nil {
		if err := sess.Close(ctx, "sdk_close"); err != nil {
			logger.Errorf("Error closing session: %v", err)
		}
		if wj != nil {
			wj.unregisterRunner(sess)
		}
		j.mu.Lock()
		j.activeSession = nil
		j.mu.Unlock()
	}
	for _, cb := range cbs {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("shutdown callback panicked: %v", r)
				}
			}()
			cb()
		}()
	}
}

// GetRoomID returns the room id, creating a room via the signaling API if unset.
func (j *JobContext) GetRoomID() (string, error) {
	if j.RoomOptions.RoomID != "" {
		return j.RoomOptions.RoomID, nil
	}
	if envRoom := strings.TrimSpace(os.Getenv("ZRT_ROOM_ID")); envRoom != "" {
		j.RoomOptions.RoomID = envRoom
		return envRoom, nil
	}
	token, err := ResolveAuthToken(j.RoomOptions.AuthToken)
	if err != nil {
		return "", fmt.Errorf("no ZRT auth available to create a room: %w", err)
	}
	roomID, err := createRoomStatic(token, j.RoomOptions.SignalingBaseURL)
	if err != nil {
		return "", err
	}
	j.RoomOptions.RoomID = roomID
	logger.Infof("Created room: %s", roomID)
	return roomID, nil
}

func createRoomStatic(authToken, signalingBaseURL string) (string, error) {
	base := signalingBaseURL
	if base == "" {
		base = os.Getenv("ZRT_SIGNALING_URL")
	}
	if base == "" {
		base = "api.videosdk.live"
	}
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	base = strings.TrimPrefix(strings.TrimPrefix(base, "https://"), "http://")
	url := "https://" + base + "/v2/rooms"
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", authToken)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create room: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to create room: %w", err)
	}
	roomID, _ := data["roomId"].(string)
	if roomID == "" {
		return "", fmt.Errorf("room creation returned no roomId: %s", string(body))
	}
	return roomID, nil
}

// errRegistrationProbeComplete aborts the probe entrypoint after capture.
var errRegistrationProbeComplete = errors.New("zrt: registration probe complete")

type runnerInfo struct {
	session   *AgentSession
	roomOpts  *RoomOptions
	sessionID string
}

// WorkerJob runs an agent entrypoint and connects it to the runtime.
type WorkerJob struct {
	entrypoint    EntrypointFunc
	jobctxFactory func() *JobContext
	options       *WorkerOptions

	mu          sync.Mutex
	currentJobs map[string]*runnerInfo
	legacyReg   *legacyBackendRegistration
}

// NewWorkerJob builds a WorkerJob. jobctxFactory may be nil.
func NewWorkerJob(entrypoint EntrypointFunc, jobctxFactory func() *JobContext, options *WorkerOptions) *WorkerJob {
	if options == nil {
		options = NewWorkerOptions()
	}
	return &WorkerJob{entrypoint: entrypoint, jobctxFactory: jobctxFactory, options: options, currentJobs: map[string]*runnerInfo{}}
}

// Options returns the worker options.
func (w *WorkerJob) Options() *WorkerOptions { return w.options }

func (w *WorkerJob) registerRunner(s *AgentSession, ro *RoomOptions) {
	sid := s.SessionID()
	if sid == "" {
		sid = fmt.Sprintf("runner_%p", s)
	}
	w.mu.Lock()
	w.currentJobs[sid] = &runnerInfo{session: s, roomOpts: ro, sessionID: sid}
	w.mu.Unlock()
}

func (w *WorkerJob) unregisterRunner(s *AgentSession) {
	w.mu.Lock()
	for k, info := range w.currentJobs {
		if info.session == s {
			delete(w.currentJobs, k)
			break
		}
	}
	w.mu.Unlock()
}

// Start runs the worker until interrupted. Blocks.
func (w *WorkerJob) Start() error {
	if w.options.Logger != nil {
		SetLogger(NewSlogLogger(w.options.Logger))
	} else {
		SetLogLevelString(w.options.LogLevel)
	}
	if w.options.Register {
		return w.runRegistered()
	}
	return w.run()
}

func (w *WorkerJob) buildJobContext() *JobContext {
	var ctx *JobContext
	if w.jobctxFactory != nil {
		ctx = w.jobctxFactory()
	} else {
		ctx = NewJobContext(NewRoomOptions(), nil)
	}
	ctx.workerJob = w
	return ctx
}

func (w *WorkerJob) run() error {
	ctx := w.buildJobContext()
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Infof("Received shutdown signal (Ctrl+C)")
		cancel()
	}()

	err := w.entrypoint(rootCtx, ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Errorf("Entrypoint error: %v", err)
	}
	ctx.shutdown(context.Background())
	logger.Infof("Cleanup complete. Exiting.")
	return nil
}

func (w *WorkerJob) runRegistered() error {
	runtimeAddr := os.Getenv("ZRT_RUNTIME_ADDRESS")
	if runtimeAddr == "" {
		runtimeAddr = "localhost:50051"
	}
	// Registration probe: run the entrypoint to capture agent + pipeline.
	probeCtx := w.buildProbeContext()
	probeCtx.registrationProbe = true
	probeRoot, probeCancel := context.WithCancel(context.Background())
	err := w.entrypoint(probeRoot, probeCtx)
	probeCancel()
	if err != nil && !errors.Is(err, errRegistrationProbeComplete) {
		logger.Errorf("Registration probe failed: %v", err)
		return err
	}
	agentTemplate := probeCtx.capturedAgent
	pipeline := probeCtx.capturedPipeline
	defaultRecording := probeCtx.capturedRecording
	if agentTemplate == nil || pipeline == nil {
		return fmt.Errorf("register=true requires the entrypoint to construct an AgentSession (with agent + pipeline) so registration can capture the config")
	}
	resolvedToken, _ := ResolveAuthToken(w.options.AuthToken)

	registry := newAgentRegistry(runtimeAddr, w.entrypoint, w.jobctxFactory, agentTemplate, pipeline, orDefault(w.options.AgentID, "ZeroRuntimeAgent"), maxInt(1, w.options.MaxProcesses), map[string]string{}, resolvedToken, defaultRecording, w.options.InitializeTimeout)

	var legacy *legacyBackendRegistration
	if resolvedToken != "" && w.options.SignalingBaseURL != "" {
		base := w.options.SignalingBaseURL
		if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
			base = "https://" + base
		}
		legacy = newLegacyBackendRegistration(resolvedToken, orDefault(w.options.AgentID, "ZeroRuntimeAgent"), base, w.options.LoadThreshold, maxInt(1, w.options.MaxProcesses), w.entrypoint, w.jobctxFactory)
		w.mu.Lock()
		w.legacyReg = legacy
		w.mu.Unlock()
	}

	shutdownCh := make(chan struct{})
	var shutdownOnce sync.Once
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		count := 0
		for range sigCh {
			count++
			if count == 1 {
				logger.Infof("Received shutdown signal — beginning drain (press Ctrl+C again to force-exit)")
				go registry.beginDrain("sigterm")
				shutdownOnce.Do(func() { close(shutdownCh) })
			} else {
				logger.Warnf("Received shutdown signal again — forcing immediate exit")
				shutdownOnce.Do(func() { close(shutdownCh) })
				return
			}
		}
	}()

	if legacy != nil {
		legacy.start()
	}

	supervisorDone := make(chan struct{})
	go func() {
		defer close(supervisorDone)
		attempt := 0
		for {
			select {
			case <-shutdownCh:
				return
			default:
			}
			runCtx, runCancel := context.WithCancel(context.Background())
			go func() {
				<-shutdownCh
				runCancel()
			}()
			err := registry.run(runCtx)
			runCancel()
			if err != nil {
				logger.Errorf("Registry stream error: %v", err)
			}
			select {
			case <-shutdownCh:
				return
			default:
			}
			attempt++
			delay := time.Duration(minInt(attempt*2, 10)) * time.Second
			logger.Warnf("Runtime gRPC stream ended — reconnecting in %.1fs (attempt %d).", delay.Seconds(), attempt)
			select {
			case <-time.After(delay):
			case <-shutdownCh:
				return
			}
		}
	}()

	if registry.waitForRegistered(w.options.InitializeTimeout) {
		logger.Infof("Registration confirmed within initialize_timeout=%.1fs", w.options.InitializeTimeout.Seconds())
	} else {
		logger.Warnf("Runtime did not confirm registration within initialize_timeout=%.1fs — continuing to wait in the background.", w.options.InitializeTimeout.Seconds())
	}

	<-shutdownCh
	<-supervisorDone
	if legacy != nil {
		legacy.stop()
	}
	registry.stop()
	logger.Infof("Registered agent shutdown complete.")
	return nil
}

func (w *WorkerJob) buildProbeContext() *JobContext {
	ctx := w.buildJobContext()
	if len(ctx.Metadata) == 0 {
		ctx.Metadata = map[string]any{"callId": "__probe__", "sipCallTo": "+0000000000", "sipCallFrom": "+0000000000", "callType": "outbound", "webhook_url": ""}
	}
	return ctx
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
