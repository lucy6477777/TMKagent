import { useEffect, useRef, useState } from 'react'
import { Square, Play, Volume2, VolumeX } from 'lucide-react'
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
  interim: string
  pipelineState: PipelineState
  sendCmd: (cmd: WSCommand) => void
  showStopToast?: boolean
  onToastDismissed?: () => void
}

export function StreamPage({ pairs, interim, pipelineState, sendCmd, showStopToast, onToastDismissed }: Props) {
  const [sourceLang, setSourceLang] = useState('zh')
  const [targetLang, setTargetLang] = useState('en')
  const [ttsEnabled, setTtsEnabled] = useState(false)
  const [toastVisible, setToastVisible] = useState(false)
  const toastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const prevPairsLen = useRef(0)
  const styleRef = useRef<HTMLStyleElement | null>(null)

  const isRunning = pipelineState === 'listening' || pipelineState === 'processing'

  // Inject CSS animations once
  useEffect(() => {
    if (styleRef.current) return
    const style = document.createElement('style')
    style.textContent = `
      @keyframes interimBreathe {
        0%, 100% { opacity: 0.45; }
        50% { opacity: 0.9; }
      }
      @keyframes dotPulse {
        0%, 80%, 100% { opacity: 0.2; transform: translateY(0); }
        40% { opacity: 1; transform: translateY(-4px); }
      }
    `
    document.head.appendChild(style)
    styleRef.current = style
    return () => { style.remove(); styleRef.current = null }
  }, [])

  // Speak new pairs via Web Speech API when TTS enabled
  useEffect(() => {
    if (!ttsEnabled || pairs.length <= prevPairsLen.current) {
      prevPairsLen.current = pairs.length
      return
    }
    const newPairs = pairs.slice(prevPairsLen.current)
    prevPairsLen.current = pairs.length
    for (const pair of newPairs) {
      const utterance = new SpeechSynthesisUtterance(pair.target)
      utterance.lang = targetLang === 'zh' ? 'zh-CN' : targetLang === 'ja' ? 'ja-JP' : targetLang === 'es' ? 'es-ES' : 'en-US'
      window.speechSynthesis.speak(utterance)
    }
  }, [pairs, ttsEnabled, targetLang])

  useEffect(() => {
    if (!ttsEnabled) window.speechSynthesis.cancel()
  }, [ttsEnabled])

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
            position: 'fixed', top: 60, right: 16,
            background: '#1F2937', color: '#FFFFFF',
            padding: '10px 16px', borderRadius: 6, fontSize: 14,
            fontFamily: "'IBM Plex Sans', sans-serif",
            zIndex: 100, display: 'flex', alignItems: 'center', gap: 10,
          }}
        >
          已停止实时翻译
          <button
            onClick={() => { setToastVisible(false); onToastDismissed?.() }}
            aria-label="关闭提示"
            style={{ background: 'none', border: 'none', color: '#9CA3AF', cursor: 'pointer', padding: 0, fontSize: 16, lineHeight: 1 }}
          >×</button>
        </div>
      )}

      {/* Control bar */}
      <div style={{
        height: 56, background: '#FFFFFF', borderBottom: '1px solid #E5E7EB',
        display: 'flex', alignItems: 'center', gap: 12, padding: '0 24px', flexShrink: 0,
      }}>
        <select aria-label="源语言" value={sourceLang} onChange={e => setSourceLang(e.target.value)} disabled={isRunning} style={selectStyle}>
          {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
        </select>

        <span style={{ color: '#9CA3AF', fontSize: 14 }}>→</span>

        <select aria-label="目标语言" value={targetLang} onChange={e => setTargetLang(e.target.value)} disabled={isRunning} style={selectStyle}>
          {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
        </select>

        <button
          onClick={handleToggle}
          disabled={pipelineState === 'processing'}
          aria-label={isRunning ? '停止实时翻译' : '开始实时翻译'}
          style={{
            height: 36, padding: '0 16px', border: 'none', borderRadius: 6,
            background: isRunning ? '#DC2626' : '#2563EB', color: '#FFFFFF',
            fontFamily: "'IBM Plex Sans', sans-serif", fontWeight: 500, fontSize: 14,
            cursor: pipelineState === 'processing' ? 'not-allowed' : 'pointer',
            display: 'flex', alignItems: 'center', gap: 6,
            transition: 'background 150ms ease',
            opacity: pipelineState === 'processing' ? 0.6 : 1,
          }}
        >
          {isRunning ? <><Square size={14} strokeWidth={1.5} /> 停止</> : <><Play size={14} strokeWidth={1.5} /> 开始</>}
        </button>

        <button
          onClick={() => setTtsEnabled(v => !v)}
          aria-label={ttsEnabled ? '关闭声音' : '开启声音'}
          aria-pressed={ttsEnabled}
          style={{
            height: 36, padding: '0 12px',
            border: `1px solid ${ttsEnabled ? '#BFDBFE' : '#D1D5DB'}`,
            borderRadius: 6,
            background: ttsEnabled ? '#EFF6FF' : '#FFFFFF',
            color: ttsEnabled ? '#2563EB' : '#9CA3AF',
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontSize: 13, fontWeight: 500, cursor: 'pointer',
            display: 'flex', alignItems: 'center', gap: 5,
            transition: 'all 150ms ease',
          }}
        >
          {ttsEnabled ? <><Volume2 size={14} strokeWidth={1.5} /> 声音</> : <><VolumeX size={14} strokeWidth={1.5} /> 声音</>}
        </button>
      </div>

      {/* ── Style C: Live interim zone ── */}
      <div style={{
        flexShrink: 0,
        minHeight: 140,
        background: '#FFFFFF',
        borderBottom: '1px solid #E5E7EB',
        padding: '24px 32px',
        display: 'flex',
        alignItems: 'center',
      }}>
        {!isRunning && pairs.length === 0 ? (
          <p style={{
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontSize: 15, color: '#CBD5E1',
          }}>
            点击「开始」后将在此实时显示识别文字
          </p>
        ) : interim ? (
          <p style={{
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontSize: 28, fontWeight: 600,
            color: '#94A3B8',
            lineHeight: 1.35,
            animation: 'interimBreathe 1.8s ease-in-out infinite',
            margin: 0,
          }}>
            {interim}
            <span style={{ marginLeft: 4, display: 'inline-block', opacity: 0.5 }}>…</span>
          </p>
        ) : isRunning ? (
          /* Waiting dots */
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            {[0, 0.18, 0.36].map((delay, i) => (
              <span key={i} style={{
                display: 'inline-block', width: 8, height: 8,
                borderRadius: '50%', background: '#CBD5E1',
                animation: `dotPulse 1.2s ease-in-out ${delay}s infinite`,
              }} />
            ))}
          </div>
        ) : (
          <p style={{
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontSize: 15, color: '#CBD5E1',
          }}>
            点击「开始」后将在此实时显示识别文字
          </p>
        )}
      </div>

      {/* ── History feed ── */}
      <div style={{ flex: 1, overflow: 'hidden', paddingTop: 8 }}>
        <SubtitleFeed
          pairs={pairs}
          emptyMessage={isRunning ? '翻译结果将在此显示' : ''}
        />
      </div>
    </div>
  )
}
