// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plexusone/omniskill/mcp/oauth2"
)

func newExternalAuthTestRuntime() *Runtime {
	return New(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}, nil)
}

func testExternalVerifier() oauth2.TokenVerifier {
	return oauth2.TokenVerifierFunc(func(_ context.Context, token string) (*oauth2.TokenInfo, error) {
		if token != "valid-external-token" {
			return nil, errors.New("invalid token")
		}
		return &oauth2.TokenInfo{
			AccessToken: token,
			TokenType:   "Bearer",
			Subject:     "user:alice",
			Actor:       []string{"agent:worker"},
			ExpiresAt:   time.Now().Add(time.Hour),
		}, nil
	})
}

func TestServeHTTP_ExternalAuthValidation(t *testing.T) {
	tests := []struct {
		name    string
		opts    *HTTPServerOptions
		wantErr string
	}{
		{
			name: "mutually_exclusive_with_oauth2",
			opts: &HTTPServerOptions{
				Addr: "localhost:0",
				ExternalAuth: &ExternalAuthOptions{
					Verifier:             testExternalVerifier(),
					AuthorizationServers: []string{"https://as.example"},
				},
				OAuth2: &OAuth2Options{Users: map[string]string{"u": "p"}},
			},
			wantErr: "mutually exclusive",
		},
		{
			name: "missing_verifier",
			opts: &HTTPServerOptions{
				Addr: "localhost:0",
				ExternalAuth: &ExternalAuthOptions{
					AuthorizationServers: []string{"https://as.example"},
				},
			},
			wantErr: "Verifier is required",
		},
		{
			name: "missing_authorization_servers",
			opts: &HTTPServerOptions{
				Addr: "localhost:0",
				ExternalAuth: &ExternalAuthOptions{
					Verifier: testExternalVerifier(),
				},
			},
			wantErr: "AuthorizationServers is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := newExternalAuthTestRuntime()
			_, err := rt.ServeHTTP(context.Background(), tt.opts)
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestServeHTTP_ExternalAuth(t *testing.T) {
	rt := newExternalAuthTestRuntime()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan *HTTPServerResult, 1)
	errChan := make(chan error, 1)
	ready := make(chan *HTTPServerResult, 1)

	go func() {
		result, err := rt.ServeHTTP(ctx, &HTTPServerOptions{
			Addr: "localhost:0",
			ExternalAuth: &ExternalAuthOptions{
				Verifier:             testExternalVerifier(),
				AuthorizationServers: []string{"https://keycloak.example/realms/agents"},
				ScopesSupported:      []string{"docs:read"},
			},
			OnReady: func(result *HTTPServerResult) {
				ready <- result
			},
		})
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}()

	var result *HTTPServerResult
	select {
	case result = <-ready:
	case err := <-errChan:
		t.Fatalf("server failed to start: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server to start")
	}

	baseURL := "http://" + result.LocalAddr

	t.Run("metadata_advertises_external_as", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/.well-known/oauth-protected-resource")
		if err != nil {
			t.Fatalf("fetching metadata: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("closing body: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var metadata map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
			t.Fatalf("decoding metadata: %v", err)
		}
		if metadata["resource"] != result.LocalURL {
			t.Errorf("expected resource %q, got %v", result.LocalURL, metadata["resource"])
		}
		servers, ok := metadata["authorization_servers"].([]any)
		if !ok || len(servers) != 1 || servers[0] != "https://keycloak.example/realms/agents" {
			t.Errorf("unexpected authorization_servers: %v", metadata["authorization_servers"])
		}
	})

	t.Run("unauthorized_without_token", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, result.LocalURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("closing body: %v", err)
			}
		}()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
		wwwAuth := resp.Header.Get("WWW-Authenticate")
		wantMeta := `resource_metadata="` + baseURL + `/.well-known/oauth-protected-resource"`
		if !strings.Contains(wwwAuth, wantMeta) {
			t.Fatalf("expected WWW-Authenticate to contain %q, got %q", wantMeta, wwwAuth)
		}
	})

	t.Run("authorized_with_valid_token", func(t *testing.T) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, result.LocalURL, strings.NewReader(`{}`))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer valid-external-token")
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("closing body: %v", err)
			}
		}()
		// The token passed verification if we got past the auth middleware:
		// anything but 401 means the MCP handler received the request.
		if resp.StatusCode == http.StatusUnauthorized {
			t.Fatal("expected authenticated request to pass auth middleware, got 401")
		}
	})

	cancel()
	select {
	case <-resultChan:
	case err := <-errChan:
		t.Fatalf("server error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server shutdown")
	}
}
