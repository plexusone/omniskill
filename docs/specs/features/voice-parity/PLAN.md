# Voice Feature Parity Implementation Plan

Full voice calling feature parity for omniagent+omnivoice with openclaw.

## Status

| Phase | Module | Package | Status |
|-------|--------|---------|--------|
| 1 | omnivoice-core | `storage/` | Completed |
| 1 | omnivoice-core | `bargein/` | Completed |
| 2 | omni-openai | `omnivoice/realtime/` | Completed |
| 2 | omni-google | `omnivoice/` | Completed |
| 3 | omniskill | `voicetools/` | Completed |

## Summary

| Feature | Module | Location |
|---------|--------|----------|
| OpenAI Realtime API | omni-openai | `omnivoice/realtime/` |
| Gemini Live API | omni-google | `omnivoice/` |
| Barge-in Detection | omnivoice-core | `bargein/` |
| Call State Persistence | omnivoice-core | `storage/` |
| Agent Consult Tools | omniskill | `voicetools/` |

---

## Phase 1: Core Infrastructure - Completed

### omnivoice-core/storage/

Session state persistence with pluggable backends:

```go
type SessionStore interface {
    Save(ctx context.Context, session *SessionState) error
    Load(ctx context.Context, sessionID string) (*SessionState, error)
    Delete(ctx context.Context, sessionID string) error
    ListActive(ctx context.Context) ([]string, error)
    UpdateHeartbeat(ctx context.Context, sessionID string) error
    Close() error
}
```

**Implementations:**

- `MemoryStore` - In-memory for development/testing
- `RedisStore` - Redis-backed for production multi-instance

**Files:**

- `store.go` - Interface definitions
- `types.go` - SessionState, Turn, SessionMetrics
- `memory.go` - In-memory implementation
- `redis.go` - Redis implementation

### omnivoice-core/bargein/

Barge-in detection for natural interruption handling:

```go
type Detector struct {
    config        Config
    ttsPipeline   interface{ IsActive() bool; Stop() }
    onInterrupt   func(Event)
}

func (d *Detector) AttachTTS(p interface{ IsActive() bool; Stop() })
func (d *Detector) AttachSTTEvents(events <-chan STTEvent)
func (d *Detector) OnInterrupt(handler func(Event))
```

**Interruption Modes:**

- `ModeImmediate` - Interrupt as soon as user speech detected
- `ModeAfterSentence` - Wait for natural pause
- `ModeDisabled` - Ignore user speech during agent output

---

## Phase 2: Realtime APIs - Completed

### omni-openai/omnivoice/realtime/

OpenAI Realtime API WebSocket client for native voice-to-voice (~100ms latency):

```go
type RealtimeProvider struct {
    client *Client
    config Config
}

func (p *RealtimeProvider) ProcessAudioStream(
    ctx context.Context,
    audioIn <-chan []byte,
    config ProcessConfig,
) (<-chan AudioChunk, <-chan Transcript, error)
```

**Audio Format:**

- Input: PCM16 24kHz mono
- Output: PCM16 24kHz mono

**Features:**

- Session management via WebSocket
- Function calling support
- Voice Activity Detection (VAD)
- 11 voice options

### omni-google/omnivoice/

Gemini Live API WebSocket client for real-time voice-to-voice:

```go
type RealtimeProvider struct {
    apiKey string
    config Config
}

func (p *RealtimeProvider) ProcessAudioStream(
    ctx context.Context,
    audioIn <-chan []byte,
    config ProcessConfig,
) (<-chan AudioChunk, <-chan Transcript, error)
```

**Audio Format:**

- Input: PCM16 16kHz mono
- Output: PCM16 24kHz mono

**Features:**

- Bidirectional WebSocket streaming
- Function calling support
- Google Search grounding
- Code execution
- 5 voice options

---

## Phase 3: Voice Tools - Completed

### omniskill/voicetools/

AI agent tools for voice call control:

```go
func NewVoiceSkill(callCtx CallContext) *skill.Skill {
    return skill.New("voice_control",
        skill.WithTool(NewTransferCallTool(callCtx)),
        skill.WithTool(NewHoldCallTool(callCtx)),
        skill.WithTool(NewUnholdCallTool(callCtx)),
        skill.WithTool(NewConsultAgentTool(callCtx)),
        skill.WithTool(NewConferenceTool(callCtx)),
    )
}
```

**Tools:**

| Tool | Description |
|------|-------------|
| `transfer_call` | Transfer to number or queue |
| `hold_call` | Place caller on hold |
| `unhold_call` | Resume from hold |
| `consult_agent` | Query specialist without transfer |
| `add_to_conference` | Add participants |

---

## Verification

All packages pass tests:

```bash
cd ~/go/src/github.com/plexusone/omnivoice-core && go test ./storage/... ./bargein/...
cd ~/go/src/github.com/plexusone/omni-openai && go test ./omnivoice/realtime/...
cd ~/go/src/github.com/plexusone/omni-google && go test ./omnivoice/...
cd ~/go/src/github.com/plexusone/omniskill && go test ./voicetools/...
```

All packages pass linting:

```bash
golangci-lint run
```

---

## Dependencies Added

| Module | Dependency | Purpose |
|--------|------------|---------|
| omnivoice-core | `github.com/redis/go-redis/v9` | Redis session store |
| omni-openai | `github.com/gorilla/websocket` | WebSocket client |
| omni-google | `github.com/gorilla/websocket` | WebSocket client |

---

## Environment Variables

```bash
# OpenAI Realtime
OPENAI_API_KEY=sk-...

# Gemini Live
GOOGLE_API_KEY=...
# or GEMINI_API_KEY=...

# Redis (optional, for persistence)
REDIS_URL=redis://localhost:6379
```

---

## Documentation

Updated documentation for each module:

| Module | README | MkDocs Pages |
|--------|--------|--------------|
| omnivoice-core | Updated with storage/, bargein/ | storage.md, bargein.md |
| omni-openai | Updated with realtime/ | providers/realtime.md |
| omni-google | Updated with omnivoice/ | omnivoice/index.md |
| omniskill | Updated with voicetools/ | voicetools/index.md |
