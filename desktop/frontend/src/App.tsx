import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import AppShell from './components/AppShell'
import ReminderOverlay from './features/reminders/ReminderOverlay'
import PlannerPage from './pages/PlannerPage'
import SettingsPage from './pages/SettingsPage'
import { hasOverlayRuntime } from './services/overlay'

function hasOverlayPreview() {
  return (
    import.meta.env.DEV &&
    new URLSearchParams(globalThis.location.search).get('overlay') === 'preview'
  )
}

export default function App() {
  const overlayRuntime = hasOverlayRuntime()
  const overlayPreview = hasOverlayPreview()

  if (overlayRuntime || overlayPreview) {
    return <ReminderOverlay preview={overlayPreview && !overlayRuntime} />
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppShell />}>
          <Route index element={<PlannerPage />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
