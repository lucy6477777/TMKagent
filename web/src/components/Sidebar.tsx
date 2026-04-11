import { AudioLines, FileText, Settings } from 'lucide-react'

type Page = 'stream' | 'transcript'

interface NavItem {
  id: Page | 'settings'
  label: string
  Icon: typeof AudioLines
}

const NAV_ITEMS: NavItem[] = [
  { id: 'stream',     label: 'Stream',     Icon: AudioLines },
  { id: 'transcript', label: 'Transcript', Icon: FileText },
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
              onClick={() => onChange(id as Page)}
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
                borderLeft: isActive ? '4px solid #1E3A5F' : '4px solid transparent',
                background: isActive ? '#1E3A5F' : 'transparent',
                color: isActive ? '#FFFFFF' : '#374151',
                fontFamily: "'IBM Plex Sans', sans-serif",
                fontWeight: isActive ? 500 : 400,
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

      {/* Divider + Settings placeholder */}
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
            color: '#374151',
            fontFamily: "'IBM Plex Sans', sans-serif",
            fontWeight: 400,
            fontSize: 14,
            cursor: 'pointer',
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
