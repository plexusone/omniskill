package voicetools

import (
	"context"
	"errors"
	"fmt"

	"github.com/plexusone/omniskill/skill"
)

// ConferenceTool adds participants to a conference call.
type ConferenceTool struct {
	callCtx CallContext
}

// NewConferenceTool creates a new add_to_conference tool.
func NewConferenceTool(callCtx CallContext) *ConferenceTool {
	return &ConferenceTool{callCtx: callCtx}
}

// Name returns the tool name.
func (t *ConferenceTool) Name() string {
	return "add_to_conference"
}

// Description returns the tool description.
func (t *ConferenceTool) Description() string {
	return "Add one or more participants to the call, creating a conference. Use this when multiple people need to be on the same call together."
}

// Parameters returns the tool parameters.
func (t *ConferenceTool) Parameters() map[string]skill.Parameter {
	return map[string]skill.Parameter{
		"participants": {
			Type:        "array",
			Description: "Phone numbers in E.164 format to add to the conference",
			Required:    true,
			Items: &skill.Parameter{
				Type:        "string",
				Description: "Phone number in E.164 format (e.g., +15551234567)",
			},
		},
	}
}

// Call executes the tool.
func (t *ConferenceTool) Call(ctx context.Context, params map[string]any) (any, error) {
	if t.callCtx == nil {
		return nil, errors.New("call context not available")
	}

	participantsRaw, ok := params["participants"].([]any)
	if !ok || len(participantsRaw) == 0 {
		return nil, errors.New("at least one participant is required")
	}

	participants := make([]string, 0, len(participantsRaw))
	for _, p := range participantsRaw {
		if s, ok := p.(string); ok {
			participants = append(participants, s)
		}
	}

	if len(participants) == 0 {
		return nil, errors.New("no valid participants provided")
	}

	call := t.callCtx.GetCall()
	transport := t.callCtx.GetTransport()

	if transport == nil {
		return nil, errors.New("telephony transport not available")
	}

	if err := transport.Conference(call.ID(), participants); err != nil {
		return nil, fmt.Errorf("conference failed: %w", err)
	}

	return map[string]any{
		"status":       "success",
		"message":      fmt.Sprintf("Added %d participant(s) to the call", len(participants)),
		"participants": participants,
	}, nil
}

var _ skill.Tool = (*ConferenceTool)(nil)
