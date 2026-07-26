// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package clawhub

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// DiscoverOptions configures capability-based discovery.
type DiscoverOptions struct {
	// Page is the page number (1-indexed).
	Page int

	// PerPage is the number of results per page.
	PerPage int

	// MinStars filters to skills with at least this many stars.
	MinStars int

	// SecurityVerified filters to skills that passed security review.
	SecurityVerified bool
}

// Discover searches ClawHub for skills providing the given capability.
// The capability should be in "category:action" format (e.g., "file:read").
func (h *Hub) Discover(ctx context.Context, capability string, opts *DiscoverOptions) (*SearchResult, error) {
	if opts == nil {
		opts = &DiscoverOptions{}
	}
	if opts.PerPage <= 0 {
		opts.PerPage = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	params := url.Values{
		"capability": {capability},
		"page":       {fmt.Sprintf("%d", opts.Page)},
		"perPage":    {fmt.Sprintf("%d", opts.PerPage)},
	}

	if opts.MinStars > 0 {
		params.Set("minStars", fmt.Sprintf("%d", opts.MinStars))
	}
	if opts.SecurityVerified {
		params.Set("securityState", string(SecurityStateVerified))
	}

	var result SearchResult
	if err := h.get(ctx, "/skills/discover?"+params.Encode(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DiscoverMultiple searches for skills providing any of the given capabilities.
func (h *Hub) DiscoverMultiple(ctx context.Context, capabilities []string, opts *DiscoverOptions) (*SearchResult, error) {
	if len(capabilities) == 0 {
		return &SearchResult{}, nil
	}

	if opts == nil {
		opts = &DiscoverOptions{}
	}
	if opts.PerPage <= 0 {
		opts.PerPage = 20
	}
	if opts.Page <= 0 {
		opts.Page = 1
	}

	params := url.Values{
		"capabilities": {strings.Join(capabilities, ",")},
		"page":         {fmt.Sprintf("%d", opts.Page)},
		"perPage":      {fmt.Sprintf("%d", opts.PerPage)},
	}

	if opts.MinStars > 0 {
		params.Set("minStars", fmt.Sprintf("%d", opts.MinStars))
	}
	if opts.SecurityVerified {
		params.Set("securityState", string(SecurityStateVerified))
	}

	var result SearchResult
	if err := h.get(ctx, "/skills/discover?"+params.Encode(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// RecommendOptions configures skill recommendations.
type RecommendOptions struct {
	// MaxResults is the maximum number of recommendations.
	MaxResults int

	// IncludeRationale includes explanation for why each skill was recommended.
	IncludeRationale bool

	// SecurityVerified filters to verified skills only.
	SecurityVerified bool
}

// Recommendation represents a skill recommendation with optional rationale.
type Recommendation struct {
	// Skill is the recommended skill.
	Skill SkillInfo `json:"skill"`

	// Score is the relevance score (0.0-1.0).
	Score float64 `json:"score"`

	// Rationale explains why this skill was recommended.
	Rationale string `json:"rationale,omitempty"`

	// MatchedCapabilities are the capabilities that matched the need.
	MatchedCapabilities []string `json:"matchedCapabilities,omitempty"`
}

// RecommendResult contains skill recommendations.
type RecommendResult struct {
	// Recommendations is the list of recommended skills.
	Recommendations []Recommendation `json:"recommendations"`

	// ParsedNeed is how the system interpreted the need.
	ParsedNeed string `json:"parsedNeed,omitempty"`

	// InferredCapabilities are capabilities inferred from the need.
	InferredCapabilities []string `json:"inferredCapabilities,omitempty"`
}

// RecommendSkills recommends skills based on a natural language description of need.
// Example: "I need to read and write files, and make HTTP requests"
func (h *Hub) RecommendSkills(ctx context.Context, need string, opts *RecommendOptions) (*RecommendResult, error) {
	if opts == nil {
		opts = &RecommendOptions{}
	}
	if opts.MaxResults <= 0 {
		opts.MaxResults = 10
	}

	params := url.Values{
		"need":       {need},
		"maxResults": {fmt.Sprintf("%d", opts.MaxResults)},
	}

	if opts.IncludeRationale {
		params.Set("includeRationale", "true")
	}
	if opts.SecurityVerified {
		params.Set("securityState", string(SecurityStateVerified))
	}

	var result RecommendResult
	if err := h.get(ctx, "/skills/recommend?"+params.Encode(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CategoryInfo describes a capability category.
type CategoryInfo struct {
	// Name is the category name (e.g., "file", "code", "ai").
	Name string `json:"name"`

	// Description describes what this category covers.
	Description string `json:"description"`

	// SkillCount is the number of skills in this category.
	SkillCount int `json:"skillCount"`

	// Capabilities lists capabilities in this category.
	Capabilities []string `json:"capabilities"`
}

// ListCategories returns all capability categories available in ClawHub.
func (h *Hub) ListCategories(ctx context.Context) ([]CategoryInfo, error) {
	var categories []CategoryInfo
	if err := h.get(ctx, "/capabilities/categories", &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

// ListCapabilities returns all capabilities in a category.
func (h *Hub) ListCapabilities(ctx context.Context, category string) ([]string, error) {
	var capabilities []string
	path := "/capabilities/" + url.PathEscape(category)
	if err := h.get(ctx, path, &capabilities); err != nil {
		return nil, err
	}
	return capabilities, nil
}

// SkillsByCategory returns skills grouped by their primary capability category.
type SkillsByCategory struct {
	// Category is the capability category.
	Category string `json:"category"`

	// Skills are the skills in this category.
	Skills []SkillInfo `json:"skills"`
}

// Browse returns skills organized by capability category for exploration.
func (h *Hub) Browse(ctx context.Context, opts *DiscoverOptions) ([]SkillsByCategory, error) {
	if opts == nil {
		opts = &DiscoverOptions{}
	}
	if opts.PerPage <= 0 {
		opts.PerPage = 10 // Skills per category
	}

	params := url.Values{
		"perCategory": {fmt.Sprintf("%d", opts.PerPage)},
	}

	if opts.MinStars > 0 {
		params.Set("minStars", fmt.Sprintf("%d", opts.MinStars))
	}
	if opts.SecurityVerified {
		params.Set("securityState", string(SecurityStateVerified))
	}

	var result []SkillsByCategory
	if err := h.get(ctx, "/skills/browse?"+params.Encode(), &result); err != nil {
		return nil, err
	}

	return result, nil
}
