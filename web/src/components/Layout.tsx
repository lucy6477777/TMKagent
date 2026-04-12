import type { PipelineState } from '../types/ws'
import { useViewport } from '../hooks/useViewport'
import { Topbar } from './Topbar'
import { Sidebar, type Page } from './Sidebar'

interface Props {
  pipelineState: PipelineState
  currentPage: Page
  onNavigate: (page: Page) => void
  children: React.ReactNode
}

export function Layout({ pipelineState, currentPage, onNavigate, children }: Props) {
  const { isMobile } = useViewport()

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100dvh' }}>
      <Topbar pipelineState={pipelineState} isMobile={isMobile} />
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden', flexDirection: isMobile ? 'column' : 'row' }}>
        {!isMobile && <Sidebar current={currentPage} onChange={onNavigate} isMobile={false} />}
        <main
          id="main-content"
          style={{
            flex: 1,
            overflowY: 'auto',
            background: '#F8F9FA',
            minHeight: 0,
            paddingBottom: isMobile ? 8 : 0,
          }}
        >
          {children}
        </main>
        {isMobile && <Sidebar current={currentPage} onChange={onNavigate} isMobile />}
      </div>
    </div>
  )
}
