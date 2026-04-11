import { useEffect, useRef } from 'react'
import { AudioLines, FileText } from 'lucide-react'
import type { Page } from '../components/Sidebar'

interface Props {
  onNavigate: (page: Page) => void
}

const FEATURE_CARDS = [
  {
    id: 'stream' as Page,
    Icon: AudioLines,
    title: '实时同声传译',
    desc: '打开麦克风，说话即翻译。Deepgram 流式识别，字幕在说话过程中实时跳出，延迟低至 200ms。',
    cta: '进入实时翻译',
    accentColor: '#2563EB',
    bgColor: '#EFF6FF',
  },
  {
    id: 'transcript' as Page,
    Icon: FileText,
    title: '文件批量转录',
    desc: '拖拽上传 WAV / MP3 / M4A，Whisper 自动识别并翻译，输出双语对照结果，支持一键复制。',
    cta: '进入文件转录',
    accentColor: '#7C3AED',
    bgColor: '#F5F3FF',
  },
]

const TECH_BADGES = [
  { label: 'Deepgram ASR', color: '#2563EB' },
  { label: 'GPT-4o mini', color: '#059669' },
  { label: 'OpenAI TTS', color: '#D97706' },
  { label: 'LiveKit RTC', color: '#7C3AED' },
  { label: 'Go WebSocket', color: '#64748B' },
]

// Animated waveform bars using CSS animation injected once
const WAVE_DELAYS = [0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.4, 0.3, 0.2, 0.1, 0.05, 0.15, 0.25, 0.35, 0.45]
const WAVE_HEIGHTS = [20, 40, 65, 85, 100, 90, 70, 50, 80, 60, 35, 55, 75, 45, 25]

export function HomePage({ onNavigate }: Props) {
  const styleRef = useRef<HTMLStyleElement | null>(null)

  useEffect(() => {
    if (styleRef.current) return
    const style = document.createElement('style')
    style.textContent = `
      @keyframes waveBar {
        0%, 100% { transform: scaleY(0.3); opacity: 0.25; }
        50% { transform: scaleY(1); opacity: 0.6; }
      }
      @keyframes orbFloat1 {
        0%, 100% { transform: translate(0, 0) scale(1); }
        50% { transform: translate(-18px, 14px) scale(1.06); }
      }
      @keyframes orbFloat2 {
        0%, 100% { transform: translate(0, 0); }
        50% { transform: translate(12px, -20px); }
      }
      @keyframes orbFloat3 {
        0%, 100% { transform: translate(0, 0); }
        50% { transform: translate(-10px, 12px); }
      }
      @keyframes cursorBlink {
        0%, 100% { opacity: 1; }
        50% { opacity: 0; }
      }
      @keyframes pulseDot {
        0%, 100% { opacity: 1; transform: scale(1); }
        50% { opacity: 0.4; transform: scale(0.65); }
      }
      @keyframes cardHover {
        to { transform: translateY(-3px); }
      }
    `
    document.head.appendChild(style)
    styleRef.current = style
    return () => { style.remove(); styleRef.current = null }
  }, [])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', minHeight: '100%', background: '#F8FAFC' }}>

      {/* ── Hero ── */}
      <div style={{ position: 'relative', overflow: 'hidden', padding: '48px 48px 36px', minHeight: 300 }}>

        {/* Dot grid background */}
        <div style={{
          position: 'absolute', inset: 0, pointerEvents: 'none',
          backgroundImage: 'radial-gradient(circle, #CBD5E1 1px, transparent 1px)',
          backgroundSize: '22px 22px',
          opacity: 0.35,
        }} />

        {/* Animated orbs */}
        <div style={{
          position: 'absolute', top: -60, right: 80, width: 280, height: 280,
          borderRadius: '50%', filter: 'blur(50px)', pointerEvents: 'none',
          background: 'radial-gradient(circle, rgba(37,99,235,0.14) 0%, transparent 70%)',
          animation: 'orbFloat1 7s ease-in-out infinite',
        }} />
        <div style={{
          position: 'absolute', bottom: -30, right: 260, width: 200, height: 200,
          borderRadius: '50%', filter: 'blur(40px)', pointerEvents: 'none',
          background: 'radial-gradient(circle, rgba(16,185,129,0.09) 0%, transparent 70%)',
          animation: 'orbFloat2 9s ease-in-out infinite',
        }} />
        <div style={{
          position: 'absolute', top: 40, left: 240, width: 160, height: 160,
          borderRadius: '50%', filter: 'blur(36px)', pointerEvents: 'none',
          background: 'radial-gradient(circle, rgba(59,130,246,0.1) 0%, transparent 70%)',
          animation: 'orbFloat3 8s ease-in-out infinite',
        }} />

        {/* Large waveform — right side */}
        <div style={{
          position: 'absolute', right: 48, top: '50%', transform: 'translateY(-50%)',
          display: 'flex', alignItems: 'center', gap: 4, height: 120, zIndex: 1,
        }}>
          {WAVE_BARS_DATA.map((h, i) => (
            <div
              key={i}
              style={{
                width: 4, borderRadius: 99,
                background: '#2563EB',
                height: `${h}%`,
                animation: `waveBar 1.4s ease-in-out ${WAVE_DELAYS[i % WAVE_DELAYS.length]}s infinite`,
                transformOrigin: 'bottom',
              }}
            />
          ))}
        </div>

        {/* Hero content */}
        <div style={{ position: 'relative', zIndex: 2, maxWidth: 500 }}>
          {/* Eyebrow */}
          <div style={{
            display: 'inline-flex', alignItems: 'center', gap: 6,
            fontSize: 11, fontWeight: 700, color: '#2563EB',
            background: '#EFF6FF', border: '1px solid #BFDBFE',
            padding: '4px 12px', borderRadius: 99, marginBottom: 18,
            fontFamily: "'IBM Plex Sans', sans-serif",
            letterSpacing: '0.03em',
          }}>
            <span style={{
              width: 6, height: 6, borderRadius: '50%', background: '#10B981',
              animation: 'pulseDot 1.8s ease-in-out infinite', display: 'inline-block',
            }} />
            Deepgram + GPT-4o mini
          </div>

          {/* Title */}
          <h1 style={{
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontSize: 40, fontWeight: 800, lineHeight: 1.1,
            letterSpacing: '-0.04em', color: '#0F172A', marginBottom: 12,
          }}>
            你开口<br />
            <span style={{ color: '#2563EB' }}>世界懂你</span>
          </h1>

          <p style={{
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontSize: 13, color: '#475569', lineHeight: 1.65,
            marginBottom: 28,
          }}>
            实时同声传译 · 文件批量转录<br />
            基于 Deepgram 流式 ASR，延迟低至 200ms
          </p>

          {/* Live demo strip */}
          <div style={{
            display: 'inline-flex', flexDirection: 'column', gap: 5,
            background: 'rgba(255,255,255,0.9)', backdropFilter: 'blur(8px)',
            border: '1px solid #E2E8F0', borderRadius: 10,
            padding: '10px 16px',
            boxShadow: '0 2px 8px rgba(37,99,235,0.06)',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 2 }}>
              <span style={{
                width: 5, height: 5, borderRadius: '50%', background: '#10B981',
                animation: 'pulseDot 1.5s ease-in-out infinite', display: 'inline-block',
              }} />
              <span style={{
                fontSize: 9, fontWeight: 700, textTransform: 'uppercase',
                letterSpacing: '0.1em', color: '#94A3B8',
                fontFamily: "'IBM Plex Sans', sans-serif",
              }}>正在识别</span>
            </div>
            <div style={{
              fontSize: 13, fontWeight: 600, color: '#94A3B8', fontStyle: 'italic',
              fontFamily: "'IBM Plex Sans', sans-serif",
              display: 'flex', alignItems: 'center',
            }}>
              今天天气真的很不错
              <span style={{
                display: 'inline-block', width: 1.5, height: 13,
                background: '#94A3B8', borderRadius: 1, marginLeft: 2,
                animation: 'cursorBlink 0.9s ease-in-out infinite',
              }} />
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 3 }}>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 7 }}>
                <span style={{
                  fontSize: 8, fontWeight: 800, padding: '1px 5px', borderRadius: 3,
                  background: '#FEF3C7', color: '#92400E',
                  fontFamily: "'IBM Plex Sans', sans-serif",
                }}>ZH</span>
                <span style={{ fontSize: 11, color: '#0F172A', fontFamily: "'IBM Plex Sans', sans-serif" }}>
                  欢迎使用 mini-tmk-agent
                </span>
              </div>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 7 }}>
                <span style={{
                  fontSize: 8, fontWeight: 800, padding: '1px 5px', borderRadius: 3,
                  background: '#DBEAFE', color: '#1E40AF',
                  fontFamily: "'IBM Plex Sans', sans-serif",
                }}>EN</span>
                <span style={{ fontSize: 11, color: '#0F172A', fontFamily: "'IBM Plex Sans', sans-serif" }}>
                  Welcome to mini-tmk-agent
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* ── Feature cards ── */}
      <div style={{ padding: '0 48px 32px', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        {FEATURE_CARDS.map(({ id, Icon, title, desc, cta, accentColor, bgColor }) => (
          <button
            key={id}
            onClick={() => onNavigate(id)}
            style={{
              background: '#FFFFFF',
              border: `1.5px solid #E2E8F0`,
              borderLeft: `3px solid ${accentColor}`,
              borderRadius: 12,
              padding: '20px 20px 20px 18px',
              cursor: 'pointer',
              textAlign: 'left',
              transition: 'border-color 0.2s, box-shadow 0.2s, transform 0.2s',
              fontFamily: "'IBM Plex Sans', sans-serif",
            }}
            onMouseEnter={e => {
              const el = e.currentTarget
              el.style.borderColor = accentColor
              el.style.boxShadow = `0 8px 24px rgba(0,0,0,0.08)`
              el.style.transform = 'translateY(-2px)'
            }}
            onMouseLeave={e => {
              const el = e.currentTarget
              el.style.borderColor = '#E2E8F0'
              el.style.borderLeftColor = accentColor
              el.style.boxShadow = 'none'
              el.style.transform = 'translateY(0)'
            }}
          >
            <div style={{
              width: 36, height: 36, borderRadius: 9,
              background: bgColor,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              marginBottom: 12,
            }}>
              <Icon size={18} strokeWidth={1.5} color={accentColor} />
            </div>
            <div style={{ fontSize: 14, fontWeight: 700, color: '#0F172A', marginBottom: 6 }}>
              {title}
            </div>
            <div style={{ fontSize: 12, color: '#64748B', lineHeight: 1.6, marginBottom: 14 }}>
              {desc}
            </div>
            <div style={{ fontSize: 12, fontWeight: 700, color: accentColor, display: 'flex', alignItems: 'center', gap: 4 }}>
              {cta} <span>→</span>
            </div>
          </button>
        ))}
      </div>

      {/* ── Tech strip ── */}
      <div style={{
        padding: '12px 48px',
        borderTop: '1px solid #E2E8F0',
        background: '#FFFFFF',
        display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap',
        marginTop: 'auto',
      }}>
        <span style={{
          fontSize: 9, fontWeight: 700, textTransform: 'uppercase',
          letterSpacing: '0.1em', color: '#94A3B8', marginRight: 4,
          fontFamily: "'IBM Plex Sans', sans-serif",
        }}>技术栈</span>
        {TECH_BADGES.map(({ label, color }) => (
          <div key={label} style={{
            display: 'flex', alignItems: 'center', gap: 5,
            fontSize: 10, fontWeight: 600, color: '#475569',
            background: '#F8FAFC', border: '1px solid #E2E8F0',
            padding: '3px 9px', borderRadius: 5,
            fontFamily: "'IBM Plex Sans', sans-serif",
          }}>
            <span style={{ width: 5, height: 5, borderRadius: '50%', background: color, display: 'inline-block' }} />
            {label}
          </div>
        ))}
      </div>

    </div>
  )
}

// 15 bars with varying heights for the waveform
const WAVE_BARS_DATA = WAVE_HEIGHTS
