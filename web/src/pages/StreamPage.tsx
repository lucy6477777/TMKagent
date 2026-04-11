import { useEffect, useRef, useState } from 'react'
import { Square, Play } from 'lucide-react'
import type { PipelineState, SubtitlePair, WSCommand } from '../types/ws'
import { SubtitleFeed } from '../components/SubtitleFeed'

const LANGUAGES = [
  { value: 'zh', label: '中文' },
  { value: 'en', label: 'English' },
  { value: 'es', label: 'Español' },
  { value: 'ja', label: '日本語' },
]

interface Props {
  pairs: SubtitlePair[]
  pipelineState: PipelineState
  sendCmd: (cmd: WSCommand) => void
  /** True when we just switched FROM this page while streaming — shows toast */
  showStopToast?: boolean
  onToastDismissed?: () => void
}

export function StreamPage({ pairs, pipelineState, sendCmd, showStopToast, onToastDismissed }: Props) {
  const [sourceLang, setSourceLang] = useState('zh')
  const [targetLang, setTargetLang] = useState('en')
  const [toastVisible, setToastVisible] = useState(false)
  const toastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const isRunning = pipelineState === 'listening' || pipelineState === 'processing'

  // Show toast when prop arrives
  useEffect(() => {
    if (showStopToast) {
      setToastVisible(true)
      toastTimerRef.current = setTimeout(() => {
        setToastVisible(false)
        onToastDismissed?.()
      }, 4000)
    }
    return () => {
      if (toastTimerRef.current) clearTimeout(toastTimerRef.current)
    }
  }, [showStopToast, onToastDismissed])

  const handleToggle = () => {
    if (isRunning) {
      sendCmd({ type: 'cmd', action: 'stop' })
    } else {
      sendCmd({ type: 'cmd', action: 'start_stream', sourceLang, targetLang })
    }
  }

  const selectStyle: React.CSSProperties = {
    height: 36,
    padding: '0 8px',
    border: '1px solid #D1D5DB',
    borderRadius: 6,
    background: '#FFFFFF',
    color: '#374151',
    fontFamily: "'IBM Plex Sans', sans-serif",
    fontSize: 14,
    cursor: 'pointer',
    appearance: 'auto',
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* Toast */}
      {toastVisible && (
        <div
          role="status"
          aria-live="polite"
          style={{
            position: 'fixed',
            top: 60,
            right: 16,
            background: '#1F2937',
            color: '#FFFFFF',
            padding: '10px 16px',
            borderRadius: 6,
            fontSize: 14,
            fontFamily: "'IBM Plex Sans', sans-serif",
            zIndex: 100,
            display: 'flex',
            alignItems: 'center',
            gap: 10,
          }}
        >
          已停止实时翻译
          <button
            onClick={() => { setToastVisible(false); onToastDismissed?.() }}
            aria-label="关闭提示"
            style={{ background: 'none', border: 'none', color: '#9CA3AF', cursor: 'pointer', padding: 0, fontSize: 16, lineHeight: 1 }}
          >
            ×
          </button>
        </div>
      )}

      {/* Control bar */}
      <div
        style={{
          height: 56,
          background: '#FFFFFF',
          borderBottom: '1px solid #E5E7EB',
          display: 'flex',
          alignItems: 'center',
          gap: 16,
          padding: '0 24px',
          flexShrink: 0,
        }}
      >
        <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14, color: '#374151', fontFamily: "'IBM Plex Sans', sans-serif" }}>
          <span className="sr-only">源语言</span>
          <select
            aria-label="源语言"
            value={sourceLang}
            onChange={e => setSourceLang(e.target.value)}
            disabled={isRunning}
            style={selectStyle}
          >
            {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
          </select>
        </label>

        <span style={{ color: '#9CA3AF', fontSize: 14 }}>→</span>

        <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14, color: '#374151', fontFamily: "'IBM Plex Sans', sans-serif" }}>
          <span className="sr-only">目标语言</span>
          <select
            aria-label="目标语言"
            value={targetLang}
            onChange={e => setTargetLang(e.target.value)}
            disabled={isRunning}
            style={selectStyle}
          >
            {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
          </select>
        </label>

        <button
          onClick={handleToggle}
          disabled={pipelineState === 'processing'}
          aria-label={isRunning ? '停止实时翻译' : '开始实时翻译'}
          style={{
            height: 36,
            padding: '0 16px',
            border: 'none',
            borderRadius: 6,
            background: isRunning ? '#DC2626' : '#1E3A5F',
            color: '#FFFFFF',
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontWeight: 500,
            fontSize: 14,
            cursor: pipelineState === 'processing' ? 'not-allowed' : 'pointer',
            display: 'flex',
            alignItems: 'center',
            gap: 6,
            transition: 'background 150ms ease',
            opacity: pipelineState === 'processing' ? 0.6 : 1,
          }}
        >
          {isRunning
            ? <><Square size={14} strokeWidth={1.5} /> 停止</>
            : <><Play size={14} strokeWidth={1.5} /> 开始</>}
        </button>
      </div>

      {/* Subtitle feed — fills remaining height */}
      <div style={{ flex: 1, overflow: 'hidden', paddingTop: 8 }}>
        <SubtitleFeed
          pairs={pairs}
          emptyMessage="点击「开始」后字幕将在此显示"
        />
      </div>
    </div>
  )
}
