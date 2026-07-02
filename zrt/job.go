package zrt

import (
	"cmp"
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
	// Video records the participant's camera track.
	Video bool
	// ScreenShare records the participant's screen-share track.
	ScreenShare bool
}

// WebSocketConfig configures a WebSocket transport.
type WebSocketConfig struct {
	Port int    // default 8080
	Path string // default "/ws"
}

// WebRTCConfig configures a WebRTC transport.
type WebRTCConfig struct {
	// SignalingURL is the signaling server URL.
	SignalingURL  string
	SignalingType string // default "websocket"
	// ICEServers lists the ICE (STUN/TURN) servers used for connectivity.
	ICEServers []any
}

// TracesOptions configures trace export.
type TracesOptions struct {
	// Enabled turns trace export on.
	Enabled bool
	// ExportURL is the endpoint traces are exported to.
	ExportURL string
	// ExportHeaders are extra HTTP headers sent with each export request.
	ExportHeaders map[string]string
}

// MetricsOptions configures metric export.
type MetricsOptions struct {
	// Enabled turns metric export on.
	Enabled bool
	// ExportURL is the endpoint metrics are exported to.
	ExportURL string
	// ExportHeaders are extra HTTP headers sent with each export request.
	ExportHeaders map[string]string
}

// LoggingOptions configures log export.
type LoggingOptions struct {
	// Enabled turns log export on.
	Enabled bool
	Level   string // default "INFO"
	// ExportURL is the endpoint logs are exported to.
	ExportURL string
	// ExportHeaders are extra HTTP headers sent with each export request.
	ExportHeaders map[string]string
}

// ObservabilityOptions groups traces/metrics/logs.
type ObservabilityOptions struct {
	// Traces configures trace export; nil leaves traces unconfigured.
	Traces *TracesOptions
	// Metrics configures metric export; nil leaves metrics unconfigured.
	Metrics *MetricsOptions
	// Logs configures log export; nil leaves logs unconfigured.
	Logs *LoggingOptions
}

// PubSubPublishConfig configures publishing a message on a room pubsub topic.
// Persistence is requested through Options (e.g. {"persist": true}).
type PubSubPublishConfig struct {
	Topic string // e.g. "CHAT", "AGENT_EVENT"
	// Message is the text payload published on the topic.
	Message string
	// Options are publish options forwarded to the room (e.g. {"persist": true}).
	Options map[string]any
	// Payload is an optional structured payload sent alongside Message.
	Payload any
}

// NewPubSubPublishConfig returns a config for the given topic.
func NewPubSubPublishConfig(topic string) *PubSubPublishConfig {
	return &PubSubPublishConfig{Topic: topic, Message: ""}
}

// RoomOptions configures the room/session the agent joins.
type RoomOptions struct {
	// RoomID is the room to join. When empty, a room is created on demand
	// (see JobContext.RoomID) from ZRT_ROOM_ID or the signaling API.
	RoomID string
	// AuthToken authorizes joining/creating the room; falls back to the
	// ambient ZRT auth when empty.
	AuthToken string
	// Name is the agent's display name in the room. Defaults to "Agent".
	Name string
	// AgentParticipantID is a fixed participant id to publish the agent under.
	AgentParticipantID string
	// Playground allows playground sessions. Defaults to true.
	Playground bool
	// Vision enables camera-frame capture for the session.
	Vision bool
	// Recording enables session recording.
	Recording bool
	// RecordingOptions selects which tracks are recorded when Recording is on.
	RecordingOptions *RecordingOptions
	// JoinMeeting makes the agent join the meeting. Defaults to true.
	JoinMeeting bool
	// OnRoomError is called with room-level errors.
	OnRoomError func(any)
	// OnSessionEnd is called when the session ends.
	OnSessionEnd func(any)
	// AutoEndSession ends the session automatically when the caller leaves.
	// Defaults to true.
	AutoEndSession bool
	// SessionTimeoutSeconds is a hard cap on session duration. Defaults to 60.
	SessionTimeoutSeconds int
	// NoParticipantTimeoutSeconds ends the session if no participant joins
	// within this many seconds. Defaults to 180.
	NoParticipantTimeoutSeconds int
	// SignalingBaseURL overrides the signaling base URL. Defaults to
	// ZRT_SIGNALING_URL or "api.videosdk.live".
	SignalingBaseURL string
	// BackgroundAudio enables background-audio playback.
	BackgroundAudio bool
	// AudioListenerEnabled delivers raw input audio to listeners.
	AudioListenerEnabled bool
	// SendLogsToDashboard forwards SDK logs to the dashboard. Defaults to true.
	SendLogsToDashboard bool
	// DashboardLogLevel sets the level for dashboard logs. Defaults to "INFO".
	DashboardLogLevel string
	// Traces configures trace export for the session.
	Traces *TracesOptions
	// Metrics configures metric export for the session.
	Metrics *MetricsOptions
	// Logs configures log export for the session.
	Logs *LoggingOptions
	// WebSocket configures a WebSocket transport.
	WebSocket *WebSocketConfig
	// WebRTC configures a WebRTC transport.
	WebRTC *WebRTCConfig
	// Observability groups traces/metrics/logs configuration.
	Observability *ObservabilityOptions
}

// NewRoomOptions returns RoomOptions with the default values.
func NewRoomOptions() *RoomOptions {
	return &RoomOptions{
		Name:                        "Agent",
		Playground:                  true,
		JoinMeeting:                 true,
		AutoEndSession:              true,
		SessionTimeoutSeconds:       60,
		NoParticipantTimeoutSeconds: 180,
		SignalingBaseURL:            cmp.Or(os.Getenv("ZRT_SIGNALING_URL"), "api.videosdk.live"),
		SendLogsToDashboard:         true,
		DashboardLogLevel:           "INFO",
	}
}

// WorkerOptions configures the worker.
type WorkerOptions struct {
	// NumIdleProcesses is the number of idle processes kept warm. Defaults to 1.
	NumIdleProcesses int
	// InitializeTimeout bounds agent/model initialization. Defaults to 10s.
	InitializeTimeout time.Duration
	// CloseTimeout bounds session shutdown. Defaults to 60s.
	CloseTimeout time.Duration
	// MaxProcesses is the maximum concurrent sessions this worker accepts.
	// Defaults to 1.
	MaxProcesses int
	// AgentID is the id the agent registers under. Defaults to "ZeroRuntimeAgent".
	AgentID string
	// AuthToken authorizes registration; falls back to the ambient ZRT auth
	// when empty.
	AuthToken string
	// MaxRetry is the maximum registration retry attempts. Defaults to 16.
	MaxRetry int
	// LoadThreshold is the fraction of MaxProcesses (0–1) above which the
	// worker reports itself busy. Defaults to 0.75.
	LoadThreshold float64
	// Register makes Start register with the ZRT registry and serve dispatched
	// jobs instead of running a single local job.
	Register bool
	// SignalingBaseURL overrides the signaling base URL. Defaults to
	// ZRT_SIGNALING_URL or "api.videosdk.live".
	SignalingBaseURL string
	// Host is the interface the debug endpoint binds. Defaults to "0.0.0.0".
	Host string
	// Port is the port the debug endpoint binds. Defaults to 8081.
	Port int
	// LogLevel sets the built-in logger verbosity. Defaults to "INFO".
	LogLevel string

	// DebugEnabled starts a local HTTP debug endpoint (health/worker/stats) on
	// Host:Port while the worker runs. Off by default; Serve turns it on.
	DebugEnabled bool

	// OnReady, when set, is called once registration is confirmed (registered/serve
	// mode). It may block / call zrt.Invoke; it runs on its own goroutine and panics
	// are recovered and logged.
	OnReady func()

	// Logger routes SDK logs through a standard library *slog.Logger. When set,
	// it takes precedence over LogLevel (level filtering is handled by the slog
	// handler). When nil, the SDK uses its built-in logger at LogLevel.
	Logger *slog.Logger
}

// NewWorkerOptions returns WorkerOptions with the default values.
func NewWorkerOptions() *WorkerOptions {
	return &WorkerOptions{
		NumIdleProcesses:  1,
		InitializeTimeout: 10 * time.Second,
		CloseTimeout:      60 * time.Second,
		MaxProcesses:      1,
		AgentID:           "ZeroRuntimeAgent",
		MaxRetry:          16,
		LoadThreshold:     0.75,
		SignalingBaseURL:  cmp.Or(os.Getenv("ZRT_SIGNALING_URL"), "api.videosdk.live"),
		Host:              "0.0.0.0",
		Port:              8081,
		LogLevel:          "INFO",
	}
}

// JobContext carries per-job configuration and session state.
type JobContext struct {
	// RoomOptions configures the room/session for this job.
	RoomOptions *RoomOptions
	// Metadata carries the job's dispatch metadata.
	Metadata map[string]any

	mu                sync.Mutex
	shutdownCallbacks []func()
	shuttingDown      bool
	activeSession     *AgentSession
	workerJob         *WorkerJob
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

// RoomID returns the room id, creating a room via the signaling API if unset.
func (j *JobContext) RoomID() (string, error) {
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
	base := cmp.Or(signalingBaseURL, os.Getenv("ZRT_SIGNALING_URL"), "api.videosdk.live")
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

type runnerInfo struct {
	session   *AgentSession
	roomOpts  *RoomOptions
	sessionID string
}

// WorkerJob runs an agent entrypoint for each incoming job.
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

func (w *WorkerJob) getStats() map[string]any {
	w.mu.Lock()
	currentJobs := len(w.currentJobs)
	legacy := w.legacyReg
	w.mu.Unlock()

	maxProc := max(1, w.options.MaxProcesses)
	workerID := "unregistered"
	connected := false
	draining := false
	if legacy != nil {
		wid, conn, drain, active := legacy.stats()
		if wid != "" {
			workerID = wid
		}
		connected = conn
		draining = drain
		if active > currentJobs {
			currentJobs = active
		}
	}
	return map[string]any{
		"agent_id":          cmp.Or(w.options.AgentID, "ZeroRuntimeAgent"),
		"worker_id":         workerID,
		"current_jobs":      currentJobs,
		"active_jobs":       currentJobs,
		"max_processes":     maxProc,
		"worker_load":       float64(currentJobs) / float64(maxProc),
		"load_threshold":    w.options.LoadThreshold,
		"backend_connected": connected,
		"connected":         connected,
		"draining":          draining,
		"register":          w.options.Register,
		"log_level":         w.options.LogLevel,
	}
}

// runnersSnapshot lists the worker's active sessions for the debug endpoint.
func (w *WorkerJob) runnersSnapshot() []map[string]any {
	w.mu.Lock()
	defer w.mu.Unlock()
	runners := make([]map[string]any, 0, len(w.currentJobs))
	for id, info := range w.currentJobs {
		room := "unknown"
		if info.roomOpts != nil && info.roomOpts.RoomID != "" {
			room = info.roomOpts.RoomID
		}
		runners = append(runners, map[string]any{
			"id": id, "room": room, "status": "running", "task_id": id, "session_id": info.sessionID,
		})
	}
	if len(runners) == 0 {
		runners = append(runners, map[string]any{"id": "worker_main", "room": "main_worker", "status": "idle", "task_id": "worker_main"})
	}
	return runners
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

	defer w.stopDebugServer(w.startDebugServer())

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

// runRegistered registers the agent with the ZRT registry and serves dispatched
// sessions, running the entrypoint for each job. It blocks until shutdown.
func (w *WorkerJob) runRegistered() error {
	resolvedToken, _ := ResolveAuthToken(w.options.AuthToken)
	if resolvedToken == "" {
		return fmt.Errorf("zrt.Serve registers over the WebSocket registry and needs auth " +
			"(ZRT_AUTH_TOKEN / ZRT_API_KEY + ZRT_SECRET_KEY)")
	}

	base := cmp.Or(w.options.SignalingBaseURL, os.Getenv("ZRT_SIGNALING_URL"), "api.videosdk.live")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	agentID := cmp.Or(w.options.AgentID, "ZeroRuntimeAgent")
	legacy := newLegacyBackendRegistration(resolvedToken, agentID, base, w.options.LoadThreshold, max(1, w.options.MaxProcesses), w.entrypoint, w.jobctxFactory)
	w.mu.Lock()
	w.legacyReg = legacy
	w.mu.Unlock()

	defer w.stopDebugServer(w.startDebugServer())

	shutdownCh := make(chan struct{})
	var shutdownOnce sync.Once
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Infof("Received shutdown signal — stopping registration")
		shutdownOnce.Do(func() { close(shutdownCh) })
	}()

	if legacy.start() {
		logger.Infof("Agent registered with registry: agent_id=%s worker_id=%s", agentID, legacy.workerID)
		if w.options.OnReady != nil {
			// OnReady may call zrt.Invoke; run it off the registration goroutine.
			go func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Errorf("OnReady callback panicked: %v", r)
					}
				}()
				w.options.OnReady()
			}()
		}
	} else {
		logger.Warnf("Registry did not confirm registration on first attempt — retrying in the background.")
	}

	<-shutdownCh
	legacy.stop()
	logger.Infof("Registered agent shutdown complete.")
	return nil
}
