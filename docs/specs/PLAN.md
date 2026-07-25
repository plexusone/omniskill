# OmniSkill Implementation Plan

## Current State

The core library is built and in production use (aha-studio's MCP server with 40+ tools runs on it).

| Area | Status | Notes |
|---|---|---|
| skill/ (Skill, Tool, FuncTool, CommandTool) | Built | Tested |
| role/ (Role, behaviors, policies, delegation, workflows, metrics) | Built | Tested; adoption still light |
| registry/ (InMemory) | Built | Tested |
| mcp/server (Runtime, stdio, HTTP/SSE, ngrok, OAuth) | Built | Tested |
| mcp/client | Built | Remote servers as local skills |
| mcp/oauth2 (OAuth 2.1 + PKCE) | Built | |
| loader/ (SKILL.md + Go, UnifiedLoader) | Built | Tested |
| pack/ (SkillPack embed.FS) | Built | Interface only; few published packs |
| installer/ (GitHub, ClawHub sources) | Built | Tested |
| clawhub/ (Hub, manifest, resolver, security) | Built | Marketplace still maturing |
| voicetools/ (transfer, hold, consult, conference) | Built | Tested |
| features/voice-parity | In progress | See `features/voice-parity/PLAN.md` |

## Phase 1: Interface Hardening

Goal: stabilize the public contract before wider ecosystem adoption.

1. Review `skill.Skill`, `skill.Tool`, and `role.Role` for gaps surfaced by aha-studio and omniagent usage; resolve before declaring interface stability.
2. Document tool parameter semantics (enum, default, nested objects) and align with the SDK's JSON Schema generation.
3. Add package-level doc.go files for role/, loader/, pack/, installer/, clawhub/ matching the quality of the root doc.go.
4. Establish a deprecation policy for pre-1.0 changes.

## Phase 2: Role Adoption

Goal: move roles from defined-but-lightly-used to a first-class agent building block.

1. Ship 2-3 reference roles (e.g., code-reviewer, meeting-pm) exercising behaviors, policies, and workflows end to end.
2. Wire delegation (`DelegationConfig`, budgets, retry policies) into omniagent execution.
3. Add role validation: required-skill resolution errors at Init with actionable messages.

## Phase 3: Pack and Distribution Tooling

Goal: make publishing and consuming skills routine.

1. CLI or go:generate helper to scaffold a `SkillPack` from a skills/ directory.
2. Pack validation: SKILL.md frontmatter linting, requirement resolution dry-run.
3. Version traceability checks (pack version equals source commit for derived packs).
4. Publish the first public packs and register them in ClawHub.

## Phase 4: Transport and Security Depth

Goal: production-grade remote serving.

1. Harden OAuth 2.1 flows (token lifetimes, refresh, metadata endpoints) against MCP spec updates.
2. Evaluate WebSocket transport as the SDK adds support.
3. Structured logging via `slog` across server/client paths.
4. Rate limiting and per-tool authorization hooks on the HTTP server.

## Phase 5: Ecosystem Integration

Goal: OmniSkill as the standard capability layer across PlexusOne.

1. Migrate remaining PlexusOne agent projects with bespoke tool layers onto omniskill.
2. `mcp/client` bridging recipes: mount a remote MCP server into a local registry.
3. Cross-repo examples with aha-studio and omniagent kept compile-tested.

## Sequencing and Dependencies

- Phase 1 precedes everything; interface changes get more expensive as adoption grows.
- Phase 2 and Phase 3 are independent and can proceed in parallel.
- Phase 4 depends on `modelcontextprotocol/go-sdk` release cadence for transport work.
- Voice parity work (features/voice-parity) proceeds on its own track.

## Verification

Each phase lands with: `go test ./...` green, `golangci-lint run` clean, examples compile-tested, and CHANGELOG.json updated per the structured-changelog workflow.
