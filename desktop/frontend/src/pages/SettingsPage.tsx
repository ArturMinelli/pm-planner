import { useEffect, useState } from 'react'
import * as backend from '../services/backend'
import { builtinPlannerAnchors } from '../util/plannerDefaults'
import {
  enforceAnchorOrder,
  type PlannerAnchorTimes,
  validatePlannerAnchors,
} from '../util/plannerTimes'
import { Banner, Page, PageHeader } from '../components/ui'
import AccountSettingsForm from '../features/settings/AccountSettingsForm'
import PlannerDefaultsForm from '../features/settings/PlannerDefaultsForm'

export default function SettingsPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [ttl, setTTL] = useState<string>('')
  const [anchors, setAnchors] = useState<PlannerAnchorTimes>(builtinPlannerAnchors)
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
    </Page>
  )
}
