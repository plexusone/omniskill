# OmniSkill — Roadmap

**Initiative:** `INIT-OMNISKILL-001`
**Repository:** `github.com/plexusone/omniskill`
**Status:** Executing — 4 of 5 phases completed (Phases 2 and 5 in progress)

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-OMNISKILL-NNN`. Phase status is derived from member RMIs — a phase is complete only when all its required RMIs are complete. Execution detail lives in [PLAN.md](PLAN.md); feature-level plans live under `features/` (e.g., `features/voice-parity/PLAN.md`).

## Phase 1 — Interface Hardening

**Theme:** Stabilize the public contract before wider ecosystem adoption.
**Status:** Complete — 5 of 5 items completed

- [x] `RMI-OMNISKILL-001` Interface gap review of `skill.Skill`, `skill.Tool`, and `role.Role` from aha-studio and omniagent usage
  - Acceptance: gaps resolved or explicitly deferred with rationale; interfaces declared stable for pre-1.0
- [x] `RMI-OMNISKILL-002` Document tool parameter semantics (enum, default, nested objects) aligned with SDK JSON Schema generation
- [x] `RMI-OMNISKILL-003` Package-level doc.go for role/, loader/, pack/, installer/, clawhub/ matching root doc.go quality
- [x] `RMI-OMNISKILL-004` Deprecation policy for pre-1.0 interface changes
  - Depends on: `RMI-OMNISKILL-001`
- [x] `RMI-OMNISKILL-005` Voice-parity close-out: verify all modules against the parity checklist and reconcile status docs
  - Acceptance: `features/voice-parity/PLAN.md` status and PLAN.md current-state table agree; residual gaps become new RMIs

## Phase 2 — Role Adoption

**Theme:** Move roles from defined-but-lightly-used to a first-class agent building block.
**Status:** In Progress — 2 of 3 items completed

- [x] `RMI-OMNISKILL-006` Ship 2-3 reference roles (e.g., code-reviewer, meeting-pm) exercising behaviors, policies, and workflows end to end
  - Depends on: `RMI-OMNISKILL-001`
- [ ] `RMI-OMNISKILL-007` Wire delegation (`DelegationConfig`, budgets, retry policies) into omniagent execution
  - Depends on: `RMI-OMNISKILL-006`
  - Cross-repo: requires changes in `github.com/plexusone/omniagent`
- [x] `RMI-OMNISKILL-008` Role validation: required-skill resolution errors at Init with actionable messages
  - Depends on: `RMI-OMNISKILL-001`

## Phase 3 — Pack and Distribution Tooling

**Theme:** Make publishing and consuming skills routine.
**Status:** Complete — 5 of 5 items completed

- [x] `RMI-OMNISKILL-009` SkillPack scaffolding helper (CLI or go:generate) from a skills/ directory
- [x] `RMI-OMNISKILL-010` Pack validation: SKILL.md frontmatter linting, requirement resolution dry-run
  - Depends on: `RMI-OMNISKILL-009`
- [x] `RMI-OMNISKILL-011` Version traceability checks (pack version equals source commit for derived packs)
  - Depends on: `RMI-OMNISKILL-010`
- [x] `RMI-OMNISKILL-012` Publish first public packs and register them in ClawHub
  - Depends on: `RMI-OMNISKILL-010`
  - Note: publish bundle preparation implemented; actual ClawHub registration pending ClawHub availability
- [x] `RMI-OMNISKILL-013` Installer robustness: version pinning for GitHub and ClawHub requirement sources

## Phase 4 — Transport and Security Depth

**Theme:** Production-grade remote serving.
**Status:** Complete — 4 of 4 items completed

- [x] `RMI-OMNISKILL-014` Harden OAuth 2.1 flows (token lifetimes, refresh, metadata endpoints) against MCP spec updates
  - Added: token revocation endpoint (RFC 7009), updated metadata to advertise revocation support
- [x] `RMI-OMNISKILL-015` Evaluate WebSocket transport as `modelcontextprotocol/go-sdk` adds support
  - Status: go-sdk does not yet support WebSocket; stub with config types and integration notes ready
- [x] `RMI-OMNISKILL-016` Structured logging via `slog` across server and client paths
- [x] `RMI-OMNISKILL-017` Rate limiting and per-tool authorization hooks on the HTTP server

## Phase 5 — Ecosystem Integration

**Theme:** OmniSkill as the standard capability layer across PlexusOne.
**Status:** In Progress — 3 of 4 items completed

- [x] `RMI-OMNISKILL-018` Migrate remaining PlexusOne agent projects with bespoke tool layers onto omniskill
  - Depends on: `RMI-OMNISKILL-004`
  - Delivered: migration guide (`docs/migration/`) and adapter utilities (`migration/` package)
- [x] `RMI-OMNISKILL-019` `mcp/client` bridging recipes: mount a remote MCP server into a local registry
  - Delivered: `mcp/bridge/` package with Bridge, RemoteTool, auto-refresh
- [ ] `RMI-OMNISKILL-020` Cross-repo examples with aha-studio and omniagent kept compile-tested
  - Depends on: `RMI-OMNISKILL-007`
  - Blocked: requires omniagent delegation wiring
- [x] `RMI-OMNISKILL-021` Runtime skill discovery: agents query registry/ClawHub for capabilities by need, not by name
  - Depends on: `RMI-OMNISKILL-012`
  - Delivered: DiscoveryRegistry, Capability constants, ClawHub discover/recommend

## Sequencing

- Phase 1 precedes everything; interface changes get more expensive as adoption grows.
- Phases 2 and 3 are independent and can proceed in parallel after Phase 1.
- Phase 4 depends on `modelcontextprotocol/go-sdk` release cadence for transport work.
- Each phase lands with `go test ./...` green, `golangci-lint run` clean, examples compile-tested, and CHANGELOG.json updated.
