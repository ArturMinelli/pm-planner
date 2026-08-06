import { StrictMode, useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { defaultLocale, initI18n, isAppLocale } from './i18n'
import * as backend from './services/backend'

function Bootstrap() {
  const [ready, setReady] = useState(false)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      let locale = defaultLocale
      try {
        const config = await backend.getConfig()
        if (config.locale && isAppLocale(config.locale)) {
          locale = config.locale
        }
      } catch {
        // fall back to default locale
      }
      await initI18n(locale)
      if (!cancelled) setReady(true)
    })()
    return () => {
      cancelled = true
    }
  }, [])

  if (!ready) return null
  return <App />
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Bootstrap />
  </StrictMode>,
)
