// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// TokenVerifier validates an externally-issued access token (for example a
// JWT minted by an enterprise IdP or authorization server such as Keycloak)
// and returns the token's identity information.
//
// Implementations own all protocol-specific verification: signature checks
// against the issuer's JWKS, issuer/audience validation, expiry, and claim
// extraction. The oauth2 package stays protocol-agnostic; adapters such as
// github.com/aistandardsio/agent-protocols/adapters/omniskill supply
// ID-JAG-aware implementations.
type TokenVerifier interface {
	// VerifyToken validates the raw bearer token and returns its TokenInfo.
	// Implementations must return an error for invalid, expired, or
	// mis-audienced tokens.
	VerifyToken(ctx context.Context, token string) (*TokenInfo, error)
}

// TokenVerifierFunc adapts a function to the TokenVerifier interface.
type TokenVerifierFunc func(ctx context.Context, token string) (*TokenInfo, error)

// VerifyToken implements TokenVerifier.
func (f TokenVerifierFunc) VerifyToken(ctx context.Context, token string) (*TokenInfo, error) {
	return f(ctx, token)
}

// ExternalBearerMiddleware returns middleware that validates Bearer tokens
// using an external TokenVerifier and stores the resulting TokenInfo in the
// request context. This makes the wrapped handler a pure OAuth resource
// server: no local authorization server is involved.
//
// Unauthorized responses carry a WWW-Authenticate header pointing at the
// RFC 9728 protected resource metadata URL, per the MCP authorization
// specification, so MCP clients can discover the authorization servers.
func ExternalBearerMiddleware(verifier TokenVerifier, resourceMetadataURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer resource_metadata="%s"`, resourceMetadataURL))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			tokenInfo, err := verifier.VerifyToken(ctx, token)
			if err != nil {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer resource_metadata="%s", error="invalid_token"`, resourceMetadataURL))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx = SetTokenInfoContext(ctx, tokenInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ExternalProtectedResourceMetadata describes a protected resource whose
// authorization is delegated to external authorization servers (RFC 9728).
type ExternalProtectedResourceMetadata struct {
	// Resource is the protected resource identifier (typically the MCP
	// endpoint URL). Also used as the expected token audience by verifiers.
	Resource string

	// AuthorizationServers lists the issuer URLs of the external
	// authorization servers that can issue tokens for this resource.
	AuthorizationServers []string

	// ScopesSupported optionally lists the scopes this resource understands.
	ScopesSupported []string
}

// ExternalProtectedResourceMetadataHandler serves RFC 9728 protected resource
// metadata for a resource governed by external authorization servers. Mount
// at /.well-known/oauth-protected-resource.
func ExternalProtectedResourceMetadataHandler(meta ExternalProtectedResourceMetadata) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		metadata := map[string]any{
			"resource":                 meta.Resource,
			"authorization_servers":    meta.AuthorizationServers,
			"bearer_methods_supported": []string{"header"},
		}
		if len(meta.ScopesSupported) > 0 {
			metadata["scopes_supported"] = meta.ScopesSupported
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(metadata)
	})
}
