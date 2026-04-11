import type { PipelineState } from '../types/ws'

const STATE_CONFIG: Record<PipelineState, { color: string; label: string; pulse: boolean }> = {
  listening:  { color: '#10B981', label: '监听中', pulse: true },
  processing: { color: '#F59E0B', label: '处理中', pulse: false },
  idle:       { color: '#9CA3AF', label: '就绪',   pulse: false },
  error:      { color: '#DC2626', label: '错误',   pulse: false },
}

interface Props {
  state: PipelineState
}

export function StatusBadge({ state }: Props) {
  const { color, label, pulse } = STATE_CONFIG[state]
  return (
    <span style={{ display: 'flex', alignItems: 'center', gap: '6px', fontSize: '14px', color: '#374151' }}>
      <span
        className={pulse ? 'status-pulse' : undefined}
        style={{
          width: 10, height: 10, borderRadius: '50%',
          backgroundColor: color,
          display: 'inline-block', flexShrink: 0,
        }}
      />
      {label}
    </span>
  )
}
