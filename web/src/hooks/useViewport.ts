import { useEffect, useState } from 'react'

function getViewportWidth() {
  if (typeof window === 'undefined') return 1280
  return window.innerWidth
}

export function useViewport() {
  const [width, setWidth] = useState(getViewportWidth)

  useEffect(() => {
    const onResize = () => setWidth(window.innerWidth)
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  return {
    width,
    isMobile: width <= 768,
    isNarrow: width <= 960,
  }
}
