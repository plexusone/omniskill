// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package registry

import (
	"strings"
	"sync"

	"github.com/plexusone/omniskill/skill"
)

// CapabilityProvider is implemented by skills that declare their capabilities.
// Skills implementing this interface enable capability-based discovery.
type CapabilityProvider interface {
	// Capabilities returns the capabilities this skill provides.
	Capabilities() []Capability
}

// KeywordProvider is implemented by skills that declare keywords/tags.
// Keywords are used for text-based search and capability inference.
type KeywordProvider interface {
	// Keywords returns searchable keywords/tags for this skill.
	Keywords() []string
}

// DiscoveryRegistry extends Registry with capability-based discovery.
type DiscoveryRegistry interface {
	Registry

	// FindByCapability returns skills that provide the given capability.
	FindByCapability(cap Capability) []skill.Skill

	// FindByCapabilities returns skills that provide ALL of the given capabilities.
	FindByCapabilities(caps []Capability) []skill.Skill

	// FindByAnyCapability returns skills that provide ANY of the given capabilities.
	FindByAnyCapability(caps []Capability) []skill.Skill

	// FindByKeyword returns skills matching the keyword in name, description, or tags.
	FindByKeyword(keyword string) []skill.Skill

	// FindByCategory returns skills with capabilities in the given category.
	FindByCategory(category string) []skill.Skill

	// RegisterCapabilities registers explicit capabilities for a skill.
	// This is used when a skill doesn't implement CapabilityProvider.
	RegisterCapabilities(skillName string, caps []Capability)
}

// DiscoveryInMemory extends InMemory with capability-based discovery.
type DiscoveryInMemory struct {
	*InMemory

	// capabilityIndex maps capabilities to skill names
	capMu           sync.RWMutex
	capabilityIndex map[Capability][]string

	// explicitCaps stores explicitly registered capabilities per skill
	explicitCaps map[string][]Capability
}

// NewDiscovery creates a new in-memory registry with discovery support.
func NewDiscovery() *DiscoveryInMemory {
	return &DiscoveryInMemory{
		InMemory:        New(),
		capabilityIndex: make(map[Capability][]string),
		explicitCaps:    make(map[string][]Capability),
	}
}

// Register adds a skill to the registry and indexes its capabilities.
func (r *DiscoveryInMemory) Register(s skill.Skill) error {
	if err := r.InMemory.Register(s); err != nil {
		return err
	}

	// Index capabilities
	r.indexSkillCapabilities(s)
	return nil
}

// Unregister removes a skill and its capability index entries.
func (r *DiscoveryInMemory) Unregister(name string) error {
	// Get the skill first to remove its capability index entries
	s, err := r.InMemory.Get(name)
	if err != nil {
		return err
	}

	// Remove from capability index
	r.removeFromCapabilityIndex(s)

	// Remove explicit capabilities
	r.capMu.Lock()
	delete(r.explicitCaps, name)
	r.capMu.Unlock()

	return r.InMemory.Unregister(name)
}

// RegisterCapabilities registers explicit capabilities for a skill.
func (r *DiscoveryInMemory) RegisterCapabilities(skillName string, caps []Capability) {
	r.capMu.Lock()
	defer r.capMu.Unlock()

	r.explicitCaps[skillName] = caps
	for _, cap := range caps {
		r.addToCapabilityIndex(skillName, cap)
	}
}

// FindByCapability returns skills that provide the given capability.
func (r *DiscoveryInMemory) FindByCapability(cap Capability) []skill.Skill {
	r.capMu.RLock()
	names := r.capabilityIndex[cap]
	r.capMu.RUnlock()

	var skills []skill.Skill
	for _, name := range names {
		if s, err := r.InMemory.Get(name); err == nil {
			skills = append(skills, s)
		}
	}
	return skills
}

// FindByCapabilities returns skills that provide ALL of the given capabilities.
func (r *DiscoveryInMemory) FindByCapabilities(caps []Capability) []skill.Skill {
	if len(caps) == 0 {
		return nil
	}

	// Get skills with the first capability
	candidates := r.FindByCapability(caps[0])
	if len(caps) == 1 {
		return candidates
	}

	// Filter to those with all capabilities
	var result []skill.Skill
	for _, s := range candidates {
		if r.skillHasAllCapabilities(s, caps[1:]) {
			result = append(result, s)
		}
	}
	return result
}

// FindByAnyCapability returns skills that provide ANY of the given capabilities.
func (r *DiscoveryInMemory) FindByAnyCapability(caps []Capability) []skill.Skill {
	seen := make(map[string]bool)
	var result []skill.Skill

	for _, cap := range caps {
		for _, s := range r.FindByCapability(cap) {
			name := s.Name()
			if !seen[name] {
				seen[name] = true
				result = append(result, s)
			}
		}
	}
	return result
}

// FindByKeyword returns skills matching the keyword in name, description, or tags.
func (r *DiscoveryInMemory) FindByKeyword(keyword string) []skill.Skill {
	keyword = strings.ToLower(keyword)
	var result []skill.Skill

	for _, s := range r.InMemory.List() {
		if r.skillMatchesKeyword(s, keyword) {
			result = append(result, s)
		}
	}
	return result
}

// FindByCategory returns skills with capabilities in the given category.
func (r *DiscoveryInMemory) FindByCategory(category string) []skill.Skill {
	category = strings.ToLower(category)
	seen := make(map[string]bool)
	var result []skill.Skill

	r.capMu.RLock()
	for cap, names := range r.capabilityIndex {
		if cap.MatchesCategory(category) {
			for _, name := range names {
				if !seen[name] {
					seen[name] = true
					if s, err := r.InMemory.Get(name); err == nil {
						result = append(result, s)
					}
				}
			}
		}
	}
	r.capMu.RUnlock()

	return result
}

// GetCapabilities returns all capabilities for a skill.
func (r *DiscoveryInMemory) GetCapabilities(skillName string) []Capability {
	s, err := r.InMemory.Get(skillName)
	if err != nil {
		return nil
	}
	return r.getSkillCapabilities(s)
}

// indexSkillCapabilities indexes a skill's capabilities.
func (r *DiscoveryInMemory) indexSkillCapabilities(s skill.Skill) {
	caps := r.getSkillCapabilities(s)
	name := s.Name()

	r.capMu.Lock()
	defer r.capMu.Unlock()

	for _, cap := range caps {
		r.addToCapabilityIndex(name, cap)
	}
}

// removeFromCapabilityIndex removes a skill from the capability index.
func (r *DiscoveryInMemory) removeFromCapabilityIndex(s skill.Skill) {
	caps := r.getSkillCapabilities(s)
	name := s.Name()

	r.capMu.Lock()
	defer r.capMu.Unlock()

	for _, cap := range caps {
		names := r.capabilityIndex[cap]
		for i, n := range names {
			if n == name {
				r.capabilityIndex[cap] = append(names[:i], names[i+1:]...)
				break
			}
		}
		if len(r.capabilityIndex[cap]) == 0 {
			delete(r.capabilityIndex, cap)
		}
	}
}

// addToCapabilityIndex adds a skill to the capability index (must hold capMu).
func (r *DiscoveryInMemory) addToCapabilityIndex(skillName string, cap Capability) {
	names := r.capabilityIndex[cap]
	for _, n := range names {
		if n == skillName {
			return // Already indexed
		}
	}
	r.capabilityIndex[cap] = append(names, skillName)
}

// getSkillCapabilities returns all capabilities for a skill.
func (r *DiscoveryInMemory) getSkillCapabilities(s skill.Skill) []Capability {
	var caps []Capability

	// Check explicit capabilities first
	r.capMu.RLock()
	if explicit, ok := r.explicitCaps[s.Name()]; ok {
		caps = append(caps, explicit...)
	}
	r.capMu.RUnlock()

	// Check if skill implements CapabilityProvider
	if cp, ok := s.(CapabilityProvider); ok {
		caps = append(caps, cp.Capabilities()...)
	}

	// Infer from keywords if available
	if kp, ok := s.(KeywordProvider); ok {
		caps = append(caps, CapabilityFromKeywords(kp.Keywords())...)
	}

	// Deduplicate
	seen := make(map[Capability]bool)
	var result []Capability
	for _, cap := range caps {
		if !seen[cap] {
			seen[cap] = true
			result = append(result, cap)
		}
	}

	return result
}

// skillHasAllCapabilities checks if a skill has all the given capabilities.
func (r *DiscoveryInMemory) skillHasAllCapabilities(s skill.Skill, caps []Capability) bool {
	skillCaps := r.getSkillCapabilities(s)
	capSet := make(map[Capability]bool)
	for _, cap := range skillCaps {
		capSet[cap] = true
	}

	for _, cap := range caps {
		if !capSet[cap] {
			return false
		}
	}
	return true
}

// skillMatchesKeyword checks if a skill matches the keyword.
func (r *DiscoveryInMemory) skillMatchesKeyword(s skill.Skill, keyword string) bool {
	// Match name
	if strings.Contains(strings.ToLower(s.Name()), keyword) {
		return true
	}

	// Match description
	if strings.Contains(strings.ToLower(s.Description()), keyword) {
		return true
	}

	// Match keywords if available
	if kp, ok := s.(KeywordProvider); ok {
		for _, kw := range kp.Keywords() {
			if strings.Contains(strings.ToLower(kw), keyword) {
				return true
			}
		}
	}

	// Match capability strings
	for _, cap := range r.getSkillCapabilities(s) {
		if strings.Contains(string(cap), keyword) {
			return true
		}
	}

	return false
}

// Ensure DiscoveryInMemory implements DiscoveryRegistry.
var _ DiscoveryRegistry = (*DiscoveryInMemory)(nil)
