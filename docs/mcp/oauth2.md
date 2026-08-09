# OAuth 2.1 Authentication

The `mcp/oauth2` package provides an OAuth 2.1 Authorization Server for authenticated MCP servers. This is required for public MCP servers that integrate with services like ChatGPT.com.

## Overview

The OAuth2 implementation supports:

- Authorization Code Flow with PKCE (RFC 7636)
- Dynamic Client Registration (RFC 7591)
- Authorization Server Metadata (RFC 8414)
- Bearer token authentication

## Quick Start

Enable OAuth2 when serving over HTTP:

```go
result, err := rt.ServeHTTP(ctx, &runtime.HTTPServerOptions{
    Addr: ":8080",
    OAuth2: &runtime.OAuth2Options{
        Users: map[string]string{
            "admin": "password",
        },
    },
    OnReady: func(r *runtime.HTTPServerResult) {
        fmt.Printf("Client ID: %s\n", r.OAuth2.ClientID)
        fmt.Printf("Client Secret: %s\n", r.OAuth2.ClientSecret)
    },
})
```

## Endpoints

When OAuth2 is enabled, these endpoints are automatically configured:

| Endpoint | Purpose |
|----------|---------|
| `/.well-known/oauth-authorization-server` | Server metadata (RFC 8414) |
| `/authorize` | Authorization endpoint |
| `/token` | Token endpoint |
| `/register` | Dynamic client registration |

## Configuration

### Basic Configuration

```go
OAuth2: &runtime.OAuth2Options{
    // Username/password pairs for authentication
    Users: map[string]string{
        "user1": "password1",
        "user2": "password2",
    },
}
```

### With Pre-configured Clients

```go
OAuth2: &runtime.OAuth2Options{
    Users: map[string]string{"admin": "password"},

    // Pre-configure a client
    ClientID:     "my-client-id",
    ClientSecret: "my-client-secret",
}
```

### Custom Token Lifetime

```go
OAuth2: &runtime.OAuth2Options{
    Users:           map[string]string{"admin": "password"},
    TokenExpiration: 24 * time.Hour,  // Default is 1 hour
}
```

## OAuth2 Flow

### 1. Client Registration

Clients register dynamically:

```http
POST /register
Content-Type: application/json

{
  "client_name": "My MCP Client",
  "redirect_uris": ["https://myapp.com/callback"]
}
```

Response:

```json
{
  "client_id": "generated-client-id",
  "client_secret": "generated-client-secret"
}
```

### 2. Authorization

Redirect user to authorize:

```
GET /authorize?
  response_type=code&
  client_id=CLIENT_ID&
  redirect_uri=https://myapp.com/callback&
  code_challenge=CHALLENGE&
  code_challenge_method=S256&
  state=STATE
```

User authenticates, then redirected to:

```
https://myapp.com/callback?code=AUTH_CODE&state=STATE
```

### 3. Token Exchange

Exchange code for token:

```http
POST /token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=AUTH_CODE&
redirect_uri=https://myapp.com/callback&
client_id=CLIENT_ID&
client_secret=CLIENT_SECRET&
code_verifier=VERIFIER
```

Response:

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### 4. Authenticated Requests

Include token in MCP requests:

```http
GET /mcp
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

## Server Metadata

Clients can discover OAuth2 configuration:

```http
GET /.well-known/oauth-authorization-server
```

Response:

```json
{
  "issuer": "https://myserver.com",
  "authorization_endpoint": "https://myserver.com/authorize",
  "token_endpoint": "https://myserver.com/token",
  "registration_endpoint": "https://myserver.com/register",
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code"],
  "code_challenge_methods_supported": ["S256"]
}
```

## PKCE (Proof Key for Code Exchange)

PKCE is required for all authorization requests:

1. Generate code verifier (random string)
2. Create code challenge: `BASE64URL(SHA256(code_verifier))`
3. Send challenge in authorization request
4. Send verifier in token request

Example (Go):

```go
import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
)

// Generate verifier
verifier := make([]byte, 32)
rand.Read(verifier)
codeVerifier := base64.RawURLEncoding.EncodeToString(verifier)

// Create challenge
hash := sha256.Sum256([]byte(codeVerifier))
codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
```

## Custom Authentication

For custom authentication logic, use the OAuth2 package directly:

```go
import "github.com/plexusone/omniskill/mcp/oauth2"

authServer, err := oauth2.New(&oauth2.Config{
    Issuer: "https://myserver.com",

    // Custom authenticator
    Authenticator: func(username, password string) bool {
        // Check against database, LDAP, etc.
        return validateCredentials(username, password)
    },
})

// Use with HTTP server
http.Handle("/authorize", authServer.AuthorizeHandler())
http.Handle("/token", authServer.TokenHandler())
```

## ChatGPT.com Integration

For ChatGPT.com, OAuth2 is required. Configure your server:

```go
rt.ServeHTTP(ctx, &runtime.HTTPServerOptions{
    Addr:          ":443",
    NgrokAuthtoken: os.Getenv("NGROK_AUTHTOKEN"),
    OAuth2: &runtime.OAuth2Options{
        Users: map[string]string{
            os.Getenv("OAUTH_USER"): os.Getenv("OAUTH_PASSWORD"),
        },
    },
    OnReady: func(r *runtime.HTTPServerResult) {
        fmt.Println("Configure in ChatGPT:")
        fmt.Printf("  MCP URL: %s\n", r.PublicURL)
        fmt.Printf("  Client ID: %s\n", r.OAuth2.ClientID)
        fmt.Printf("  Client Secret: %s\n", r.OAuth2.ClientSecret)
    },
})
```

## Token Revocation

*Added in v0.11.0.* `Server.RevocationHandler()` implements [RFC 7009](https://datatracker.ietf.org/doc/html/rfc7009) token revocation, mounted by default at `/oauth/revoke`:

```go
srv, err := oauth2.New(&oauth2.Config{Issuer: "https://myserver.com"})
mux.Handle("/oauth/revoke", srv.RevocationHandler())
```

The endpoint accepts an access or refresh token via `token` (and optional `token_type_hint`) and revokes it. Per the spec, it always returns `200 OK` regardless of whether the token was found, to avoid leaking token validity to callers.

## External Resource Server Mode

*Added in v0.12.0.* Instead of running the built-in authorization server, a server can validate tokens issued by an external authorization server (an enterprise IdP, or an ID-JAG / MCP Enterprise-Managed Authorization deployment) by implementing `oauth2.TokenVerifier`:

```go
type TokenVerifier interface {
    VerifyToken(ctx context.Context, token string) (*TokenInfo, error)
}
```

Implementations own all protocol-specific verification — signature checks against the issuer's JWKS, issuer/audience validation, expiry, and claim extraction. Wire it up with `ExternalAuth` instead of `OAuth2`:

```go
result, err := rt.ServeHTTP(ctx, &runtime.HTTPServerOptions{
    Addr: ":8080",
    ExternalAuth: &runtime.ExternalAuthOptions{
        Verifier:             myJWTVerifier,
        AuthorizationServers: []string{"https://idp.example.com"},
        ScopesSupported:      []string{"mcp:read", "mcp:write"},
    },
})
```

`ExternalAuth` is mutually exclusive with `OAuth` and `OAuth2`. Only `/.well-known/oauth-protected-resource` is mounted (RFC 9728), advertising `AuthorizationServers`; no local `/authorize`, `/token`, or `/register` endpoints are exposed. Unauthenticated or invalid requests get a `401` with a `WWW-Authenticate: Bearer resource_metadata="..."` header pointing MCP clients at that metadata.

A plain function can be adapted to the interface with `oauth2.TokenVerifierFunc`:

```go
verifier := oauth2.TokenVerifierFunc(func(ctx context.Context, token string) (*oauth2.TokenInfo, error) {
    claims, err := verifyJWT(token) // your JWKS-backed verification
    if err != nil {
        return nil, err
    }
    return &oauth2.TokenInfo{
        Subject: claims.Subject,
        Scope:   claims.Scope,
        Actor:   claims.ActorChain, // RFC 8693 "act" delegation chain, outermost first
        Claims:  claims.Raw,
    }, nil
})
```

`TokenInfo.Actor` and `TokenInfo.Claims` are only populated for externally-verified tokens; read the delegation chain from a request with `oauth2.GetActorFromContext(ctx)`. Use `Claims` for custom policy decisions, e.g. with `ToolAuthorizer` (see the [server middleware guide](server.md)).

## Security Notes

1. **HTTPS Required** - Always use HTTPS in production
2. **Strong Secrets** - Use cryptographically random client secrets
3. **Token Expiration** - Set appropriate token lifetime
4. **PKCE Required** - Never disable PKCE
5. **Validate Redirect URIs** - Strictly validate registered redirect URIs
