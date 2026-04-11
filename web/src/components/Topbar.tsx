import type { PipelineState } from '../types/ws'
import { StatusBadge } from './StatusBadge'

interface Props {
  pipelineState: PipelineState
}

export function Topbar({ pipelineState }: Props) {
  return (
    <header
      style={{
        height: 48,
        background: '#FFFFFF',
        borderBottom: '1px solid #E5E7EB',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 24px',
        position: 'sticky',
        top: 0,
        zIndex: 40,
      }}
    >
      <span
        style={{
          fontFamily: "'IBM Plex Sans', sans-serif",
          fontWeight: 600,
          fontSize: 18,
          color: '#1E3A5F',
          letterSpacing: '-0.02em',
        }}
      >
        mini-tmk-agent
      </span>
      <StatusBadge state={pipelineState} />
    </header>
  )
}
