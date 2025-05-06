package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type ctxKey string

const (
	traceKey  ctxKey = "traceID"
	loggerKey ctxKey = "logger"
)

// getTraceID reads X-Trace-ID or makes a new one
func getTraceID(r *http.Request) (context.Context, string) {
	id := r.Header.Get("X-Trace-ID")
	slog.Debug("Header.Get X-Trace-ID", "X-Trace-ID", id)
	if id == "" {
		id = fmt.Sprintf("trace-%d", time.Now().UnixNano())
	}
	return context.WithValue(r.Context(), traceKey, id), id
}

// WithLoggingAndTrace injects traceID + slog.Logger into every request
func WithLoggingAndTrace(next http.Handler) http.Handler {
	base := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, tid := getTraceID(r)
		w.Header().Set("X-Trace-ID", tid)
		logger := base.With("traceID", tid)
		ctx = context.WithValue(ctx, loggerKey, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerFromContext retrieves the request‐scoped logger
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if lg, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return lg
	}
	return slog.Default()
}
