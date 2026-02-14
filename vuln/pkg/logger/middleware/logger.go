package middleware

import (
    "log/slog"
    "net/http"
    "time"
)

func Log(log *slog.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        log := log.With(
            slog.String("component", "middleware/logger"),
        )

        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            ww := &responseWriter{ResponseWriter: w, status: 200}

            next.ServeHTTP(ww, r)

            log.Info("request completed",
                slog.String("method", r.Method),
                slog.String("path", r.URL.Path),
                slog.String("remote_addr", r.RemoteAddr),
                slog.String("user_agent", r.UserAgent()),
                slog.Int("status", ww.status),
                slog.Int("bytes", ww.bytes),
                slog.String("duration", time.Since(start).String()),
            )
        })
    }
}

type responseWriter struct {
    http.ResponseWriter
    status int
    bytes  int
}

func (w *responseWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
    n, err := w.ResponseWriter.Write(b)
    w.bytes += n
    return n, err
}

func WithLogging(log *slog.Logger, next func(ctx context.Context) error) func(ctx context.Context) error {
    return func(ctx context.Context) error {
        start := time.Now()
        log.Info("scraper started")

        err := next(ctx)

        log.Info("scraper finished",
            slog.String("duration", time.Since(start).String()),
            slog.Any("error", err),
        )

        return err
    }
}
