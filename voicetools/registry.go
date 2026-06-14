package voicetools

import (
	"context"

	"github.com/plexusone/omniskill/skill"
)

// VoiceSkill is a skill that provides voice call control tools.
type VoiceSkill struct {
	skill.BaseSkill
	callCtx CallContext
}

// NewVoiceSkill creates a new voice control skill with all tools.
func NewVoiceSkill(callCtx CallContext) *VoiceSkill {
	s := &VoiceSkill{
		BaseSkill: skill.BaseSkill{
			SkillName:        "voice_control",
			SkillDescription: "Voice call control tools for transfers, holds, conferencing, and agent consultation",
		},
		callCtx: callCtx,
	}

	s.SkillTools = []skill.Tool{
		NewTransferCallTool(callCtx),
		NewHoldCallTool(callCtx),
		NewUnholdCallTool(callCtx),
		NewConsultAgentTool(callCtx),
		NewConferenceTool(callCtx),
	}

	return s
}

// Init initializes the skill.
func (s *VoiceSkill) Init(ctx context.Context) error {
	// Validate call context is available
	if s.callCtx == nil {
		return ErrNoCallContext
	}
	return nil
}

var _ skill.Skill = (*VoiceSkill)(nil)
