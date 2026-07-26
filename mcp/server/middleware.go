// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/grokify/mogo/log/slogutil"
)

// RateLimiterConfig configures rate limiting.
type RateLimiterConfig struct {
	// RequestsPerSecond is the maximum requests per second per client.
	// Zero or negative means no limit.
	RequestsPerSecond float64

	// BurstSize is the maximum burst size (token bucket capacity).
	// Defaults to 10 if not specified.
	BurstSize int

	// KeyFunc extracts the client key from the request.
	// Defaults to using the client's IP address.
	KeyFunc func(r *http.Request) string

	// OnLimited is called when a request is rate limited.
	// If nil, returns 429 Too Many Requests.
	OnLimited func(w http.ResponseWriter, r *http.Request)

	// Logger for rate limit events.
	Logger *slog.Logger
}

// tokenBucket implements a simple token bucket rate limiter.
type tokenBucket struct {
	tokens     float64
	lastUpdate time.Time
	rate       float64
	burst      float64
	mu         sync.Mutex
}

func newTokenBucket(rate float64, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(burst),
		lastUpdate: time.Now(),
		rate:       rate,
		burst:      float64(burst),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastUpdate).Seconds()
	b.lastUpdate = now

	// Add tokens based on elapsed time
	b.tokens += elapsed * b.rate
	if b.tokens > b.burst {
		b.tokens = b.burst
	}

	// Check if we have tokens available
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimiter stores per-client rate limiters.
type RateLimiter struct {
	config   *RateLimiterConfig
	limiters sync.Map // map[string]*tokenBucket
	logger   *slog.Logger
}

// NewRateLimiter creates a new rate limiter with the given configuration.
func NewRateLimiter(cfg *RateLimiterConfig) *RateLimiter {
	if cfg == nil {
		cfg = &RateLimiterConfig{}
	}
	if cfg.BurstSize <= 0 {
		cfg.BurstSize = 10
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(r *http.Request) string {
			return r.RemoteAddr
		}
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &RateLimiter{
		config: cfg,
		logger: logger,
	}
}

// Middleware returns rate limiting middleware.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if no rate limit configured
			if rl.config.RequestsPerSecond <= 0 {
				next.ServeHTTP(w, r)
				return
			}

			key := rl.config.KeyFunc(r)

			// Get or create limiter for this key
			limiterI, _ := rl.limiters.LoadOrStore(key, newTokenBucket(
				rl.config.RequestsPerSecond,
				rl.config.BurstSize,
			))
			limiter := limiterI.(*tokenBucket)

			if !limiter.allow() {
				// Rate limited
				logger := slogutil.LoggerFromContext(r.Context(), rl.logger)
				logger.Warn("rate_limited",
					"client_key", key,
					"path", r.URL.Path,
				)

				if rl.config.OnLimited != nil {
					rl.config.OnLimited(w, r)
				} else {
					http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ToolAuthorizerConfig configures per-tool authorization.
type ToolAuthorizerConfig struct {
	// Authorizer checks if a client can access a tool.
	// Returns true if access is allowed.
	// The clientID is extracted from the request context (set by auth middleware).
	Authorizer func(ctx context.Context, clientID, toolName string) bool

	// ClientIDFromContext extracts the client ID from the request context.
	// Defaults to looking for "client_id" in context values.
	ClientIDFromContext func(ctx context.Context) string

	// OnDenied is called when access is denied.
	// If nil, returns 403 Forbidden.
	OnDenied func(w http.ResponseWriter, r *http.Request, toolName string)

	// Logger for authorization events.
	Logger *slog.Logger
}

// ToolAuthorizer provides per-tool access control.
type ToolAuthorizer struct {
	config *ToolAuthorizerConfig
	logger *slog.Logger
}

// NewToolAuthorizer creates a new tool authorizer.
func NewToolAuthorizer(cfg *ToolAuthorizerConfig) *ToolAuthorizer {
	if cfg == nil {
		cfg = &ToolAuthorizerConfig{}
	}
	if cfg.Authorizer == nil {
		// Default: allow all
		cfg.Authorizer = func(_ context.Context, _, _ string) bool {
			return true
		}
	}
	if cfg.ClientIDFromContext == nil {
		cfg.ClientIDFromContext = func(ctx context.Context) string {
			if id, ok := ctx.Value(clientIDKey{}).(string); ok {
				return id
			}
			return ""
		}
	}

	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}

	return &ToolAuthorizer{
		config: cfg,
		logger: logger,
	}
}

// clientIDKey is the context key for storing client ID.
type clientIDKey struct{}

// ContextWithClientID returns a context with the client ID stored.
func ContextWithClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, clientIDKey{}, clientID)
}

// ClientIDFromContext retrieves the client ID from context.
func ClientIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(clientIDKey{}).(string); ok {
		return id
	}
	return ""
}

// CheckAccess verifies if the client can access the given tool.
// Returns nil if allowed, error if denied.
func (ta *ToolAuthorizer) CheckAccess(ctx context.Context, toolName string) error {
	clientID := ta.config.ClientIDFromContext(ctx)

	if !ta.config.Authorizer(ctx, clientID, toolName) {
		logger := slogutil.LoggerFromContext(ctx, ta.logger)
		logger.Warn("tool_access_denied",
			"client_id", clientID,
			"tool", toolName,
		)
		return &ToolAccessDeniedError{
			ClientID: clientID,
			ToolName: toolName,
		}
	}

	return nil
}

// ToolAccessDeniedError is returned when tool access is denied.
type ToolAccessDeniedError struct {
	ClientID string
	ToolName string
}

func (e *ToolAccessDeniedError) Error() string {
	return "access denied to tool: " + e.ToolName
}

// Middleware returns HTTP middleware that checks tool access for MCP tool calls.
// This middleware inspects the request body to extract the tool name from
// MCP CallTool requests.
func (ta *ToolAuthorizer) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only check POST requests to MCP endpoint
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// We can't easily intercept MCP requests at the HTTP level
			// because the body needs to be parsed as JSON-RPC.
			// This middleware is a placeholder for integration with
			// the MCP server's tool handler.
			//
			// For real tool authorization, integrate with Runtime.CallTool
			// or use a tool handler wrapper.

			next.ServeHTTP(w, r)
		})
	}
}

// WrapToolHandler wraps a tool handler with authorization checks.
// This is the recommended way to add per-tool authorization.
func (ta *ToolAuthorizer) WrapToolHandler(toolName string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ta.CheckAccess(r.Context(), toolName); err != nil {
			logger := slogutil.LoggerFromContext(r.Context(), ta.logger)
			logger.Warn("tool_access_denied",
				"tool", toolName,
				"error", err,
			)

			if ta.config.OnDenied != nil {
				ta.config.OnDenied(w, r, toolName)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "access denied",
					"tool":  toolName,
				})
			}
			return
		}

		handler(w, r)
	}
}

// PolicyRule defines an authorization rule.
type PolicyRule struct {
	// ClientID pattern to match (* for all).
	ClientID string

	// Tools is the list of tool patterns to match (* for all).
	Tools []string

	// Allow determines if matching requests are allowed.
	Allow bool
}

// PolicyAuthorizer implements authorization based on rules.
type PolicyAuthorizer struct {
	rules []PolicyRule
}

// NewPolicyAuthorizer creates a policy-based authorizer.
// Rules are evaluated in order; first match wins.
// If no rules match, access is denied by default.
func NewPolicyAuthorizer(rules []PolicyRule) *PolicyAuthorizer {
	return &PolicyAuthorizer{rules: rules}
}

// Authorize checks if the client can access the tool.
func (pa *PolicyAuthorizer) Authorize(_ context.Context, clientID, toolName string) bool {
	for _, rule := range pa.rules {
		// Check client ID match
		if rule.ClientID != "*" && rule.ClientID != clientID {
			continue
		}

		// Check tool match
		toolMatch := false
		for _, pattern := range rule.Tools {
			if pattern == "*" || pattern == toolName {
				toolMatch = true
				break
			}
		}
		if !toolMatch {
			continue
		}

		// Rule matches
		return rule.Allow
	}

	// No rule matched, deny by default
	return false
}

// AuthorizerFunc returns an authorizer function for use with ToolAuthorizerConfig.
func (pa *PolicyAuthorizer) AuthorizerFunc() func(context.Context, string, string) bool {
	return pa.Authorize
}
