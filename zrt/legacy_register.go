package zrt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	legacyDefaultCapabilities  = []string{"room", "voice", "stt", "tts"}
	legacyStatusInterval       = 30 * time.Second
	legacyReconnectMaxAttempts = 16
	legacyReconnectMaxDelay    = 10 * time.Second
	// legacyPongWait is how long the recv loop waits for any frame (or pong)
	// before treating the connection as dead. It must exceed legacyStatusInterval
	// (the ping cadence) so a single missed pong does not drop a live connection.
	legacyPongWait = 60 * time.Second
)

type legacyJobInfo struct {
	ctx    *JobContext
	cancel context.CancelFunc
}

// legacyBackendRegistration registers the worker over the legacy WebSocket
// dispatch path.
type legacyBackendRegistration struct {
	authToken     string
	agentID       string
	apiBaseURL    string
	loadThreshold float64
	maxProcesses  int
	capabilities  []string
	entrypoint    EntrypointFunc
	jobctxFactory func() *JobContext

	mu         sync.Mutex
	activeJobs map[string]*legacyJobInfo
	draining   bool
	conn       *websocket.Conn
	workerID   string
	closed     bool

	firstAttempt chan struct{}
	firstOnce    sync.Once
	stopCh       chan struct{}
	sendMu       sync.Mutex
}

func newLegacyBackendRegistration(authToken, agentID, apiBaseURL string, loadThreshold float64, maxProcesses int, entrypoint EntrypointFunc, jobctxFactory func() *JobContext) *legacyBackendRegistration {
	if agentID == "" {
		agentID = "ZeroRuntimeAgent"
	}
	return &legacyBackendRegistration{
		authToken:     authToken,
		agentID:       agentID,
		apiBaseURL:    apiBaseURL,
		loadThreshold: loadThreshold,
		maxProcesses:  maxProcesses,
		capabilities:  legacyDefaultCapabilities,
		entrypoint:    entrypoint,
		jobctxFactory: jobctxFactory,
		activeJobs:    map[string]*legacyJobInfo{},
		firstAttempt:  make(chan struct{}),
		stopCh:        make(chan struct{}),
	}
}

// WorkerID returns the assigned worker id ("" until registered).
func (l *legacyBackendRegistration) WorkerID() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.workerID
}

func (l *legacyBackendRegistration) start() bool {
	go l.supervisor()
	select {
	case <-l.firstAttempt:
	case <-time.After(30 * time.Second):
	}
	return l.WorkerID() != ""
}

func (l *legacyBackendRegistration) stop() {
	l.mu.Lock()
	l.closed = true
	conn := l.conn
	wid := l.workerID
	l.mu.Unlock()
	close(l.stopCh)
	if conn != nil {
		_ = l.sendJSON(map[string]any{"type": "status_update", "worker_id": wid, "status": "offline", "load": 0.0, "job_count": 0})
		conn.Close()
	}
}

func (l *legacyBackendRegistration) supervisor() {
	attempt := 0
	for {
		l.mu.Lock()
		closed := l.closed
		l.mu.Unlock()
		if closed {
			l.firstOnce.Do(func() { close(l.firstAttempt) })
			return
		}
		registryURL, err := l.fetchAgentInitConfig()
		if err == nil {
			err = l.openWS(registryURL)
		}
		if err != nil {
			if attempt == 0 {
				logger.Warnf("Legacy register skipped: %v", err)
			} else {
				logger.Warnf("Legacy reconnect failed: %v", err)
			}
		} else {
			attempt = 0
			l.firstOnce.Do(func() { close(l.firstAttempt) })
			l.runOneConnection()
		}
		l.firstOnce.Do(func() { close(l.firstAttempt) })
		l.mu.Lock()
		closed = l.closed
		l.mu.Unlock()
		if closed {
			return
		}
		attempt++
		if attempt > legacyReconnectMaxAttempts {
			logger.Errorf("Legacy backend registration gave up after %d attempts (runtime gRPC registration unaffected)", legacyReconnectMaxAttempts)
			return
		}
		delay := time.Duration(minInt(attempt*2, int(legacyReconnectMaxDelay.Seconds()))) * time.Second
		select {
		case <-time.After(delay):
		case <-l.stopCh:
			return
		}
	}
}

func (l *legacyBackendRegistration) fetchAgentInitConfig() (string, error) {
	base := os.Getenv("ZRT_API_BASE_URL")
	if base == "" {
		base = l.apiBaseURL
	}
	if base == "" {
		base = "https://api.videosdk.live"
	}
	endpoint := strings.TrimRight(base, "/") + "/v2/agent/init-config"
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", l.authToken)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("init-config failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if ok, _ := data["success"].(bool); !ok {
		msg, _ := data["message"].(string)
		return "", fmt.Errorf("init-config returned error: %s", msg)
	}
	inner, _ := data["data"].(map[string]any)
	registryURL, _ := inner["registryUrl"].(string)
	if registryURL == "" {
		return "", fmt.Errorf("init-config response missing data.registryUrl")
	}
	return registryURL, nil
}

var nonWorkerIDChars = regexp.MustCompile(`[^A-Za-z0-9_-]`)

func sanitizeAgentID(agentID string) string {
	s := nonWorkerIDChars.ReplaceAllString(agentID, "")
	if s == "" {
		return "default"
	}
	return s
}

func (l *legacyBackendRegistration) loadCachedWorkerID() string {
	envKey := "ZRT_WORKER_ID_" + strings.ToUpper(orDefault(l.agentID, "default"))
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".zrt-agents", "worker-ids", sanitizeAgentID(l.agentID)+".worker_id")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (l *legacyBackendRegistration) saveCachedWorkerID(workerID string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".zrt-agents", "worker-ids")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, sanitizeAgentID(l.agentID)+".worker_id"), []byte(workerID), 0o644)
}

func (l *legacyBackendRegistration) openWS(registryURL string) error {
	u, err := url.Parse(registryURL)
	if err != nil {
		return err
	}
	scheme := u.Scheme
	if scheme == "" {
		scheme = "wss"
	} else if strings.HasPrefix(scheme, "http") {
		scheme = strings.Replace(scheme, "http", "wss", 1)
	}
	base := strings.TrimRight(scheme+"://"+u.Host+u.Path, "/") + "/"
	agentURL := base + "agent"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+l.authToken)
	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	conn, _, err := dialer.Dial(agentURL, header)
	if err != nil {
		return err
	}

	priorWorkerID := l.loadCachedWorkerID()
	registerFrame := map[string]any{
		"type":           "register",
		"worker_id":      priorWorkerID,
		"agent_name":     l.agentID,
		"capabilities":   l.capabilities,
		"registry_uuid":  "default",
		"token":          l.authToken,
		"load_threshold": l.loadThreshold,
		"max_processes":  l.maxProcesses,
	}
	l.mu.Lock()
	l.conn = conn
	l.mu.Unlock()
	if err := l.sendJSON(registerFrame); err != nil {
		conn.Close()
		return err
	}
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, msg, err := conn.ReadMessage()
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		conn.Close()
		return err
	}
	var resp map[string]any
	if err := json.Unmarshal(msg, &resp); err != nil {
		conn.Close()
		return err
	}
	if t, _ := resp["type"].(string); t != "register" {
		conn.Close()
		return fmt.Errorf("unexpected legacy register response type: %v", resp["type"])
	}
	if ok, _ := resp["success"].(bool); !ok {
		conn.Close()
		m, _ := resp["message"].(string)
		return fmt.Errorf("legacy register rejected: %s", m)
	}
	assigned, _ := resp["worker_id"].(string)
	if assigned == "" {
		assigned = priorWorkerID
	}
	if assigned == "" {
		assigned = uuid.NewString()
	}
	l.mu.Lock()
	l.workerID = assigned
	l.mu.Unlock()
	if assigned != priorWorkerID {
		l.saveCachedWorkerID(assigned)
	}
	return nil
}

func (l *legacyBackendRegistration) runOneConnection() {
	done := make(chan struct{})
	go func() {
		l.statusLoop(done)
	}()
	l.recvLoop()
	close(done)
	// Close the connection when the recv loop ends (e.g. the peer dropped it),
	// otherwise the next reconnect overwrites l.conn and leaks this socket.
	l.mu.Lock()
	c := l.conn
	l.conn = nil
	l.mu.Unlock()
	if c != nil {
		c.Close()
	}
}

func (l *legacyBackendRegistration) recvLoop() {
	l.mu.Lock()
	conn := l.conn
	l.mu.Unlock()
	if conn == nil {
		return
	}
	// Keepalive: without a read deadline a silently dropped connection (no
	// FIN/RST, e.g. an idle NAT/load-balancer timeout) would block ReadMessage
	// forever. statusLoop sends periodic pings; each pong (and any other frame)
	// extends the deadline.
	conn.SetReadDeadline(time.Now().Add(legacyPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(legacyPongWait))
		return nil
	})
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		conn.SetReadDeadline(time.Now().Add(legacyPongWait))
		var data map[string]any
		if err := json.Unmarshal(msg, &data); err != nil {
			continue
		}
		l.handleInbound(data)
	}
}

func (l *legacyBackendRegistration) statusLoop(done <-chan struct{}) {
	ticker := time.NewTicker(legacyStatusInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.sendStatusUpdate()
			l.sendPing()
		case <-done:
			return
		case <-l.stopCh:
			return
		}
	}
}

// sendPing sends a websocket control ping so a silently half-open connection is
// detected via the recv loop's read deadline. WriteControl is safe to call
// concurrently with the loop's reads and other writes.
func (l *legacyBackendRegistration) sendPing() {
	l.mu.Lock()
	conn := l.conn
	l.mu.Unlock()
	if conn == nil {
		return
	}
	_ = conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
}

func (l *legacyBackendRegistration) handleInbound(data map[string]any) {
	switch data["type"] {
	case "availability_request":
		l.handleAvailability(data)
	case "job_assignment":
		go l.handleJobAssignment(data)
	case "job_termination":
		go l.handleJobTermination(data)
	case "pong":
	default:
		logger.Debugf("legacy: unhandled inbound type %v", data["type"])
	}
}

func (l *legacyBackendRegistration) handleAvailability(data map[string]any) {
	jobID, _ := data["job_id"].(string)
	l.mu.Lock()
	available := l.entrypoint != nil && !l.draining && !l.closed && len(l.activeJobs) < l.maxProcesses
	l.mu.Unlock()
	if available {
		_ = l.sendJSON(map[string]any{"type": "availability_response", "job_id": jobID, "available": true, "token": l.authToken})
		return
	}
	reason := "at capacity"
	if l.entrypoint == nil {
		reason = "no entrypoint registered (presence-only)"
	} else if l.draining {
		reason = "draining"
	}
	_ = l.sendJSON(map[string]any{"type": "availability_response", "job_id": jobID, "available": false, "error": reason})
}

func (l *legacyBackendRegistration) handleJobAssignment(data map[string]any) {
	jobID, _ := data["job_id"].(string)
	if l.entrypoint == nil {
		l.sendJobUpdate(jobID, "failed", "no entrypoint registered")
		return
	}
	l.mu.Lock()
	if _, exists := l.activeJobs[jobID]; exists {
		l.mu.Unlock()
		return
	}
	l.mu.Unlock()
	jobCtx := l.buildJobContext(data)
	ctx, cancel := context.WithCancel(context.Background())
	l.mu.Lock()
	l.activeJobs[jobID] = &legacyJobInfo{ctx: jobCtx, cancel: cancel}
	l.mu.Unlock()
	l.sendJobUpdate(jobID, "running", "")
	go func() {
		status, errMsg := "completed", ""
		func() {
			defer func() {
				if r := recover(); r != nil {
					status, errMsg = "failed", fmt.Sprintf("%v", r)
				}
			}()
			if err := l.entrypoint(ctx, jobCtx); err != nil {
				if ctx.Err() != nil {
					status, errMsg = "completed", "terminated"
				} else {
					status, errMsg = "failed", err.Error()
				}
			}
		}()
		jobCtx.shutdown(context.Background())
		l.mu.Lock()
		delete(l.activeJobs, jobID)
		l.mu.Unlock()
		l.sendJobUpdate(jobID, status, errMsg)
		l.sendStatusUpdate()
	}()
}

func (l *legacyBackendRegistration) handleJobTermination(data map[string]any) {
	jobID, _ := data["job_id"].(string)
	l.mu.Lock()
	info := l.activeJobs[jobID]
	l.mu.Unlock()
	if info == nil {
		logger.Warnf("legacy: job_termination for unknown job %s", jobID)
		return
	}
	info.ctx.shutdown(context.Background())
	time.AfterFunc(2*time.Second, info.cancel)
}

func (l *legacyBackendRegistration) buildJobContext(data map[string]any) *JobContext {
	var base *JobContext
	if l.jobctxFactory != nil {
		base = l.jobctxFactory()
	} else {
		base = NewJobContext(NewRoomOptions(), nil)
	}
	ro := base.RoomOptions
	if ro == nil {
		ro = NewRoomOptions()
	}
	if rid, _ := data["room_id"].(string); rid != "" {
		ro.RoomID = rid
	}
	if tok, _ := data["token"].(string); tok != "" {
		ro.AuthToken = tok
	} else if ro.AuthToken == "" {
		ro.AuthToken = l.authToken
	}
	metadata, _ := data["metadata"].(map[string]any)
	if metadata == nil {
		metadata = map[string]any{}
	}
	return NewJobContext(ro, metadata)
}

func (l *legacyBackendRegistration) sendJobUpdate(jobID, status, errMsg string) {
	frame := map[string]any{"type": "job_update", "job_id": jobID, "status": status}
	if errMsg != "" {
		frame["error"] = errMsg
	}
	_ = l.sendJSON(frame)
}

func (l *legacyBackendRegistration) sendStatusUpdate() {
	l.mu.Lock()
	jobCount := len(l.activeJobs)
	wid := l.workerID
	draining := l.draining
	l.mu.Unlock()
	load := float64(jobCount) / float64(maxInt(1, l.maxProcesses))
	if load > 1.0 {
		load = 1.0
	}
	status := "available"
	if draining {
		status = "draining"
	}
	_ = l.sendJSON(map[string]any{"type": "status_update", "worker_id": wid, "agent_name": l.agentID, "status": status, "load": load, "job_count": jobCount})
}

func (l *legacyBackendRegistration) sendJSON(payload any) error {
	l.mu.Lock()
	conn := l.conn
	l.mu.Unlock()
	if conn == nil {
		return nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	l.sendMu.Lock()
	defer l.sendMu.Unlock()
	return conn.WriteMessage(websocket.TextMessage, b)
}
