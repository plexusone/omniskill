// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	wrapped := LoggingMiddleware(logger)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Check that log output contains expected fields
	output := buf.String()
	if output == "" {
		t.Fatal("no log output")
	}

	// Parse the completion log entry (second line)
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	if len(lines) < 2 {
		t.Fatal("expected at least 2 log lines")
	}

	var entry map[string]any
	if err := json.Unmarshal(lines[1], &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v", err)
	}

	if entry["msg"] != "request completed" {
		t.Errorf("msg = %v, want %q", entry["msg"], "request completed")
	}
	if entry["method"] != "GET" {
		t.Errorf("method = %v, want %q", entry["method"], "GET")
	}
	if entry["path"] != "/test" {
		t.Errorf("path = %v, want %q", entry["path"], "/test")
	}
	if entry["status"] != float64(200) {
		t.Errorf("status = %v, want %v", entry["status"], 200)
	}
}

func TestLoggingMiddleware_ErrorStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	wrapped := LoggingMiddleware(logger)(handler)

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	// Check that error status is logged at ERROR level
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	if len(lines) < 1 {
		t.Fatal("expected log output")
	}

	var entry map[string]any
	if err := json.Unmarshal(lines[0], &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v", err)
	}

	if entry["level"] != "ERROR" {
		t.Errorf("level = %v, want %q", entry["level"], "ERROR")
	}
}

func TestLoggerMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The middleware should have injected a logger into context
		w.WriteHeader(http.StatusOK)
	})

	wrapped := LoggerMiddleware(logger)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "test-123")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestLogToolCall(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	LogToolCall(ctx, logger, "test-tool", true, 100*time.Millisecond, nil)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v", err)
	}

	if entry["msg"] != "tool_call" {
		t.Errorf("msg = %v, want %q", entry["msg"], "tool_call")
	}
	if entry["tool"] != "test-tool" {
		t.Errorf("tool = %v, want %q", entry["tool"], "test-tool")
	}
	if entry["success"] != true {
		t.Errorf("success = %v, want %v", entry["success"], true)
	}
}

func TestLogAuth(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := context.Background()
	LogAuth(ctx, logger, "token_issued", true, "client-123", nil)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v", err)
	}

	if entry["msg"] != "auth" {
		t.Errorf("msg = %v, want %q", entry["msg"], "auth")
	}
	if entry["event"] != "token_issued" {
		t.Errorf("event = %v, want %q", entry["event"], "token_issued")
	}
	if entry["client_id"] != "client-123" {
		t.Errorf("client_id = %v, want %q", entry["client_id"], "client-123")
	}
}

func TestResponseWriter_Unwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	w := &responseWriter{ResponseWriter: rec}

	if w.Unwrap() != rec {
		t.Error("Unwrap should return underlying ResponseWriter")
	}
}
