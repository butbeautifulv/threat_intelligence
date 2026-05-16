package audit

import (
	"log/slog"
	"time"
)

// Logger records tool invocations.
type Logger struct {
	log *slog.Logger
}

func New(l *slog.Logger) *Logger {
	return &Logger{log: l}
}

func (a *Logger) ToolRun(subject, tool, target, jobID string, success bool, errMsg string) {
	if a == nil || a.log == nil {
		return
	}
	a.log.Info("engage tool run",
		slog.String("subject", subject),
		slog.String("tool", tool),
		slog.String("target", target),
		slog.String("job_id", jobID),
		slog.Bool("success", success),
		slog.String("error", errMsg),
		slog.Time("at", time.Now().UTC()),
	)
}
