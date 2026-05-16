package components

import (
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(env string) *slog.Logger {
	return newLogger(env, os.Stdout)
}

// SetupMCPLogger logs to stderr so stdio MCP JSON-RPC on stdout stays clean.
func SetupMCPLogger(env string) *slog.Logger {
	return newLogger(env, os.Stderr)
}

func newLogger(env string, w *os.File) *slog.Logger {
	switch env {
	case envLocal:
		return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
