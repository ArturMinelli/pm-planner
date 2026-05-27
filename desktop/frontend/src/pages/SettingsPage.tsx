import { useEffect, useState } from 'react'
import * as backend from '../services/backend'

export default function SettingsPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [ttl, setTTL] = useState<string>('')
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
      } catch (e) {
        setErr(e instanceof Error ? e.message : String(e))
      }
    })()
  }, [])

  async function persist(): Promise<boolean> {
    setMsg(null)
    setErr(null)
    if (!backend.hasWailsRuntime()) {
      setErr(
        'Configurações são gravadas pelo app desktop. Use pm-desktop aqui.',
      )
      return false
    }
    const cache_ttl_hours = ttl.trim() === '' ? undefined : Number(ttl)
    if (ttl.trim() !== '' && !Number.isFinite(cache_ttl_hours)) {
      setErr('TTL deve ser um número válido.')
      return false
    }
    try {
      await backend.saveConfig({
        email,
        password,
        cache_ttl_hours,
      })
      return true
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
      return false
    }
  }

  const save = async () => {
    setBusy(true)
    try {
      const ok = await persist()
      if (ok) setMsg('Salvo em ~/.config/pm/config.yaml')
    } finally {
      setBusy(false)
    }
  }

  const test = async () => {
    setMsg(null)
    setErr(null)
    setBusy(true)
    try {
      if (!backend.hasWailsRuntime()) {
        setErr('Autenticação só funciona dentro do desktop.')
        return
      }
      const ok = await persist()
      if (!ok) return
      const problem = await backend.pingAuth()
      if (problem) setErr(problem)
      else setMsg('Credenciais / sessão ok.')
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
            onClick={() => void save()}
            disabled={busy}
          >
            Salvar
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
    </section>
  )
}
