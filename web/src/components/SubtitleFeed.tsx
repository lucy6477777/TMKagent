import { useEffect, useRef, useState } from 'react'
import { Check, Copy } from 'lucide-react'
import type { SubtitlePair } from '../types/ws'

interface Props {
  pairs: SubtitlePair[]
  showCopy?: boolean
  emptyMessage?: string
}

interface PairItemProps {
  pair: SubtitlePair
  isLatest: boolean
  showCopy: boolean
  isNew: boolean
}

function PairItem({ pair, isLatest, showCopy, isNew }: PairItemProps) {
  const [copied, setCopied] = useState(false)
  const [hovered, setHovered] = useState(false)
  const [visible, setVisible] = useState(!isNew)

  useEffect(() => {
    if (isNew) {
      // Trigger animation on mount
      const raf = requestAnimationFrame(() => setVisible(true))
      return () => cancelAnimationFrame(raf)
    }
  }, [isNew])

  const handleCopy = () => {
    const text = `${pair.source}\n${pair.target}`
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }

  const prefersReducedMotion =
    typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches

  return (
    <article
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        padding: '16px 0',
        borderBottom: '1px solid #E5E7EB',
        borderLeft: isLatest ? '2px solid #1E3A5F' : '2px solid transparent',
        paddingLeft: 12,
        opacity: isLatest ? 1 : 0.55,
        transition: 'opacity 600ms ease',
        position: 'relative',
        transform: isNew && !prefersReducedMotion && !visible
          ? 'translateY(8px)' : 'translateY(0)',
        ...(isNew && !prefersReducedMotion ? {
          transition: visible
            ? 'transform 200ms ease-out, opacity 200ms ease-out'
            : 'none',
          opacity: visible ? (isLatest ? 1 : 0.55) : 0,
        } : {}),
      }}
    >
      {/* Source line */}
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10, marginBottom: 8 }}>
        <span style={{
          flexShrink: 0,
          fontFamily: "'IBM Plex Mono', monospace",
          fontWeight: 500,
          fontSize: 11,
          letterSpacing: '0.05em',
          background: '#64748B',
          color: '#FFFFFF',
          borderRadius: 4,
          padding: '2px 8px',
        }}>SRC</span>
        <span style={{
          fontFamily: "'IBM Plex Mono', monospace",
          fontWeight: 400,
          fontSize: 18,
          color: '#111827',
          lineHeight: 1.6,
        }}>{pair.source}</span>
      </div>

      {/* Target line */}
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 10 }}>
        <span style={{
          flexShrink: 0,
          fontFamily: "'IBM Plex Mono', monospace",
          fontWeight: 500,
          fontSize: 11,
          letterSpacing: '0.05em',
          background: '#1E3A5F',
          color: '#FFFFFF',
          borderRadius: 4,
          padding: '2px 8px',
        }}>TGT</span>
        <span style={{
          fontFamily: "'IBM Plex Mono', monospace",
          fontWeight: 400,
          fontSize: 18,
          color: '#111827',
          lineHeight: 1.6,
        }}>{pair.target}</span>
      </div>

      {/* Copy button — hover only */}
      {showCopy && hovered && (
        <button
          onClick={handleCopy}
          aria-label="复制字幕"
          title="复制字幕"
          style={{
            position: 'absolute',
            top: 12,
            right: 0,
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            padding: 4,
            color: '#9CA3AF',
            display: 'flex',
            alignItems: 'center',
          }}
        >
          {copied
            ? <Check size={14} strokeWidth={1.5} color="#10B981" />
            : <Copy size={14} strokeWidth={1.5} />}
        </button>
      )}
    </article>
  )
}

export function SubtitleFeed({ pairs, showCopy = false, emptyMessage = '等待字幕...' }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const prevLengthRef = useRef(pairs.length)

  // Track which pair IDs are new (just arrived)
  const newIdsRef = useRef<Set<string>>(new Set())

  useEffect(() => {
    if (pairs.length > prevLengthRef.current) {
      // Mark newly added pairs
      const newPairs = pairs.slice(prevLengthRef.current)
      newPairs.forEach(p => newIdsRef.current.add(p.id))
      // Auto-scroll
      const el = containerRef.current
      if (el) {
        el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
      }
    }
    prevLengthRef.current = pairs.length
  }, [pairs])

  return (
    <div
      ref={containerRef}
      role="log"
      aria-live="polite"
      aria-label="翻译字幕"
      style={{ height: '100%', overflowY: 'auto', padding: '0 24px' }}
    >
      {pairs.length === 0 ? (
        <p style={{
          color: '#9CA3AF',
          fontFamily: "'IBM Plex Sans', sans-serif",
          fontSize: 14,
          marginTop: 48,
          textAlign: 'center',
        }}>
          {emptyMessage}
        </p>
      ) : (
        pairs.map((pair, i) => (
          <PairItem
            key={pair.id}
            pair={pair}
            isLatest={i === pairs.length - 1}
            showCopy={showCopy}
            isNew={newIdsRef.current.has(pair.id)}
          />
        ))
      )}
    </div>
  )
}
