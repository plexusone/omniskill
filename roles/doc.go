// Copyright 2025 John Wang. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package roles provides reference role implementations.
//
// These roles demonstrate how to use the role package's features:
//   - Behaviors: context-aware actions
//   - Policies: governance rules
//   - Workflows: structured action sequences
//   - Metrics: success measurements
//   - Delegation: sub-agent orchestration
//
// # Available Roles
//
// [CodeReviewer] reviews code changes with configurable strictness:
//
//	reviewer := roles.NewCodeReviewer(roles.CodeReviewerConfig{
//	    Strictness: roles.StrictnessBalanced,
//	})
//	err := reviewer.Init(ctx, skills)
//
// [MeetingPM] manages meeting preparation and follow-up:
//
//	pm := roles.NewMeetingPM()
//	err := pm.Init(ctx, skills)
//
// # Using Reference Roles
//
// Reference roles can be used directly or as templates for custom roles.
// Each role documents its required and optional skills.
package roles
