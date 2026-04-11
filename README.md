# mini-tmk-agent

A simultaneous interpretation CLI agent built with Go. Supports real-time microphone translation, audio file transcription, and TTS audio output.

Powered by OpenAI Whisper-1 (ASR) / Deepgram Nova-2 (streaming ASR), GPT-4o-mini (translation), and OpenAI TTS-1 (text-to-speech).

---

## Prerequisites

**macOS:**
```bash
brew install portaudio
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt install portaudio19-dev libportaudio2
```

**Go 1.21+** and an **OpenAI API key** are required. A **Deepgram API key** is optional but recommended for low-latency streaming.

---

## Install

```bash
git clone https://github.com/lucyliuu/mini-tmk-agent
cd mini-tmk-agent
make build
```

The binary is created at `bin/mini-tmk-agent`.

---

## Configure

```bash
export OPENAI_API_KEY=sk-...

# Optional: Deepgram key enables streaming ASR with interim results (~200ms latency)
# Sign up at https://console.deepgram.com ($200 free credits, no credit card)
export DEEPGRAM_API_KEY=dg-...

# Optional: override API endpoint (e.g. for a proxy)
export OPENAI_BASE_URL=https://api.openai.com/v1
```

---

## Usage

### Stream Mode — Real-time simultaneous interpretation

Speak into your microphone. The terminal displays source and translated text simultaneously.

```bash
# Basic (Whisper batch ASR — higher latency)
./bin/mini-tmk-agent stream --source-lang zh --target-lang en

# With Deepgram streaming ASR (low latency, words appear as you speak)
DEEPGRAM_API_KEY=dg-... ./bin/mini-tmk-agent stream --source-lang zh --target-lang en
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

Press `Ctrl+C` to stop.

**Supported language codes:** `zh` (Chinese) · `en` (English) · `es` (Spanish) · `ja` (Japanese)

### Transcript Mode — Transcribe an audio file

```bash
./bin/mini-tmk-agent transcript --file speech.mp3 --output result.txt
./bin/mini-tmk-agent transcript --file audio.pcm --output result.txt --source-lang zh
```

**Supported input formats:** `.wav` · `.mp3` · `.pcm`

### Web UI

```bash
./bin/mini-tmk-agent web --port 8080
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

# Integration tests — requires OPENAI_API_KEY
OPENAI_API_KEY=sk-... make test-integration
```

---

## Architecture

```
Stream mode (Deepgram streaming ASR):
  mic → portaudio → raw PCM frames → Deepgram WebSocket
    interim results → terminal (overwriting gray text)
    final results → GPT-4o-mini translate → terminal display
                                          → TTS goroutine → PortAudio playback

Stream mode (Whisper batch ASR, fallback):
  mic → portaudio → VAD (RMS energy) → audioCh
  audioCh → Whisper-1 ASR → asrCh
  asrCh → GPT-4o-mini translate → translateCh
  translateCh → terminal display → TTS goroutine → PortAudio playback

Transcript mode:
  audio file (.wav/.mp3/.pcm) → Whisper-1 ASR → .txt output
```

**Streaming ASR (Deepgram):** Audio frames are streamed to Deepgram via WebSocket in real-time. The server handles VAD and endpointing, returning interim results (words appear as you speak) and final results (complete utterance). This reduces perceived latency from ~4s to ~200ms.

**TTS:** Translation text is sent to OpenAI TTS-1 asynchronously via a dedicated goroutine. Audio is streamed back in PCM 24kHz format and played through PortAudio. A non-blocking channel ensures TTS never blocks the main pipeline.

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
  file.go                     WAV/MP3/PCM file reading + PCMToWAV
internal/asr/
  whisper.go                  Whisper-1 batch ASR client (Client interface)
  deepgram.go                 Deepgram streaming ASR (StreamClient/StreamSession)
internal/translate/openai.go  GPT-4o-mini translation (Client interface)
internal/tts/
  tts.go                      TTS client interface + OpenAI TTS-1 implementation
  player.go                   PortAudio PCM playback with IsPlaying flag
internal/pipeline/
  stream.go                   Streaming + batch real-time pipelines with TTS
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

See [`docs/superpowers/specs/`](docs/superpowers/specs/) for the full architecture design document.
