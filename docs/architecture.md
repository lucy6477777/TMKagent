# Architecture

## Goals

- Keep the core translation pipeline small and understandable
- Support both CLI and Web UI entry points
- Separate I/O concerns from pipeline logic where possible
- Make external services replaceable behind interfaces for testing

## High-Level Components

### CLI Layer

- `cmd/mini-tmk-agent/main.go`
- `cmd/mini-tmk-agent/web.go`

Responsibilities:

- parse flags
- load config
- wire dependencies
- choose runtime mode

### Config Layer

- `config/config.go`

Responsibilities:

- load `.env`
- merge environment variables
- apply CLI flag overrides

### Audio Layer

- `internal/audio/capture.go`
- `internal/audio/file.go`
- `internal/audio/vad.go`

Responsibilities:

- microphone capture via PortAudio
- local audio file normalization
- VAD logic for chunking

### ASR Layer

- `internal/asr/deepgram.go`
- `internal/asr/whisper.go`
- `internal/asr/simplify.go`

Responsibilities:

- streaming ASR for realtime mode
- batch ASR for transcript mode
- zh text normalization helpers

### Translation Layer

- `internal/translate/openai.go`

Responsibilities:

- convert recognized source text to target language
- optionally normalize zh source output

### TTS Layer

- `internal/tts/tts.go`
- `internal/tts/player.go`

Responsibilities:

- synthesize speech from translated text
- stream PCM output to PortAudio

### RTC Relay Layer

- `internal/rtc/livekit.go`

Responsibilities:

- publish and receive realtime subtitle messages over LiveKit data channels

### Web Layer

- `internal/web/server.go`
- `internal/web/handler.go`
- `web/src/*`

Responsibilities:

- serve SPA assets
- expose upload and websocket endpoints
- translate browser actions into pipeline commands

## Runtime Flows

### Transcript Flow

```text
audio file -> ReadAudioFile -> Whisper -> optional translation -> text output / websocket pairs
```

### Stream Flow

```text
microphone -> PortAudio capture -> Deepgram stream -> translation -> terminal/Web UI -> optional TTS
```

### RTC Flow

```text
speaker mic -> Deepgram -> translation -> LiveKit relay -> listener UI / listener TTS
```

## Web Protocol

### HTTP

- `POST /upload`: upload one audio file for transcript mode
- `GET /api/info`: provide local/public access info for QR generation
- `GET /`: serve SPA entry and client assets

### WebSocket

Commands:

- `start_stream`
- `stop`
- `transcript`
- `rtc_speaker_start`
- `rtc_join`
- `rtc_stop`

Messages:

- `status`
- `interim`
- `pair`
- `progress`
- `error`

## Design Constraints

- external API failures should degrade gracefully where possible
- terminal and web clients should share the same conceptual message model
- RTC listener mode must avoid microphone dependencies
- uploaded transcript files are processed one-at-a-time per active websocket session

## Testing Strategy

The repo uses three layers:

- unit tests for pure logic and adapters
- local mock-server tests for OpenAI-compatible clients
- tagged integration tests for real API calls

See [testing.md](testing.md) for the full workflow.

## Documentation Policy

The README is the project entry page.

Longer-form documents live in `docs/`:

- `usage.md`: operator-facing setup and commands
- `architecture.md`: design overview and boundaries
- `testing.md`: quality gates and coverage workflow

Historical implementation notes and UI specs remain in `docs/superpowers/specs/`.
