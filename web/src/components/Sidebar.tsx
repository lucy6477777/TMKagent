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
  isMobile?: boolean
}

export function Sidebar({ current, onChange, isMobile = false }: Props) {
  return (
    <nav
      aria-label="主导航"
      style={{
        width: isMobile ? '100%' : 180,
        background: '#FFFFFF',
        borderRight: isMobile ? 'none' : '1px solid #E5E7EB',
        borderTop: isMobile ? '1px solid #E5E7EB' : 'none',
        display: 'flex',
        flexDirection: isMobile ? 'row' : 'column',
        flexShrink: 0,
        overflowX: isMobile ? 'auto' : 'visible',
        position: isMobile ? 'sticky' : 'static',
        bottom: isMobile ? 0 : 'auto',
        zIndex: isMobile ? 30 : 'auto',
        paddingBottom: isMobile ? 'max(env(safe-area-inset-bottom), 8px)' : 0,
        boxShadow: isMobile ? '0 -8px 20px rgba(15, 23, 42, 0.06)' : 'none',
      }}
    >
      <div style={{ flex: 1, paddingTop: isMobile ? 0 : 12, display: 'flex', flexDirection: isMobile ? 'row' : 'column' }}>
        {NAV_ITEMS.map(({ id, label, Icon }) => {
          const isActive = id === current
          return (
            <button
              key={id}
              onClick={() => onChange(id)}
              aria-current={isActive ? 'page' : undefined}
              style={{
                width: isMobile ? '25%' : '100%',
                minWidth: isMobile ? 80 : '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: isMobile ? 'center' : 'flex-start',
                flexDirection: isMobile ? 'column' : 'row',
                gap: 10,
                height: isMobile ? 60 : 40,
                paddingLeft: isMobile ? 8 : (isActive ? 16 : 20),
                paddingRight: isMobile ? 8 : 16,
                border: 'none',
                borderLeft: isMobile ? 'none' : (isActive ? '4px solid #2563EB' : '4px solid transparent'),
                borderTop: isMobile ? (isActive ? '3px solid #2563EB' : '3px solid transparent') : 'none',
                background: isActive ? '#EFF6FF' : 'transparent',
                color: isActive ? '#2563EB' : '#374151',
                fontFamily: "'IBM Plex Sans', sans-serif",
                fontWeight: isActive ? 600 : 400,
                fontSize: isMobile ? 12 : 14,
                cursor: 'pointer',
                textAlign: 'center',
                transition: 'background 150ms ease, color 150ms ease',
              }}
            >
              <Icon size={isMobile ? 17 : 16} strokeWidth={1.5} />
              {label}
            </button>
          )
        })}
      </div>

      {!isMobile && (
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
      )}
    </nav>
  )
}
