import { useCallback, useEffect, useState } from 'react'
import * as backend from '../services/backend'
import type { ReminderSettings, UpdateStatus } from '../types'
import {
  builtinPlannerAnchors,
  DEFAULT_BALANCE_CREDIT_MULTIPLIER,
  DEFAULT_MAX_DAILY_EXTRA_MINUTES,
} from '../util/plannerDefaults'
import {
  enforceAnchorOrder,
  type PlannerAnchorTimes,
  validatePlannerAnchors,
} from '../util/plannerTimes'
import {
  normalizeLoginCredential,
  validateLoginCredential,
} from '../util/loginCredential'
import { defaultReminderSettings, normalizeReminderSettings } from '../util/reminderSettings'
import { Banner, Page, PageHeader, Toast } from '../components/ui'
import AccountSettingsForm from '../features/settings/AccountSettingsForm'
import PlannerDefaultsForm from '../features/settings/PlannerDefaultsForm'
import ReminderSettingsForm from '../features/settings/ReminderSettingsForm'
import UpdateSettingsForm from '../features/settings/UpdateSettingsForm'

type SettingsToast = {
  tone: 'success' | 'error'
  message: string
}

const TOAST_AUTO_DISMISS_MS = 4200
export default function SettingsPage() {
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
    setToast({
      tone: 'error',
      message:
        errorOrMessage instanceof Error ? errorOrMessage.message : String(errorOrMessage),
    })
  }, [])

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
        const c = await backend.getConfig()
        const builtins = builtinPlannerAnchors()
        setLogin(c.login ?? '')
        setPassword(c.password ?? '')
        setMaxDailyExtraHours(
          String(
            (c.max_daily_extra_minutes ?? DEFAULT_MAX_DAILY_EXTRA_MINUTES) / 60,
          ),
        )
        setBalanceCreditMultiplier(
          String(c.balance_credit_multiplier ?? DEFAULT_BALANCE_CREDIT_MULTIPLIER),
        )
        if (c.planner) {
          setAnchors({
            in1: c.planner.in1 ?? builtins.in1,
            out1: c.planner.out1 ?? builtins.out1,
            in2: c.planner.in2 ?? builtins.in2,
            out2: c.planner.out2 ?? builtins.out2,
          })
        }
        setReminders(normalizeReminderSettings(c.reminders))
        if (backend.hasWailsRuntime()) {
          const status = await backend.getReminderStatus()
          setReminders(status.settings)
        }
      } catch (e) {
        showError(e)
      }
    })()
  }, [showError])

  // An update runs while the app is closed, so its outcome is only reported here.
  useEffect(() => {
    ;(async () => {
      try {
        const result = await backend.consumeUpdateResult()
        if (!result) return
        if (result.ok) showSuccess(result.message)
        else showError(result.message)
      } catch (e) {
        showError(e)
      }
    })()
  }, [showError, showSuccess])

  const applyAnchors = (next: PlannerAnchorTimes) => {
    setAnchors(enforceAnchorOrder(next))
  }

  const saveAccount = async () => {
    dismissToast()
    if (!backend.hasWailsRuntime()) {
      showError(
        'Configurações são gravadas pelo app desktop. Use pm-desktop aqui.',
      )
      return
    }
    const loginError = validateLoginCredential(login)
    if (loginError) {
      showError(loginError)
      return
    }
    setBusy(true)
    try {
      await backend.mergeAndSave({
        login: normalizeLoginCredential(login),
        password,
      })
      showSuccess('Conta salva no arquivo de configuração compartilhado.')
    } catch (e) {
      showError(e)
    } finally {
      setBusy(false)
    }
  }

  const savePlanner = async () => {
    dismissToast()
    if (!backend.hasWailsRuntime()) {
      showError(
        'Configurações são gravadas pelo app desktop. Use pm-desktop aqui.',
      )
      return
    }
    const anchorErr = validatePlannerAnchors(anchors)
    if (anchorErr) {
      showError(anchorErr)
      return
    }
    const maxDailyExtraMinutes = Math.round(Number(maxDailyExtraHours) * 60)
    if (
      !Number.isFinite(maxDailyExtraMinutes) ||
      maxDailyExtraMinutes < 1 ||
      maxDailyExtraMinutes > 24 * 60
    ) {
      showError('O limite diário deve estar entre 1 minuto e 24 horas.')
      return
    }
    const multiplier = Number(balanceCreditMultiplier)
    if (!Number.isFinite(multiplier) || multiplier < 1.0 || multiplier > 3.0) {
      showError('O multiplicador de crédito deve estar entre 1,0 e 3,0.')
      return
    }
    setBusy(true)
    try {
      await backend.mergeAndSave({
        planner: { ...anchors },
        max_daily_extra_minutes: maxDailyExtraMinutes,
        balance_credit_multiplier: multiplier,
      })
      showSuccess('Configurações do planner salvas no arquivo compartilhado.')
    } catch (e) {
      showError(e)
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
      showError(
        'Configurações são gravadas pelo app desktop. Use pm-desktop aqui.',
      )
      return
    }
    const normalized = normalizeReminderSettings(reminders)
    setBusy(true)
    try {
      await backend.saveReminderSettings(normalized)
      await refreshReminderSettings()
      showSuccess('Lembretes salvos e daemon sincronizado.')
    } catch (e) {
      showError(e)
    } finally {
      setBusy(false)
    }
  }

  const checkUpdate = async () => {
    dismissToast()
    setCheckingUpdate(true)
    try {
      setUpdateStatus(await backend.checkForUpdate())
    } catch (e) {
      showError(e)
    } finally {
      setCheckingUpdate(false)
    }
  }

  const startUpdate = async () => {
    dismissToast()
    setUpdating(true)
    try {
      await backend.startUpdate()
      showSuccess('Atualizando: o app será fechado e reaberto automaticamente.')
    } catch (e) {
      // The app stays open when the updater fails to start.
      showError(e)
      setUpdating(false)
    }
  }

  const test = async () => {
    dismissToast()
    if (!backend.hasWailsRuntime()) {
      showError('Autenticação só funciona dentro do desktop.')
      return
    }
    if (!login.trim() || !password) {
      showError('Informe login (e-mail ou CPF) e senha para testar.')
      return
    }
    const loginError = validateLoginCredential(login)
    if (loginError) {
      showError(loginError)
      return
    }
    setBusy(true)
    try {
      const problem = await backend.testAuth(
        normalizeLoginCredential(login),
        password,
      )
      if (problem) showError(problem)
      else showSuccess('Credenciais válidas.')
    } catch (e) {
      showError(e)
    } finally {
      setBusy(false)
    }
  }

  const devHint =
    backend.hasWailsRuntime() ?
      ''
    : 'Modo navegador: edição apenas local; gravar/API exige o shell Wails.'

  return (
    <Page narrow>
      <PageHeader
        title="Configurações"
        description={
          <>
            Usa o mesmo arquivo de configuração da CLI, resolvido por{' '}
            <code className="pill" translate="no">
              pm
            </code>{' '}
            em cada plataforma.
          </>
        }
      />

      {devHint ? <Banner>{devHint}</Banner> : null}

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
