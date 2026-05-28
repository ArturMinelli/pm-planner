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
        setEmail(c.email)
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
    if (ttl.trim() !== '' && !Number.isFinite(cache_ttl_hours)) {
      setErr('TTL deve ser um número válido.')
      return
    }
    setBusy(true)
    try {
      await backend.mergeAndSave({
        email,
        password,
        cache_ttl_hours,
      })
      setMsg('Conta salva em ~/.config/pm/config.yaml')
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
      setMsg('Horários do planner salvos em ~/.config/pm/config.yaml')
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
    <section className="page narrow">
      <header className="page-header">
        <h1>Configurações</h1>
        <p className="muted">
          Mesmo arquivo YAML usado pela CLI (
          <code className="pill">~/.config/pm/config.yaml</code>).
        </p>
      </header>

      {devHint && <div className="banner muted">{devHint}</div>}
      {err && (
        <div className="banner error" role="alert">
          {err}
        </div>
      )}
      {msg && (
        <div className="banner ok" role="status">
          {msg}
        </div>
      )}

      <div className="card">
        <h2 className="card-title">Conta e sessão</h2>
        <div className="stack">
          <label className="field">
            <span>E-mail (login)</span>
            <input
              autoComplete="username email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </label>
          <label className="field">
            <span>Senha</span>
            <input
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </label>
          <label className="field">
            <span>Cache TTL (horas)</span>
            <input
              placeholder="Opcional · padrão 8"
              value={ttl}
              onChange={(e) => setTTL(e.target.value)}
            />
            <small className="hint">
              Quando a API não envia validade da sessão, limita há quantas
              horas o arquivo <code className="pill">session.json</code> é
              reutilizado.
            </small>
          </label>
        </div>
        <div className="btn-row">
          <button
            type="button"
            className="btn primary"
            onClick={() => void saveAccount()}
            disabled={busy}
          >
            Salvar conta
          </button>
          <button
            type="button"
            className="btn ghost"
            onClick={() => void test()}
            disabled={busy || !backend.hasWailsRuntime()}
          >
            Testar login
          </button>
        </div>
      </div>

      <div className="card">
        <h2 className="card-title">Horários padrão do planner</h2>
        <p className="muted card-intro">
          Âncoras usadas para atribuir batidas aos quatro campos e para preencher
          Entrada 1 quando não há registro correspondente. A CLI (
          <code className="pill">pm plan</code>) usa os mesmos valores.
        </p>
        <div className="stack planner-anchor-grid">
          {ANCHOR_FIELDS.map(({ field, label }) => (
            <label key={field} className="field">
              <span>{label}</span>
              <PlannerTimeInput
                value={anchors[field]}
                onChange={(value) =>
                  applyAnchors({ ...anchors, [field]: value })
                }
                onStep={(delta) =>
                  applyAnchors(adjustPlannerAnchor(anchors, field, delta))
                }
                onBlurNormalize={() => applyAnchors(anchors)}
              />
            </label>
          ))}
        </div>
        <small className="hint">
          Padrão de fábrica: {builtins.in1}, {builtins.out1}, {builtins.in2},{' '}
          {builtins.out2}. Intervalo mínimo de 15 minutos entre horários
          consecutivos.
        </small>
        <div className="btn-row">
          <button
            type="button"
            className="btn ghost"
            onClick={() => applyAnchors(builtinPlannerAnchors())}
            disabled={busy}
          >
            Restaurar padrões
          </button>
          <button
            type="button"
            className="btn primary"
            onClick={() => void savePlanner()}
            disabled={busy}
          >
            Salvar horários
          </button>
        </div>
      </div>
    </section>
  )
}
