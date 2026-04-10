# mini-tmk-agent

A simultaneous interpretation CLI agent built with Go. Supports real-time microphone translation and audio file transcription.

Powered by OpenAI Whisper-1 (ASR) and GPT-4o-mini (translation).

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

**Go 1.21+** and an **OpenAI API key** are also required.

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
# Optional: override API endpoint (e.g. for a proxy)
export OPENAI_BASE_URL=https://api.openai.com/v1
```

---

## Usage

### Stream Mode — Real-time simultaneous interpretation

Speak into your microphone. The terminal displays source and translated text simultaneously.

```bash
./bin/mini-tmk-agent stream --source-lang zh --target-lang en
```

Output example:
```
[SRC] 你好，欢迎来到时空壶
[TGT] Hello, welcome to Timekettle
```

Press `Ctrl+C` to stop.

**Supported language codes:** `zh` (Chinese) · `en` (English) · `es` (Spanish) · `ja` (Japanese)

### Transcript Mode — Transcribe an audio file

```bash
./bin/mini-tmk-agent transcript --file speech.mp3 --output result.txt
./bin/mini-tmk-agent transcript --file audio.pcm --output result.txt --source-lang zh
```

**Supported input formats:** `.wav` · `.mp3` · `.pcm`

### Global flags

```
--api-key string    Override OPENAI_API_KEY environment variable
--base-url string   Override OPENAI_BASE_URL environment variable
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
Stream mode:
  mic → portaudio callback → VAD (RMS energy) → audioCh
  audioCh → Whisper-1 ASR → asrCh
  asrCh → GPT-4o-mini translate → translateCh
  translateCh → ANSI terminal display

Transcript mode:
  audio file (.wav/.mp3/.pcm) → Whisper-1 ASR → .txt output
```

Four goroutines connected by typed channels enable pipeline parallelism: ASR and translation of consecutive speech chunks run concurrently.

**VAD:** Pure Go RMS energy threshold — no external dependency. Detects speech start/end and emits complete chunks.

**Graceful shutdown:** `Ctrl+C` triggers `context.Cancel()`, propagates through all goroutines via channel cascade.

---

## Project Structure

```
cmd/mini-tmk-agent/main.go    CLI entry point (cobra)
config/config.go              API key + base URL loading
internal/audio/
  capture.go                  portaudio microphone capture
  vad.go                      RMS energy VAD state machine
  file.go                     WAV/MP3/PCM file reading + PCMToWAV
internal/asr/whisper.go       Whisper-1 ASR client interface
internal/translate/openai.go  GPT-4o-mini translation client interface
internal/pipeline/
  stream.go                   4-goroutine real-time pipeline
  transcript.go               Sequential file transcription
internal/display/terminal.go  ANSI coloured bilingual output
tests/unit/                   Unit tests (no API key needed)
tests/integration/            Integration tests (real API)
testdata/hello_zh.wav         Test audio fixture
```

---

## Design

See [`docs/superpowers/specs/`](docs/superpowers/specs/) for the full architecture design document.
