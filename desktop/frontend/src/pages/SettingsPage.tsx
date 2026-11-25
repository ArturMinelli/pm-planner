import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Trans, useTranslation } from 'react-i18next'
import * as backend from '../services/backend'
import { useConfig } from '../context/ConfigContext'
import type { PmConfig, ReminderSettings, UpdateStatus } from '../types'
import {
  builtinPlannerAnchors,
  DEFAULT_BALANCE_CREDIT_MULTIPLIER,
} from '../util/plannerDefaults'
import {
  anchorLabelKey,
  anchorLabelParams,
  enforceAnchorOrder,
  type PlannerAnchorTimes,
  validatePlannerAnchors,
} from '../util/plannerTimes'
import {
  normalizeLoginCredential,
  validateLoginCredential,
} from '../util/loginCredential'
import { defaultReminderSettings, normalizeReminderSettings } from '../util/reminderSettings'
import { translateMessage } from '../i18n/translateMessage'
import {
  applyTheme,
  isProtectedPreset,
  PRESETS,
  PROTECTED_PRESET_IDS,
  presetIdForColors,
  resolveEffectiveColors,
  resolvePresets,
  slugifyThemeId,
  visiblePresets,
} from '../util/theme'
import { Page, PageHeader, Toast } from '../components/ui'
import AccountSettingsForm from '../features/settings/AccountSettingsForm'
import PlannerDefaultsForm from '../features/settings/PlannerDefaultsForm'
import ReminderSettingsForm from '../features/settings/ReminderSettingsForm'
import ThemeSettingsForm from '../features/settings/ThemeSettingsForm'
import UpdateSettingsForm from '../features/settings/UpdateSettingsForm'

type SettingsToast = {
  tone: 'success' | 'error'
  message: string
}

const TOAST_AUTO_DISMISS_MS = 4200

export default function SettingsPage() {
  const { t } = useTranslation()
  const { config, mergeAndSave, saveReminderSettings: persistReminderSettings } = useConfig()
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [balanceCreditMultiplier, setBalanceCreditMultiplier] = useState(
    String(DEFAULT_BALANCE_CREDIT_MULTIPLIER),
  )
  const [anchors, setAnchors] = useState<PlannerAnchorTimes>(builtinPlannerAnchors)
  const [reminders, setReminders] = useState<ReminderSettings>(
    defaultReminderSettings,
  )
  const [themeDraft, setThemeDraft] = useState<Record<string, string>>({})
  const presets = useMemo(() => resolvePresets(config?.themes), [config?.themes])
  const hiddenThemes = useMemo(() => config?.hidden_themes ?? [], [config?.hidden_themes])
  // Removed themes stay resolvable by id; they only disappear from the picker.
  // Claro and Escuro are never hidden, so a stale entry for them is dropped on the next save.
  const withoutProtected = (ids: string[]) => ids.filter((hidden) => !PROTECTED_PRESET_IDS.includes(hidden))
  const shownPresets = useMemo(() => visiblePresets(presets, hiddenThemes), [presets, hiddenThemes])
  const selectedId = presetIdForColors(presets, themeDraft)
  const configRef = useRef(config)
  const [toast, setToast] = useState<SettingsToast | null>(null)
  const [busy, setBusy] = useState(false)
  const [updateStatus, setUpdateStatus] = useState<UpdateStatus | null>(null)
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [checkError, setCheckError] = useState<string | null>(null)
  const [updating, setUpdating] = useState(false)
  const usesDesktopShell = backend.hasWailsRuntime()

  const dismissToast = useCallback(() => {
    setToast(null)
  }, [])

  const showSuccess = useCallback((message: string) => {
    setToast({ tone: 'success', message })
  }, [])

  const showError = useCallback((errorOrMessage: unknown) => {
    if (typeof errorOrMessage === 'object' && errorOrMessage && 'key' in errorOrMessage) {
      setToast({
        tone: 'error',
        message: translateMessage(t, errorOrMessage as { key: string; params?: Record<string, string> }) ?? t('errors.generic'),
      })
      return
    }
    const message = errorOrMessage instanceof Error ? errorOrMessage.message : String(errorOrMessage)
    setToast({
      tone: 'error',
      message: message.startsWith('errors.') ? t(message) : message,
    })
  }, [t])

  useEffect(() => {
    if (!toast) return
    const timeout = window.setTimeout(() => {
      setToast(null)
    }, TOAST_AUTO_DISMISS_MS)
    return () => window.clearTimeout(timeout)
  }, [toast])

  useEffect(() => {
    if (!config) return

    const builtins = builtinPlannerAnchors()
    // Hydrate editable form fields from the shared config store.
    // eslint-disable-next-line react-hooks/set-state-in-effect -- settings forms are local drafts until Save
    setLogin(config.login ?? '')
    setPassword(config.password ?? '')
    setBalanceCreditMultiplier(
      String(config.balance_credit_multiplier ?? DEFAULT_BALANCE_CREDIT_MULTIPLIER),
    )
    if (config.planner) {
      setAnchors({
        in1: config.planner.in1 ?? builtins.in1,
        out1: config.planner.out1 ?? builtins.out1,
        in2: config.planner.in2 ?? builtins.in2,
        out2: config.planner.out2 ?? builtins.out2,
      })
    }
    setReminders(normalizeReminderSettings(config.reminders))
    setThemeDraft(resolveEffectiveColors(config) ?? {})

    void (async () => {
      try {
        const status = await backend.getReminderStatus()
        setReminders(status.settings)
      } catch (error) {
        showError(error)
      }
    })()
  }, [config, showError])

  useEffect(() => {
    configRef.current = config
  }, [config])

  // Drop unsaved live previews when leaving the page.
  useEffect(() => {
    return () => applyTheme(resolveEffectiveColors(configRef.current ?? {}))
  }, [])

  useEffect(() => {
    if (!usesDesktopShell) return
    ;(async () => {
      try {
        const result = await backend.consumeUpdateResult()
        if (!result) return
        const message = translateMessage(t, result.message)
        if (result.ok) showSuccess(message ?? t('update.result.success'))
        else showError(message ?? t('errors.generic'))
      } catch (error) {
        showError(error)
      }
    })()
  }, [showError, showSuccess, t, usesDesktopShell])

  const applyAnchors = (next: PlannerAnchorTimes) => {
    setAnchors(enforceAnchorOrder(next))
  }

  const saveAccount = async () => {
    dismissToast()
    const loginErrorKey = validateLoginCredential(login)
    if (loginErrorKey) {
      showError(t(loginErrorKey))
      return
    }
    setBusy(true)
    try {
      await mergeAndSave({
        login: normalizeLoginCredential(login),
        password,
      })
      showSuccess(t('settings.account.saved'))
    } catch (error) {
      showError(error)
    } finally {
      setBusy(false)
    }
  }

  const savePlanner = async () => {
    dismissToast()
    const anchorError = validatePlannerAnchors(anchors)
    if (anchorError) {
      if (anchorError.key === 'errors.validation.invalidAnchorTime') {
        const fieldIndex = Number(anchorError.params.label) - 1
        showError(
          t(anchorError.key, {
            label: t(anchorLabelKey(fieldIndex), anchorLabelParams(fieldIndex)),
          }),
        )
      } else if ('params' in anchorError) {
        showError(t(anchorError.key, anchorError.params))
      } else {
        showError(t(anchorError.key))
      }
      return
    }
    const multiplier = Number(balanceCreditMultiplier)
    if (!Number.isFinite(multiplier) || multiplier < 1.0 || multiplier > 3.0) {
      showError(t('settings.planner.multiplierError'))
      return
    }
    setBusy(true)
    try {
      await mergeAndSave({
        planner: { ...anchors },
        balance_credit_multiplier: multiplier,
      })
      showSuccess(t('settings.planner.saved'))
    } catch (error) {
      showError(error)
    } finally {
      setBusy(false)
    }
  }

  const refreshReminderSettings = async () => {
    const status = await backend.getReminderStatus()
    setReminders(status.settings)
  }

  const saveReminders = async () => {
    dismissToast()
    const normalized = normalizeReminderSettings(reminders)
    setBusy(true)
    try {
      await persistReminderSettings(normalized)
      await refreshReminderSettings()
      showSuccess(t('settings.reminders.saved'))
    } catch (error) {
      showError(error)
    } finally {
      setBusy(false)
    }
  }

  const previewTheme = (colors: Record<string, string>) => {
    setThemeDraft(colors)
    applyTheme(colors)
  }

  const selectPreset = (id: string) => {
    previewTheme(presets.find((preset) => preset.id === id)?.colors ?? {})
  }

  const fallbackColors = (themes: PmConfig['themes'] = config?.themes) =>
    resolveEffectiveColors({ default_theme: config?.default_theme, themes }) ?? {}

  const saveTheme = async (
    patch: Partial<PmConfig>,
    next?: Record<string, string>,
    doneKey = 'settings.theme.saved',
  ) => {
    dismissToast()
    setBusy(true)
    try {
      await mergeAndSave(patch)
      if (next) previewTheme(next)
      showSuccess(t(doneKey))
    } catch (error) {
      showError(error)
    } finally {
      setBusy(false)
    }
  }

  // With a preset selected, store only its id so later edits to the preset apply.
  const persistTheme = () =>
    saveTheme(selectedId ? { theme: {}, theme_name: selectedId } : { theme: themeDraft, theme_name: '' })

  const resetTheme = () => saveTheme({ theme: {}, theme_name: '' }, fallbackColors())

  const saveAsPreset = (name: string, colors: Record<string, string>) => {
    const id = slugifyThemeId(name)
    if (!id) return
    void saveTheme({
      themes: { ...(config?.themes ?? {}), [id]: { label: name.trim(), colors } },
      hidden_themes: withoutProtected(hiddenThemes.filter((hidden) => hidden !== id)),
      theme: {},
      theme_name: id,
    })
  }

  const deletePreset = (id: string) => {
    const target = presets.find((preset) => preset.id === id)
    if (target && isProtectedPreset(target)) return
    const themes = { ...(config?.themes ?? {}) }
    delete themes[id]
    // Built-ins live in code, so removing one means hiding it from the picker.
    const hidden = withoutProtected(
      PRESETS.some((preset) => preset.id === id) ? [...new Set([...hiddenThemes, id])] : hiddenThemes,
    )
    const selectedGone = config?.theme_name === id
    void saveTheme(
      selectedGone ? { themes, hidden_themes: hidden, theme_name: '' } : { themes, hidden_themes: hidden },
      selectedGone ? fallbackColors(themes) : undefined,
      'settings.theme.deleted',
    )
  }

  const checkUpdate = useCallback(async () => {
    setCheckingUpdate(true)
    setCheckError(null)
    try {
      setUpdateStatus(await backend.checkForUpdate())
    } catch (error) {
      if (typeof error === 'object' && error && 'key' in error) {
        const message =
          translateMessage(t, error as { key: string; params?: Record<string, string> }) ??
          t('errors.generic')
        setCheckError(message)
      } else {
        const message = error instanceof Error ? error.message : String(error)
        setCheckError(message.startsWith('errors.') ? t(message) : message)
      }
      showError(error)
    } finally {
      setCheckingUpdate(false)
    }
  }, [showError, t])

  useEffect(() => {
    void checkUpdate()
  }, [checkUpdate])

  const startUpdate = async () => {
    dismissToast()
    setUpdating(true)
    try {
      await backend.startUpdate()
      showSuccess(t('settings.updates.starting'))
    } catch (error) {
      showError(error)
      setUpdating(false)
    }
  }

  const test = async () => {
    dismissToast()
    if (!login.trim() || !password) {
      showError(t('settings.account.testMissing'))
      return
    }
    const loginErrorKey = validateLoginCredential(login)
    if (loginErrorKey) {
      showError(t(loginErrorKey))
      return
    }
    setBusy(true)
    try {
      const result = await backend.testAuth(
        normalizeLoginCredential(login),
        password,
      )
      if (result.error) {
        showError(translateMessage(t, result.error) ?? t('errors.generic'))
      } else {
        showSuccess(t('settings.account.testSuccess'))
      }
    } catch (error) {
      showError(error)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Page>
      <PageHeader
        title={t('settings.title')}
        description={
          <Trans i18nKey="settings.description" components={{ code: <code className="pill" translate="no" /> }} />
        }
      />

      <div className="settings-grid">
        <AccountSettingsForm
          login={login}
          password={password}
          busy={busy}
          onLoginChange={setLogin}
          onPasswordChange={setPassword}
          onSave={() => void saveAccount()}
          onTest={() => void test()}
        />

        <PlannerDefaultsForm
          anchors={anchors}
          balanceCreditMultiplier={balanceCreditMultiplier}
          busy={busy}
          onAnchorsChange={applyAnchors}
          onBalanceCreditMultiplierChange={setBalanceCreditMultiplier}
          onSave={() => void savePlanner()}
        />

        <ReminderSettingsForm
          settings={reminders}
          busy={busy}
          onSettingsChange={setReminders}
          onSave={() => void saveReminders()}
        />

        <ThemeSettingsForm
          colors={themeDraft}
          presets={shownPresets}
          selectedId={selectedId}
          busy={busy}
          onChange={previewTheme}
          onSelectPreset={selectPreset}
          onSaveAs={saveAsPreset}
          onDeletePreset={deletePreset}
          onSave={() => void persistTheme()}
          onReset={() => void resetTheme()}
        />
      </div>

      <UpdateSettingsForm
        status={updateStatus}
        checking={checkingUpdate}
        checkError={checkError}
        updating={updating}
        canApplyUpdate={usesDesktopShell}
        onUpdate={() => void startUpdate()}
      />

      {toast ? (
        <Toast tone={toast.tone} onDismiss={dismissToast}>
          {toast.message}
        </Toast>
      ) : null}
    </Page>
  )
}
