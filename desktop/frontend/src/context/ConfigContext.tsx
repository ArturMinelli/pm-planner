import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import * as backend from '../services/backend'
import { notifyConfigChanged } from '../services/browserReminders'
import type { PmConfig, ReminderSettings } from '../types'
import { normalizeReminderSettings } from '../util/reminderSettings'

export type ConfigContextValue = {
  config: PmConfig | null
  loading: boolean
  configRevision: number
  reloadConfig: () => Promise<void>
  mergeAndSave: (patch: Partial<PmConfig>) => Promise<void>
  saveReminderSettings: (settings: ReminderSettings) => Promise<void>
}

const ConfigContext = createContext<ConfigContextValue | null>(null)

export function ConfigProvider({ children }: { children: ReactNode }) {
  const [config, setConfig] = useState<PmConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [configRevision, setConfigRevision] = useState(0)

  const reloadConfig = useCallback(async () => {
    const loaded = await backend.getConfig()
    setConfig(loaded)
    setConfigRevision((revision) => revision + 1)
  }, [])

  useEffect(() => {
    let cancelled = false

    const loadInitialConfig = async () => {
      setLoading(true)
      try {
        const loaded = await backend.getConfig()
        if (cancelled) return
        setConfig(loaded)
        setConfigRevision(1)
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadInitialConfig()

    return () => {
      cancelled = true
    }
  }, [])

  const mergeAndSave = useCallback(async (patch: Partial<PmConfig>) => {
    const current = config ?? (await backend.getConfig())
    const merged: PmConfig = { ...current, ...patch }
    if (patch.reminders) {
      merged.reminders = normalizeReminderSettings(patch.reminders)
    }
    await backend.saveConfig(merged)
    setConfig(merged)
    setConfigRevision((revision) => revision + 1)
    notifyConfigChanged()
  }, [config])

  const saveReminderSettings = useCallback(async (settings: ReminderSettings) => {
    const normalized = normalizeReminderSettings(settings)
    await backend.saveReminderSettings(normalized)
    const loaded = await backend.getConfig()
    setConfig(loaded)
    setConfigRevision((revision) => revision + 1)
    notifyConfigChanged()
  }, [])

  const value = useMemo(
    () => ({
      config,
      loading,
      configRevision,
      reloadConfig,
      mergeAndSave,
      saveReminderSettings,
    }),
    [config, loading, configRevision, reloadConfig, mergeAndSave, saveReminderSettings],
  )

  return <ConfigContext.Provider value={value}>{children}</ConfigContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components -- hook is co-located with its provider
export function useConfig(): ConfigContextValue {
  const context = useContext(ConfigContext)
  if (!context) {
    throw new Error('useConfig must be used within ConfigProvider')
  }
  return context
}
