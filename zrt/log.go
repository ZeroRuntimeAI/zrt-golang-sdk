package zrt

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

// LogLevel controls SDK log verbosity.
type LogLevel int32

const (
	// LevelDebug logs everything, including debug detail.
	LevelDebug LogLevel = iota
	// LevelInfo logs informational messages and above.
	LevelInfo
	// LevelWarning logs warnings and above.
	LevelWarning
	// LevelError logs errors and above.
	LevelError
	// LevelCritical logs only critical messages.
	LevelCritical
)

// Logger is the SDK logging interface.
//
// Prefer wiring a standard library *slog.Logger via WorkerOptions.Logger. Use
// this interface to adapt a non-slog logger.
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// slogLogger adapts a *slog.Logger to the SDK Logger interface.
type slogLogger struct{ l *slog.Logger }

// NewSlogLogger adapts a standard library *slog.Logger to the SDK Logger
// interface. A nil logger falls back to slog.Default().
func NewSlogLogger(l *slog.Logger) Logger {
	if l == nil {
		l = slog.Default()
	}
	return &slogLogger{l: l}
}

func (s *slogLogger) logf(lvl slog.Level, format string, args ...any) {
	if !s.l.Enabled(context.Background(), lvl) {
		return
	}
	s.l.Log(context.Background(), lvl, fmt.Sprintf(format, args...))
}
func (s *slogLogger) Debugf(format string, args ...any) { s.logf(slog.LevelDebug, format, args...) }
func (s *slogLogger) Infof(format string, args ...any)  { s.logf(slog.LevelInfo, format, args...) }
func (s *slogLogger) Warnf(format string, args ...any)  { s.logf(slog.LevelWarn, format, args...) }
func (s *slogLogger) Errorf(format string, args ...any) { s.logf(slog.LevelError, format, args...) }

type stdLogger struct {
	level int32 // LogLevel
	out   *log.Logger
}

func newStdLogger() *stdLogger {
	return &stdLogger{level: int32(LevelInfo), out: log.New(os.Stdout, "", log.LstdFlags)}
}

func (l *stdLogger) setLevel(lvl LogLevel) { atomic.StoreInt32(&l.level, int32(lvl)) }
func (l *stdLogger) enabled(lvl LogLevel) bool {
	return int32(lvl) >= atomic.LoadInt32(&l.level)
}
func (l *stdLogger) logf(lvl LogLevel, tag, format string, args ...any) {
	if !l.enabled(lvl) {
		return
	}
	l.out.Output(3, fmt.Sprintf("zrt - "+tag+" - "+format, args...))
}
func (l *stdLogger) Debugf(format string, args ...any) { l.logf(LevelDebug, "DEBUG", format, args...) }
func (l *stdLogger) Infof(format string, args ...any)  { l.logf(LevelInfo, "INFO", format, args...) }
func (l *stdLogger) Warnf(format string, args ...any) {
	l.logf(LevelWarning, "WARNING", format, args...)
}
func (l *stdLogger) Errorf(format string, args ...any) { l.logf(LevelError, "ERROR", format, args...) }

var defaultLogger = newStdLogger()

// logger is the package-level logger used throughout the SDK.
var logger Logger = defaultLogger

// SetLogger replaces the process-wide SDK logger.
//
// Deprecated: prefer WorkerOptions.Logger (a *slog.Logger) so logging is
// configured per worker rather than through global mutable state. To route the
// global logger through slog, use SetLogger(NewSlogLogger(l)).
func SetLogger(l Logger) {
	if l != nil {
		logger = l
	}
}

// SetLogLevel sets the verbosity of the default logger. Accepts a LogLevel.
func SetLogLevel(level LogLevel) { defaultLogger.setLevel(level) }

// SetLogLevelString sets the level from a string ("DEBUG", "INFO", ...).
func SetLogLevelString(level string) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		defaultLogger.setLevel(LevelDebug)
	case "INFO":
		defaultLogger.setLevel(LevelInfo)
	case "WARNING", "WARN":
		defaultLogger.setLevel(LevelWarning)
	case "ERROR":
		defaultLogger.setLevel(LevelError)
	case "CRITICAL":
		defaultLogger.setLevel(LevelCritical)
	}
}
