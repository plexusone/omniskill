# GitHub Skill

The `github` package provides a skill for AI agents to interact with GitHub repositories, including issues, pull requests, and code search.

## Installation

```go
import "github.com/plexusone/omniskill/github"
```

## Quick Start

```go
import (
    "github.com/plexusone/omniskill/github"
    "github.com/plexusone/omniskill/mcp/server"
)

// Create GitHub skill
ghSkill := github.New(github.Config{
    Token: os.Getenv("GITHUB_TOKEN"),
})

// Initialize
if err := ghSkill.Init(ctx); err != nil {
    log.Fatal(err)
}
defer ghSkill.Close()

// Register with MCP server
rt := server.New(impl, nil)
rt.RegisterSkill(ghSkill)
```

## Configuration

```go
type Config struct {
    // Token is the GitHub personal access token (required)
    Token string

    // BaseURL is the GitHub API base URL
    // Defaults to https://api.github.com
    // Set for GitHub Enterprise (e.g., https://github.mycompany.com/api/v3)
    BaseURL string

    // DefaultOwner is the default repository owner
    DefaultOwner string

    // DefaultRepo is the default repository name
    DefaultRepo string
}
```

### GitHub Enterprise

```go
ghSkill := github.New(github.Config{
    Token:   os.Getenv("GHE_TOKEN"),
    BaseURL: "https://github.mycompany.com/api/v3",
})
```

### Default Repository

Set defaults to avoid specifying owner/repo on every call:

```go
ghSkill := github.New(github.Config{
    Token:        os.Getenv("GITHUB_TOKEN"),
    DefaultOwner: "plexusone",
    DefaultRepo:  "omniskill",
})
```

## Available Tools

### Issue Tools

| Tool | Description |
|------|-------------|
| `list_issues` | List issues with filters (state, labels) |
| `get_issue` | Get details of a specific issue |
| `create_issue` | Create a new issue |
| `update_issue` | Update an existing issue (title, body, state, labels) |
| `add_issue_comment` | Add a comment to an issue |

### Pull Request Tools

| Tool | Description |
|------|-------------|
| `list_pull_requests` | List PRs with filters (state, base, head) |
| `get_pull_request` | Get details of a specific PR |
| `add_pull_request_comment` | Add a comment to a PR |

### Search Tools

| Tool | Description |
|------|-------------|
| `search_code` | Search code across repositories |
| `search_issues` | Search issues and pull requests |

## Tool Parameters

### list_issues

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `owner` | string | No | Repository owner (uses default if not set) |
| `repo` | string | No | Repository name (uses default if not set) |
| `state` | string | No | Issue state: "open", "closed", "all" (default: "open") |
| `labels` | string | No | Comma-separated label names |
| `per_page` | integer | No | Results per page, max 100 (default: 30) |

### create_issue

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `owner` | string | No | Repository owner |
| `repo` | string | No | Repository name |
| `title` | string | Yes | Issue title |
| `body` | string | No | Issue body (markdown) |
| `labels` | string | No | Comma-separated label names |
| `assignees` | string | No | Comma-separated assignee usernames |

### search_code

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query (GitHub code search syntax) |
| `per_page` | integer | No | Results per page, max 100 (default: 30) |

## Example Usage

### Direct Tool Calls (Library Mode)

```go
// List open bugs
result, err := rt.CallTool(ctx, "list_issues", map[string]any{
    "owner":  "plexusone",
    "repo":   "omniskill",
    "state":  "open",
    "labels": "bug",
})

// Create an issue
result, err := rt.CallTool(ctx, "create_issue", map[string]any{
    "owner":  "plexusone",
    "repo":   "omniskill",
    "title":  "Fix login bug",
    "body":   "Users cannot log in when...",
    "labels": "bug,high-priority",
})

// Search code
result, err := rt.CallTool(ctx, "search_code", map[string]any{
    "query": "repo:plexusone/omniskill language:go func Init",
})
```

### With MCP Server

When exposed via MCP, AI assistants like Claude can use the tools directly:

```
User: List all open issues labeled "bug" in the omniskill repo

Claude: [Uses list_issues tool with owner="plexusone", repo="omniskill", state="open", labels="bug"]

Here are the open bug issues:
1. #42 - Login fails on timeout
2. #38 - Config parsing error
```

## Required Permissions

The GitHub token needs appropriate scopes:

| Scope | Required For |
|-------|--------------|
| `repo` | Private repository access |
| `public_repo` | Public repository access (issues, PRs) |
| `read:org` | Organization repository listing |

For read-only operations on public repos, `public_repo` is sufficient.

## See Also

- [GitHub REST API Documentation](https://docs.github.com/en/rest)
- [GitHub Code Search Syntax](https://docs.github.com/en/search-github/searching-on-github/searching-code)
- [Creating a Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
