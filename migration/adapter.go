// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package migration

import (
	"context"

	"github.com/plexusone/omniskill/skill"
)

// LegacyTool is the minimal interface for legacy tool types.
// Tools implementing this interface can be wrapped with AdaptTool.
type LegacyTool interface {
	// Name returns the tool name.
	Name() string

	// Execute runs the tool with the given arguments.
	Execute(args map[string]any) (any, error)
}

// LegacyToolWithDescription extends LegacyTool with a description method.
type LegacyToolWithDescription interface {
	LegacyTool
	Description() string
}

// LegacyToolWithParameters extends LegacyTool with parameter definitions.
type LegacyToolWithParameters interface {
	LegacyTool
	Parameters() map[string]LegacyParameter
}

// LegacyParameter represents a legacy parameter definition.
type LegacyParameter struct {
	Type        string
	Description string
	Required    bool
	Default     any
}

// ToolAdapter wraps a legacy tool to implement skill.Tool.
type ToolAdapter struct {
	legacy      LegacyTool
	description string
	parameters  map[string]skill.Parameter
}

// AdaptTool wraps a legacy tool to implement skill.Tool.
//
// The adapter extracts description and parameters if the legacy tool
// implements the optional interfaces (LegacyToolWithDescription,
// LegacyToolWithParameters).
func AdaptTool(legacy LegacyTool) *ToolAdapter {
	adapter := &ToolAdapter{
		legacy:      legacy,
		description: "Adapted legacy tool: " + legacy.Name(),
	}

	// Extract description if available
	if desc, ok := legacy.(LegacyToolWithDescription); ok {
		adapter.description = desc.Description()
	}

	// Extract parameters if available
	if params, ok := legacy.(LegacyToolWithParameters); ok {
		adapter.parameters = convertParameters(params.Parameters())
	}

	return adapter
}

// Name returns the tool name.
func (a *ToolAdapter) Name() string {
	return a.legacy.Name()
}

// Description returns the tool description.
func (a *ToolAdapter) Description() string {
	return a.description
}

// Parameters returns the tool parameters.
func (a *ToolAdapter) Parameters() map[string]skill.Parameter {
	return a.parameters
}

// Call runs the legacy tool.
// The context is ignored since legacy tools don't support cancellation.
func (a *ToolAdapter) Call(ctx context.Context, args map[string]any) (any, error) {
	return a.legacy.Execute(args)
}

// IsAdapted returns true, indicating this is an adapted legacy tool.
func (a *ToolAdapter) IsAdapted() bool {
	return true
}

// LegacySkill is the minimal interface for legacy skill types.
type LegacySkill interface {
	// Name returns the skill name.
	Name() string

	// Tools returns the skill's tools.
	Tools() []LegacyTool
}

// LegacySkillWithDescription extends LegacySkill with a description method.
type LegacySkillWithDescription interface {
	LegacySkill
	Description() string
}

// LegacySkillWithInit extends LegacySkill with initialization.
type LegacySkillWithInit interface {
	LegacySkill
	Init() error
}

// LegacySkillWithClose extends LegacySkill with cleanup.
type LegacySkillWithClose interface {
	LegacySkill
	Close() error
}

// SkillAdapter wraps a legacy skill to implement skill.Skill.
type SkillAdapter struct {
	legacy      LegacySkill
	description string
	tools       []skill.Tool
}

// AdaptSkill wraps a legacy skill to implement skill.Skill.
func AdaptSkill(legacy LegacySkill) *SkillAdapter {
	adapter := &SkillAdapter{
		legacy:      legacy,
		description: "Adapted legacy skill: " + legacy.Name(),
	}

	// Extract description if available
	if desc, ok := legacy.(LegacySkillWithDescription); ok {
		adapter.description = desc.Description()
	}

	// Adapt all tools
	for _, lt := range legacy.Tools() {
		adapter.tools = append(adapter.tools, AdaptTool(lt))
	}

	return adapter
}

// Name returns the skill name.
func (a *SkillAdapter) Name() string {
	return a.legacy.Name()
}

// Description returns the skill description.
func (a *SkillAdapter) Description() string {
	return a.description
}

// Version returns empty string for legacy skills.
func (a *SkillAdapter) Version() string {
	return ""
}

// Tools returns the adapted tools.
func (a *SkillAdapter) Tools() []skill.Tool {
	return a.tools
}

// Init initializes the skill if the legacy skill supports it.
func (a *SkillAdapter) Init(ctx context.Context) error {
	if init, ok := a.legacy.(LegacySkillWithInit); ok {
		return init.Init()
	}
	return nil
}

// Close closes the skill if the legacy skill supports it.
func (a *SkillAdapter) Close() error {
	if closer, ok := a.legacy.(LegacySkillWithClose); ok {
		return closer.Close()
	}
	return nil
}

// IsAdapted returns true, indicating this is an adapted legacy skill.
func (a *SkillAdapter) IsAdapted() bool {
	return true
}

// convertParameters converts legacy parameters to skill.Parameter.
func convertParameters(legacy map[string]LegacyParameter) map[string]skill.Parameter {
	params := make(map[string]skill.Parameter, len(legacy))
	for name, lp := range legacy {
		params[name] = skill.Parameter{
			Type:        lp.Type,
			Description: lp.Description,
			Required:    lp.Required,
			Default:     lp.Default,
		}
	}
	return params
}

// Ensure adapters implement interfaces.
var (
	_ skill.Tool  = (*ToolAdapter)(nil)
	_ skill.Skill = (*SkillAdapter)(nil)
)
