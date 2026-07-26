// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package installer

import (
	"testing"
)

func TestParseSemVer(t *testing.T) {
	tests := []struct {
		input   string
		want    SemVer
		wantErr bool
	}{
		{"1.2.3", SemVer{1, 2, 3, "", ""}, false},
		{"v1.2.3", SemVer{1, 2, 3, "", ""}, false},
		{"0.0.1", SemVer{0, 0, 1, "", ""}, false},
		{"1.0.0-alpha", SemVer{1, 0, 0, "alpha", ""}, false},
		{"1.0.0-beta.1", SemVer{1, 0, 0, "beta.1", ""}, false},
		{"1.0.0+build", SemVer{1, 0, 0, "", "build"}, false},
		{"1.0.0-alpha+build", SemVer{1, 0, 0, "alpha", "build"}, false},
		{"1.0", SemVer{1, 0, 0, "", ""}, false},
		{"1", SemVer{1, 0, 0, "", ""}, false},
		{"invalid", SemVer{}, true},
		{"", SemVer{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSemVer(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSemVer(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil && *got != tt.want {
				t.Errorf("ParseSemVer(%q) = %v, want %v", tt.input, *got, tt.want)
			}
		})
	}
}

func TestSemVerCompare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
	}

	for _, tt := range tests {
		t.Run(tt.a+" vs "+tt.b, func(t *testing.T) {
			a, _ := ParseSemVer(tt.a)
			b, _ := ParseSemVer(tt.b)
			got := a.Compare(*b)
			if got != tt.want {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestParseVersionConstraint(t *testing.T) {
	tests := []struct {
		input    string
		wantType ConstraintType
		wantOp   string
		wantErr  bool
	}{
		{"1.2.3", ConstraintExact, "", false},
		{"v1.2.3", ConstraintExact, "", false},
		{"latest", ConstraintLatest, "", false},
		{"*", ConstraintLatest, "", false},
		{"", ConstraintLatest, "", false},
		{"^1.2.3", ConstraintCaret, "", false},
		{"~1.2.3", ConstraintTilde, "", false},
		{">=1.0.0", ConstraintRange, ">=", false},
		{"<=2.0.0", ConstraintRange, "<=", false},
		{">1.0.0", ConstraintRange, ">", false},
		{"<2.0.0", ConstraintRange, "<", false},
		{"=1.5.0", ConstraintRange, "=", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseVersionConstraint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersionConstraint(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil {
				if got.Type != tt.wantType {
					t.Errorf("Type = %v, want %v", got.Type, tt.wantType)
				}
				if got.Op != tt.wantOp {
					t.Errorf("Op = %v, want %v", got.Op, tt.wantOp)
				}
			}
		})
	}
}

func TestConstraintMatches(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		// Exact
		{"1.2.3", "1.2.3", true},
		{"1.2.3", "1.2.4", false},

		// Latest
		{"latest", "999.0.0", true},
		{"*", "0.0.1", true},

		// Range
		{">=1.0.0", "1.0.0", true},
		{">=1.0.0", "2.0.0", true},
		{">=1.0.0", "0.9.0", false},
		{"<=2.0.0", "1.0.0", true},
		{"<=2.0.0", "2.0.0", true},
		{"<=2.0.0", "2.0.1", false},
		{">1.0.0", "1.0.1", true},
		{">1.0.0", "1.0.0", false},
		{"<2.0.0", "1.9.9", true},
		{"<2.0.0", "2.0.0", false},

		// Caret (major version fixed for >= 1.0.0)
		{"^1.2.3", "1.2.3", true},
		{"^1.2.3", "1.9.9", true},
		{"^1.2.3", "2.0.0", false},
		{"^1.2.3", "1.2.2", false},
		{"^0.2.3", "0.2.3", true},
		{"^0.2.3", "0.2.9", true},
		{"^0.2.3", "0.3.0", false},

		// Tilde (minor version fixed)
		{"~1.2.3", "1.2.3", true},
		{"~1.2.3", "1.2.9", true},
		{"~1.2.3", "1.3.0", false},
		{"~1.2.3", "1.2.2", false},
	}

	for _, tt := range tests {
		t.Run(tt.constraint+" matches "+tt.version, func(t *testing.T) {
			c, err := ParseVersionConstraint(tt.constraint)
			if err != nil {
				t.Fatal(err)
			}
			v, err := ParseSemVer(tt.version)
			if err != nil {
				t.Fatal(err)
			}

			got := c.Matches(v)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePinnedSource(t *testing.T) {
	tests := []struct {
		input      string
		wantURL    string
		wantType   ConstraintType
		wantPinned string
	}{
		{
			input:      "github.com/user/repo@v1.2.3",
			wantURL:    "https://github.com/user/repo.git",
			wantType:   ConstraintExact,
			wantPinned: "v1.2.3",
		},
		{
			input:    "github.com/user/repo@^1.0.0",
			wantURL:  "https://github.com/user/repo.git",
			wantType: ConstraintCaret,
		},
		{
			input:    "github.com/user/repo@latest",
			wantURL:  "https://github.com/user/repo.git",
			wantType: ConstraintLatest,
		},
		{
			input:      "github.com/user/repo@main",
			wantURL:    "https://github.com/user/repo.git",
			wantPinned: "main", // Branch, not semver
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePinnedSource(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			if got.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tt.wantURL)
			}

			if tt.wantType != "" && (got.Constraint == nil || got.Constraint.Type != tt.wantType) {
				var gotType ConstraintType
				if got.Constraint != nil {
					gotType = got.Constraint.Type
				}
				t.Errorf("Constraint.Type = %v, want %v", gotType, tt.wantType)
			}

			if tt.wantPinned != "" && got.Pinned != tt.wantPinned {
				t.Errorf("Pinned = %q, want %q", got.Pinned, tt.wantPinned)
			}
		})
	}
}

func TestResolveVersion(t *testing.T) {
	available := []string{"v1.0.0", "v1.1.0", "v1.2.0", "v2.0.0"}

	tests := []struct {
		constraint string
		want       string
		wantErr    bool
	}{
		{"latest", "v2.0.0", false},
		{"^1.0.0", "v1.2.0", false},
		{"~1.1.0", "v1.1.0", false},
		{">=1.1.0", "v2.0.0", false},
		{"<2.0.0", "v1.2.0", false},
		{">=3.0.0", "", true}, // No match
	}

	for _, tt := range tests {
		t.Run(tt.constraint, func(t *testing.T) {
			pinned, err := ParsePinnedSource("github.com/user/repo@" + tt.constraint)
			if err != nil {
				t.Fatal(err)
			}

			got, err := pinned.ResolveVersion(available)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
