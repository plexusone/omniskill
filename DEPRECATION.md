# Deprecation Policy

## Pre-1.0 (Current)

OmniSkill is pre-1.0 software. During this phase:

- **Breaking changes are expected.** We aggressively refine interfaces and design to create the best go-forward API.
- **No backward compatibility guarantees.** Consumers should pin to specific versions and review changelogs before upgrading.
- **Changes are documented in release notes.** All breaking changes appear in CHANGELOG.md with migration guidance where helpful.
- **Primary consumers are internal.** The PlexusOne ecosystem (aha-studio, omniagent, etc.) updates in lockstep, so breaking changes are low-friction.

## Post-1.0 (Future)

Once v1.0 is released, we will adopt stricter semver practices:

- Breaking changes only in major versions
- Deprecated APIs marked with `// Deprecated:` comments
- Migration windows before removal

This policy will be updated when we approach v1.0 and have external adopters.
