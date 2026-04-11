# Web UI 设计文档

**日期：** 2026-04-11  
**项目：** mini-tmk-agent  
**状态：** 已批准，待实现

---

## 概述

为 mini-tmk-agent 添加 Web UI，通过 `mini-tmk-agent web` 子命令启动。前端用 React + Vite，后端用 Go embed 打包产物进二进制，WebSocket 实时推送字幕，零外部部署依赖。

---

## 架构

### 整体结构

```
mini-tmk-agent/
├── cmd/mini-tmk-agent/main.go         # 新增 web 子命令
├── internal/
│   └── web/
│       ├── server.go                  # HTTP + WebSocket server
│       ├── handler.go                 # WS 消息协议 & pipeline 桥接
│       └── static/                   # embed.FS 目标（web/dist 复制至此）
├── web/                               # React 项目根
│   ├── src/
│   │   ├── App.tsx
│   │   ├── pages/
│   │   │   ├── StreamPage.tsx         # 实时字幕视图
│   │   │   └── TranscriptPage.tsx     # 文件上传 + 转录历史
│   │   └── components/
│   │       ├── Sidebar.tsx
│   │       └── SubtitleFeed.tsx
│   ├── package.json
│   └── vite.config.ts
└── Makefile                           # 新增 web-build target
```

### 数据流

```
[麦克风] → RunStream goroutines → WS push → React SubtitleFeed
[音频文件] → HTTP POST /upload → RunTranscript → WS push → React TranscriptPage
```

现有 `RunStream` / `RunTranscript` pipeline 代码零改动。Go web 层把 `display.Pair` 序列化为 JSON 推送给前端。

---

## WebSocket 消息协议

### 服务端 → 前端

```jsonc
// 字幕推送（stream 和 transcript 模式统一）
{ "type": "pair", "source": "你好世界", "target": "Hello world", "ts": 1712800000 }

// 状态通知
{ "type": "status", "state": "listening" }   // listening | processing | idle | error

// 转录进度（大文件分块时）
{ "type": "progress", "current": 3, "total": 8 }

// 错误
{ "type": "error", "msg": "ASR failed: ..." }
```

### 前端 → 服务端

```jsonc
// 启动实时翻译
{ "type": "cmd", "action": "start_stream", "sourceLang": "zh", "targetLang": "en" }

// 停止当前 pipeline
{ "type": "cmd", "action": "stop" }

// 触发文件转录（文件通过 POST /upload 上传，WS 推结果）
{ "type": "cmd", "action": "transcript", "sourceLang": "zh", "targetLang": "en" }
```

### 并发约束

每次只允许一个 pipeline 运行（stream 或 transcript）。收到新指令时，cancel 当前 pipeline context，再启动新的。UI 切换时必须显示明确提示（如「已停止实时翻译，开始转录文件」），不允许静默切换。

---

## Go 服务端

### HTTP 路由

```
GET  /          → serve embed React dist（SPA fallback → index.html）
GET  /assets/*  → serve embed 静态资源
GET  /ws        → WebSocket 升级
POST /upload    → 接收音频文件（multipart）→ 触发 transcript pipeline
```

### embed 集成

```go
//go:embed static
var staticFiles embed.FS
```

`web/dist` 构建产物复制到 `internal/web/static/`，随 `go build` 打进二进制。

### 新增依赖

- `github.com/gorilla/websocket` — 唯一新增第三方依赖

### CLI 子命令

```bash
mini-tmk-agent web --port 8080 --api-key sk-...
# 启动后打印：Web UI running at http://localhost:8080
```

`--port` 默认 8080，`--api-key` / `--base-url` 与其他子命令保持一致，优先读环境变量。

---

## 前端设计

### 视觉风格

浅色系，冷白极简 + 石墨蓝白组合。

| 元素 | 颜色 |
|------|------|
| 页面背景 | `#F8F9FA` |
| Sidebar 背景 | `#FFFFFF` + 右侧 1px `#E5E7EB` 分割线 |
| 选中导航项 | `#1E3A5F` 背景，白字 |
| SRC 标签胶囊 | `#64748B`（蓝灰） |
| TGT 标签胶囊 | `#1E3A5F`（深蓝） |
| 字幕正文 | `#111827` |
| 状态"监听中" | `#10B981` 绿点 + 脉冲动画 |
| 主按钮 | `#1E3A5F`，hover `#2D5A8E` |

### 布局

```
┌──────────────────────────────────────────────────────┐
│  ● mini-tmk-agent              [状态: 监听中...]      │  ← Topbar
├──────────┬───────────────────────────────────────────┤
│          │                                           │
│  Stream  │  [SRC] 你好，这是一段实时翻译的文字         │
│          │  [TGT] Hello, this is real-time text      │
│  ──────  │                                           │
│          │  [SRC] 请问今天天气怎么样？                 │
│ Transcript  [TGT] What's the weather like today?    │
│          │                                           │
│  ──────  │                                           │
│  设置    │                                           │
└──────────┴───────────────────────────────────────────┘
```

### Stream 页（StreamPage.tsx）

- 源语言 / 目标语言下拉选择
- 开始 / 停止按钮
- `SubtitleFeed` 组件：新字幕从底部插入向上滚动，最新一条高亮显示，旧条目渐淡

### Transcript 页（TranscriptPage.tsx）

- 拖拽上传区（接受 WAV / MP3）
- 上传后显示进度条（对应 `progress` 消息）
- 转录结果列表：双语对照展示（pair 类型），支持一键复制

### Transcript 翻译策略

Transcript 模式同样调用翻译，输出 `pair`（双语对照），与 Stream 模式统一，UI 复用同一展示组件。

---

## Makefile

```makefile
make web-build   # cd web && npm run build，复制 dist/ → internal/web/static/
make build       # 先跑 web-build，再 go build
```

---

## 技术栈

| 层 | 选型 |
|----|------|
| 前端框架 | React 18 + TypeScript |
| 构建工具 | Vite |
| 后端 WebSocket | gorilla/websocket |
| 静态文件 | Go embed.FS |
| 样式 | Tailwind CSS（与视觉规范配色对应） |

---

## 不在本期范围内

- 多用户 / 多连接并发
- 用户认证
- 历史记录持久化（刷新后消失）
- TTS 播放（加分项，核心功能完成后再做）
