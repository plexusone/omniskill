package voicetools

import (
	"context"
	"errors"
	"fmt"

	"github.com/plexusone/omniskill/skill"
)

// TransferCallTool transfers the call to another number or agent queue.
type TransferCallTool struct {
	callCtx CallContext
}

// NewTransferCallTool creates a new transfer_call tool.
func NewTransferCallTool(callCtx CallContext) *TransferCallTool {
	return &TransferCallTool{callCtx: callCtx}
}

// Name returns the tool name.
func (t *TransferCallTool) Name() string {
	return "transfer_call"
}

// Description returns the tool description.
func (t *TransferCallTool) Description() string {
	return "Transfer the current call to another phone number or agent queue. Use this when the caller needs to speak with a different department or person."
}

// Parameters returns the tool parameters.
func (t *TransferCallTool) Parameters() map[string]skill.Parameter {
	return map[string]skill.Parameter{
		"target": {
			Type:        "string",
			Description: "The destination phone number in E.164 format (e.g., +15551234567) or a queue ID (e.g., sales, support)",
			Required:    true,
		},
		"announce": {
			Type:        "string",
			Description: "Optional message to announce to the transfer recipient before connecting",
			Required:    false,
		},
		"warm": {
			Type:        "boolean",
			Description: "If true, perform a warm transfer (stay on line until the target answers). If false, perform a blind transfer (disconnect immediately).",
			Default:     false,
			Required:    false,
		},
	}
}

// Call executes the tool.
func (t *TransferCallTool) Call(ctx context.Context, params map[string]any) (any, error) {
	if t.callCtx == nil {
		return nil, errors.New("call context not available")
	}

	target, ok := params["target"].(string)
	if !ok || target == "" {
		return nil, errors.New("target is required")
	}

	announce, _ := params["announce"].(string)
	warm, _ := params["warm"].(bool)

	call := t.callCtx.GetCall()
	transport := t.callCtx.GetTransport()

	if transport == nil {
		return nil, errors.New("telephony transport not available")
	}

	if err := transport.Transfer(call.ID(), target, announce, warm); err != nil {
		return nil, fmt.Errorf("transfer failed: %w", err)
	}

	return map[string]any{
		"status":  "success",
		"message": fmt.Sprintf("Call transferred to %s", target),
		"warm":    warm,
	}, nil
}

// Ensure TransferCallTool implements Tool.
var _ skill.Tool = (*TransferCallTool)(nil)
