# Voice Tools

The `voicetools` package provides AI agent tools for controlling voice calls, enabling agents to transfer calls, place callers on hold, consult specialist agents, and manage conference calls.

## Overview

Voice tools enable AI agents to:

- **Transfer calls** - Route to human agents or other queues
- **Hold/Unhold** - Manage caller wait states with music
- **Consult agents** - Query specialists without transferring
- **Conference** - Add participants to multi-party calls

## Quick Start

```go
import (
    "github.com/plexusone/omniskill/voicetools"
    runtime "github.com/plexusone/omniskill/mcp/server"
)

// Create call context with transport and agent registry
callCtx := voicetools.NewCallContext(call, transport, agentRegistry)

// Create voice skill with all tools
voiceSkill := voicetools.NewVoiceSkill(callCtx)

// Register with MCP server
rt := runtime.New(impl, nil)
rt.RegisterSkill(voiceSkill)

// Or use tools directly
transferTool := voicetools.NewTransferCallTool(callCtx)
result, err := transferTool.Call(ctx, map[string]any{
    "target": "+15551234567",
    "warm":   true,
})
```

## Interfaces

### CallContext

Provides access to call state and transport:

```go
type CallContext interface {
    GetCall() Call
    GetTransport() TelephonyTransport
    GetAgentRegistry() AgentRegistry
}
```

### Call

Represents the current call:

```go
type Call interface {
    ID() string
    From() string
    To() string
    Direction() string
    Duration() time.Duration
    Metadata() map[string]string
}
```

### TelephonyTransport

Implements call control operations:

```go
type TelephonyTransport interface {
    Transfer(callID, target, announce string, warm bool) error
    Hold(callID, music string) error
    Unhold(callID string) error
    Conference(callID string, participants []string) error
    Mute(callID string) error
    Unmute(callID string) error
    Hangup(callID string) error
}
```

### AgentRegistry

Provides access to specialist agents:

```go
type AgentRegistry interface {
    GetAgent(id string) (AgentConfig, bool)
    ListAgents() []string
    ConsultAgent(agentID, query string, context map[string]any) (string, error)
}
```

## Available Tools

### transfer_call

Transfer the current call to another number or agent queue:

```go
// Tool parameters
{
    "target":   "+15551234567",  // Required: E.164 number or queue ID
    "announce": "Connecting...", // Optional: Message during transfer
    "warm":     true             // Optional: Stay on line until answered
}

// Result
{
    "status":  "success",
    "call_id": "CA123456",
    "target":  "+15551234567"
}
```

**Usage by AI agent:**

> "I'll transfer you to our billing department now."
> *Agent calls transfer_call with target="billing-queue"*

### hold_call

Place the caller on hold:

```go
// Tool parameters
{
    "music": "jazz"  // Optional: Hold music style
}

// Result
{
    "status":  "success",
    "call_id": "CA123456"
}
```

**Usage by AI agent:**

> "Please hold while I look that up for you."
> *Agent calls hold_call*

### unhold_call

Resume the call from hold:

```go
// Tool parameters (none required)
{}

// Result
{
    "status":  "success",
    "call_id": "CA123456"
}
```

### consult_agent

Query a specialist AI agent without transferring the call:

```go
// Tool parameters
{
    "agent_id": "billing",                    // Required: Specialist agent ID
    "query":    "What is the customer balance?" // Required: Question for specialist
}

// Result
{
    "status":   "success",
    "agent_id": "billing",
    "response": "The customer's balance is $142.50 due on Jan 15."
}
```

**Usage by AI agent:**

> "Let me check your balance..."
> *Agent calls consult_agent to get info from billing specialist*
> "Your current balance is $142.50, due on January 15th."

### add_to_conference

Add participants to a conference call:

```go
// Tool parameters
{
    "participants": ["+15551234567", "+15559876543"]  // Required: Numbers to add
}

// Result
{
    "status":       "success",
    "call_id":      "CA123456",
    "participants": ["+15551234567", "+15559876543"]
}
```

## Integration with Voice Gateways

### Twilio

```go
import (
    "github.com/plexusone/omnivoice-core/callsystem/twilio"
    "github.com/plexusone/omniskill/voicetools"
)

// Twilio transport implements TelephonyTransport
twilioGateway := twilio.NewGateway(config)

twilioGateway.OnCall(func(call twilio.Call) {
    // Create call context
    callCtx := voicetools.NewCallContext(
        adaptCall(call),
        adaptTransport(twilioGateway),
        agentRegistry,
    )

    // Register voice tools
    voiceSkill := voicetools.NewVoiceSkill(callCtx)
    runtime.RegisterSkill(voiceSkill)
})
```

### LiveKit

```go
import (
    "github.com/plexusone/omnivoice-core/callsystem/livekit"
    "github.com/plexusone/omniskill/voicetools"
)

livekitGateway := livekit.NewGateway(config)

livekitGateway.OnParticipantJoined(func(participant livekit.Participant) {
    callCtx := voicetools.NewCallContext(
        adaptParticipant(participant),
        adaptTransport(livekitGateway),
        agentRegistry,
    )

    voiceSkill := voicetools.NewVoiceSkill(callCtx)
    runtime.RegisterSkill(voiceSkill)
})
```

## Agent Registry Implementation

```go
type MyAgentRegistry struct {
    agents map[string]*Agent
}

func (r *MyAgentRegistry) GetAgent(id string) (voicetools.AgentConfig, bool) {
    agent, ok := r.agents[id]
    if !ok {
        return voicetools.AgentConfig{}, false
    }
    return voicetools.AgentConfig{
        ID:          agent.ID,
        Name:        agent.Name,
        Description: agent.Description,
    }, true
}

func (r *MyAgentRegistry) ListAgents() []string {
    ids := make([]string, 0, len(r.agents))
    for id := range r.agents {
        ids = append(ids, id)
    }
    return ids
}

func (r *MyAgentRegistry) ConsultAgent(agentID, query string, context map[string]any) (string, error) {
    agent, ok := r.agents[agentID]
    if !ok {
        return "", fmt.Errorf("agent not found: %s", agentID)
    }

    // Call the specialist agent
    response, err := agent.Query(query, context)
    return response, err
}
```

## Error Handling

All tools return errors for:

- Missing required parameters
- Invalid parameter types
- Transport failures
- Agent not found (consult_agent)

```go
result, err := tool.Call(ctx, params)
if err != nil {
    switch {
    case errors.Is(err, voicetools.ErrNoCallContext):
        // CallContext not provided
    case errors.Is(err, voicetools.ErrAgentNotFound):
        // Specialist agent not registered
    default:
        // Transport or other error
    }
}
```

## Best Practices

1. **Validate before transfer** - Confirm with user before transferring
2. **Use warm transfers** - Stay on line when possible
3. **Announce holds** - Tell caller what you're doing
4. **Consult before escalating** - Try specialist agents first
5. **Handle errors gracefully** - Inform user if action fails
