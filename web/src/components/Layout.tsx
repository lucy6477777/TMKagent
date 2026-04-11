import type { PipelineState } from '../types/ws'
import { Topbar } from './Topbar'
import { Sidebar, type Page } from './Sidebar'

interface Props {
  pipelineState: PipelineState
  currentPage: Page
  onNavigate: (page: Page) => void
  children: React.ReactNode
}

export function Layout({ pipelineState, currentPage, onNavigate, children }: Props) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100dvh' }}>
      <Topbar pipelineState={pipelineState} />
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        <Sidebar current={currentPage} onChange={onNavigate} />
        <main
          id="main-content"
          style={{ flex: 1, overflowY: 'auto', background: '#F8F9FA' }}
        >
          {children}
        </main>
      </div>
    </div>
  )
}
