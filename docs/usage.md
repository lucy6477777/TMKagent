# Usage Guide

## Who This Is For

This guide is for someone cloning the repo for the first time and trying to get a working local setup quickly, without reading the code first.

## Prerequisites

- Go 1.21+
- Node.js + npm
- PortAudio
- `OPENAI_API_KEY`
- `DEEPGRAM_API_KEY` for realtime stream mode
- LiveKit credentials only for RTC relay mode

Install PortAudio:

```bash
# macOS
brew install portaudio

# Ubuntu / Debian
sudo apt install portaudio19-dev libportaudio2
```

## First-Time Setup

```bash
git clone https://github.com/lucyliuu/mini-tmk-agent
cd mini-tmk-agent
cp .env.example .env
make build
```

Edit `.env` and set at least:

```bash
OPENAI_API_KEY=sk-...
```

## Smoke Test

The fastest way to verify the repo is healthy is transcript mode, because it only needs OpenAI:

```bash
./bin/mini-tmk-agent transcript --file testdata/hello_zh.wav --output result.txt --source-lang zh
cat result.txt
```

## Environment Variables

| Variable | Required | Used by |
| --- | --- | --- |
| `OPENAI_API_KEY` | Yes | transcript, translation, TTS, Web UI |
| `DEEPGRAM_API_KEY` | Stream only | realtime stream and Web realtime pages |
| `LIVEKIT_URL` | RTC only | cross-device relay |
| `LIVEKIT_API_KEY` | RTC only | cross-device relay |
| `LIVEKIT_API_SECRET` | RTC only | cross-device relay |
| `OPENAI_BASE_URL` | Optional | proxy / compatible OpenAI endpoint |
| `WEB_PUBLIC_BASE_URL` | Optional | phone-friendly RTC QR code |

## CLI Modes

### Transcript

```bash
./bin/mini-tmk-agent transcript --file speech.mp3 --output result.txt
./bin/mini-tmk-agent transcript --file speech.m4a --output result.txt
./bin/mini-tmk-agent transcript --file audio.pcm --output result.txt --source-lang zh
```

Supported input formats:

- `.wav`
- `.mp3`
- `.m4a`
- `.pcm`

### Stream

```bash
./bin/mini-tmk-agent stream --source-lang zh --target-lang en
```

Requires:

- `OPENAI_API_KEY`
- `DEEPGRAM_API_KEY`
- microphone access

With TTS:

```bash
./bin/mini-tmk-agent stream --source-lang zh --target-lang en --tts
./bin/mini-tmk-agent stream --source-lang zh --target-lang en --tts --tts-speaker-mode
```

### RTC Relay

Speaker:

```bash
./bin/mini-tmk-agent stream --source-lang zh --target-lang en --room my-room --role speaker
```

Listener:

```bash
./bin/mini-tmk-agent stream --source-lang zh --target-lang en --room my-room --role listener --tts
```

Requires:

- `OPENAI_API_KEY`
- `DEEPGRAM_API_KEY` for speaker side
- LiveKit credentials

## Web UI

Start the server:

```bash
./bin/mini-tmk-agent web --port 8080
```

Open:

```text
http://localhost:8080
```

Feature requirements:

- `文件转录`: `OPENAI_API_KEY`
- `实时翻译`: `OPENAI_API_KEY` + `DEEPGRAM_API_KEY`
- `RTC`: `OPENAI_API_KEY` + `DEEPGRAM_API_KEY` + LiveKit credentials

To use the RTC listener entry on another device in the same LAN:

```bash
WEB_PUBLIC_BASE_URL=http://YOUR_LAN_IP:8080
```

## Common Problems

### `OPENAI_API_KEY is not set`

Create `.env` and fill in your real key.

### `DEEPGRAM_API_KEY is required for stream mode`

This is expected for realtime translation. Transcript mode does not need Deepgram.

### Web app shows a placeholder page

Rebuild embedded frontend assets:

```bash
make web-build
```

### `make build` fails before Go compilation

The build also bundles the React frontend, so Node.js and npm must be installed.
