// Package voicetools provides voice call control tools for AI agents.
//
// These tools enable AI agents to perform telephony operations during voice
// conversations, including call transfers, hold management, agent consultation,
// and conferencing.
//
// # Available Tools
//
//   - transfer_call: Transfer the call to another number or agent queue
//   - hold_call: Place the caller on hold
//   - unhold_call: Take the caller off hold
//   - consult_agent: Query a specialist agent without transferring
//   - add_to_conference: Add another participant to the call
//
// # Usage
//
//	// Create a CallContext with your telephony transport
//	ctx := voicetools.NewCallContext(call, transport, agentRegistry)
//
//	// Create the voice skill with all tools
//	skill := voicetools.NewVoiceSkill(ctx)
//
//	// Register with MCP server
//	server.RegisterSkill(skill)
//
// # Integration
//
// The tools require a CallContext that provides access to:
//   - The current call (for call metadata)
//   - The telephony transport (for SIP/WebRTC operations)
//   - An agent registry (for specialist agent lookup)
//
// See the context.go file for interface definitions.
package voicetools
