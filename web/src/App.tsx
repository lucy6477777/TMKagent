import { useCallback, useState } from 'react'
import { Layout } from './components/Layout'
import { type Page } from './components/Sidebar'
import { HomePage } from './pages/HomePage'
import { StreamPage } from './pages/StreamPage'
import { TranscriptPage } from './pages/TranscriptPage'
import { RtcPage } from './pages/RtcPage'
import { useWebSocket } from './hooks/useWebSocket'

const WS_URL = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`

export default function App() {
  const [page, setPage] = useState<Page>('home')
  const [showStopToast, setShowStopToast] = useState(false)

  const { pipelineState, pairs, interim, progress, sendCmd, clearPairs } = useWebSocket(WS_URL)

  const isRunning = pipelineState === 'listening' || pipelineState === 'processing'

  const handleNavigate = useCallback((nextPage: Page) => {
    if (nextPage === page) return
    if ((page === 'stream' || page === 'rtc') && isRunning) {
      sendCmd({ type: 'cmd', action: 'stop' })
      if (page === 'stream') setShowStopToast(true)
    }
    clearPairs()
    setPage(nextPage)
  }, [page, isRunning, sendCmd, clearPairs])

  return (
    <Layout
      pipelineState={pipelineState}
      currentPage={page}
      onNavigate={handleNavigate}
    >
      {page === 'home' && (
        <HomePage onNavigate={handleNavigate} />
      )}
      {page === 'stream' && (
        <StreamPage
          pairs={pairs}
          interim={interim}
          pipelineState={pipelineState}
          sendCmd={sendCmd}
          showStopToast={showStopToast}
          onToastDismissed={() => setShowStopToast(false)}
        />
      )}
      {page === 'transcript' && (
        <TranscriptPage
          pairs={pairs}
          pipelineState={pipelineState}
          progress={progress}
          sendCmd={sendCmd}
          clearPairs={clearPairs}
        />
      )}
      {page === 'rtc' && (
        <RtcPage
          pairs={pairs}
          interim={interim}
          pipelineState={pipelineState}
          sendCmd={sendCmd}
          clearPairs={clearPairs}
        />
      )}
    </Layout>
  )
}
