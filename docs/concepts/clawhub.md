# ClawHub Marketplace

ClawHub is a skill marketplace for discovering, sharing, and installing AI agent skills. OmniSkill integrates with ClawHub to provide:

- **Skill Discovery** - Search and browse published skills
- **Version Resolution** - Automatic dependency resolution
- **Security Scanning** - Check skills for security issues
- **Signature Verification** - Verify skill authenticity

## Overview

ClawHub skills use the `CLAWHUB.json` manifest format:

```json
{
  "name": "weather",
  "version": "1.2.0",
  "description": "Weather forecasting skill",
  "author": "plexusone",
  "repository": "github.com/plexusone/weather-skill",
  "license": "MIT",
  "dependencies": [],
  "permissions": ["network"],
  "signature": "..."
}
```

## Quick Start

### Install from ClawHub

```go
import "github.com/plexusone/omniskill/installer"

si := installer.NewSkillInstaller()

// Install official ClawHub skill
skill, err := si.Install(ctx, "@clawhub/weather")

// Install community skill
skill, err := si.Install(ctx, "@user/my-skill")
```

### CLI Commands

```bash
# Search for skills
omniagent skills search weather

# Install from ClawHub
omniagent skills install @clawhub/weather

# Install from GitHub
omniagent skills install github.com/user/skill

# List installed skills
omniagent skills list --installed

# Update a skill
omniagent skills update weather

# Remove a skill
omniagent skills remove weather
```

## Manifest Format

### CLAWHUB.json

```go
type Manifest struct {
    Name         string       `json:"name"`
    Version      string       `json:"version"`
    Description  string       `json:"description"`
    Author       string       `json:"author"`
    Repository   string       `json:"repository"`
    License      string       `json:"license"`
    Dependencies []Dependency `json:"dependencies"`
    Permissions  []string     `json:"permissions"`
    Signature    string       `json:"signature,omitempty"`
}
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique skill identifier |
| `version` | Yes | Semantic version (e.g., `1.2.0`) |
| `description` | Yes | Human-readable description |
| `author` | Yes | Author or organization |
| `repository` | Yes | Source repository URL |
| `license` | Yes | SPDX license identifier |
| `dependencies` | No | Other skills this depends on |
| `permissions` | No | Required permissions |
| `signature` | No | Cryptographic signature |

### Dependencies

```json
{
  "dependencies": [
    {
      "name": "http-client",
      "version": ">=1.0.0",
      "optional": false
    }
  ]
}
```

### Permissions

| Permission | Description |
|------------|-------------|
| `network` | HTTP/HTTPS requests |
| `filesystem` | File read/write |
| `shell` | Execute shell commands |
| `browser` | Browser automation |
| `docker` | Container operations |

## Hub Client

### Searching Skills

```go
import "github.com/plexusone/omniskill/clawhub"

hub := clawhub.NewHub(clawhub.Config{
    // Optional: custom registry URL
    // BaseURL: "https://registry.clawhub.ai",
})

// Search by query
results, err := hub.Search(ctx, "weather")

for _, r := range results {
    fmt.Printf("%s v%s - %s\n", r.Name, r.Version, r.Description)
}
```

### Getting Skill Info

```go
// Get skill details
info, err := hub.Get(ctx, "@clawhub/weather")

fmt.Printf("Name: %s\n", info.Name)
fmt.Printf("Version: %s\n", info.Version)
fmt.Printf("Downloads: %d\n", info.Downloads)
```

### Version Resolution

```go
// List available versions
versions, err := hub.Versions(ctx, "@clawhub/weather")

// Resolve version constraint
version, err := hub.Resolve(ctx, "@clawhub/weather", ">=1.0.0 <2.0.0")
```

## Dependency Resolution

The resolver handles complex dependency graphs:

```go
import "github.com/plexusone/omniskill/clawhub"

resolver := clawhub.NewResolver(hub)

// Resolve all dependencies for a skill
deps, err := resolver.Resolve(ctx, "@clawhub/weather@1.2.0")

for _, dep := range deps {
    fmt.Printf("%s@%s\n", dep.Name, dep.Version)
}
```

### Resolution Options

```go
resolver := clawhub.NewResolver(hub)

// Allow pre-release versions
resolver.AllowPrerelease = true

// Prefer latest compatible versions
resolver.Strategy = clawhub.StrategyLatest

// Or prefer minimal satisfying versions
resolver.Strategy = clawhub.StrategyMinimal
```

### Conflict Detection

```go
deps, err := resolver.Resolve(ctx, skill)
if err != nil {
    if conflicts, ok := err.(*clawhub.ConflictError); ok {
        for _, c := range conflicts.Conflicts {
            fmt.Printf("Conflict: %s requires %s, but %s requires %s\n",
                c.RequiredBy[0], c.Version1,
                c.RequiredBy[1], c.Version2)
        }
    }
}
```

## Security

### Security Scanning

```go
import "github.com/plexusone/omniskill/clawhub"

scanner := clawhub.NewSecurityScanner()

// Scan a skill directory
report, err := scanner.Scan(ctx, "./skills/weather")

if len(report.Issues) > 0 {
    for _, issue := range report.Issues {
        fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.File, issue.Message)
    }
}
```

### Issue Severities

| Severity | Description |
|----------|-------------|
| `critical` | Must fix before installation |
| `high` | Security vulnerability |
| `medium` | Potential security concern |
| `low` | Best practice violation |
| `info` | Informational |

### Security Checks

The scanner performs these checks:

- **Shell Injection** - Unsafe command construction
- **Path Traversal** - Directory escape attempts
- **Credential Exposure** - Hardcoded secrets
- **Network Access** - Undeclared network requests
- **File Operations** - Undeclared filesystem access

### Signature Verification

```go
verifier := clawhub.NewSignatureVerifier()

// Add trusted public keys
verifier.AddKey("clawhub-official", officialPublicKey)
verifier.AddKey("my-org", orgPublicKey)

// Verify skill signature
valid, signer, err := verifier.Verify(ctx, "./skills/weather")
if valid {
    fmt.Printf("Verified: signed by %s\n", signer)
} else {
    fmt.Println("Warning: signature verification failed")
}
```

## GitHub Integration

Install skills directly from GitHub releases:

```go
import "github.com/plexusone/omniskill/installer"

si := installer.NewSkillInstaller()

// Install from GitHub (clones repository)
skill, err := si.Install(ctx, "github.com/user/skill")

// Install specific release
skill, err := si.Install(ctx, "github.com/user/skill@v1.2.0")

// Install from release tarball
si.UseReleases = true
skill, err := si.Install(ctx, "github.com/user/skill@v1.2.0")
```

### GitHub Release Assets

When `UseReleases` is enabled, the installer:

1. Fetches release metadata from GitHub API
2. Downloads the release tarball
3. Extracts and installs the skill
4. Verifies CLAWHUB.json manifest

### Authentication

For private repositories:

```go
si := installer.NewSkillInstaller()
si.GitHubToken = os.Getenv("GITHUB_TOKEN")

skill, err := si.Install(ctx, "github.com/private-org/skill")
```

## Installer Integration

The `SkillInstaller` supports multiple source types:

| Source | Format | Example |
|--------|--------|---------|
| ClawHub | `@namespace/skill` | `@clawhub/weather` |
| GitHub | `github.com/user/repo` | `github.com/user/skill` |
| Git URL | `https://...` | `https://gitlab.com/user/skill.git` |
| Local | Path | `./my-skill` |

### ClawHub Source

```go
si := installer.NewSkillInstaller()

// Configure ClawHub
si.ClawHubConfig = &clawhub.Config{
    BaseURL: "https://registry.clawhub.ai",
}

// Install from ClawHub
skill, err := si.Install(ctx, "@clawhub/weather")
```

### Source Detection

The installer automatically detects source type:

```go
// ClawHub (starts with @)
si.Install(ctx, "@clawhub/weather")

// GitHub (github.com prefix)
si.Install(ctx, "github.com/user/skill")

// Git URL (https:// prefix)
si.Install(ctx, "https://gitlab.com/user/skill.git")

// Local path (everything else)
si.Install(ctx, "./local-skill")
```

## Best Practices

### For Skill Authors

1. **Use Semantic Versioning** - Follow semver for compatibility
2. **Declare Permissions** - List all required permissions
3. **Sign Releases** - Use signature verification for trust
4. **Document Dependencies** - Specify version constraints
5. **Security Review** - Run scanner before publishing

### For Skill Users

1. **Check Security** - Review security scan results
2. **Verify Signatures** - Trust only verified skills
3. **Pin Versions** - Use exact versions in production
4. **Review Permissions** - Understand what the skill can do
5. **Update Regularly** - Keep skills updated for security fixes

## See Also

- [Installer](installer.md) - Skill installation management
- [Skills](skills.md) - Skill interface and lifecycle
- [Loader](loader.md) - Loading SKILL.md format
