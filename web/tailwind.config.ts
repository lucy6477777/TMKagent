import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        'bg-page': '#F8F9FA',
        'bg-sidebar': '#FFFFFF',
        'nav-active': '#1E3A5F',
        'nav-text': '#374151',
        'tag-src': '#64748B',
        'tag-tgt': '#1E3A5F',
        'btn-primary': '#1E3A5F',
        'btn-hover': '#2D5A8E',
        'status-live': '#10B981',
        'status-idle': '#9CA3AF',
        'brand': '#1E3A5F',
      },
      fontFamily: {
        sans: ['IBM Plex Sans', 'system-ui', 'sans-serif'],
        mono: ['IBM Plex Mono', 'Menlo', 'monospace'],
      },
    },
  },
  plugins: [],
} satisfies Config
