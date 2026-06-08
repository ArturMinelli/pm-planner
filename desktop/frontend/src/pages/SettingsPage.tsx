import { useCallback, useEffect, useState } from 'react'
import * as backend from '../services/backend'
import type { ReminderSettings, ReminderStatus } from '../types'
import { builtinPlannerAnchors } from '../util/plannerDefaults'
import {
  enforceAnchorOrder,
  type PlannerAnchorTimes,
  validatePlannerAnchors,
} from '../util/plannerTimes'
import { defaultReminderSettings, normalizeReminderSettings } from '../util/reminderSettings'
import { Banner, Page, PageHeader, Toast } from '../components/ui'
import AccountSettingsForm from '../features/settings/AccountSettingsForm'
import PlannerDefaultsForm from '../features/settings/PlannerDefaultsForm'
import ReminderSettingsForm from '../features/settings/ReminderSettingsForm'

type SettingsToast = {
  tone: 'success' | 'error'
  message: string
}

const TOAST_AUTO_DISMISS_MS = 4200

export default function SettingsPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [ttl, setTTL] = useState<string>('')
  const [anchors, setAnchors] = useState<PlannerAnchorTimes>(builtinPlannerAnchors)
  const [reminders, setReminders] = useState<ReminderSettings>(
    defaultReminderSettings,
  )
  const [reminderStatus, setReminderStatus] = useState<ReminderStatus | null>(null)
  const [toast, setToast] = useState<SettingsToast | null>(null)
  const [busy, setBusy] = useState(false)

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
        setEmail(c.email ?? '')
        setPassword(c.password ?? '')
        if (c.cache_ttl_hours && c.cache_ttl_hours > 0) {
          setTTL(String(c.cache_ttl_hours))
        }
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
          setReminderStatus(status)
          setReminders(status.settings)
        }
      } catch (e) {
        showError(e)
      }
    })()
  }, [showError])

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
    if (!email.trim()) {
      showError('Informe o e-mail.')
      return
    }
    const cache_ttl_hours = ttl.trim() === '' ? undefined : Number(ttl)
    if (
      ttl.trim() !== '' &&
      (!Number.isFinite(cache_ttl_hours) || Number(cache_ttl_hours) <= 0)
    ) {
      showError('TTL deve ser um número positivo em horas.')
      return
    }
    setBusy(true)
    try {
      await backend.mergeAndSave({
        email,
        password,
        cache_ttl_hours,
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
    setBusy(true)
    try {
      await backend.mergeAndSave({
        planner: { ...anchors },
      })
      showSuccess('Horários do planner salvos no arquivo de configuração compartilhado.')
    } catch (e) {
      showError(e)
    } finally {
      setBusy(false)
    }
  }

  const refreshReminderStatus = async () => {
    if (!backend.hasWailsRuntime()) return
    const status = await backend.getReminderStatus()
    setReminderStatus(status)
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
    if (
      normalized.enabled &&
      !normalized.native_notifications &&
      !normalized.popup_notifications
    ) {
      showError('Ative pelo menos um canal de lembrete.')
      return
    }
    setBusy(true)
    try {
      await backend.saveReminderSettings(normalized)
      await refreshReminderStatus()
      showSuccess('Lembretes salvos e daemon sincronizado.')
    } catch (e) {
      showError(e)
    } finally {
      setBusy(false)
    }
  }

  const requestReminderPermission = async () => {
    dismissToast()
    if (!backend.hasWailsRuntime()) {
      showError('Permissão de notificações só funciona dentro do desktop.')
      return
    }
    setBusy(true)
    try {
      const status = await backend.requestNotificationPermission()
      await refreshReminderStatus()
      if (status.authorized) {
        showSuccess('Notificações autorizadas.')
      } else {
        showError(status.detail || 'Notificações ainda não autorizadas.')
      }
    } catch (e) {
      showError(e)
    } finally {
      setBusy(false)
    }
  }

  const testReminder = async () => {
    dismissToast()
    if (!backend.hasWailsRuntime()) {
      showError('Teste de lembrete só funciona dentro do desktop.')
      return
    }
    setBusy(true)
    try {
      await backend.sendTestReminder()
      showSuccess('Lembrete de teste enviado.')
    } catch (e) {
      showError(e)
    } finally {
      setBusy(false)
    }
  }

  const test = async () => {
    dismissToast()
    if (!backend.hasWailsRuntime()) {
      showError('Autenticação só funciona dentro do desktop.')
      return
    }
    if (!email.trim() || !password) {
      showError('Informe e-mail e senha para testar.')
      return
    }
    setBusy(true)
    try {
      const problem = await backend.testAuth(email, password)
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
        email={email}
        password={password}
        ttl={ttl}
        busy={busy}
        canTest={backend.hasWailsRuntime()}
        onEmailChange={setEmail}
        onPasswordChange={setPassword}
        onTTLChange={setTTL}
        onSave={() => void saveAccount()}
        onTest={() => void test()}
      />

      <PlannerDefaultsForm
        anchors={anchors}
        busy={busy}
        onAnchorsChange={applyAnchors}
        onSave={() => void savePlanner()}
      />

      <ReminderSettingsForm
        settings={reminders}
        status={reminderStatus}
        busy={busy}
        canUseRuntime={backend.hasWailsRuntime()}
        onSettingsChange={setReminders}
        onSave={() => void saveReminders()}
        onRequestPermission={() => void requestReminderPermission()}
        onTest={() => void testReminder()}
      />

      {toast ? (
        <Toast tone={toast.tone} onDismiss={dismissToast}>
          {toast.message}
        </Toast>
      ) : null}
    </Page>
  )
}
