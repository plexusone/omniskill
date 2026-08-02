# MCP Bridge

*Added in v0.11.0.* The `mcp/bridge` package mounts a remote MCP server as a local skill in the omniskill registry. Tools from the remote server become callable through the standard `skill.Tool` interface, so agent code doesn't need to know whether a tool is local or remote.

## Basic Usage

```go
import (
    "github.com/plexusone/omniskill/mcp/bridge"
    "github.com/plexusone/omniskill/mcp/client"
    "github.com/plexusone/omniskill/registry"
)

// Connect to a remote MCP server
c := client.New("my-client", "1.0.0", nil)
cmd := exec.Command("npx", "-y", "@modelcontextprotocol/server-github")
session, err := c.ConnectCommand(ctx, cmd, nil)
if err != nil {
    return err
}
defer session.Close()

// Bridge the session to a local skill
b := bridge.NewBridge(session)
remoteSkill, err := b.ToSkill(ctx)
if err != nil {
    return err
}

// Register in local registry
reg := registry.New()
reg.Register(remoteSkill)

// Now use tools through the registry
tool, err := reg.GetTool("github.create_issue")
result, err := tool.Call(ctx, map[string]any{
    "owner": "myorg",
    "repo":  "myrepo",
    "title": "Bug report",
})
```

## Naming

By default, the bridged skill's name is derived from the remote MCP server's implementation name (`InitializeResult.ServerInfo.Name`). Override it with `WithName`:

```go
b := bridge.NewBridge(session, bridge.WithName("gh"))
```

The remote server's implementation version becomes the skill's `Version()`.

## Tool Conversion

`ToSkill` fetches the tool list from the MCP server and wraps each one in a `RemoteTool`:

- Tool name and description are preserved.
- JSON Schema input schemas become `skill.Parameter` maps.
- Tool calls are proxied through the MCP session; `RemoteTool.Call` wraps MCP transport and execution errors with context.
- Results are returned as-is from the MCP server.

## Lifecycle

`Bridge` and the returned `RemoteSkill` do not own the underlying `client.Session` — `RemoteSkill.Init`/`Close` are no-ops. The caller is responsible for closing the session (e.g., via `defer session.Close()`).

## See Also

- [MCP Client](client.md) - Connecting to remote MCP servers
- [Registry](../concepts/registry.md) - Skill registration and discovery
- [Migration](../migration/README.md) - Moving bespoke tool layers to omniskill
