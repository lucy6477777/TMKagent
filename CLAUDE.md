# CLAUDE.md — mini-tmk-agent

## 项目概述

Go CLI 同声传译 Agent，面试作业项目。支持两种运行模式 + Web UI：

- `stream`：实时麦克风采集 → Deepgram 流式 ASR（interim + final）→ GPT 翻译 → 终端双语显示 → 可选 TTS 语音输出
- `transcript`：音频文件（WAV/MP3/M4A/PCM）→ Whisper ASR → 纯文本输出
- `web`：Go HTTP server + WebSocket + 嵌入式 React SPA

目标平台：Linux / macOS 演示环境。

---

## 行为准则

> 这些准则以谨慎优先于速度。对于简单任务，酌情使用。

### 1. 先思考，再编码

**不要假设。不要隐藏困惑。主动暴露权衡。**

开始实现前：

- 明确陈述你的假设。如有不确定，先问。
- 如果存在多种理解，逐一列出，不要私自选择。
- 如果有更简单的方案，说出来。必要时主动反驳。
- 如果某处不清晰，停下来，说明哪里不清晰，然后提问。

### 2. 简单优先

**最少的代码解决问题。不写投机性代码。**

- 不做超出要求的功能。
- 单次使用的代码不要抽象。
- 不要引入没有被要求的"灵活性"或"可配置性"。
- 不要为不可能发生的场景写错误处理。
- 如果写了 200 行但 50 行就够，重写。

自问：「一个高级工程师会觉得这过度复杂吗？」如果是，简化。

### 3. 手术式修改

**只动必须动的地方。只清理自己制造的混乱。**

修改现有代码时：

- 不要「顺手优化」周边代码、注释或格式。
- 不要重构没有出问题的东西。
- 沿用已有风格，即使你有不同偏好。
- 发现无关的死代码，提及它——不要直接删除。

当你的改动产生孤儿时：

- 移除**你的改动**导致的无用 import / 变量 / 函数。
- 不要移除原本就存在的死代码，除非被要求。

检验标准：每一行改动都应能直接追溯到用户的请求。

### 4. 目标驱动执行

**定义成功标准。循环直到验证通过。**

将任务转化为可验证的目标：

- 「添加校验」→「为非法输入写测试，然后让测试通过」
- 「修复 bug」→「写一个能复现 bug 的测试，然后让它通过」
- 「重构 X」→「确保重构前后测试均通过」

对于多步骤任务，先列简要计划：

[步骤] → 验证：[检查项]
[步骤] → 验证：[检查项]
[步骤] → 验证：[检查项]

明确的成功标准让你能独立循环推进；模糊的标准（「让它能跑」）会导致频繁需要澄清。

---

## 技术栈

| 用途 | 选型 |
| ------------- | ---------------------------------------------------------- |
| 语言 | Go 1.21+ |
| CLI 框架 | `github.com/spf13/cobra` |
| 麦克风采集 | `github.com/gordonklaus/portaudio`（CGo，需 libportaudio） |
| 流式 ASR | Deepgram Nova-2（WebSocket 流式，interim results + endpointing） |
| 文件 ASR | OpenAI Whisper-1（batch，用于 transcript 模式） |
| 翻译 | OpenAI GPT-4o-mini |
| TTS | OpenAI TTS-1（PCM 24kHz 流式播放） |
| Go OpenAI SDK | `github.com/sashabaranov/go-openai` |
| WebSocket | `github.com/gorilla/websocket`（Deepgram 连接 + Web UI） |
| 并发模型 | goroutine + channel 流水线 |
| 前端 | React + TypeScript + Vite + Tailwind（嵌入到 Go 二进制） |

---

## 项目结构

```
mini-tmk-agent/
├── cmd/mini-tmk-agent/
│   ├── main.go              # CLI 入口：stream, transcript 子命令
│   └── web.go               # web 子命令
├── internal/
│   ├── audio/
│   │   ├── capture.go       # PortAudio 麦克风采集 (16kHz mono)
│   │   ├── vad.go           # RMS 能量 VAD（仅 web handler 使用）
│   │   └── file.go          # WAV/MP3/M4A/PCM 文件读取 + PCMToWAV
│   ├── asr/
│   │   ├── deepgram.go      # StreamClient/StreamSession 接口 + Deepgram WebSocket 实现
│   │   ├── whisper.go       # Whisper-1 batch Client 接口（transcript 模式）
│   │   └── simplify.go      # 繁→简中文映射
│   ├── translate/
│   │   └── openai.go        # GPT-4o-mini 翻译 Client 接口
│   ├── tts/
│   │   ├── tts.go           # TTS Client 接口 + OpenAI TTS-1 实现（流式 PCM）
│   │   └── player.go        # PortAudio PCM 播放（persistent stream, IsPlaying flag）
│   ├── pipeline/
│   │   ├── stream.go        # Deepgram 流式 pipeline：interim/final → 翻译 → 显示 → TTS
│   │   ├── transcript.go    # 文件转录 pipeline
│   │   └── metrics.go       # JSONL 延迟/成本指标
│   ├── display/
│   │   └── terminal.go      # ANSI 终端：PrintInterim（覆盖刷新）+ PrintFinal（固定显示）
│   └── web/
│       ├── server.go        # HTTP server + 嵌入式静态文件
│       ├── handler.go       # WebSocket 消息处理
│       └── static/          # 前端编译产物（make web-build 生成）
├── web/                     # React + TypeScript 前端源码
├── config/config.go         # API Key 加载（OpenAI + Deepgram）
├── tests/unit/              # 单元测试（无需 API Key）
├── tests/integration/       # 集成测试（需真实 API）
├── testdata/                # hello_zh.wav 测试音频
├── Makefile
└── README.md
```

---

## Stream 模式架构

```
麦克风 → PortAudio (16kHz) → PCM 帧 → Deepgram WebSocket
                                          ↓
                                    interim result → 终端灰色覆盖显示（字在跳）
                                    final result → GPT 翻译 → 终端固定显示 [SRC]/[TGT]
                                                                    ↓
                                                              TTS goroutine（异步）
                                                                    ↓
                                                              OpenAI TTS-1 → PCM 流
                                                                    ↓
                                                              PortAudio 播放（边收边播）
```

关键设计：

- **Deepgram 服务端做 VAD + endpointing**，客户端不再需要 VAD
- **interim results** 让文字在说话过程中实时跳出（~200ms 延迟）
- **TTS 异步**：翻译完成后文字立刻显示，TTS 通过 non-blocking channel 送给独立 goroutine
- **persistent PortAudio 输出流**：播放器只开一次，所有句子复用，避免 ALSA 重复初始化
- **underrun 非致命**：ALSA 缓冲区偶尔空了就跳过继续播，不中断整句

### TTS 三档模式

1. 无 TTS（默认）：`stream --source-lang zh --target-lang en`
2. 耳机模式（全双工）：`stream --tts` — 麦克风和 TTS 同时工作
3. 扬声器模式（半双工）：`stream --tts --tts-speaker-mode` — TTS 播放时暂停麦克风输入

---

## 代码规范

- 所有 goroutine 必须通过 `context.Context` 控制生命周期，支持优雅退出
- channel 发送必须使用 non-blocking send（`select { case ch <- v: default: }`）当处于 portaudio 回调中或 TTS 队列满时
- 错误处理：API 失败（ASR/GPT/TTS）不中断 stream 主流程，记录日志后跳过
- 启动时致命错误（API Key 未设置、麦克风打开失败）明确提示解决方法
- 接口化 ASR（`Client` + `StreamClient`）、翻译（`Client`）和 TTS（`Client`），方便单元测试注入 mock

---

## 配置方式

通过环境变量（不要硬编码 API Key）：

```bash
export OPENAI_API_KEY=sk-...                           # 必须：翻译 + TTS + transcript ASR
export DEEPGRAM_API_KEY=dg-...                         # 必须（stream 模式）：流式 ASR
export OPENAI_BASE_URL=https://api.openai.com/v1       # 可选，默认值
```

CLI flag `--api-key`、`--base-url`、`--deepgram-api-key` 可覆盖环境变量。

---

## 测试规范

- **单元测试**：不需要 API Key，使用 mock interface，`go test ./...` 直接跑
- **集成测试**：需要真实 API Key，通过 `go test -tags integration ./tests/integration/...` 触发
- 集成测试用 `testdata/hello_zh.wav`（短音频，控制 API 费用）

---

## Makefile 规范

```makefile
make build              # 编译（含前端）→ bin/mini-tmk-agent
make web-build          # 仅编译前端 → internal/web/static/
make web-dev            # 前端开发服务器
make test               # 仅单元测试，无需 API Key
make test-integration   # 集成测试，需要 OPENAI_API_KEY
make lint               # golangci-lint run
make clean              # rm -rf bin/
```

---

## 支持语言

`zh`（中文）、`en`（英语）、`es`（西班牙语）、`ja`（日语）

---

## 加分项状态

- [x] TTS：OpenAI `tts-1` 模型流式播放译文音频（三档模式）
- [x] Web UI：Go HTTP server + WebSocket + React SPA
- [ ] RTC 跨端：火山引擎 RTC 或声网 RTC
- [ ] 更多语言支持

---

## 系统依赖（README 中说明）

```bash
# macOS
brew install portaudio

# Linux (Ubuntu/Debian)
sudo apt install libportaudio2 portaudio19-dev
```
