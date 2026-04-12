import type { PipelineState } from '../types/ws'
import { StatusBadge } from './StatusBadge'

interface Props {
  pipelineState: PipelineState
  isMobile?: boolean
}

export function Topbar({ pipelineState, isMobile = false }: Props) {
  return (
    <header
      style={{
        height: isMobile ? 52 : 48,
        background: '#FFFFFF',
        borderBottom: '1px solid #E5E7EB',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: isMobile ? '0 14px' : '0 24px',
        paddingTop: isMobile ? 'env(safe-area-inset-top)' : 0,
        position: 'sticky',
        top: 0,
        zIndex: 40,
      }}
    >
      <span
        style={{
          fontFamily: "'IBM Plex Sans', sans-serif",
          fontWeight: 600,
          fontSize: isMobile ? 16 : 18,
          color: '#2563EB',
          letterSpacing: '-0.02em',
        }}
      >
        mini-tmk-agent
      </span>
      <StatusBadge state={pipelineState} />
    </header>
  )
}
