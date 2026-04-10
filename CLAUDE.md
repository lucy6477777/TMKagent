# CLAUDE.md — mini-tmk-agent

## 项目概述

Go CLI 同声传译 Agent，面试作业项目。支持两种模式：

- `stream`：实时麦克风采集 → Whisper ASR → GPT 翻译 → 终端双语显示
- `transcript`：音频文件（WAV/MP3/PCM）→ Whisper ASR → 纯文本输出

目标平台：Linux 演示环境：开发者自己的机器。

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

## 技术栈（已确定，不要更换）

| 用途          | 选型                                                       |
| ------------- | ---------------------------------------------------------- |
| 语言          | Go 1.21+                                                   |
| CLI 框架      | `github.com/spf13/cobra`                                   |
| 麦克风采集    | `github.com/gordonklaus/portaudio`（CGo，需 libportaudio） |
| LLM 平台      | OpenAI（Whisper-1 做 ASR，GPT-4o mini 做翻译）             |
| Go OpenAI SDK | `github.com/sashabaranov/go-openai`                        |
| 并发模型      | goroutine + channel 流水线                                 |
| VAD           | 纯 Go 能量门限（RMS），无外部依赖                          |

**禁止**在未经讨论的情况下切换上述任何选型（例如：不要换成 malgo、不要换成 ffmpeg 子进程采集、不要换 LLM 平台）。

---

## 项目结构

mini-tmk-agent/
├── cmd/mini-tmk-agent/main.go # CLI 入口
├── internal/
│ ├── audio/ # capture.go, vad.go, file.go
│ ├── asr/ # whisper.go
│ ├── translate/ # openai.go
│ ├── pipeline/ # stream.go, transcript.go
│ └── display/ # terminal.go
├── config/config.go
├── tests/unit/
├── tests/integration/
├── testdata/ # hello_zh.wav 等测试音频
├── Makefile
└── README.md

---

## 代码规范

- 所有 goroutine 必须通过 `context.Context` 控制生命周期，支持优雅退出
- channel 发送必须使用 non-blocking send（`select { case ch <- v: default: }`）当处于 portaudio 回调中
- 错误处理：API 失败（Whisper/GPT）不中断 stream 主流程，记录日志后跳过当前块
- 启动时致命错误（API Key 未设置、麦克风打开失败）用 `log.Fatal` 明确提示解决方法
- 接口化 ASR 和翻译客户端，方便单元测试注入 mock

---

## 配置方式

通过环境变量（不要硬编码 API Key）：

```bash
export OPENAI_API_KEY=sk-...
export OPENAI_BASE_URL=https://api.openai.com/v1   # 可选，默认值
```

CLI flag `--api-key` 和 `--base-url` 可覆盖环境变量。

---

## 测试规范

- **单元测试**：不需要 API Key，使用 mock interface，`go test ./...` 直接跑
- **集成测试**：需要真实 API Key，通过 `INTEGRATION=1 go test ./tests/integration/...` 触发
- 集成测试用 `testdata/hello_zh.wav`（短音频，控制 API 费用）

---

## Makefile 规范

```makefile
make build              # go build -o bin/mini-tmk-agent
make test               # 仅单元测试，无需 API Key
make test-integration   # 集成测试，需要 OPENAI_API_KEY
make lint               # golangci-lint run
make clean              # rm -rf bin/
```

---

## 支持语言

`zh`（中文）、`en`（英语）、`es`（西班牙语）、`ja`（日语）

---

## 加分项（核心功能完成后再做，不要提前引入）

1. TTS：OpenAI `tts-1` 模型播放译文音频
2. RTC 跨端：火山引擎 RTC 或声网 RTC
3. Web UI：Go HTTP server + SSE 推送
4. 更多语言支持

---

## 系统依赖（README 中说明）

```bash
# macOS
brew install portaudio

# Linux (Ubuntu/Debian)
sudo apt install libportaudio2 libportaudio-dev
```
