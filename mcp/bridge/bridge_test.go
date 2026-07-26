// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package bridge

import (
	"errors"
	"testing"

	"github.com/plexusone/omniskill/skill"
)

func TestSchemaToParameter(t *testing.T) {
	tests := []struct {
		name     string
		schema   map[string]any
		required bool
		want     skill.Parameter
	}{
		{
			name: "simple string",
			schema: map[string]any{
				"type":        "string",
				"description": "A string parameter",
			},
			required: true,
			want: skill.Parameter{
				Type:        "string",
				Description: "A string parameter",
				Required:    true,
			},
		},
		{
			name: "number with constraints",
			schema: map[string]any{
				"type":    "number",
				"minimum": float64(0),
				"maximum": float64(100),
			},
			required: false,
			want: skill.Parameter{
				Type:    "number",
				Minimum: ptrFloat64(0),
				Maximum: ptrFloat64(100),
			},
		},
		{
			name: "string with enum",
			schema: map[string]any{
				"type": "string",
				"enum": []any{"a", "b", "c"},
			},
			required: false,
			want: skill.Parameter{
				Type: "string",
				Enum: []any{"a", "b", "c"},
			},
		},
		{
			name: "string with format",
			schema: map[string]any{
				"type":   "string",
				"format": "date-time",
			},
			required: false,
			want: skill.Parameter{
				Type:   "string",
				Format: "date-time",
			},
		},
		{
			name: "string with length constraints",
			schema: map[string]any{
				"type":      "string",
				"minLength": float64(1),
				"maxLength": float64(255),
			},
			required: false,
			want: skill.Parameter{
				Type:      "string",
				MinLength: ptrInt(1),
				MaxLength: ptrInt(255),
			},
		},
		{
			name: "with default",
			schema: map[string]any{
				"type":    "string",
				"default": "hello",
			},
			required: false,
			want: skill.Parameter{
				Type:    "string",
				Default: "hello",
			},
		},
		{
			name: "array with items",
			schema: map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			required: false,
			want: skill.Parameter{
				Type: "array",
				Items: &skill.Parameter{
					Type: "string",
				},
			},
		},
		{
			name: "object with properties",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{
						"type": "string",
					},
					"age": map[string]any{
						"type": "integer",
					},
				},
				"required": []any{"name"},
			},
			required: false,
			want: skill.Parameter{
				Type: "object",
				Properties: map[string]skill.Parameter{
					"name": {Type: "string", Required: true},
					"age":  {Type: "integer"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schemaToParameter(tt.schema, tt.required)

			if got.Type != tt.want.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Description != tt.want.Description {
				t.Errorf("Description = %q, want %q", got.Description, tt.want.Description)
			}
			if got.Required != tt.want.Required {
				t.Errorf("Required = %v, want %v", got.Required, tt.want.Required)
			}
			if got.Format != tt.want.Format {
				t.Errorf("Format = %q, want %q", got.Format, tt.want.Format)
			}
			if tt.want.Minimum != nil && (got.Minimum == nil || *got.Minimum != *tt.want.Minimum) {
				t.Errorf("Minimum = %v, want %v", got.Minimum, tt.want.Minimum)
			}
			if tt.want.Maximum != nil && (got.Maximum == nil || *got.Maximum != *tt.want.Maximum) {
				t.Errorf("Maximum = %v, want %v", got.Maximum, tt.want.Maximum)
			}
			if tt.want.MinLength != nil && (got.MinLength == nil || *got.MinLength != *tt.want.MinLength) {
				t.Errorf("MinLength = %v, want %v", got.MinLength, tt.want.MinLength)
			}
			if tt.want.MaxLength != nil && (got.MaxLength == nil || *got.MaxLength != *tt.want.MaxLength) {
				t.Errorf("MaxLength = %v, want %v", got.MaxLength, tt.want.MaxLength)
			}
			if len(tt.want.Enum) > 0 && len(got.Enum) != len(tt.want.Enum) {
				t.Errorf("Enum len = %d, want %d", len(got.Enum), len(tt.want.Enum))
			}
			if tt.want.Default != nil && got.Default != tt.want.Default {
				t.Errorf("Default = %v, want %v", got.Default, tt.want.Default)
			}
			if tt.want.Items != nil && (got.Items == nil || got.Items.Type != tt.want.Items.Type) {
				t.Errorf("Items.Type = %v, want %v", got.Items, tt.want.Items)
			}
			if len(tt.want.Properties) > 0 && len(got.Properties) != len(tt.want.Properties) {
				t.Errorf("Properties len = %d, want %d", len(got.Properties), len(tt.want.Properties))
			}
		})
	}
}

func TestGetRequiredFields(t *testing.T) {
	schema := map[string]any{
		"type":     "object",
		"required": []any{"name", "email"},
	}

	got := getRequiredFields(schema)

	if !got["name"] {
		t.Error("expected name to be required")
	}
	if !got["email"] {
		t.Error("expected email to be required")
	}
	if got["optional"] {
		t.Error("optional should not be required")
	}
}

func TestToolError(t *testing.T) {
	err := &ToolError{
		ToolName: "test_tool",
		Content:  "something went wrong",
	}

	if !errors.Is(err, err) {
		t.Error("ToolError should match itself")
	}

	msg := err.Error()
	if msg != `tool "test_tool" error: something went wrong` {
		t.Errorf("unexpected error message: %s", msg)
	}

	if !IsToolError(err) {
		t.Error("IsToolError should return true for ToolError")
	}

	if IsToolError(errors.New("regular error")) {
		t.Error("IsToolError should return false for regular error")
	}
}

func TestRemoteSkillInterface(t *testing.T) {
	s := &RemoteSkill{
		name:        "test-remote",
		description: "A test remote skill",
		version:     "1.0.0",
		tools:       []skill.Tool{},
	}

	if s.Name() != "test-remote" {
		t.Errorf("Name() = %q, want %q", s.Name(), "test-remote")
	}

	if s.Description() != "A test remote skill" {
		t.Errorf("Description() = %q, want %q", s.Description(), "A test remote skill")
	}

	if s.Version() != "1.0.0" {
		t.Errorf("Version() = %q, want %q", s.Version(), "1.0.0")
	}

	if len(s.Tools()) != 0 {
		t.Errorf("Tools() len = %d, want 0", len(s.Tools()))
	}

	// Init and Close should be no-ops
	if err := s.Init(nil); err != nil {
		t.Errorf("Init() error = %v", err)
	}

	if err := s.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrInt(i int) *int {
	return &i
}
