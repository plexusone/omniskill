package clawhub

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

// DependencyResolver resolves skill dependencies.
type DependencyResolver struct {
	hub   *Hub
	cache map[string]*SkillInfo
}

// NewDependencyResolver creates a new dependency resolver.
func NewDependencyResolver(hub *Hub) *DependencyResolver {
	return &DependencyResolver{
		hub:   hub,
		cache: make(map[string]*SkillInfo),
	}
}

// ResolvedDependency represents a fully resolved dependency.
type ResolvedDependency struct {
	// Name is the dependency name.
	Name string

	// Version is the resolved version.
	Version string

	// Repository is the source repository URL.
	Repository string

	// Optional indicates if this is an optional dependency.
	Optional bool

	// Transitive indicates if this is a transitive dependency.
	Transitive bool
}

// ResolutionError represents an error during dependency resolution.
type ResolutionError struct {
	// Dependency is the name of the dependency that failed.
	Dependency string

	// Constraint is the version constraint that couldn't be satisfied.
	Constraint string

	// Reason is the error reason.
	Reason string
}

func (e *ResolutionError) Error() string {
	return fmt.Sprintf("cannot resolve %s (%s): %s", e.Dependency, e.Constraint, e.Reason)
}

// Resolve resolves all dependencies for a manifest.
func (r *DependencyResolver) Resolve(ctx context.Context, manifest *Manifest) ([]ResolvedDependency, error) {
	var resolved []ResolvedDependency
	seen := make(map[string]bool)

	for _, dep := range manifest.Dependencies {
		deps, err := r.resolveDep(ctx, dep, seen, false)
		if err != nil {
			if dep.Optional {
				// Skip optional dependencies that fail to resolve
				continue
			}
			return nil, err
		}
		resolved = append(resolved, deps...)
	}

	return resolved, nil
}

// resolveDep resolves a single dependency and its transitive dependencies.
func (r *DependencyResolver) resolveDep(ctx context.Context, dep Dependency, seen map[string]bool, transitive bool) ([]ResolvedDependency, error) {
	if seen[dep.Name] {
		return nil, nil // Already resolved
	}
	seen[dep.Name] = true

	// Fetch skill info
	info, err := r.fetchSkill(ctx, dep.Name)
	if err != nil {
		return nil, &ResolutionError{
			Dependency: dep.Name,
			Constraint: dep.Version,
			Reason:     err.Error(),
		}
	}

	// Find a version that satisfies the constraint
	version, err := r.findMatchingVersion(ctx, dep.Name, dep.Version, info.Versions)
	if err != nil {
		return nil, &ResolutionError{
			Dependency: dep.Name,
			Constraint: dep.Version,
			Reason:     err.Error(),
		}
	}

	resolved := []ResolvedDependency{{
		Name:       dep.Name,
		Version:    version,
		Repository: info.Repository,
		Optional:   dep.Optional,
		Transitive: transitive,
	}}

	// Resolve transitive dependencies
	versionInfo, err := r.hub.GetVersion(ctx, dep.Name, version)
	if err == nil && versionInfo != nil {
		// Parse manifest for transitive deps (would need to fetch from repo)
		// For now, we skip transitive resolution
	}

	return resolved, nil
}

// fetchSkill fetches skill info, using cache if available.
func (r *DependencyResolver) fetchSkill(ctx context.Context, name string) (*SkillInfo, error) {
	if info, ok := r.cache[name]; ok {
		return info, nil
	}

	info, err := r.hub.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	r.cache[name] = info
	return info, nil
}

// findMatchingVersion finds a version that satisfies the constraint.
func (r *DependencyResolver) findMatchingVersion(ctx context.Context, name, constraint string, versions []string) (string, error) {
	if constraint == "" || constraint == "*" || constraint == "latest" {
		// Return the latest version
		if len(versions) > 0 {
			return versions[0], nil
		}
		info, err := r.hub.Get(ctx, name)
		if err != nil {
			return "", err
		}
		return info.Version, nil
	}

	// Normalize constraint
	constraint = normalizeConstraint(constraint)

	// Check each version against the constraint
	for _, v := range versions {
		if matchesConstraint(v, constraint) {
			return v, nil
		}
	}

	return "", fmt.Errorf("no version satisfies constraint %q", constraint)
}

// normalizeConstraint normalizes a version constraint.
func normalizeConstraint(c string) string {
	c = strings.TrimSpace(c)
	if !strings.HasPrefix(c, "v") && !strings.HasPrefix(c, "^") &&
		!strings.HasPrefix(c, "~") && !strings.HasPrefix(c, ">") &&
		!strings.HasPrefix(c, "<") && !strings.HasPrefix(c, "=") {
		c = "v" + c
	}
	return c
}

// matchesConstraint checks if a version satisfies a constraint.
func matchesConstraint(version, constraint string) bool {
	// Ensure version has v prefix
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	// Handle caret (^) - compatible with version
	if strings.HasPrefix(constraint, "^") {
		minVersion := "v" + strings.TrimPrefix(constraint, "^")
		if !semver.IsValid(minVersion) {
			return false
		}
		// ^1.2.3 means >=1.2.3 and <2.0.0
		if semver.Compare(version, minVersion) < 0 {
			return false
		}
		major := semver.Major(minVersion)
		nextMajor := incrementMajor(major)
		return semver.Compare(version, nextMajor) < 0
	}

	// Handle tilde (~) - approximately equivalent
	if strings.HasPrefix(constraint, "~") {
		minVersion := "v" + strings.TrimPrefix(constraint, "~")
		if !semver.IsValid(minVersion) {
			return false
		}
		// ~1.2.3 means >=1.2.3 and <1.3.0
		if semver.Compare(version, minVersion) < 0 {
			return false
		}
		minor := semver.MajorMinor(minVersion)
		nextMinor := incrementMinor(minor)
		return semver.Compare(version, nextMinor) < 0
	}

	// Handle exact version
	if strings.HasPrefix(constraint, "=") {
		exactVersion := "v" + strings.TrimPrefix(constraint, "=")
		return semver.Compare(version, exactVersion) == 0
	}

	// Handle >= constraint
	if strings.HasPrefix(constraint, ">=") {
		minVersion := "v" + strings.TrimPrefix(constraint, ">=")
		return semver.Compare(version, minVersion) >= 0
	}

	// Handle > constraint
	if strings.HasPrefix(constraint, ">") {
		minVersion := "v" + strings.TrimPrefix(constraint, ">")
		return semver.Compare(version, minVersion) > 0
	}

	// Handle <= constraint
	if strings.HasPrefix(constraint, "<=") {
		maxVersion := "v" + strings.TrimPrefix(constraint, "<=")
		return semver.Compare(version, maxVersion) <= 0
	}

	// Handle < constraint
	if strings.HasPrefix(constraint, "<") {
		maxVersion := "v" + strings.TrimPrefix(constraint, "<")
		return semver.Compare(version, maxVersion) < 0
	}

	// Default: exact match
	return semver.Compare(version, constraint) == 0
}

// incrementMajor increments the major version.
func incrementMajor(version string) string {
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) > 0 {
		if major := parts[0]; major != "" {
			n, err := strconv.Atoi(major)
			if err != nil {
				return "v999.0.0"
			}
			return fmt.Sprintf("v%d.0.0", n+1)
		}
	}
	return "v999.0.0"
}

// incrementMinor increments the minor version.
func incrementMinor(version string) string {
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) >= 2 {
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return version + ".999.0"
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return version + ".999.0"
		}
		return fmt.Sprintf("v%d.%d.0", major, minor+1)
	}
	return version + ".999.0"
}
