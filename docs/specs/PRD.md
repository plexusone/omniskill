# OmniSkill Product Requirements

## Product Summary

OmniSkill is a Go library that lets developers define an AI agent capability once and deliver it everywhere: as an in-process function call, as an MCP tool for Claude Desktop and other MCP clients, or as part of a reusable agent persona. It removes the per-project boilerplate of tool schemas, MCP transports, authentication, and skill distribution.

## Problem

Every agent project reinvents the same plumbing:

- Tool definitions and parameter schemas are rewritten per framework.
- MCP servers require transport, auth, and registration code that is identical across projects.
- Prompt-centric skills (markdown) and programmatic skills (Go) live in separate systems.
- There is no standard way to package, version, or install skills across agents.

## Personas

| Persona | Needs |
|---|---|
| Agent developer | Compose skills into agents and personas; call tools in-process without protocol overhead |
| Tool author | Define tools (Go functions or CLI wrappers) once; get MCP exposure for free |
| MCP integrator | Stand up authenticated MCP servers (stdio/HTTP/SSE) and consume remote MCP servers as local skills |
| Skill publisher | Package skills as versioned packs and distribute them via ClawHub or GitHub |

## Use Cases

1. **Define an agent skill.** Bundle related tools (`FuncTool`, `CommandTool`) into a named `Skill` with lifecycle hooks, then register it once for both library and MCP use. Example: aha-studio's AhaSkill exposes 40+ Aha! product-management tools.

2. **Expose CLI tools via MCP.** Wrap an existing CLI as a `CommandTool` and serve it to Claude Desktop over stdio or to web clients over HTTP/SSE with OAuth 2.1.

3. **Library-mode invocation.** Call `rt.CallTool(ctx, ...)` directly in-process for tests and embedded use, with no JSON-RPC round trip.

4. **Build agent personas.** Define a `Role` with a system prompt, required skills, behaviors, policies, metrics, delegation rules, and workflows; initialize it with resolved skills from the registry.

5. **Distribute skills.** Embed SKILL.md markdown skills in a binary via `SkillPack`, discover them with the loader, and install requirements from GitHub or ClawHub.

6. **Consume remote capabilities.** Connect to an external MCP server via `mcp/client` and use its tools as if they were local skills.

## Requirements

### Functional

| ID | Requirement |
|---|---|
| F1 | Define tools with typed parameters (type, required, enum, default) and Go handlers |
| F2 | Wrap CLI commands as tools with structured results |
| F3 | Group tools into skills with `Init`/`Close` lifecycle |
| F4 | Register and discover skills/tools via a registry with `skill.tool` naming |
| F5 | Serve skills over MCP: stdio, HTTP, SSE; optional ngrok tunneling |
| F6 | Secure HTTP MCP with OAuth 2.1 (PKCE) |
| F7 | Invoke tools directly in library mode without transport |
| F8 | Consume remote MCP servers as local skills |
| F9 | Load skills from SKILL.md (OpenClaw format) and Go constructors via one loader |
| F10 | Package markdown skills as embedded, version-traceable packs |
| F11 | Resolve and install skill requirements from GitHub and ClawHub |
| F12 | Define roles with system prompts, required skills, behaviors, policies, metrics, delegation, and workflows |

### Non-Functional

| ID | Requirement |
|---|---|
| N1 | Pure Go; MCP protocol delegated to the official `modelcontextprotocol/go-sdk` |
| N2 | Skills behave identically across library mode and all MCP transports |
| N3 | Minimal core interfaces; extended capabilities are opt-in interfaces |
| N4 | Thread-safe registry and runtime |
| N5 | Marketplace skills carry security state and dependency manifests |

## Success Criteria

- A skill written for one consumer (e.g., aha-studio MCP) is reusable unchanged by another (e.g., omniagent).
- New tool exposure to Claude Desktop requires only skill registration, no transport code.
- Skills can be authored in markdown by non-Go authors and loaded by the same runtime as Go skills.
- PlexusOne agent projects standardize on OmniSkill instead of bespoke tool layers.

## Out of Scope

- LLM invocation and model routing (omnillm).
- Agent orchestration loops (omniagent).
- Domain-specific tools themselves (they live in consuming repos such as aha-studio).
