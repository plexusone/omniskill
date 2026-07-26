// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/grokify/mogo/log/slogutil"
)

// LoggingMiddleware returns middleware that logs HTTP requests.
//
// It logs request start (at DEBUG level) and completion (at INFO level),
// including method, path, status code, and duration.
//
// The logger is retrieved from context if available (via slogutil.LoggerFromContext),
// falling back to the provided logger or slog.Default().
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get logger from context or use provided
			reqLogger := slogutil.LoggerFromContext(r.Context(), logger)

			// Log request start at debug level
			reqLogger.Debug("request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			// Wrap response writer to capture status
			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			level := slog.LevelInfo
			if wrapped.status >= 500 {
				level = slog.LevelError
			} else if wrapped.status >= 400 {
				level = slog.LevelWarn
			}

			reqLogger.Log(r.Context(), level, "request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"duration_ms", duration.Milliseconds(),
				"bytes", wrapped.bytes,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

// WriteHeader captures the status code.
func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Write captures bytes written.
func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// Unwrap returns the underlying ResponseWriter for middleware compatibility.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// LoggerMiddleware injects a logger into the request context.
//
// This allows handlers to retrieve a request-scoped logger via
// slogutil.LoggerFromContext(ctx) with request-specific attributes
// already attached.
func LoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create request-scoped logger with request ID if available
			reqLogger := logger.With(
				"method", r.Method,
				"path", r.URL.Path,
			)

			// Add request ID if present in header
			if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
				reqLogger = reqLogger.With("request_id", reqID)
			}

			// Store in context
			ctx := slogutil.ContextWithLogger(r.Context(), reqLogger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LogEvent logs a structured event with the given level.
func LogEvent(ctx context.Context, logger *slog.Logger, level slog.Level, event string, attrs ...any) {
	l := slogutil.LoggerFromContext(ctx, logger)
	l.Log(ctx, level, event, attrs...)
}

// LogToolCall logs a tool invocation.
func LogToolCall(ctx context.Context, logger *slog.Logger, toolName string, success bool, duration time.Duration, err error) {
	l := slogutil.LoggerFromContext(ctx, logger)

	attrs := []any{
		"tool", toolName,
		"success", success,
		"duration_ms", duration.Milliseconds(),
	}

	level := slog.LevelInfo
	if err != nil {
		level = slog.LevelError
		attrs = append(attrs, "error", err.Error())
	}

	l.Log(ctx, level, "tool_call", attrs...)
}

// LogAuth logs an authentication event.
func LogAuth(ctx context.Context, logger *slog.Logger, event string, success bool, clientID string, err error) {
	l := slogutil.LoggerFromContext(ctx, logger)

	attrs := []any{
		"event", event,
		"success", success,
	}

	if clientID != "" {
		attrs = append(attrs, "client_id", clientID)
	}

	level := slog.LevelInfo
	if !success {
		level = slog.LevelWarn
		if err != nil {
			attrs = append(attrs, "error", err.Error())
		}
	}

	l.Log(ctx, level, "auth", attrs...)
}
