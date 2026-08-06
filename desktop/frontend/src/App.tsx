import { useEffect } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import AppShell from './components/AppShell'
import PlannerPage from './pages/PlannerPage'
import SettingsPage from './pages/SettingsPage'
import {
  startBrowserReminderScheduler,
  stopBrowserReminderScheduler,
} from './services/browserReminders'

export default function App() {
  useEffect(() => {
    startBrowserReminderScheduler()
    return () => stopBrowserReminderScheduler()
  }, [])

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
