package voicetools

import (
	"context"
	"errors"
	"fmt"

	"github.com/plexusone/omniskill/skill"
)

// HoldCallTool places the caller on hold.
type HoldCallTool struct {
	callCtx CallContext
}

// NewHoldCallTool creates a new hold_call tool.
func NewHoldCallTool(callCtx CallContext) *HoldCallTool {
	return &HoldCallTool{callCtx: callCtx}
}

// Name returns the tool name.
func (t *HoldCallTool) Name() string {
	return "hold_call"
}

// Description returns the tool description.
func (t *HoldCallTool) Description() string {
	return "Place the caller on hold. Use this when you need to perform a task that requires the caller to wait, such as looking up information or consulting with another agent."
}

// Parameters returns the tool parameters.
func (t *HoldCallTool) Parameters() map[string]skill.Parameter {
	return map[string]skill.Parameter{
		"music": {
			Type:        "string",
			Description: "URL to hold music or 'default' for the provider's default hold music",
			Default:     "default",
			Required:    false,
		},
	}
}

// Call executes the tool.
func (t *HoldCallTool) Call(ctx context.Context, params map[string]any) (any, error) {
	if t.callCtx == nil {
		return nil, errors.New("call context not available")
	}

	music, _ := params["music"].(string)
	if music == "" {
		music = "default"
	}

	call := t.callCtx.GetCall()
	transport := t.callCtx.GetTransport()

	if transport == nil {
		return nil, errors.New("telephony transport not available")
	}

	if err := transport.Hold(call.ID(), music); err != nil {
		return nil, fmt.Errorf("hold failed: %w", err)
	}

	return map[string]any{
		"status":  "success",
		"message": "Caller placed on hold",
	}, nil
}

var _ skill.Tool = (*HoldCallTool)(nil)

// UnholdCallTool takes the caller off hold.
type UnholdCallTool struct {
	callCtx CallContext
}

// NewUnholdCallTool creates a new unhold_call tool.
func NewUnholdCallTool(callCtx CallContext) *UnholdCallTool {
	return &UnholdCallTool{callCtx: callCtx}
}

// Name returns the tool name.
func (t *UnholdCallTool) Name() string {
	return "unhold_call"
}

// Description returns the tool description.
func (t *UnholdCallTool) Description() string {
	return "Take the caller off hold and resume the conversation. Use this after completing the task that required the caller to wait."
}

// Parameters returns the tool parameters.
func (t *UnholdCallTool) Parameters() map[string]skill.Parameter {
	return map[string]skill.Parameter{}
}

// Call executes the tool.
func (t *UnholdCallTool) Call(ctx context.Context, params map[string]any) (any, error) {
	if t.callCtx == nil {
		return nil, errors.New("call context not available")
	}

	call := t.callCtx.GetCall()
	transport := t.callCtx.GetTransport()

	if transport == nil {
		return nil, errors.New("telephony transport not available")
	}

	if err := transport.Unhold(call.ID()); err != nil {
		return nil, fmt.Errorf("unhold failed: %w", err)
	}

	return map[string]any{
		"status":  "success",
		"message": "Caller taken off hold",
	}, nil
}

var _ skill.Tool = (*UnholdCallTool)(nil)
