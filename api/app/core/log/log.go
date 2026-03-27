package log

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"localhost/app/core/config"
	"localhost/app/core/config/rule"
)

type state struct {
	mu     sync.Mutex
	writer *dailyFileWriter
}

var global state

// Load creates the dual-output structured logger and sets it as the slog
// default. It must be called after [config.Load].
//
// Configuration keys:
//
//   - "log.level" (env LOG_LEVEL, default "info") — minimum level for both
//     console and file sinks.
//   - "log.format" (env LOG_FORMAT, default "json") — console output format.
//     "text" uses slog.TextHandler (human-readable, good for development).
//     "json" uses slog.JSONHandler (machine-readable, good for production).
//     File output is always JSONL regardless of this setting.
//   - "log.retention_days" (env LOG_RETENTION_DAYS, default 30) — number of
//     days to keep log files. Files older than this are deleted on rotation.
//
// Both sinks include source location. On invalid config the function panics.
func Load() {
	config.SetDefaults(config.Values{
		"log.level":          "info",
		"log.format":         "json",
		"log.retention_days": 30,
	})

	config.SetRule("log.level", rule.Required, rule.InFold("debug", "info", "warn", "warning", "error"))
	config.SetRule("log.format", rule.Required, rule.InFold("text", "json"))
	config.SetRule("log.retention_days", rule.Required, rule.Between(1, 365))

	var level slog.LevelVar
	if err := parseLevel(&level, config.GetString("log.level")); err != nil {
		panic(fmt.Sprintf("log: %v", err))
	}

	consoleHandler := newConsoleHandler(config.GetString("log.format"), &level)

	logsDir := filepath.Join(config.GetString("data_dir"), "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		panic(fmt.Sprintf("log: creating logs directory: %v", err))
	}

	writer := newDailyFileWriter(logsDir, config.GetInt("log.retention_days"))
	if prev := swapWriter(writer); prev != nil {
		_ = prev.Close()
	}

	fileHandler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level:     &level,
		AddSource: true,
	})

	root := newContextHandler(newMergedHandler(consoleHandler, fileHandler))
	slog.SetDefault(slog.New(root))
}

// FxPrinter returns an fx.Option that redirects fx's internal logging through
// the slog default logger. Call after Load so the configured handler is active.
func FxPrinter() fx.Option {
	return fx.WithLogger(func() fxevent.Logger {
		return &fxevent.SlogLogger{Logger: slog.Default()}
	})
}

// Flush writes any buffered log data to the underlying file. Safe to call
// concurrently; returns nil when no writer is active.
func Flush() error {
	global.mu.Lock()
	writer := global.writer
	global.mu.Unlock()
	if writer != nil {
		return writer.Flush()
	}
	return nil
}

// Close flushes and closes the file writer. Safe to call multiple times.
func Close() error {
	global.mu.Lock()
	writer := global.writer
	global.writer = nil
	global.mu.Unlock()
	if writer != nil {
		return writer.Close()
	}
	return nil
}

func swapWriter(next *dailyFileWriter) *dailyFileWriter {
	global.mu.Lock()
	prev := global.writer
	global.writer = next
	global.mu.Unlock()
	return prev
}

func newConsoleHandler(format string, level *slog.LevelVar) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "text":
		return slog.NewTextHandler(os.Stdout, opts)
	case "json":
		return slog.NewJSONHandler(os.Stdout, opts)
	default:
		panic(fmt.Sprintf("log: unknown format %q (expected \"text\" or \"json\")", format))
	}
}

func parseLevel(lv *slog.LevelVar, s string) error {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "DEBUG":
		lv.Set(slog.LevelDebug)
	case "INFO":
		lv.Set(slog.LevelInfo)
	case "WARN", "WARNING":
		lv.Set(slog.LevelWarn)
	case "ERROR":
		lv.Set(slog.LevelError)
	default:
		return fmt.Errorf("unknown level %q", s)
	}
	return nil
}
