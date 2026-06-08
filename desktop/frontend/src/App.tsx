import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import AppShell from './components/AppShell'
import ReminderOverlay from './features/reminders/ReminderOverlay'
import PlannerPage from './pages/PlannerPage'
import SettingsPage from './pages/SettingsPage'
import { hasOverlayRuntime } from './services/overlay'

export default function App() {
  if (hasOverlayRuntime()) {
    return <ReminderOverlay />
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
