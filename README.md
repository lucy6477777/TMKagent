# mini-tmk-agent

A simultaneous interpretation CLI agent built with Go. Supports real-time microphone translation, audio file transcription, and TTS audio output.

Powered by OpenAI Whisper-1 (ASR) / Deepgram Nova-2 (streaming ASR), GPT-4o-mini (translation), OpenAI TTS-1 (text-to-speech), and LiveKit (RTC cross-device relay).

---

## Prerequisites

You need:

- **Go 1.21+**
- **Node.js + npm** for `make build` and the embedded Web UI bundle
- **PortAudio** for microphone capture and TTS playback
- **OPENAI_API_KEY** for all modes
- **DEEPGRAM_API_KEY** for real-time stream mode
- **`LIVEKIT_URL` / `LIVEKIT_API_KEY` / `LIVEKIT_API_SECRET`** only for RTC relay

**macOS:**

```bash
brew install portaudio
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt install portaudio19-dev libportaudio2
```

If you only want file transcription, `OPENAI_API_KEY` is enough.

---

## Quick Start

1. Clone and build:

```bash
git clone https://github.com/lucyliuu/mini-tmk-agent
cd mini-tmk-agent
make build
```

2. Create `.env`:

```bash
cp .env.example .env
```

3. Fill in at least:

```bash
OPENAI_API_KEY=sk-...
```

4. Smoke test with the included sample audio:

```bash
./bin/mini-tmk-agent transcript --file testdata/hello_zh.wav --output result.txt --source-lang zh
cat result.txt
```

If that works, your local setup is good.

---

## Build Output

After `make build`, the binary is created at `bin/mini-tmk-agent`.

---

## Configure

```bash
cp .env.example .env
# Edit .env and fill in your API keys
```

The program automatically loads `.env` on startup — no need to `export` manually. Open `.env` and fill in:

| Variable              | Required?                | Used by                              | Where to get it                                             |
| --------------------- | ------------------------ | ------------------------------------ | ----------------------------------------------------------- |
| `OPENAI_API_KEY`      | Yes                      | transcript / web / translation / TTS | [platform.openai.com](https://platform.openai.com/api-keys) |
| `DEEPGRAM_API_KEY`    | Yes for real-time stream | CLI `stream`, Web realtime pages     | [console.deepgram.com](https://console.deepgram.com)        |
| `LIVEKIT_URL`         | Optional                 | RTC relay only                       | [cloud.livekit.io](https://cloud.livekit.io)                |
| `LIVEKIT_API_KEY`     | Optional                 | RTC relay only                       | Same as above                                               |
| `LIVEKIT_API_SECRET`  | Optional                 | RTC relay only                       | Same as above                                               |
| `OPENAI_BASE_URL`     | Optional                 | Custom proxy / compatible endpoint   | Your provider                                               |
| `WEB_PUBLIC_BASE_URL` | Optional                 | Web QR code / phone access           | Your own public URL                                         |

> **Note:** `.env` is git-ignored. Your keys stay local and never get pushed to GitHub.

---

## Usage

### Stream Mode — Real-time simultaneous interpretation

Speak into your microphone. The terminal displays source and translated text simultaneously.

This mode requires:

- `OPENAI_API_KEY`
- `DEEPGRAM_API_KEY`
- a working microphone
- PortAudio installed

```bash
# Basic real-time translation
./bin/mini-tmk-agent stream --source-lang zh --target-lang en

# If you prefer CLI flags over .env
./bin/mini-tmk-agent --deepgram-api-key dg-... stream --source-lang zh --target-lang en
```

Output example:

```
[...] 今天天气        ← interim (gray, updating as you speak)
[SRC] 今天天气真好
[TGT] The weather is nice today
```

**With TTS (text-to-speech) output:**

```bash
# Headphone mode (full-duplex: mic + TTS work simultaneously)
./bin/mini-tmk-agent stream --source-lang zh --target-lang en --tts

# Speaker mode (half-duplex: mic pauses during TTS playback)
./bin/mini-tmk-agent stream --source-lang zh --target-lang en --tts --tts-speaker-mode

# Custom TTS voice
./bin/mini-tmk-agent stream --source-lang zh --target-lang en --tts --tts-voice nova
```

> **Headphones recommended** when using TTS. Without them, TTS audio may be picked up by the microphone, causing a feedback loop. Speaker mode mitigates this by pausing the mic during playback but loses speech during that time.

**Cross-device RTC mode (LiveKit):**

```bash
# Terminal A (speaker): speaks into mic, translations sent to LiveKit room
./bin/mini-tmk-agent stream --source-lang zh --target-lang en \
  --room my-room --role speaker

# Terminal B (listener, can be on a different machine): receives translations + TTS
./bin/mini-tmk-agent stream --source-lang zh --target-lang en \
  --room my-room --role listener --tts
```

Speaker A talks in Chinese → Listener B sees English text and hears English TTS, in real time, across the internet.

Press `Ctrl+C` to stop.

**Supported language codes:** `zh` (Chinese) · `en` (English) · `es` (Spanish) · `ja` (Japanese)

### Transcript Mode — Transcribe an audio file

This mode requires only:

- `OPENAI_API_KEY`

```bash
./bin/mini-tmk-agent transcript --file speech.mp3 --output result.txt
./bin/mini-tmk-agent transcript --file speech.m4a --output result.txt
./bin/mini-tmk-agent transcript --file audio.pcm --output result.txt --source-lang zh
```

**Supported input formats:** `.wav` · `.mp3` · `.m4a` · `.pcm`

### Web UI

```bash
./bin/mini-tmk-agent web --port 8080
```

Then open:

```text
http://localhost:8080
```

What works in the Web UI:

- `文件转录`: requires `OPENAI_API_KEY`
- `实时翻译`: requires `OPENAI_API_KEY` + `DEEPGRAM_API_KEY`
- `RTC`: requires `OPENAI_API_KEY` + `DEEPGRAM_API_KEY` + LiveKit credentials

If you want to open the RTC listener page from your phone, set:

```bash
WEB_PUBLIC_BASE_URL=http://YOUR_LAN_IP:8080
```

### Global flags

```
--api-key string            Override OPENAI_API_KEY environment variable
--base-url string           Override OPENAI_BASE_URL environment variable
--deepgram-api-key string   Override DEEPGRAM_API_KEY environment variable
```

---

## Run Tests

```bash
# Unit tests — no API key required
make test

# Coverage report
make test-cover

# Integration tests — requires OPENAI_API_KEY
OPENAI_API_KEY=sk-... make test-integration
```

---

## Troubleshooting

### `OPENAI_API_KEY is not set`

Create `.env` from `.env.example`, then put your real key in:

```bash
cp .env.example .env
```

### `DEEPGRAM_API_KEY is required for stream mode`

This is expected for real-time translation. Transcript mode does not need Deepgram.

### PortAudio errors

Install PortAudio first:

```bash
# macOS
brew install portaudio

# Ubuntu / Debian
sudo apt install portaudio19-dev libportaudio2
```

### `make build` fails before Go compilation

`make build` also builds the React frontend, so Node.js and npm must be installed.

### Web UI shows a placeholder page instead of the app

Rebuild the embedded frontend assets:

```bash
make web-build
```

---

## Architecture

```
Stream mode (Deepgram streaming ASR):
  mic → portaudio → raw PCM frames → Deepgram WebSocket
    interim results → terminal (overwriting gray text)
    final results → GPT-4o-mini translate → terminal display
                                          → TTS goroutine → PortAudio playback

Transcript mode:
  audio file (.wav/.mp3/.m4a/.pcm) → Whisper-1 ASR → .txt output
```

**Streaming ASR (Deepgram):** Audio frames are streamed to Deepgram via WebSocket in real-time. The server handles VAD and endpointing, returning interim results (words appear as you speak) and final results (complete utterance). This reduces perceived latency from ~4s to ~200ms.

**TTS:** Translation text is sent to OpenAI TTS-1 asynchronously via a dedicated goroutine. Audio is streamed back in PCM 24kHz format and played through PortAudio. A non-blocking channel ensures TTS never blocks the main pipeline.

**RTC Relay (LiveKit):** In speaker mode, interim/final results are published to a LiveKit room via WebRTC Data Channel (reliable mode). In listener mode, no microphone is needed — the pipeline receives messages from the room and feeds them to display + TTS. LiveKit Cloud handles NAT traversal and global routing.

**Speaker mode:** An `atomic.Bool` flag tracks TTS playback state. When enabled, the audio-sending goroutine skips frames during playback to prevent feedback loops.

**Graceful shutdown:** `Ctrl+C` triggers `context.Cancel()`, propagates through all goroutines via channel cascade.

---

## Project Structure

```
cmd/mini-tmk-agent/
  main.go                     CLI entry point (cobra): stream, transcript, web
  web.go                      Web UI subcommand
config/config.go              API key + base URL loading (OpenAI + Deepgram)
internal/audio/
  capture.go                  PortAudio microphone capture (16kHz mono)
  vad.go                      RMS energy VAD state machine
  file.go                     WAV/MP3/M4A/PCM file reading + PCMToWAV
internal/asr/
  whisper.go                  Whisper-1 batch ASR client (Client interface)
  deepgram.go                 Deepgram streaming ASR (StreamClient/StreamSession)
internal/translate/openai.go  GPT-4o-mini translation (Client interface)
internal/tts/
  tts.go                      TTS client interface + OpenAI TTS-1 implementation
  player.go                   PortAudio PCM playback with IsPlaying flag
internal/rtc/
  livekit.go                  LiveKit RTC client (connect, send, receive via Data Channel)
internal/pipeline/
  stream.go                   Streaming real-time pipelines with TTS
  transcript.go               Sequential file transcription
  metrics.go                  JSONL latency/cost metrics
internal/display/terminal.go  ANSI terminal: interim (overwrite) + final display
internal/web/                 Web UI backend (WebSocket + embedded SPA)
web/                          React + TypeScript frontend source
tests/unit/                   Unit tests (no API key needed)
tests/integration/            Integration tests (real API)
testdata/hello_zh.wav         Test audio fixture
```

---

## Design

Project docs:

- [docs/usage.md](docs/usage.md) — installation, configuration, CLI and Web UI walkthrough
- [docs/architecture.md](docs/architecture.md) — system design, component boundaries, runtime flows
- [docs/testing.md](docs/testing.md) — test strategy, coverage workflow, how to add tests
- [`docs/superpowers/specs/`](docs/superpowers/specs/) — implementation-era design notes and UI specs
