import { useEffect, useState } from 'react'
import * as backend from '../services/backend'
import type { ReminderSettings, ReminderStatus } from '../types'
import { builtinPlannerAnchors } from '../util/plannerDefaults'
import {
  enforceAnchorOrder,
  type PlannerAnchorTimes,
  validatePlannerAnchors,
} from '../util/plannerTimes'
import { defaultReminderSettings, normalizeReminderSettings } from '../util/reminderSettings'
import { Banner, Page, PageHeader } from '../components/ui'
import AccountSettingsForm from '../features/settings/AccountSettingsForm'
import PlannerDefaultsForm from '../features/settings/PlannerDefaultsForm'
import ReminderSettingsForm from '../features/settings/ReminderSettingsForm'

export default function SettingsPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [ttl, setTTL] = useState<string>('')
  const [anchors, setAnchors] = useState<PlannerAnchorTimes>(builtinPlannerAnchors)
  const [reminders, setReminders] = useState<ReminderSettings>(
    defaultReminderSettings,
  )
  const [reminderStatus, setReminderStatus] = useState<ReminderStatus | null>(null)
  const [msg, setMsg] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

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
        setErr(e instanceof Error ? e.message : String(e))
      }
    })()
  }, [])

  const applyAnchors = (next: PlannerAnchorTimes) => {
    setAnchors(enforceAnchorOrder(next))
  }

  const saveAccount = async () => {
    setMsg(null)
    setErr(null)
    if (!backend.hasWailsRuntime()) {
      setErr(
        'Configurações são gravadas pelo app desktop. Use pm-desktop aqui.',
      )
      return
    }
    if (!email.trim()) {
      setErr('Informe o e-mail.')
      return
    }
    const cache_ttl_hours = ttl.trim() === '' ? undefined : Number(ttl)
    if (
      ttl.trim() !== '' &&
      (!Number.isFinite(cache_ttl_hours) || Number(cache_ttl_hours) <= 0)
    ) {
      setErr('TTL deve ser um número positivo em horas.')
      return
    }
    setBusy(true)
    try {
      await backend.mergeAndSave({
        email,
        password,
        cache_ttl_hours,
      })
      setMsg('Conta salva no arquivo de configuração compartilhado.')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  const savePlanner = async () => {
    setMsg(null)
    setErr(null)
    if (!backend.hasWailsRuntime()) {
      setErr(
        'Configurações são gravadas pelo app desktop. Use pm-desktop aqui.',
      )
      return
    }
    const anchorErr = validatePlannerAnchors(anchors)
    if (anchorErr) {
      setErr(anchorErr)
      return
    }
    setBusy(true)
    try {
      await backend.mergeAndSave({
        planner: { ...anchors },
      })
      setMsg('Horários do planner salvos no arquivo de configuração compartilhado.')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
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
    setMsg(null)
    setErr(null)
    if (!backend.hasWailsRuntime()) {
      setErr(
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
      setErr('Ative pelo menos um canal de lembrete.')
      return
    }
    setBusy(true)
    try {
      await backend.saveReminderSettings(normalized)
      await refreshReminderStatus()
      setMsg('Lembretes salvos e daemon sincronizado.')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  const requestReminderPermission = async () => {
    setMsg(null)
    setErr(null)
    if (!backend.hasWailsRuntime()) {
      setErr('Permissão de notificações só funciona dentro do desktop.')
      return
    }
    setBusy(true)
    try {
      const status = await backend.requestNotificationPermission()
      await refreshReminderStatus()
      setMsg(
        status.authorized
          ? 'Notificações autorizadas.'
          : status.detail || 'Notificações ainda não autorizadas.',
      )
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  const testReminder = async () => {
    setMsg(null)
    setErr(null)
    if (!backend.hasWailsRuntime()) {
      setErr('Teste de lembrete só funciona dentro do desktop.')
      return
    }
    setBusy(true)
    try {
      await backend.sendTestReminder()
      setMsg('Lembrete de teste enviado.')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  const test = async () => {
    setMsg(null)
    setErr(null)
    if (!backend.hasWailsRuntime()) {
      setErr('Autenticação só funciona dentro do desktop.')
      return
    }
    if (!email.trim() || !password) {
      setErr('Informe e-mail e senha para testar.')
      return
    }
    setBusy(true)
    try {
      const problem = await backend.testAuth(email, password)
      if (problem) setErr(problem)
      else setMsg('Credenciais válidas.')
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
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
      {err ? <Banner tone="error">{err}</Banner> : null}
      {msg ? <Banner tone="success">{msg}</Banner> : null}

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
    </Page>
  )
}
