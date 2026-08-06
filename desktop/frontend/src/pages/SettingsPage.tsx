import { useCallback, useEffect, useState } from 'react'
import { Trans, useTranslation } from 'react-i18next'
import * as backend from '../services/backend'
import type { AppLocale, ReminderSettings, UpdateStatus } from '../types'
import {
  builtinPlannerAnchors,
  DEFAULT_BALANCE_CREDIT_MULTIPLIER,
  DEFAULT_MAX_DAILY_EXTRA_MINUTES,
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
import { defaultLocale, isAppLocale } from '../i18n'
import { Banner, Page, PageHeader, Toast } from '../components/ui'
import AccountSettingsForm from '../features/settings/AccountSettingsForm'
import LanguageSettingsForm from '../features/settings/LanguageSettingsForm'
import PlannerDefaultsForm from '../features/settings/PlannerDefaultsForm'
import ReminderSettingsForm from '../features/settings/ReminderSettingsForm'
import UpdateSettingsForm from '../features/settings/UpdateSettingsForm'

type SettingsToast = {
  tone: 'success' | 'error'
  message: string
}

const TOAST_AUTO_DISMISS_MS = 4200

export default function SettingsPage() {
  const { t } = useTranslation()
  const [locale, setLocale] = useState<AppLocale>(defaultLocale)
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [maxDailyExtraHours, setMaxDailyExtraHours] = useState(
    String(DEFAULT_MAX_DAILY_EXTRA_MINUTES / 60),
  )
  const [balanceCreditMultiplier, setBalanceCreditMultiplier] = useState(
    String(DEFAULT_BALANCE_CREDIT_MULTIPLIER),
  )
  const [anchors, setAnchors] = useState<PlannerAnchorTimes>(builtinPlannerAnchors)
  const [reminders, setReminders] = useState<ReminderSettings>(
    defaultReminderSettings,
  )
  const [toast, setToast] = useState<SettingsToast | null>(null)
  const [busy, setBusy] = useState(false)
  const [updateStatus, setUpdateStatus] = useState<UpdateStatus | null>(null)
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [updating, setUpdating] = useState(false)

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
    ;(async () => {
      try {
        const config = await backend.getConfig()
        const builtins = builtinPlannerAnchors()
        if (config.locale && isAppLocale(config.locale)) {
          setLocale(config.locale)
        }
        setLogin(config.login ?? '')
        setPassword(config.password ?? '')
        setMaxDailyExtraHours(
          String(
            (config.max_daily_extra_minutes ?? DEFAULT_MAX_DAILY_EXTRA_MINUTES) / 60,
          ),
        )
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
        if (backend.hasWailsRuntime()) {
          const status = await backend.getReminderStatus()
          setReminders(status.settings)
        }
      } catch (error) {
        showError(error)
      }
    })()
  }, [showError])

  useEffect(() => {
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
  }, [showError, showSuccess, t])

  const applyAnchors = (next: PlannerAnchorTimes) => {
    setAnchors(enforceAnchorOrder(next))
  }

  const saveAccount = async () => {
    dismissToast()
    if (!backend.hasWailsRuntime()) {
      showError(t('common.desktopOnlyConfig'))
      return
    }
    const loginErrorKey = validateLoginCredential(login)
    if (loginErrorKey) {
      showError(t(loginErrorKey))
      return
    }
    setBusy(true)
    try {
      await backend.mergeAndSave({
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
    if (!backend.hasWailsRuntime()) {
      showError(t('common.desktopOnlyConfig'))
      return
    }
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
    const maxDailyExtraMinutes = Math.round(Number(maxDailyExtraHours) * 60)
    if (
      !Number.isFinite(maxDailyExtraMinutes) ||
      maxDailyExtraMinutes < 1 ||
      maxDailyExtraMinutes > 24 * 60
    ) {
      showError(t('settings.planner.dailyLimitError'))
      return
    }
    const multiplier = Number(balanceCreditMultiplier)
    if (!Number.isFinite(multiplier) || multiplier < 1.0 || multiplier > 3.0) {
      showError(t('settings.planner.multiplierError'))
      return
    }
    setBusy(true)
    try {
      await backend.mergeAndSave({
        planner: { ...anchors },
        max_daily_extra_minutes: maxDailyExtraMinutes,
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
    if (!backend.hasWailsRuntime()) return
    const status = await backend.getReminderStatus()
    setReminders(status.settings)
  }

  const saveReminders = async () => {
    dismissToast()
    if (!backend.hasWailsRuntime()) {
      showError(t('common.desktopOnlyConfig'))
      return
    }
    const normalized = normalizeReminderSettings(reminders)
    setBusy(true)
    try {
      await backend.saveReminderSettings(normalized)
      await refreshReminderSettings()
      showSuccess(t('settings.reminders.saved'))
    } catch (error) {
      showError(error)
    } finally {
      setBusy(false)
    }
  }

  const checkUpdate = async () => {
    dismissToast()
    setCheckingUpdate(true)
    try {
      setUpdateStatus(await backend.checkForUpdate())
    } catch (error) {
      showError(error)
    } finally {
      setCheckingUpdate(false)
    }
  }

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
    if (!backend.hasWailsRuntime()) {
      showError(t('common.desktopOnlyAuth'))
      return
    }
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

  const devHint = backend.hasWailsRuntime() ? '' : t('settings.browserHint')

  return (
    <Page>
      <PageHeader
        title={t('settings.title')}
        description={
          <Trans i18nKey="settings.description" components={{ code: <code className="pill" translate="no" /> }} />
        }
      />

      {devHint ? <Banner>{devHint}</Banner> : null}

      <div className="settings-grid">
        <LanguageSettingsForm
          locale={locale}
          busy={busy}
          onLocaleChange={setLocale}
          onSaved={() => showSuccess(t('settings.language.saved'))}
          onError={showError}
        />

        <AccountSettingsForm
          login={login}
          password={password}
          busy={busy}
          canTest={backend.hasWailsRuntime()}
          onLoginChange={setLogin}
          onPasswordChange={setPassword}
          onSave={() => void saveAccount()}
          onTest={() => void test()}
        />

        <PlannerDefaultsForm
          anchors={anchors}
          maxDailyExtraHours={maxDailyExtraHours}
          balanceCreditMultiplier={balanceCreditMultiplier}
          busy={busy}
          onAnchorsChange={applyAnchors}
          onMaxDailyExtraHoursChange={setMaxDailyExtraHours}
          onBalanceCreditMultiplierChange={setBalanceCreditMultiplier}
          onSave={() => void savePlanner()}
        />

        <ReminderSettingsForm
          settings={reminders}
          busy={busy}
          canUseRuntime={backend.hasWailsRuntime()}
          onSettingsChange={setReminders}
          onSave={() => void saveReminders()}
        />
      </div>

      <UpdateSettingsForm
        status={updateStatus}
        checking={checkingUpdate}
        updating={updating}
        canUseRuntime={backend.hasWailsRuntime()}
        onCheck={() => void checkUpdate()}
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
