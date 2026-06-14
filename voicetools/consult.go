package voicetools

import (
	"context"
	"errors"
	"fmt"

	"github.com/plexusone/omniskill/skill"
)

// ConsultAgentTool queries a specialist agent without transferring the call.
type ConsultAgentTool struct {
	callCtx CallContext
}

// NewConsultAgentTool creates a new consult_agent tool.
func NewConsultAgentTool(callCtx CallContext) *ConsultAgentTool {
	return &ConsultAgentTool{callCtx: callCtx}
}

// Name returns the tool name.
func (t *ConsultAgentTool) Name() string {
	return "consult_agent"
}

// Description returns the tool description.
func (t *ConsultAgentTool) Description() string {
	return "Query a specialist AI agent for information without transferring the call. Use this to get expert answers on specific topics like billing, technical support, or sales."
}

// Parameters returns the tool parameters.
func (t *ConsultAgentTool) Parameters() map[string]skill.Parameter {
	return map[string]skill.Parameter{
		"agent_id": {
			Type:        "string",
			Description: "The specialist agent to consult",
			Required:    true,
			Enum:        []any{"billing", "technical", "sales", "scheduling", "orders"},
		},
		"query": {
			Type:        "string",
			Description: "The question or information request for the specialist agent",
			Required:    true,
		},
		"context": {
			Type:        "object",
			Description: "Additional context to provide to the specialist (e.g., customer account info)",
			Required:    false,
		},
	}
}

// Call executes the tool.
func (t *ConsultAgentTool) Call(ctx context.Context, params map[string]any) (any, error) {
	if t.callCtx == nil {
		return nil, errors.New("call context not available")
	}

	agentID, ok := params["agent_id"].(string)
	if !ok || agentID == "" {
		return nil, errors.New("agent_id is required")
	}

	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, errors.New("query is required")
	}

	contextData, _ := params["context"].(map[string]any)

	registry := t.callCtx.GetAgentRegistry()
	if registry == nil {
		return nil, errors.New("agent registry not available")
	}

	// Check if agent exists
	agent, exists := registry.GetAgent(agentID)
	if !exists {
		return nil, fmt.Errorf("agent '%s' not found", agentID)
	}

	// Add call context to the query context
	call := t.callCtx.GetCall()
	if contextData == nil {
		contextData = make(map[string]any)
	}
	contextData["call_id"] = call.ID()
	contextData["caller"] = call.From()
	contextData["duration_seconds"] = call.Duration().Seconds()

	// Consult the specialist agent
	response, err := registry.ConsultAgent(agentID, query, contextData)
	if err != nil {
		return nil, fmt.Errorf("consultation failed: %w", err)
	}

	return map[string]any{
		"status":     "success",
		"agent_id":   agentID,
		"agent_name": agent.Name,
		"response":   response,
	}, nil
}

var _ skill.Tool = (*ConsultAgentTool)(nil)
