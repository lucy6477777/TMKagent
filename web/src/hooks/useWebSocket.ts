import { useCallback, useEffect, useRef, useState } from 'react'
import type { PipelineState, SubtitlePair, WSCommand, WSMessage } from '../types/ws'

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected'

interface UseWebSocketReturn {
  status: ConnectionStatus
  pipelineState: PipelineState
  pairs: SubtitlePair[]
  progress: { current: number; total: number } | null
  sendCmd: (cmd: WSCommand) => void
  clearPairs: () => void
}

const MAX_PAIRS = 100
const MAX_RETRIES = 5
const BASE_RETRY_DELAY = 1000 // ms

export function useWebSocket(url: string): UseWebSocketReturn {
  const wsRef = useRef<WebSocket | null>(null)
  const retriesRef = useRef(0)
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const unmountedRef = useRef(false)

  const [status, setStatus] = useState<ConnectionStatus>('connecting')
  const [pipelineState, setPipelineState] = useState<PipelineState>('idle')
  const [pairs, setPairs] = useState<SubtitlePair[]>([])
  const [progress, setProgress] = useState<{ current: number; total: number } | null>(null)

  const connect = useCallback(() => {
    if (unmountedRef.current) return

    setStatus('connecting')
    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      retriesRef.current = 0
      setStatus('connected')
    }

    ws.onmessage = (event) => {
      let msg: WSMessage
      try {
        msg = JSON.parse(event.data as string)
      } catch {
        return
      }

      switch (msg.type) {
        case 'pair':
          setPairs(prev => {
            const next = [...prev, {
              id: `${msg.ts}-${Math.random().toString(36).slice(2, 7)}`,
              source: msg.source,
              target: msg.target,
              ts: msg.ts,
            }]
            return next.length > MAX_PAIRS ? next.slice(-MAX_PAIRS) : next
          })
          setProgress(null)
          break
        case 'status':
          setPipelineState(msg.state)
          break
        case 'progress':
          setProgress({ current: msg.current, total: msg.total })
          break
        case 'error':
          console.error('WS error from server:', msg.msg)
          break
      }
    }

    ws.onclose = () => {
      if (unmountedRef.current) return
      setStatus('disconnected')
      setPipelineState('idle')
      // Exponential backoff retry
      if (retriesRef.current < MAX_RETRIES) {
        const delay = BASE_RETRY_DELAY * Math.pow(2, retriesRef.current)
        retriesRef.current += 1
        retryTimerRef.current = setTimeout(connect, delay)
      }
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [url])

  useEffect(() => {
    unmountedRef.current = false
    connect()
    return () => {
      unmountedRef.current = true
      if (retryTimerRef.current) clearTimeout(retryTimerRef.current)
      wsRef.current?.close()
    }
  }, [connect])

  const sendCmd = useCallback((cmd: WSCommand) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(cmd))
    }
  }, [])

  const clearPairs = useCallback(() => {
    setPairs([])
    setProgress(null)
  }, [])

  return { status, pipelineState, pairs, progress, sendCmd, clearPairs }
}
