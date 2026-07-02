package zrt

import (
	"cmp"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"time"
)

// debugEndpoints is the set of paths the debug server exposes, echoed back by the
// /debug and /api/status endpoints for discoverability.
var debugEndpoints = []string{"/", "/health", "/worker", "/stats", "/debug", "/api/status", "/debug/runners/"}

// debugServer is a lightweight HTTP endpoint exposing worker health and status
// while a worker runs.
type debugServer struct {
	host   string
	port   int
	worker *WorkerJob
	srv    *http.Server
}

func newDebugServer(host string, port int, worker *WorkerJob) *debugServer {
	return &debugServer{host: host, port: port, worker: worker}
}

// startDebugServer starts the debug endpoint when enabled, returning nil (and
// logging) when disabled or when the listener cannot bind — a failed debug
// endpoint must never take the worker down.
func (w *WorkerJob) startDebugServer() *debugServer {
	if !w.options.DebugEnabled {
		return nil
	}
	host := cmp.Or(w.options.Host, "0.0.0.0")
	port := w.options.Port
	if port == 0 {
		port = 8081
	}
	d := newDebugServer(host, port, w)
	if err := d.start(); err != nil {
		logger.Warnf("debug endpoint disabled: failed to listen on %s: %v", net.JoinHostPort(host, strconv.Itoa(port)), err)
		return nil
	}
	return d
}

func (w *WorkerJob) stopDebugServer(d *debugServer) {
	if d != nil {
		d.stop()
	}
}

func (d *debugServer) start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", d.handleHealth)
	mux.HandleFunc("/worker", d.handleWorker)
	mux.HandleFunc("/stats", d.handleStats)
	mux.HandleFunc("/debug/runners/", d.handleRunners)
	mux.HandleFunc("/debug", d.handleDebugInfo)
	mux.HandleFunc("/api/status", d.handleAPIStatus)
	mux.HandleFunc("/", d.handleDashboard)

	addr := net.JoinHostPort(d.host, strconv.Itoa(d.port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	d.srv = &http.Server{Handler: mux}
	go func() {
		if err := d.srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			logger.Warnf("debug server stopped: %v", err)
		}
	}()
	logger.Infof("Debug endpoint listening on http://%s", addr)
	return nil
}

func (d *debugServer) stop() {
	if d.srv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = d.srv.Shutdown(ctx)
}

func writeDebugJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (d *debugServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("OK"))
}

func (d *debugServer) handleWorker(w http.ResponseWriter, _ *http.Request) {
	writeDebugJSON(w, d.worker.getStats())
}

func (d *debugServer) handleStats(w http.ResponseWriter, _ *http.Request) {
	writeDebugJSON(w, d.worker.getStats())
}

func (d *debugServer) handleRunners(w http.ResponseWriter, _ *http.Request) {
	writeDebugJSON(w, map[string]any{"runners": d.worker.runnersSnapshot()})
}

func (d *debugServer) handleDebugInfo(w http.ResponseWriter, _ *http.Request) {
	stats := d.worker.getStats()
	writeDebugJSON(w, map[string]any{
		"server": map[string]any{"host": d.host, "port": d.port, "endpoints": debugEndpoints},
		"worker": map[string]any{
			"available": true,
			"options": map[string]any{
				"agent_id":      stats["agent_id"],
				"register":      stats["register"],
				"max_processes": stats["max_processes"],
			},
		},
	})
}

func (d *debugServer) handleAPIStatus(w http.ResponseWriter, _ *http.Request) {
	stats := d.worker.getStats()
	writeDebugJSON(w, map[string]any{
		"worker": map[string]any{
			"available": true,
			"connected": stats["connected"],
			"worker_id": stats["worker_id"],
			"options": map[string]any{
				"agent_id":      stats["agent_id"],
				"register":      stats["register"],
				"max_processes": stats["max_processes"],
				"log_level":     stats["log_level"],
			},
		},
		"stats":     stats,
		"server":    map[string]any{"host": d.host, "port": d.port, "endpoints": debugEndpoints},
		"timestamp": time.Now().Unix(),
	})
}

func (d *debugServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Only the exact root renders the dashboard; unknown paths 404 so typos are visible.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(debugDashboardHTML))
}

const debugDashboardHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>ZRT Worker</title>
<style>
  body{font-family:system-ui,sans-serif;margin:2rem;color:#1a1a2e}
  h1{font-size:1.25rem}
  pre{background:#0f1117;color:#e6e6e6;padding:1rem;border-radius:8px;overflow:auto}
  a{color:#4f6bed}
</style>
</head>
<body>
<h1>ZRT Worker</h1>
<p>Endpoints: <a href="/health">/health</a> · <a href="/worker">/worker</a> ·
<a href="/stats">/stats</a> · <a href="/debug">/debug</a> ·
<a href="/api/status">/api/status</a> · <a href="/debug/runners/">/debug/runners/</a></p>
<pre id="out">loading…</pre>
<script>
  async function refresh(){
    try{
      const r = await fetch('/api/status');
      document.getElementById('out').textContent = JSON.stringify(await r.json(), null, 2);
    }catch(e){ document.getElementById('out').textContent = String(e); }
  }
  refresh(); setInterval(refresh, 2000);
</script>
</body>
</html>`
