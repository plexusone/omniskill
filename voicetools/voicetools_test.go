package voicetools

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/plexusone/omniskill/skill"
)

// Mock implementations for testing

type mockCall struct {
	id        string
	from      string
	to        string
	direction string
	duration  time.Duration
	metadata  map[string]string
}

func (c *mockCall) ID() string                  { return c.id }
func (c *mockCall) From() string                { return c.from }
func (c *mockCall) To() string                  { return c.to }
func (c *mockCall) Direction() string           { return c.direction }
func (c *mockCall) Duration() time.Duration     { return c.duration }
func (c *mockCall) Metadata() map[string]string { return c.metadata }

type mockTransport struct {
	transferCalled   bool
	holdCalled       bool
	unholdCalled     bool
	conferenceCalled bool
	lastTarget       string
	lastParticipants []string
	err              error
}

func (t *mockTransport) Transfer(callID, target, announce string, warm bool) error {
	t.transferCalled = true
	t.lastTarget = target
	return t.err
}

func (t *mockTransport) Hold(callID, music string) error {
	t.holdCalled = true
	return t.err
}

func (t *mockTransport) Unhold(callID string) error {
	t.unholdCalled = true
	return t.err
}

func (t *mockTransport) Conference(callID string, participants []string) error {
	t.conferenceCalled = true
	t.lastParticipants = participants
	return t.err
}

func (t *mockTransport) Mute(callID string) error   { return nil }
func (t *mockTransport) Unmute(callID string) error { return nil }
func (t *mockTransport) Hangup(callID string) error { return nil }

type mockAgentRegistry struct {
	agents      map[string]AgentConfig
	consultResp string
	consultErr  error
}

func (r *mockAgentRegistry) GetAgent(id string) (AgentConfig, bool) {
	agent, ok := r.agents[id]
	return agent, ok
}

func (r *mockAgentRegistry) ListAgents() []string {
	ids := make([]string, 0, len(r.agents))
	for id := range r.agents {
		ids = append(ids, id)
	}
	return ids
}

func (r *mockAgentRegistry) ConsultAgent(agentID, query string, context map[string]any) (string, error) {
	return r.consultResp, r.consultErr
}

func newTestContext() (CallContext, *mockTransport, *mockAgentRegistry) {
	call := &mockCall{
		id:        "call-123",
		from:      "+15551234567",
		to:        "+15559876543",
		direction: "inbound",
		duration:  5 * time.Minute,
	}

	transport := &mockTransport{}

	registry := &mockAgentRegistry{
		agents: map[string]AgentConfig{
			"billing":   {ID: "billing", Name: "Billing Support", Description: "Handles billing questions"},
			"technical": {ID: "technical", Name: "Tech Support", Description: "Handles technical issues"},
		},
		consultResp: "Here is the information you requested.",
	}

	return NewCallContext(call, transport, registry), transport, registry
}

func TestVoiceSkill(t *testing.T) {
	ctx, _, _ := newTestContext()
	skill := NewVoiceSkill(ctx)

	if skill.Name() != "voice_control" {
		t.Errorf("expected name 'voice_control', got %q", skill.Name())
	}

	tools := skill.Tools()
	if len(tools) != 5 {
		t.Errorf("expected 5 tools, got %d", len(tools))
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools {
		toolNames[tool.Name()] = true
	}

	expected := []string{"transfer_call", "hold_call", "unhold_call", "consult_agent", "add_to_conference"}
	for _, name := range expected {
		if !toolNames[name] {
			t.Errorf("missing tool: %s", name)
		}
	}

	if err := skill.Init(context.Background()); err != nil {
		t.Errorf("Init failed: %v", err)
	}
}

func TestVoiceSkill_NoContext(t *testing.T) {
	skill := NewVoiceSkill(nil)
	err := skill.Init(context.Background())
	if err != ErrNoCallContext {
		t.Errorf("expected ErrNoCallContext, got %v", err)
	}
}

func TestTransferCallTool(t *testing.T) {
	callCtx, transport, _ := newTestContext()
	tool := NewTransferCallTool(callCtx)

	if tool.Name() != "transfer_call" {
		t.Errorf("expected name 'transfer_call', got %q", tool.Name())
	}

	params := tool.Parameters()
	if _, ok := params["target"]; !ok {
		t.Error("missing 'target' parameter")
	}
	if !params["target"].Required {
		t.Error("'target' should be required")
	}

	// Test successful transfer
	result, err := tool.Call(context.Background(), map[string]any{
		"target":   "+15551112222",
		"announce": "Connecting you now",
		"warm":     true,
	})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if !transport.transferCalled {
		t.Error("expected Transfer to be called")
	}
	if transport.lastTarget != "+15551112222" {
		t.Errorf("expected target '+15551112222', got %q", transport.lastTarget)
	}

	resultMap := result.(map[string]any)
	if resultMap["status"] != "success" {
		t.Errorf("expected status 'success', got %v", resultMap["status"])
	}
}

func TestTransferCallTool_MissingTarget(t *testing.T) {
	callCtx, _, _ := newTestContext()
	tool := NewTransferCallTool(callCtx)

	_, err := tool.Call(context.Background(), map[string]any{})
	if err == nil {
		t.Error("expected error for missing target")
	}
}

func TestTransferCallTool_TransportError(t *testing.T) {
	callCtx, transport, _ := newTestContext()
	transport.err = errors.New("connection failed")

	tool := NewTransferCallTool(callCtx)
	_, err := tool.Call(context.Background(), map[string]any{"target": "+15551112222"})
	if err == nil {
		t.Error("expected error from transport")
	}
}

func TestHoldCallTool(t *testing.T) {
	callCtx, transport, _ := newTestContext()
	tool := NewHoldCallTool(callCtx)

	if tool.Name() != "hold_call" {
		t.Errorf("expected name 'hold_call', got %q", tool.Name())
	}

	result, err := tool.Call(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if !transport.holdCalled {
		t.Error("expected Hold to be called")
	}

	resultMap := result.(map[string]any)
	if resultMap["status"] != "success" {
		t.Errorf("expected status 'success', got %v", resultMap["status"])
	}
}

func TestUnholdCallTool(t *testing.T) {
	callCtx, transport, _ := newTestContext()
	tool := NewUnholdCallTool(callCtx)

	if tool.Name() != "unhold_call" {
		t.Errorf("expected name 'unhold_call', got %q", tool.Name())
	}

	result, err := tool.Call(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if !transport.unholdCalled {
		t.Error("expected Unhold to be called")
	}

	resultMap := result.(map[string]any)
	if resultMap["status"] != "success" {
		t.Errorf("expected status 'success', got %v", resultMap["status"])
	}
}

func TestConsultAgentTool(t *testing.T) {
	callCtx, _, registry := newTestContext()
	tool := NewConsultAgentTool(callCtx)

	if tool.Name() != "consult_agent" {
		t.Errorf("expected name 'consult_agent', got %q", tool.Name())
	}

	params := tool.Parameters()
	if _, ok := params["agent_id"]; !ok {
		t.Error("missing 'agent_id' parameter")
	}
	if _, ok := params["query"]; !ok {
		t.Error("missing 'query' parameter")
	}

	result, err := tool.Call(context.Background(), map[string]any{
		"agent_id": "billing",
		"query":    "What is the customer's balance?",
	})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["status"] != "success" {
		t.Errorf("expected status 'success', got %v", resultMap["status"])
	}
	if resultMap["agent_id"] != "billing" {
		t.Errorf("expected agent_id 'billing', got %v", resultMap["agent_id"])
	}
	if resultMap["response"] != registry.consultResp {
		t.Errorf("expected response %q, got %v", registry.consultResp, resultMap["response"])
	}
}

func TestConsultAgentTool_AgentNotFound(t *testing.T) {
	callCtx, _, _ := newTestContext()
	tool := NewConsultAgentTool(callCtx)

	_, err := tool.Call(context.Background(), map[string]any{
		"agent_id": "nonexistent",
		"query":    "test",
	})
	if err == nil {
		t.Error("expected error for nonexistent agent")
	}
}

func TestConferenceTool(t *testing.T) {
	callCtx, transport, _ := newTestContext()
	tool := NewConferenceTool(callCtx)

	if tool.Name() != "add_to_conference" {
		t.Errorf("expected name 'add_to_conference', got %q", tool.Name())
	}

	result, err := tool.Call(context.Background(), map[string]any{
		"participants": []any{"+15551112222", "+15553334444"},
	})
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if !transport.conferenceCalled {
		t.Error("expected Conference to be called")
	}
	if len(transport.lastParticipants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(transport.lastParticipants))
	}

	resultMap := result.(map[string]any)
	if resultMap["status"] != "success" {
		t.Errorf("expected status 'success', got %v", resultMap["status"])
	}
}

func TestConferenceTool_EmptyParticipants(t *testing.T) {
	callCtx, _, _ := newTestContext()
	tool := NewConferenceTool(callCtx)

	_, err := tool.Call(context.Background(), map[string]any{
		"participants": []any{},
	})
	if err == nil {
		t.Error("expected error for empty participants")
	}
}

func TestToolsImplementInterface(t *testing.T) {
	callCtx, _, _ := newTestContext()

	// Verify all tools implement skill.Tool
	var _ skill.Tool = NewTransferCallTool(callCtx)
	var _ skill.Tool = NewHoldCallTool(callCtx)
	var _ skill.Tool = NewUnholdCallTool(callCtx)
	var _ skill.Tool = NewConsultAgentTool(callCtx)
	var _ skill.Tool = NewConferenceTool(callCtx)
}

func TestCallContext(t *testing.T) {
	call := &mockCall{id: "test-call"}
	transport := &mockTransport{}
	registry := &mockAgentRegistry{}

	ctx := NewCallContext(call, transport, registry)

	if ctx.GetCall().ID() != "test-call" {
		t.Error("GetCall returned wrong call")
	}
	if ctx.GetTransport() != transport {
		t.Error("GetTransport returned wrong transport")
	}
	if ctx.GetAgentRegistry() != registry {
		t.Error("GetAgentRegistry returned wrong registry")
	}
}
