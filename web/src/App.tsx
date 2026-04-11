import { useState } from 'react'

type Page = 'stream' | 'transcript'

export default function App() {
  const [page, setPage] = useState<Page>('stream')
  return (
    <div style={{ fontFamily: 'IBM Plex Sans, sans-serif' }}>
      <p>mini-tmk-agent — page: {page}</p>
      <button onClick={() => setPage('stream')}>Stream</button>
      <button onClick={() => setPage('transcript')}>Transcript</button>
    </div>
  )
}
