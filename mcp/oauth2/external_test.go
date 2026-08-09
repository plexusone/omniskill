// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testResourceMetadataURL = "http://localhost/.well-known/oauth-protected-resource"

func staticVerifier(want string, info *TokenInfo) TokenVerifier {
	return TokenVerifierFunc(func(_ context.Context, token string) (*TokenInfo, error) {
		if token != want {
			return nil, errors.New("invalid token")
		}
		return info, nil
	})
}

func TestExternalBearerMiddleware(t *testing.T) {
	info := &TokenInfo{
		AccessToken: "good-token",
		TokenType:   "Bearer",
		ClientID:    "agent-client",
		Subject:     "user:alice",
		Scope:       "docs:read",
		Actor:       []string{"agent:orchestrator", "agent:worker"},
		Claims:      map[string]any{"jti": "abc123"},
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	var gotInfo *TokenInfo
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotInfo = GetTokenInfoContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	handler := ExternalBearerMiddleware(staticVerifier("good-token", info), testResourceMetadataURL)(next)

	t.Run("missing_token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		wwwAuth := rec.Header().Get("WWW-Authenticate")
		if !strings.Contains(wwwAuth, `resource_metadata="`+testResourceMetadataURL+`"`) {
			t.Fatalf("expected resource_metadata in WWW-Authenticate, got %q", wwwAuth)
		}
	})

	t.Run("invalid_token", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer wrong-token")
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
		wwwAuth := rec.Header().Get("WWW-Authenticate")
		if !strings.Contains(wwwAuth, `error="invalid_token"`) {
			t.Fatalf("expected invalid_token error in WWW-Authenticate, got %q", wwwAuth)
		}
	})

	t.Run("valid_token", func(t *testing.T) {
		gotInfo = nil
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer good-token")
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
		if gotInfo == nil {
			t.Fatal("expected TokenInfo in context")
		}
		if gotInfo.Subject != "user:alice" {
			t.Errorf("expected subject user:alice, got %q", gotInfo.Subject)
		}
		if len(gotInfo.Actor) != 2 || gotInfo.Actor[0] != "agent:orchestrator" {
			t.Errorf("unexpected actor chain: %v", gotInfo.Actor)
		}
		if gotInfo.Claims["jti"] != "abc123" {
			t.Errorf("unexpected claims: %v", gotInfo.Claims)
		}
	})
}

func TestGetActorFromContext(t *testing.T) {
	if got := GetActorFromContext(context.Background()); got != nil {
		t.Fatalf("expected nil actor for empty context, got %v", got)
	}

	info := &TokenInfo{Subject: "user:bob", Actor: []string{"agent:calendar-bot"}}
	ctx := SetTokenInfoContext(context.Background(), info)
	got := GetActorFromContext(ctx)
	if len(got) != 1 || got[0] != "agent:calendar-bot" {
		t.Fatalf("unexpected actor chain: %v", got)
	}
}

func TestExternalProtectedResourceMetadataHandler(t *testing.T) {
	handler := ExternalProtectedResourceMetadataHandler(ExternalProtectedResourceMetadata{
		Resource:             "http://localhost:8080/mcp",
		AuthorizationServers: []string{"https://keycloak.example/realms/agents"},
		ScopesSupported:      []string{"docs:read"},
	})

	t.Run("get", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var metadata map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &metadata); err != nil {
			t.Fatalf("decoding metadata: %v", err)
		}
		if metadata["resource"] != "http://localhost:8080/mcp" {
			t.Errorf("unexpected resource: %v", metadata["resource"])
		}
		servers, ok := metadata["authorization_servers"].([]any)
		if !ok || len(servers) != 1 || servers[0] != "https://keycloak.example/realms/agents" {
			t.Errorf("unexpected authorization_servers: %v", metadata["authorization_servers"])
		}
		scopes, ok := metadata["scopes_supported"].([]any)
		if !ok || len(scopes) != 1 || scopes[0] != "docs:read" {
			t.Errorf("unexpected scopes_supported: %v", metadata["scopes_supported"])
		}
	})

	t.Run("method_not_allowed", func(t *testing.T) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/.well-known/oauth-protected-resource", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", rec.Code)
		}
	})
}
