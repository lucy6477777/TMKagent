import { useCallback, useState } from 'react'
import { Layout } from './components/Layout'
import { StreamPage } from './pages/StreamPage'
import { TranscriptPage } from './pages/TranscriptPage'
import { useWebSocket } from './hooks/useWebSocket'

type Page = 'stream' | 'transcript'

const WS_URL = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`

export default function App() {
  const [page, setPage] = useState<Page>('stream')
  const [showStopToast, setShowStopToast] = useState(false)

  const { pipelineState, pairs, progress, sendCmd, clearPairs } = useWebSocket(WS_URL)

  const isRunning = pipelineState === 'listening' || pipelineState === 'processing'

  const handleNavigate = useCallback((nextPage: Page) => {
    if (nextPage === page) return
    // If stream is running and user switches away, stop it and notify
    if (page === 'stream' && isRunning) {
      sendCmd({ type: 'cmd', action: 'stop' })
      setShowStopToast(true)
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
      {page === 'stream' && (
        <StreamPage
          pairs={pairs}
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
    </Layout>
  )
}
