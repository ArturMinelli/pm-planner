import { useEffect, useState } from 'react'
import PlannerTimeInput from '../components/PlannerTimeInput'
import * as backend from '../services/backend'
import { builtinPlannerAnchors } from '../util/plannerDefaults'
import {
  adjustPlannerAnchor,
  enforceAnchorOrder,
  type PlannerAnchorTimes,
  validatePlannerAnchors,
} from '../util/plannerTimes'
import {
  Banner,
  Button,
  Card,
  Field,
  Page,
  PageHeader,
  Stack,
} from '../components/ui'

type AnchorField = keyof PlannerAnchorTimes

const ANCHOR_FIELDS: { field: AnchorField; label: string }[] = [
  { field: 'in1', label: 'Entrada 1' },
  { field: 'out1', label: 'Saída 1' },
  { field: 'in2', label: 'Entrada 2' },
  { field: 'out2', label: 'Saída 2' },
]

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
        setEmail(c.email ?? '')
        setPassword(c.password ?? '')
        if (c.cache_ttl_hours && c.cache_ttl_hours > 0) {
          setTTL(String(c.cache_ttl_hours))
        }
        if (c.planner) {
          setAnchors({
            in1: c.planner.in1 ?? builtinPlannerAnchors().in1,
            out1: c.planner.out1 ?? builtinPlannerAnchors().out1,
            in2: c.planner.in2 ?? builtinPlannerAnchors().in2,
            out2: c.planner.out2 ?? builtinPlannerAnchors().out2,
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

  const builtins = builtinPlannerAnchors()

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

      <Card title="Conta e Sessão">
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            void saveAccount()
          }}
        >
          <Stack>
            <Field id="settings-email" label="E-mail de Login">
            <input
              id="settings-email"
              name="email"
              type="email"
              autoComplete="username email"
              spellCheck={false}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            </Field>
            <Field id="settings-password" label="Senha">
            <input
              id="settings-password"
              name="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            </Field>
            <Field
              id="settings-cache-ttl"
              label="Cache TTL em Horas"
              hint={
                <>
                  Quando a API não envia validade da sessão, limita há quantas
                  horas o arquivo{' '}
                  <code className="pill" translate="no">
                    session.json
                  </code>{' '}
                  é reutilizado.
                </>
              }
            >
            <input
              id="settings-cache-ttl"
              name="cache_ttl_hours"
              type="number"
              min="1"
              inputMode="numeric"
              autoComplete="off"
              placeholder="8…"
              value={ttl}
              onChange={(e) => setTTL(e.target.value)}
            />
            </Field>
          </Stack>
          <div className="btn-row">
          <Button
            type="submit"
            variant="primary"
            disabled={busy}
            aria-busy={busy}
          >
            {busy ? 'Salvando…' : 'Salvar Conta'}
          </Button>
          <Button
            type="button"
            variant="secondary"
            onClick={() => void test()}
            disabled={busy || !backend.hasWailsRuntime()}
            aria-busy={busy}
          >
            Testar Login
          </Button>
          </div>
        </form>
      </Card>

      <Card
        title="Horários Padrão do Planner"
        intro={
          <>
            Âncoras usadas para atribuir batidas aos quatro campos e preencher
            Entrada 1 quando não há registro correspondente. A CLI{' '}
            <code className="pill" translate="no">
              pm plan
            </code>{' '}
            usa os mesmos valores.
          </>
        }
      >
        <form
          className="form-stack"
          onSubmit={(event) => {
            event.preventDefault()
            void savePlanner()
          }}
        >
          <div className="planner-anchor-grid">
          {ANCHOR_FIELDS.map(({ field, label }) => (
            <Field
              key={field}
              id={`settings-anchor-${field}`}
              label={label}
            >
              <PlannerTimeInput
                id={`settings-anchor-${field}`}
                name={`planner_${field}`}
                value={anchors[field]}
                onChange={(value) =>
                  applyAnchors({ ...anchors, [field]: value })
                }
                onStep={(delta) =>
                  applyAnchors(adjustPlannerAnchor(anchors, field, delta))
                }
                onBlurNormalize={() => applyAnchors(anchors)}
              />
            </Field>
          ))}
          </div>
          <small className="hint">
          Padrão de fábrica: {builtins.in1}, {builtins.out1}, {builtins.in2},{' '}
          {builtins.out2}. Intervalo mínimo de 15 minutos entre horários
          consecutivos.
          </small>
          <div className="btn-row">
          <Button
            type="button"
            variant="secondary"
            onClick={() => applyAnchors(builtinPlannerAnchors())}
            disabled={busy}
          >
            Restaurar Padrões
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={busy}
            aria-busy={busy}
          >
            {busy ? 'Salvando…' : 'Salvar Horários'}
          </Button>
          </div>
        </form>
      </Card>
    </Page>
  )
}
