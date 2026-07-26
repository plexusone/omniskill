// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	// 2 requests per second, burst of 2
	bucket := newTokenBucket(2.0, 2)

	// Should allow first 2 requests (burst)
	if !bucket.allow() {
		t.Error("first request should be allowed")
	}
	if !bucket.allow() {
		t.Error("second request should be allowed")
	}

	// Third request should be denied (bucket empty)
	if bucket.allow() {
		t.Error("third request should be denied (bucket empty)")
	}

	// Wait for token to refill
	time.Sleep(600 * time.Millisecond) // 0.6s * 2/s = 1.2 tokens

	if !bucket.allow() {
		t.Error("request after refill should be allowed")
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	rl := NewRateLimiter(&RateLimiterConfig{
		RequestsPerSecond: 1,
		BurstSize:         1,
		KeyFunc: func(r *http.Request) string {
			return "test-client"
		},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Errorf("first request: status = %d, want %d", rec1.Code, http.StatusOK)
	}

	// Second immediate request should be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: status = %d, want %d", rec2.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimiter_NoLimit(t *testing.T) {
	rl := NewRateLimiter(&RateLimiterConfig{
		RequestsPerSecond: 0, // Disabled
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	// All requests should succeed
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: status = %d, want %d", i, rec.Code, http.StatusOK)
		}
	}
}

func TestRateLimiter_PerClientLimit(t *testing.T) {
	rl := NewRateLimiter(&RateLimiterConfig{
		RequestsPerSecond: 1,
		BurstSize:         1,
		KeyFunc: func(r *http.Request) string {
			return r.Header.Get("X-Client-ID")
		},
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	// Client A's first request
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.Header.Set("X-Client-ID", "client-a")
	rec1 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("client-a first: status = %d, want %d", rec1.Code, http.StatusOK)
	}

	// Client B's first request should also succeed (separate bucket)
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Client-ID", "client-b")
	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("client-b first: status = %d, want %d", rec2.Code, http.StatusOK)
	}

	// Client A's second request should be limited
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.Header.Set("X-Client-ID", "client-a")
	rec3 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("client-a second: status = %d, want %d", rec3.Code, http.StatusTooManyRequests)
	}
}

func TestToolAuthorizer_CheckAccess(t *testing.T) {
	ta := NewToolAuthorizer(&ToolAuthorizerConfig{
		Authorizer: func(_ context.Context, clientID, toolName string) bool {
			// Only admin can use admin-tool
			if toolName == "admin-tool" {
				return clientID == "admin"
			}
			return true
		},
	})

	// Regular user can access regular tool
	ctx := ContextWithClientID(context.Background(), "user")
	if err := ta.CheckAccess(ctx, "regular-tool"); err != nil {
		t.Errorf("regular user accessing regular tool: %v", err)
	}

	// Regular user cannot access admin tool
	if err := ta.CheckAccess(ctx, "admin-tool"); err == nil {
		t.Error("regular user should not access admin-tool")
	}

	// Admin can access admin tool
	adminCtx := ContextWithClientID(context.Background(), "admin")
	if err := ta.CheckAccess(adminCtx, "admin-tool"); err != nil {
		t.Errorf("admin accessing admin-tool: %v", err)
	}
}

func TestContextWithClientID(t *testing.T) {
	ctx := context.Background()

	// No client ID initially
	if id := ClientIDFromContext(ctx); id != "" {
		t.Errorf("ClientIDFromContext = %q, want empty", id)
	}

	// Add client ID
	ctx = ContextWithClientID(ctx, "test-client")
	if id := ClientIDFromContext(ctx); id != "test-client" {
		t.Errorf("ClientIDFromContext = %q, want %q", id, "test-client")
	}
}

func TestPolicyAuthorizer(t *testing.T) {
	pa := NewPolicyAuthorizer([]PolicyRule{
		// Admin can do anything
		{ClientID: "admin", Tools: []string{"*"}, Allow: true},
		// Users can use read tools
		{ClientID: "*", Tools: []string{"read", "list"}, Allow: true},
		// Block everything else by default (implicit)
	})

	tests := []struct {
		clientID string
		tool     string
		allowed  bool
	}{
		{"admin", "anything", true},
		{"admin", "write", true},
		{"user", "read", true},
		{"user", "list", true},
		{"user", "write", false},
		{"", "read", true},
		{"", "delete", false},
	}

	for _, tt := range tests {
		t.Run(tt.clientID+"-"+tt.tool, func(t *testing.T) {
			got := pa.Authorize(context.Background(), tt.clientID, tt.tool)
			if got != tt.allowed {
				t.Errorf("Authorize(%q, %q) = %v, want %v",
					tt.clientID, tt.tool, got, tt.allowed)
			}
		})
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(&RateLimiterConfig{
		RequestsPerSecond: 100,
		BurstSize:         10,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := rl.Middleware()(handler)

	var wg sync.WaitGroup
	var okCount, limitedCount int
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			mu.Lock()
			if rec.Code == http.StatusOK {
				okCount++
			} else if rec.Code == http.StatusTooManyRequests {
				limitedCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// With burst of 10 and 50 concurrent requests, expect some to be limited
	if okCount == 0 {
		t.Error("expected some requests to succeed")
	}
	if limitedCount == 0 {
		t.Error("expected some requests to be rate limited")
	}
}
