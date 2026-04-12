import { useCallback, useState } from 'react'
import { Upload } from 'lucide-react'
import type { PipelineState, SubtitlePair, WSCommand } from '../types/ws'
import { SubtitleFeed } from '../components/SubtitleFeed'

const LANGUAGES = [
  { value: 'zh', label: '中文' },
  { value: 'en', label: 'English' },
  { value: 'es', label: 'Español' },
  { value: 'ja', label: '日本語' },
]

const ACCEPT_TYPES = '.wav,.mp3,.m4a,.pcm,audio/wav,audio/x-wav,audio/mpeg,audio/mp4,audio/x-m4a'
const SUPPORTED_EXTENSIONS = new Set(['wav', 'mp3', 'm4a', 'pcm'])

interface Props {
  pairs: SubtitlePair[]
  pipelineState: PipelineState
  progress: { current: number; total: number } | null
  sendCmd: (cmd: WSCommand) => void
  clearPairs: () => void
}

export function TranscriptPage({ pairs, pipelineState, progress, sendCmd, clearPairs }: Props) {
  const [dragOver, setDragOver] = useState(false)
  const [sourceLang, setSourceLang] = useState('zh')
  const [targetLang, setTargetLang] = useState('en')
  const [fileName, setFileName] = useState<string | null>(null)
  const [uploadError, setUploadError] = useState<string | null>(null)

  const isProcessing = pipelineState === 'processing' || pipelineState === 'listening'

  const validateFile = useCallback((file: File) => {
    const extension = file.name.split('.').pop()?.toLowerCase() || ''
    if (!SUPPORTED_EXTENSIONS.has(extension)) {
      return '当前仅支持 WAV、MP3、M4A、PCM 音频文件'
    }
    return null
  }, [])

  const uploadAndTranscribe = useCallback(async (file: File) => {
    const validationError = validateFile(file)
    if (validationError) {
      setUploadError(validationError)
      setFileName(null)
      return
    }

    setUploadError(null)
    setFileName(file.name)
    clearPairs()

    const formData = new FormData()
    formData.append('file', file)

    try {
      const res = await fetch('/upload', { method: 'POST', body: formData })
      if (!res.ok) throw new Error(`Upload failed: ${res.status}`)
      // File uploaded — now trigger transcript via WS
      sendCmd({ type: 'cmd', action: 'transcript', sourceLang, targetLang })
    } catch (e) {
      setUploadError(e instanceof Error ? e.message : 'Upload failed')
      setFileName(null)
    }
  }, [sourceLang, targetLang, sendCmd, clearPairs, validateFile])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
    const file = e.dataTransfer.files[0]
    if (file) uploadAndTranscribe(file)
  }, [uploadAndTranscribe])

  const handleFileInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) uploadAndTranscribe(file)
    e.target.value = '' // reset so same file can be re-selected
  }, [uploadAndTranscribe])

  const progressPercent = progress
    ? Math.round((progress.current / progress.total) * 100)
    : 0

  const showProgress = isProcessing && progress !== null

  const selectStyle: React.CSSProperties = {
    height: 32,
    padding: '0 8px',
    border: '1px solid #D1D5DB',
    borderRadius: 6,
    background: '#FFFFFF',
    color: '#374151',
    fontFamily: "'IBM Plex Sans', sans-serif",
    fontSize: 13,
    cursor: 'pointer',
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', padding: '24px 24px 0' }}>

      {/* Language row */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16 }}>
        <select aria-label="源语言" value={sourceLang} onChange={e => setSourceLang(e.target.value)} disabled={isProcessing} style={selectStyle}>
          {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
        </select>
        <span style={{ color: '#9CA3AF', fontSize: 14 }}>→</span>
        <select aria-label="目标语言" value={targetLang} onChange={e => setTargetLang(e.target.value)} disabled={isProcessing} style={selectStyle}>
          {LANGUAGES.map(l => <option key={l.value} value={l.value}>{l.label}</option>)}
        </select>
      </div>

      {/* Upload zone */}
      <label
        htmlFor="file-upload"
        onDragOver={e => { e.preventDefault(); setDragOver(true) }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          height: 160,
          border: dragOver ? '1.5px solid #1E3A5F' : '1.5px dashed #D1D5DB',
          borderRadius: 8,
          background: dragOver ? '#EEF2FF' : '#FFFFFF',
          cursor: isProcessing ? 'not-allowed' : 'pointer',
          transition: 'border-color 150ms ease, background 150ms ease',
          marginBottom: 0,
          flexShrink: 0,
          userSelect: 'none',
        }}
      >
        <Upload size={20} strokeWidth={1.5} color={dragOver ? '#1E3A5F' : '#9CA3AF'} />
        <span style={{
          marginTop: 12,
          fontSize: 14,
          fontFamily: "'IBM Plex Sans', sans-serif",
          color: '#6B7280',
        }}>
          {fileName
            ? `已选择：${fileName}`
            : '拖拽 WAV / MP3 / M4A / PCM 至此，或'}
        </span>
        {!fileName && (
          <span style={{ fontSize: 14, fontFamily: "'IBM Plex Sans', sans-serif", color: '#1E3A5F', marginTop: 2 }}>
            点击选择文件
          </span>
        )}
        <input
          id="file-upload"
          type="file"
          accept={ACCEPT_TYPES}
          onChange={handleFileInput}
          disabled={isProcessing}
          style={{ position: 'absolute', width: 1, height: 1, opacity: 0, overflow: 'hidden' }}
          aria-label="选择音频文件（WAV、MP3、M4A 或 PCM）"
        />
      </label>

      {/* Error */}
      {uploadError && (
        <p role="alert" style={{ color: '#DC2626', fontSize: 13, fontFamily: "'IBM Plex Sans', sans-serif", margin: '8px 0 0' }}>
          {uploadError}
        </p>
      )}

      {/* Progress bar */}
      {showProgress && (
        <div style={{ marginTop: 16, flexShrink: 0 }}>
          <div style={{ height: 4, background: '#E5E7EB', borderRadius: 2, overflow: 'hidden' }}>
            <div style={{
              height: '100%',
              width: `${progressPercent}%`,
              background: '#1E3A5F',
              borderRadius: 2,
              transition: 'width 300ms ease',
            }} />
          </div>
          <p style={{
            margin: '6px 0 0',
            fontSize: 13,
            fontFamily: "'IBM Plex Sans', sans-serif",
            color: '#6B7280',
          }}>
            处理中 {progress!.current}/{progress!.total}
          </p>
        </div>
      )}

      {/* Results */}
      <div style={{ flex: 1, overflow: 'hidden', marginTop: 20, paddingBottom: 24 }}>
        <SubtitleFeed
          pairs={pairs}
          showCopy={true}
          emptyMessage="上传音频文件后，转录结果将在此显示"
        />
      </div>
    </div>
  )
}
