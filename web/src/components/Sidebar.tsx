import { AudioLines, FileText, House, Radio, Settings } from 'lucide-react'

export type Page = 'home' | 'stream' | 'transcript' | 'rtc'

interface NavItem {
  id: Page
  label: string
  Icon: typeof AudioLines
}

const NAV_ITEMS: NavItem[] = [
  { id: 'home',       label: '首页',     Icon: House },
  { id: 'stream',     label: '实时翻译', Icon: AudioLines },
  { id: 'transcript', label: '文件转录', Icon: FileText },
  { id: 'rtc',        label: '跨端协作', Icon: Radio },
]

interface Props {
  current: Page
  onChange: (page: Page) => void
}

export function Sidebar({ current, onChange }: Props) {
  return (
    <nav
      aria-label="主导航"
      style={{
        width: 180,
        background: '#FFFFFF',
        borderRight: '1px solid #E5E7EB',
        display: 'flex',
        flexDirection: 'column',
        flexShrink: 0,
      }}
    >
      <div style={{ flex: 1, paddingTop: 12 }}>
        {NAV_ITEMS.map(({ id, label, Icon }) => {
          const isActive = id === current
          return (
            <button
              key={id}
              onClick={() => onChange(id)}
              aria-current={isActive ? 'page' : undefined}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                height: 40,
                paddingLeft: isActive ? 16 : 20,
                paddingRight: 16,
                border: 'none',
                borderLeft: isActive ? '4px solid #2563EB' : '4px solid transparent',
                background: isActive ? '#EFF6FF' : 'transparent',
                color: isActive ? '#2563EB' : '#374151',
                fontFamily: "'IBM Plex Sans', sans-serif",
                fontWeight: isActive ? 600 : 400,
                fontSize: 14,
                cursor: 'pointer',
                textAlign: 'left',
                transition: 'background 150ms ease, color 150ms ease',
              }}
            >
              <Icon size={16} strokeWidth={1.5} />
              {label}
            </button>
          )
        })}
      </div>

      <div style={{ borderTop: '1px solid #E5E7EB', padding: '12px 0' }}>
        <button
          style={{
            width: '100%',
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            height: 40,
            paddingLeft: 20,
            paddingRight: 16,
            border: 'none',
            borderLeft: '4px solid transparent',
            background: 'transparent',
            color: '#9CA3AF',
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontWeight: 400,
            fontSize: 14,
            cursor: 'not-allowed',
            textAlign: 'left',
          }}
          disabled
        >
          <Settings size={16} strokeWidth={1.5} />
          设置
        </button>
      </div>
    </nav>
  )
}
