package voicetools

import "errors"

// Common errors.
var (
	// ErrNoCallContext is returned when CallContext is not available.
	ErrNoCallContext = errors.New("call context not available")

	// ErrNoTransport is returned when TelephonyTransport is not available.
	ErrNoTransport = errors.New("telephony transport not available")

	// ErrNoRegistry is returned when AgentRegistry is not available.
	ErrNoRegistry = errors.New("agent registry not available")

	// ErrAgentNotFound is returned when a specified agent is not found.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrInvalidParameter is returned when a parameter is invalid.
	ErrInvalidParameter = errors.New("invalid parameter")
)
