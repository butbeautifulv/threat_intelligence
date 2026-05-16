package components

import (
	"log/slog"
	"os"
)

func SetupLogger(env string) *slog.Logger {
	return newLogger(env, os.Stdout)
}

func SetupMCPLogger(env string) *slog.Logger {
	return newLogger(env, os.Stderr)
}

func newLogger(env string, w *os.File) *slog.Logger {
	switch env {
	case "local":
		return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "dev":
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
