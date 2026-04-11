export type PipelineState = 'listening' | 'processing' | 'idle' | 'error'

export type WSMessage =
  | { type: 'pair'; source: string; target: string; ts: number }
  | { type: 'interim'; text: string }
  | { type: 'status'; state: PipelineState }
  | { type: 'progress'; current: number; total: number }
  | { type: 'error'; msg: string }

export type WSCommand =
  | { type: 'cmd'; action: 'start_stream'; sourceLang: string; targetLang: string }
  | { type: 'cmd'; action: 'stop' }
  | { type: 'cmd'; action: 'transcript'; sourceLang: string; targetLang: string }
  | { type: 'cmd'; action: 'rtc_speaker_start'; room: string; sourceLang: string; targetLang: string }
  | { type: 'cmd'; action: 'rtc_join'; room: string; role: 'listener' }
  | { type: 'cmd'; action: 'rtc_stop' }

export interface SubtitlePair {
  id: string
  source: string
  target: string
  ts: number
}
