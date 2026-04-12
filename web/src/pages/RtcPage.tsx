import { useEffect, useRef, useState } from 'react'
import { Square, Play, Volume2, VolumeX, Radio, User } from 'lucide-react'
import QRCode from 'qrcode'
import type { ConnectionStatus } from '../hooks/useWebSocket'
import { useViewport } from '../hooks/useViewport'
import type { PipelineState, SubtitlePair, WSCommand } from '../types/ws'
import { SubtitleFeed } from '../components/SubtitleFeed'

const LANGUAGES = [
  { value: 'zh', label: '中文' },
  { value: 'en', label: 'English' },
  { value: 'es', label: 'Español' },
  { value: 'ja', label: '日本語' },
]

interface Props {
  wsStatus: ConnectionStatus
  pairs: SubtitlePair[]
  interim: string
  pipelineState: PipelineState
  sendCmd: (cmd: WSCommand) => void
  clearPairs: () => void
}

interface ServerInfo {
  ip: string
  port: number
  baseURL?: string
}

type RtcTab = 'speaker' | 'listener'

function getSpeechLang(lang: string): string {
  switch (lang) {
    case 'zh':
      return 'zh-CN'
    case 'ja':
      return 'ja-JP'
    case 'es':
      return 'es-ES'
    default:
      return 'en-US'
  }
}

function getRtcQueryConfig() {
  const params = new URLSearchParams(window.location.search)
  const tab = params.get('tab')
  const room = params.get('room')
  const sourceLang = params.get('sourceLang')
  const targetLang = params.get('targetLang')
  const autoJoin = params.get('autojoin') === '1' || params.get('autoJoin') === '1'
  const resolvedTab: RtcTab = tab === 'listener' ? 'listener' : 'speaker'

  return {
    tab: resolvedTab,
    room: room?.trim() || 'demo-room',
    sourceLang: sourceLang || 'zh',
    targetLang: targetLang || 'en',
    autoJoin,
  }
}

export function RtcPage({ wsStatus, pairs, interim, pipelineState, sendCmd, clearPairs }: Props) {
  const { isMobile, isNarrow } = useViewport()
  const initialConfig = getRtcQueryConfig()
  const [tab, setTab] = useState<RtcTab>(initialConfig.tab)
  const [room, setRoom] = useState(initialConfig.room)
  const [sourceLang, setSourceLang] = useState(initialConfig.sourceLang)
  const [targetLang, setTargetLang] = useState(initialConfig.targetLang)
  const [ttsEnabled, setTtsEnabled] = useState(false)
  const [ttsError, setTtsError] = useState<string | null>(null)
  const [pendingStart, setPendingStart] = useState(false)
  const [serverInfo, setServerInfo] = useState<ServerInfo | null>(null)
  const [qrDataUrl, setQrDataUrl] = useState<string | null>(null)
  const prevPairsLen = useRef(0)
  const styleRef = useRef<HTMLStyleElement | null>(null)
  const autoJoinAttemptedRef = useRef(false)
  const ttsUnlockedRef = useRef(false)

  const isRunning = pipelineState === 'listening' || pipelineState === 'processing'
  const isProcessing = pipelineState === 'processing'
  const isBusy = pendingStart || isProcessing

  useEffect(() => {
    setPendingStart(false)
  }, [pipelineState])

  // Inject CSS animations once
  useEffect(() => {
    if (styleRef.current) return
    const style = document.createElement('style')
    style.textContent = `
      @keyframes rtcInterimBreathe {
        0%, 100% { opacity: 0.45; }
        50% { opacity: 0.9; }
      }
      @keyframes rtcDotPulse {
        0%, 80%, 100% { opacity: 0.2; transform: translateY(0); }
        40% { opacity: 1; transform: translateY(-4px); }
      }
    `
    document.head.appendChild(style)
    styleRef.current = style
    return () => { style.remove(); styleRef.current = null }
  }, [])

  // Fetch server info for QR code
  useEffect(() => {
    fetch('/api/info')
      .then(r => r.json())
      .then((info: ServerInfo) => setServerInfo(info))
      .catch(() => {/* dev mode — /api/info not available */})
  }, [])

  const listenerBaseURL = serverInfo?.baseURL || (serverInfo ? `http://${serverInfo.ip}:${serverInfo.port}` : null)
  const listenerUrl = listenerBaseURL
    ? `${listenerBaseURL}/?page=rtc&tab=listener&room=${encodeURIComponent(room.trim() || 'demo-room')}&sourceLang=${encodeURIComponent(sourceLang)}&targetLang=${encodeURIComponent(targetLang)}&autojoin=1`
    : null

  // Generate QR code when listener URL changes
  useEffect(() => {
    if (!listenerUrl) return
    QRCode.toDataURL(listenerUrl, { width: 140, margin: 1, color: { dark: '#0F172A', light: '#FFFFFF' } })
      .then(setQrDataUrl)
      .catch(() => {})
  }, [listenerUrl])

  // Speak new pairs via Web Speech API when TTS enabled
  useEffect(() => {
    if (!ttsEnabled || pairs.length <= prevPairsLen.current) {
      prevPairsLen.current = pairs.length
      return
    }
    if (!('speechSynthesis' in window)) {
      setTtsError('当前浏览器不支持语音播放')
      prevPairsLen.current = pairs.length
      return
    }

    const newPairs = pairs.slice(prevPairsLen.current)
    prevPairsLen.current = pairs.length
    const synth = window.speechSynthesis
    synth.resume()

    for (const pair of newPairs) {
      const utterance = new SpeechSynthesisUtterance(pair.target)
      utterance.lang = getSpeechLang(targetLang)

      const voices = synth.getVoices()
      const exactVoice = voices.find(voice => voice.lang.toLowerCase() === utterance.lang.toLowerCase())
      const prefix = utterance.lang.toLowerCase().split('-')[0]
      const fallbackVoice = voices.find(voice => voice.lang.toLowerCase().startsWith(prefix))
      if (exactVoice || fallbackVoice) {
        utterance.voice = exactVoice || fallbackVoice || null
      }

      utterance.onerror = () => {
        setTtsError('手机浏览器拦截了语音播放，请再点一次「声音」')
        ttsUnlockedRef.current = false
      }
      utterance.onstart = () => setTtsError(null)
      synth.speak(utterance)
    }
  }, [pairs, ttsEnabled, targetLang])

  useEffect(() => {
    if (!('speechSynthesis' in window)) return
    if (!ttsEnabled) {
      window.speechSynthesis.cancel()
      ttsUnlockedRef.current = false
    }
  }, [ttsEnabled])

  useEffect(() => {
    if (!('speechSynthesis' in window)) return
    const synth = window.speechSynthesis
    const loadVoices = () => synth.getVoices()
    loadVoices()
    synth.addEventListener?.('voiceschanged', loadVoices)
    return () => synth.removeEventListener?.('voiceschanged', loadVoices)
  }, [])

  useEffect(() => {
    if (!initialConfig.autoJoin || tab !== 'listener' || wsStatus !== 'connected' || isRunning || pendingStart) return
    const nextRoom = room.trim()
    if (!nextRoom || autoJoinAttemptedRef.current) return

    autoJoinAttemptedRef.current = true
    setPendingStart(true)
    clearPairs()
    sendCmd({ type: 'cmd', action: 'rtc_join', room: nextRoom, role: 'listener' })
  }, [initialConfig.autoJoin, tab, wsStatus, isRunning, pendingStart, room, clearPairs, sendCmd])

  const handleSpeakerToggle = () => {
    const nextRoom = room.trim()
    if (isRunning) {
      sendCmd({ type: 'cmd', action: 'rtc_stop' })
    } else {
      if (!nextRoom) return
      setPendingStart(true)
      clearPairs()
      sendCmd({ type: 'cmd', action: 'rtc_speaker_start', room: nextRoom, sourceLang, targetLang })
    }
  }

  const handleListenerToggle = () => {
    const nextRoom = room.trim()
    if (isRunning) {
      sendCmd({ type: 'cmd', action: 'rtc_stop' })
    } else {
      if (!nextRoom) return
      autoJoinAttemptedRef.current = true
      setPendingStart(true)
      clearPairs()
      sendCmd({ type: 'cmd', action: 'rtc_join', room: nextRoom, role: 'listener' })
    }
  }

  const handleTTSToggle = () => {
    if (ttsEnabled) {
      setTtsEnabled(false)
      setTtsError(null)
      return
    }

    if (!('speechSynthesis' in window)) {
      setTtsError('当前浏览器不支持语音播放')
      return
    }

    const synth = window.speechSynthesis
    const unlockUtterance = new SpeechSynthesisUtterance('.')
    unlockUtterance.volume = 0
    unlockUtterance.lang = getSpeechLang(targetLang)
    unlockUtterance.onstart = () => {
      ttsUnlockedRef.current = true
      setTtsError(null)
    }
    unlockUtterance.onerror = () => {
      ttsUnlockedRef.current = false
      setTtsError('手机浏览器没有解锁语音播放，请再点一次「声音」')
    }

    synth.cancel()
    synth.resume()
    synth.speak(unlockUtterance)
    window.setTimeout(() => synth.resume(), 0)
    setTtsEnabled(true)
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
    flex: isMobile ? 1 : undefined,
    minWidth: isMobile ? 0 : undefined,
  }

  const inputStyle: React.CSSProperties = {
    height: 36,
    padding: '0 10px',
    border: '1px solid #D1D5DB',
    borderRadius: 6,
    background: '#FFFFFF',
    color: '#374151',
    fontFamily: "'IBM Plex Sans', sans-serif",
    fontSize: 14,
    outline: 'none',
    width: isMobile ? '100%' : 160,
  }

  const controlBarStyle: React.CSSProperties = {
    height: isMobile ? 'auto' : 56,
    minHeight: isMobile ? undefined : 56,
    background: '#FFFFFF',
    borderBottom: '1px solid #E5E7EB',
    display: 'flex',
    alignItems: isMobile ? 'stretch' : 'center',
    gap: 12,
    padding: isMobile ? '12px 14px' : '0 24px',
    flexShrink: 0,
    flexWrap: 'wrap',
  }

  const mobileButtonRowStyle: React.CSSProperties = {
    width: '100%',
    display: 'flex',
    gap: 10,
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>

      {/* Tab bar */}
      <div style={{
        height: isMobile ? 40 : 44,
        background: '#FFFFFF',
        borderBottom: '1px solid #E5E7EB',
        display: 'flex',
        alignItems: 'flex-end',
        paddingLeft: isMobile ? 12 : 24,
        gap: 0,
        flexShrink: 0,
        overflowX: 'auto',
      }}>
        {(['speaker', 'listener'] as RtcTab[]).map(t => {
          const isActive = tab === t
          const Icon = t === 'speaker' ? Radio : User
          const label = t === 'speaker' ? '演讲者' : '听众'
          return (
            <button
              key={t}
              onClick={() => { setTab(t); clearPairs() }}
              style={{
                height: 40,
                padding: isMobile ? '0 12px' : '0 16px',
                border: 'none',
                borderBottom: isActive ? '2px solid #2563EB' : '2px solid transparent',
                background: 'none',
                color: isActive ? '#2563EB' : '#6B7280',
                fontFamily: "'IBM Plex Sans', sans-serif",
                fontWeight: isActive ? 600 : 400,
                fontSize: isMobile ? 13 : 14,
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                transition: 'color 150ms ease',
              }}
            >
              <Icon size={14} strokeWidth={1.5} />
              {label}
            </button>
          )
        })}
      </div>

      {tab === 'speaker' && (
        <>
          {/* Speaker control bar */}
          <div style={{
            ...controlBarStyle,
          }}>
            {isMobile ? (
              <>
                <div style={{ width: '100%', display: 'flex', alignItems: 'center', gap: 10 }}>
                  <select
                    aria-label="源语言"
                    value={sourceLang}
                    onChange={e => setSourceLang(e.target.value)}
                    disabled={isRunning}
                    style={selectStyle}
                  >
                    {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
                  </select>

                  <span style={{ color: '#9CA3AF', fontSize: 14, flexShrink: 0 }}>→</span>

                  <select
                    aria-label="目标语言"
                    value={targetLang}
                    onChange={e => setTargetLang(e.target.value)}
                    disabled={isRunning}
                    style={selectStyle}
                  >
                    {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
                  </select>
                </div>

                <input
                  aria-label="房间名"
                  placeholder="房间名"
                  value={room}
                  onChange={e => setRoom(e.target.value)}
                  disabled={isRunning}
                  style={inputStyle}
                />

                <div style={mobileButtonRowStyle}>
                  <button
                    onClick={handleSpeakerToggle}
                    disabled={isBusy || !room.trim()}
                    aria-label={isRunning ? '停止广播' : '开始广播'}
                    style={{
                      height: 36,
                      padding: '0 16px',
                      border: 'none',
                      borderRadius: 6,
                      background: isRunning ? '#DC2626' : '#2563EB',
                      color: '#FFFFFF',
                      fontFamily: "'IBM Plex Sans', sans-serif",
                      fontWeight: 500,
                      fontSize: 14,
                      cursor: isBusy || !room.trim() ? 'not-allowed' : 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: 6,
                      opacity: isBusy || !room.trim() ? 0.6 : 1,
                      transition: 'background 150ms ease',
                      flex: 1.6,
                    }}
                  >
                    {isRunning
                      ? <><Square size={14} strokeWidth={1.5} /> 停止</>
                      : pendingStart
                        ? <><Play size={14} strokeWidth={1.5} /> 启动中</>
                        : <><Play size={14} strokeWidth={1.5} /> 开始广播</>}
                  </button>

                  <button
                    onClick={handleTTSToggle}
                    aria-label={ttsEnabled ? '关闭声音' : '开启声音'}
                    aria-pressed={ttsEnabled}
                    style={{
                      height: 36,
                      padding: '0 12px',
                      border: `1px solid ${ttsEnabled ? '#BFDBFE' : '#D1D5DB'}`,
                      borderRadius: 6,
                      background: ttsEnabled ? '#EFF6FF' : '#FFFFFF',
                      color: ttsEnabled ? '#2563EB' : '#9CA3AF',
                      fontFamily: "'IBM Plex Sans', sans-serif",
                      fontSize: 13,
                      fontWeight: 500,
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: 5,
                      transition: 'all 150ms ease',
                      flex: 1,
                    }}
                  >
                    {ttsEnabled
                      ? <><Volume2 size={14} strokeWidth={1.5} /> 声音</>
                      : <><VolumeX size={14} strokeWidth={1.5} /> 声音</>}
                  </button>
                </div>
              </>
            ) : (
              <>
                <select
                  aria-label="源语言"
                  value={sourceLang}
                  onChange={e => setSourceLang(e.target.value)}
                  disabled={isRunning}
                  style={selectStyle}
                >
                  {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
                </select>

                <span style={{ color: '#9CA3AF', fontSize: 14 }}>→</span>

                <select
                  aria-label="目标语言"
                  value={targetLang}
                  onChange={e => setTargetLang(e.target.value)}
                  disabled={isRunning}
                  style={selectStyle}
                >
                  {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
                </select>

                <input
                  aria-label="房间名"
                  placeholder="房间名"
                  value={room}
                  onChange={e => setRoom(e.target.value)}
                  disabled={isRunning}
                  style={inputStyle}
                />

                <button
                  onClick={handleSpeakerToggle}
                  disabled={isBusy || !room.trim()}
                  aria-label={isRunning ? '停止广播' : '开始广播'}
                  style={{
                    height: 36,
                    padding: '0 16px',
                    border: 'none',
                    borderRadius: 6,
                    background: isRunning ? '#DC2626' : '#2563EB',
                    color: '#FFFFFF',
                    fontFamily: "'IBM Plex Sans', sans-serif",
                    fontWeight: 500,
                    fontSize: 14,
                    cursor: isBusy || !room.trim() ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 6,
                    opacity: isBusy || !room.trim() ? 0.6 : 1,
                    transition: 'background 150ms ease',
                  }}
                >
                  {isRunning
                    ? <><Square size={14} strokeWidth={1.5} /> 停止</>
                    : pendingStart
                      ? <><Play size={14} strokeWidth={1.5} /> 启动中</>
                      : <><Play size={14} strokeWidth={1.5} /> 开始广播</>}
                </button>

                <button
                  onClick={handleTTSToggle}
                  aria-label={ttsEnabled ? '关闭声音' : '开启声音'}
                  aria-pressed={ttsEnabled}
                  style={{
                    height: 36,
                    padding: '0 12px',
                    border: `1px solid ${ttsEnabled ? '#BFDBFE' : '#D1D5DB'}`,
                    borderRadius: 6,
                    background: ttsEnabled ? '#EFF6FF' : '#FFFFFF',
                    color: ttsEnabled ? '#2563EB' : '#9CA3AF',
                    fontFamily: "'IBM Plex Sans', sans-serif",
                    fontSize: 13,
                    fontWeight: 500,
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 5,
                    transition: 'all 150ms ease',
                  }}
                >
                  {ttsEnabled
                    ? <><Volume2 size={14} strokeWidth={1.5} /> 声音</>
                    : <><VolumeX size={14} strokeWidth={1.5} /> 声音</>}
                </button>
              </>
            )}
          </div>

          {/* Speaker interim zone */}
          <div style={{
            flexShrink: 0, minHeight: isMobile ? 96 : 120, background: '#FFFFFF',
            borderBottom: '1px solid #E5E7EB', padding: isMobile ? '16px 14px' : '20px 32px',
            display: 'flex', alignItems: 'center',
          }}>
            {interim ? (
              <p style={{
                fontFamily: "'IBM Plex Sans', sans-serif",
                fontSize: isMobile ? 20 : 26, fontWeight: 600, color: '#94A3B8',
                lineHeight: 1.35, animation: 'rtcInterimBreathe 1.8s ease-in-out infinite', margin: 0,
              }}>
                {interim}<span style={{ marginLeft: 4, opacity: 0.5 }}>…</span>
              </p>
            ) : isRunning ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                {[0, 0.18, 0.36].map((delay, i) => (
                  <span key={i} style={{
                    display: 'inline-block', width: 8, height: 8,
                    borderRadius: '50%', background: '#CBD5E1',
                    animation: `rtcDotPulse 1.2s ease-in-out ${delay}s infinite`,
                  }} />
                ))}
              </div>
            ) : (
              <p style={{ fontFamily: "'IBM Plex Sans', sans-serif", fontSize: isMobile ? 13 : 14, color: '#CBD5E1' }}>
                点击「开始广播」后将在此实时显示识别文字
              </p>
            )}
          </div>

          {/* Speaker body: subtitle feed + listener entry */}
          <div style={{ flex: 1, overflow: 'hidden', display: 'flex', gap: 0, flexDirection: isNarrow ? 'column' : 'row' }}>
            <div style={{ flex: 1, overflow: 'hidden', paddingTop: 8, minHeight: 0 }}>
              <SubtitleFeed
                pairs={pairs}
                emptyMessage={isRunning ? '翻译结果将在此显示' : ''}
              />
            </div>

            {!isMobile && (
            <div style={{
              width: isNarrow ? '100%' : 200,
              flexShrink: 0,
              background: '#FFFFFF',
              borderLeft: isNarrow ? 'none' : '1px solid #E5E7EB',
              borderTop: isNarrow ? '1px solid #E5E7EB' : 'none',
              padding: '20px 16px',
              display: 'flex',
              flexDirection: 'column',
              flexWrap: 'wrap',
              alignItems: 'center',
              gap: 12,
              justifyContent: 'flex-start',
            }}>
              <div style={{
                fontSize: 11, fontWeight: 700, textTransform: 'uppercase',
                letterSpacing: '0.08em', color: '#94A3B8',
                fontFamily: "'IBM Plex Sans', sans-serif",
              }}>
                听众入口
              </div>

              {qrDataUrl ? (
                <img
                  src={qrDataUrl}
                  alt="扫码加入"
                  width={isMobile ? 112 : 140}
                  height={isMobile ? 112 : 140}
                  style={{ borderRadius: 8, border: '1px solid #E2E8F0' }}
                />
              ) : (
                <div style={{
                  width: isMobile ? 112 : 140, height: isMobile ? 112 : 140, borderRadius: 8,
                  border: '1px dashed #D1D5DB',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  color: '#CBD5E1', fontSize: 11,
                  fontFamily: "'IBM Plex Sans', sans-serif",
                  textAlign: 'center', padding: 8,
                }}>
                  {window.location.protocol === 'http:' && window.location.hostname !== 'localhost'
                    ? '加载中...'
                    : '需在局域网\n访问时显示'}
                </div>
              )}

              {listenerUrl && (
                <div style={{
                  width: isMobile ? '100%' : 'auto',
                  fontSize: 10, color: '#64748B', wordBreak: 'break-all',
                  textAlign: 'center', fontFamily: "'IBM Plex Mono', monospace",
                }}>
                  {listenerUrl}
                </div>
              )}

              <div style={{
                fontSize: 11, color: '#94A3B8', textAlign: 'center',
                fontFamily: "'IBM Plex Sans', sans-serif",
                lineHeight: 1.5,
              }}>
                手机扫码<br />进入跨端协作
              </div>

              {isRunning && (
                <div style={{
                  display: 'flex', alignItems: 'center', gap: 5,
                  fontSize: 11, color: '#10B981',
                  fontFamily: "'IBM Plex Sans', sans-serif", fontWeight: 600,
                }}>
                  <span style={{
                    width: 6, height: 6, borderRadius: '50%',
                    background: '#10B981', display: 'inline-block',
                  }} />
                  直播中 · {room}
                </div>
              )}
            </div>
            )}
          </div>
        </>
      )}

      {tab === 'listener' && (
        <>
          {/* Listener control bar */}
          <div style={{
            ...controlBarStyle,
          }}>
            {isMobile ? (
              <>
                <input
                  aria-label="房间名"
                  placeholder="输入房间名"
                  value={room}
                  onChange={e => setRoom(e.target.value)}
                  disabled={isRunning}
                  style={inputStyle}
                />

                <div style={mobileButtonRowStyle}>
                  <button
                    onClick={handleListenerToggle}
                    disabled={isBusy || !room.trim()}
                    aria-label={isRunning ? '离开房间' : '加入房间'}
                    style={{
                      height: 36,
                      padding: '0 16px',
                      border: 'none',
                      borderRadius: 6,
                      background: isRunning ? '#DC2626' : '#2563EB',
                      color: '#FFFFFF',
                      fontFamily: "'IBM Plex Sans', sans-serif",
                      fontWeight: 500,
                      fontSize: 14,
                      cursor: isBusy || !room.trim() ? 'not-allowed' : 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: 6,
                      opacity: isBusy || !room.trim() ? 0.6 : 1,
                      transition: 'background 150ms ease',
                      flex: 1.5,
                    }}
                  >
                    {isRunning
                      ? <><Square size={14} strokeWidth={1.5} /> 离开</>
                      : pendingStart
                        ? <><Play size={14} strokeWidth={1.5} /> 连接中</>
                        : <><Play size={14} strokeWidth={1.5} /> 加入</>}
                  </button>

                  <button
                    onClick={handleTTSToggle}
                    aria-label={ttsEnabled ? '关闭声音' : '开启声音'}
                    aria-pressed={ttsEnabled}
                    style={{
                      height: 36,
                      padding: '0 12px',
                      border: `1px solid ${ttsEnabled ? '#BFDBFE' : '#D1D5DB'}`,
                      borderRadius: 6,
                      background: ttsEnabled ? '#EFF6FF' : '#FFFFFF',
                      color: ttsEnabled ? '#2563EB' : '#9CA3AF',
                      fontFamily: "'IBM Plex Sans', sans-serif",
                      fontSize: 13,
                      fontWeight: 500,
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      gap: 5,
                      transition: 'all 150ms ease',
                      flex: 1,
                    }}
                  >
                    {ttsEnabled
                      ? <><Volume2 size={14} strokeWidth={1.5} /> 声音</>
                      : <><VolumeX size={14} strokeWidth={1.5} /> 声音</>}
                  </button>
                </div>
              </>
            ) : (
              <>
                <input
                  aria-label="房间名"
                  placeholder="输入房间名"
                  value={room}
                  onChange={e => setRoom(e.target.value)}
                  disabled={isRunning}
                  style={inputStyle}
                />

                <button
                  onClick={handleListenerToggle}
                  disabled={isBusy || !room.trim()}
                  aria-label={isRunning ? '离开房间' : '加入房间'}
                  style={{
                    height: 36,
                    padding: '0 16px',
                    border: 'none',
                    borderRadius: 6,
                    background: isRunning ? '#DC2626' : '#2563EB',
                    color: '#FFFFFF',
                    fontFamily: "'IBM Plex Sans', sans-serif",
                    fontWeight: 500,
                    fontSize: 14,
                    cursor: isBusy || !room.trim() ? 'not-allowed' : 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 6,
                    opacity: isBusy || !room.trim() ? 0.6 : 1,
                    transition: 'background 150ms ease',
                  }}
                >
                  {isRunning
                    ? <><Square size={14} strokeWidth={1.5} /> 离开</>
                    : pendingStart
                      ? <><Play size={14} strokeWidth={1.5} /> 连接中</>
                      : <><Play size={14} strokeWidth={1.5} /> 加入</>}
                </button>

                <button
                  onClick={handleTTSToggle}
                  aria-label={ttsEnabled ? '关闭声音' : '开启声音'}
                  aria-pressed={ttsEnabled}
                  style={{
                    height: 36,
                    padding: '0 12px',
                    border: `1px solid ${ttsEnabled ? '#BFDBFE' : '#D1D5DB'}`,
                    borderRadius: 6,
                    background: ttsEnabled ? '#EFF6FF' : '#FFFFFF',
                    color: ttsEnabled ? '#2563EB' : '#9CA3AF',
                    fontFamily: "'IBM Plex Sans', sans-serif",
                    fontSize: 13,
                    fontWeight: 500,
                    cursor: 'pointer',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 5,
                    transition: 'all 150ms ease',
                  }}
                >
                  {ttsEnabled
                    ? <><Volume2 size={14} strokeWidth={1.5} /> 声音</>
                    : <><VolumeX size={14} strokeWidth={1.5} /> 声音</>}
                </button>
              </>
            )}

            {isRunning && (
              <div style={{
                width: '100%',
                display: 'flex', alignItems: 'center', gap: 6,
                fontSize: 13, color: '#10B981',
                fontFamily: "'IBM Plex Sans', sans-serif", fontWeight: 600,
              }}>
                <span style={{
                  width: 7, height: 7, borderRadius: '50%',
                  background: '#10B981', display: 'inline-block',
                }} />
                已连接 · {room}
              </div>
            )}

            {ttsError && (
              <div style={{
                width: '100%',
                fontSize: 12,
                color: '#B45309',
                fontFamily: "'IBM Plex Sans', sans-serif",
              }}>
                {ttsError}
              </div>
            )}
          </div>

          {/* Listener interim zone */}
          <div style={{
            flexShrink: 0, minHeight: isMobile ? 96 : 120, background: '#FFFFFF',
            borderBottom: '1px solid #E5E7EB', padding: isMobile ? '16px 14px' : '20px 32px',
            display: 'flex', alignItems: 'center',
          }}>
            {interim ? (
              <p style={{
                fontFamily: "'IBM Plex Sans', sans-serif",
                fontSize: isMobile ? 20 : 26, fontWeight: 600, color: '#94A3B8',
                lineHeight: 1.35, animation: 'rtcInterimBreathe 1.8s ease-in-out infinite', margin: 0,
              }}>
                {interim}<span style={{ marginLeft: 4, opacity: 0.5 }}>…</span>
              </p>
            ) : isRunning ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                {[0, 0.18, 0.36].map((delay, i) => (
                  <span key={i} style={{
                    display: 'inline-block', width: 8, height: 8,
                    borderRadius: '50%', background: '#CBD5E1',
                    animation: `rtcDotPulse 1.2s ease-in-out ${delay}s infinite`,
                  }} />
                ))}
              </div>
            ) : (
              <p style={{ fontFamily: "'IBM Plex Sans', sans-serif", fontSize: isMobile ? 13 : 14, color: '#CBD5E1' }}>
                加入房间后将在此实时显示识别文字
              </p>
            )}
          </div>

          <div style={{ flex: 1, overflow: 'hidden', paddingTop: 8, minHeight: 0 }}>
            <SubtitleFeed
              pairs={pairs}
              emptyMessage={isRunning ? '等待演讲者发送字幕...' : '输入房间名，点击「加入」接收实时字幕'}
            />
          </div>
        </>
      )}
    </div>
  )
}
