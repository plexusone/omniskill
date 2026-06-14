package voicetools

import "time"

// CallContext provides access to the current call and telephony operations.
type CallContext interface {
	// GetCall returns the current call information.
	GetCall() Call

	// GetTransport returns the telephony transport for call control.
	GetTransport() TelephonyTransport

	// GetAgentRegistry returns the registry for specialist agents.
	GetAgentRegistry() AgentRegistry
}

// Call represents the current voice call.
type Call interface {
	// ID returns the call identifier.
	ID() string

	// From returns the caller's phone number (E.164 format).
	From() string

	// To returns the called number (E.164 format).
	To() string

	// Direction returns "inbound" or "outbound".
	Direction() string

	// Duration returns the call duration.
	Duration() time.Duration

	// Metadata returns call metadata.
	Metadata() map[string]string
}

// TelephonyTransport provides call control operations.
type TelephonyTransport interface {
	// Transfer transfers the call to another destination.
	// target: E.164 phone number or queue ID
	// announce: message to play to the target before connecting (optional)
	// warm: if true, stay on line until target answers
	Transfer(callID, target string, announce string, warm bool) error

	// Hold places the call on hold.
	// music: URL to hold music or "default" for provider default
	Hold(callID string, music string) error

	// Unhold takes the call off hold.
	Unhold(callID string) error

	// Conference creates or joins a conference.
	// participants: E.164 numbers to add to the conference
	Conference(callID string, participants []string) error

	// Mute mutes the caller's audio.
	Mute(callID string) error

	// Unmute unmutes the caller's audio.
	Unmute(callID string) error

	// Hangup ends the call.
	Hangup(callID string) error
}

// AgentRegistry provides access to specialist agents.
type AgentRegistry interface {
	// GetAgent returns an agent configuration by ID.
	GetAgent(id string) (AgentConfig, bool)

	// ListAgents returns all available agent IDs.
	ListAgents() []string

	// ConsultAgent sends a query to a specialist agent and returns the response.
	ConsultAgent(agentID, query string, context map[string]any) (string, error)
}

// AgentConfig describes a specialist agent.
type AgentConfig struct {
	// ID is the agent identifier.
	ID string

	// Name is a human-readable name.
	Name string

	// Description describes what the agent specializes in.
	Description string

	// Capabilities lists what the agent can do.
	Capabilities []string
}

// callContext is the default implementation of CallContext.
type callContext struct {
	call      Call
	transport TelephonyTransport
	registry  AgentRegistry
}

// NewCallContext creates a new CallContext.
func NewCallContext(call Call, transport TelephonyTransport, registry AgentRegistry) CallContext {
	return &callContext{
		call:      call,
		transport: transport,
		registry:  registry,
	}
}

func (c *callContext) GetCall() Call                    { return c.call }
func (c *callContext) GetTransport() TelephonyTransport { return c.transport }
func (c *callContext) GetAgentRegistry() AgentRegistry  { return c.registry }
