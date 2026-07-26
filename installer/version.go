// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package installer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VersionConstraint represents a version requirement.
//
// Supported formats:
//   - Exact: "1.2.3", "v1.2.3"
//   - Range: ">=1.0.0", "<=2.0.0", ">1.0.0", "<2.0.0"
//   - Caret: "^1.2.3" (>=1.2.3, <2.0.0)
//   - Tilde: "~1.2.3" (>=1.2.3, <1.3.0)
//   - Latest: "latest", "*"
type VersionConstraint struct {
	// Raw is the original constraint string.
	Raw string

	// Type is the constraint type.
	Type ConstraintType

	// Version is the parsed semver (for exact/range/caret/tilde).
	Version *SemVer

	// Op is the comparison operator (for range constraints).
	Op string
}

// ConstraintType identifies the type of version constraint.
type ConstraintType string

const (
	// ConstraintExact matches a specific version.
	ConstraintExact ConstraintType = "exact"

	// ConstraintRange matches versions in a range.
	ConstraintRange ConstraintType = "range"

	// ConstraintCaret matches compatible versions (^).
	ConstraintCaret ConstraintType = "caret"

	// ConstraintTilde matches patch versions (~).
	ConstraintTilde ConstraintType = "tilde"

	// ConstraintLatest matches the latest version.
	ConstraintLatest ConstraintType = "latest"
)

// SemVer represents a semantic version.
type SemVer struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

// String returns the string representation.
func (v SemVer) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// Compare returns -1, 0, or 1 if v is less than, equal to, or greater than other.
func (v SemVer) Compare(other SemVer) int {
	if v.Major != other.Major {
		return intCmp(v.Major, other.Major)
	}
	if v.Minor != other.Minor {
		return intCmp(v.Minor, other.Minor)
	}
	if v.Patch != other.Patch {
		return intCmp(v.Patch, other.Patch)
	}

	// Prerelease versions have lower precedence
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}
	if v.Prerelease != other.Prerelease {
		return strings.Compare(v.Prerelease, other.Prerelease)
	}

	return 0
}

func intCmp(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// semverRegex matches semantic versions.
var semverRegex = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([a-zA-Z0-9.-]+))?(?:\+([a-zA-Z0-9.-]+))?$`)

// ParseSemVer parses a semantic version string.
func ParseSemVer(s string) (*SemVer, error) {
	matches := semverRegex.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid semver: %s", s)
	}

	major, _ := strconv.Atoi(matches[1])
	minor := 0
	patch := 0

	if matches[2] != "" {
		minor, _ = strconv.Atoi(matches[2])
	}
	if matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}

	return &SemVer{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}, nil
}

// ParseVersionConstraint parses a version constraint string.
func ParseVersionConstraint(s string) (*VersionConstraint, error) {
	s = strings.TrimSpace(s)

	// Latest
	if s == "" || s == "latest" || s == "*" {
		return &VersionConstraint{
			Raw:  s,
			Type: ConstraintLatest,
		}, nil
	}

	// Caret
	if strings.HasPrefix(s, "^") {
		v, err := ParseSemVer(strings.TrimPrefix(s, "^"))
		if err != nil {
			return nil, err
		}
		return &VersionConstraint{
			Raw:     s,
			Type:    ConstraintCaret,
			Version: v,
		}, nil
	}

	// Tilde
	if strings.HasPrefix(s, "~") {
		v, err := ParseSemVer(strings.TrimPrefix(s, "~"))
		if err != nil {
			return nil, err
		}
		return &VersionConstraint{
			Raw:     s,
			Type:    ConstraintTilde,
			Version: v,
		}, nil
	}

	// Range operators
	for _, op := range []string{">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(s, op) {
			v, err := ParseSemVer(strings.TrimPrefix(s, op))
			if err != nil {
				return nil, err
			}
			return &VersionConstraint{
				Raw:     s,
				Type:    ConstraintRange,
				Version: v,
				Op:      op,
			}, nil
		}
	}

	// Exact version
	v, err := ParseSemVer(s)
	if err != nil {
		return nil, err
	}
	return &VersionConstraint{
		Raw:     s,
		Type:    ConstraintExact,
		Version: v,
	}, nil
}

// Matches returns true if the version satisfies the constraint.
func (c *VersionConstraint) Matches(v *SemVer) bool {
	switch c.Type {
	case ConstraintLatest:
		return true

	case ConstraintExact:
		return v.Compare(*c.Version) == 0

	case ConstraintRange:
		cmp := v.Compare(*c.Version)
		switch c.Op {
		case ">=":
			return cmp >= 0
		case "<=":
			return cmp <= 0
		case ">":
			return cmp > 0
		case "<":
			return cmp < 0
		case "=":
			return cmp == 0
		}

	case ConstraintCaret:
		// ^1.2.3 means >=1.2.3, <2.0.0
		if v.Compare(*c.Version) < 0 {
			return false
		}
		// Major version must match (for v >= 1.0.0)
		if c.Version.Major > 0 {
			return v.Major == c.Version.Major
		}
		// For 0.x.y, minor must match
		if c.Version.Minor > 0 {
			return v.Major == 0 && v.Minor == c.Version.Minor
		}
		// For 0.0.x, exact match
		return v.Compare(*c.Version) == 0

	case ConstraintTilde:
		// ~1.2.3 means >=1.2.3, <1.3.0
		if v.Compare(*c.Version) < 0 {
			return false
		}
		return v.Major == c.Version.Major && v.Minor == c.Version.Minor
	}

	return false
}

// PinnedSource extends Source with version pinning.
type PinnedSource struct {
	Source

	// Constraint is the version constraint.
	Constraint *VersionConstraint

	// Pinned is the resolved/pinned version.
	Pinned string
}

// ParsePinnedSource parses a source string with version constraint.
//
// Formats:
//   - github.com/user/repo@v1.2.3 (exact version)
//   - github.com/user/repo@^1.2.0 (caret range)
//   - github.com/user/repo@~1.2.0 (tilde range)
//   - github.com/user/repo@>=1.0.0 (range)
//   - github.com/user/repo@latest (latest)
func ParsePinnedSource(s string) (*PinnedSource, error) {
	source, err := ParseSource(s)
	if err != nil {
		return nil, err
	}

	pinned := &PinnedSource{
		Source: *source,
	}

	// Parse constraint from Ref field
	if source.Ref != "" {
		constraint, err := ParseVersionConstraint(source.Ref)
		if err != nil {
			// Not a constraint, treat as exact ref (branch/tag)
			pinned.Pinned = source.Ref
		} else {
			pinned.Constraint = constraint
			// For exact constraints, pin immediately
			if constraint.Type == ConstraintExact {
				pinned.Pinned = "v" + constraint.Version.String()
			}
		}
	}

	return pinned, nil
}

// ResolveVersion resolves the constraint against available versions.
//
// Returns the best matching version or an error if no match.
func (p *PinnedSource) ResolveVersion(available []string) (string, error) {
	if p.Constraint == nil {
		// No constraint, use ref as-is
		return p.Pinned, nil
	}

	if p.Constraint.Type == ConstraintLatest {
		if len(available) == 0 {
			return "", fmt.Errorf("no versions available")
		}
		// Assume available is sorted, latest is last
		return available[len(available)-1], nil
	}

	// Find best matching version
	var best *SemVer
	var bestStr string

	for _, vs := range available {
		v, err := ParseSemVer(vs)
		if err != nil {
			continue
		}

		if !p.Constraint.Matches(v) {
			continue
		}

		if best == nil || v.Compare(*best) > 0 {
			best = v
			bestStr = vs
		}
	}

	if best == nil {
		return "", fmt.Errorf("no version matching %s", p.Constraint.Raw)
	}

	return bestStr, nil
}

// LockFile represents a version lock file for reproducible installs.
type LockFile struct {
	// Version is the lock file format version.
	Version int `json:"version"`

	// Generated is when the lock file was created.
	Generated string `json:"generated"`

	// Locked maps source URLs to pinned versions.
	Locked map[string]LockedEntry `json:"locked"`
}

// LockedEntry is a single locked dependency.
type LockedEntry struct {
	// Version is the pinned version.
	Version string `json:"version"`

	// Constraint is the original constraint.
	Constraint string `json:"constraint,omitempty"`

	// Checksum is the integrity hash.
	Checksum string `json:"checksum,omitempty"`
}
