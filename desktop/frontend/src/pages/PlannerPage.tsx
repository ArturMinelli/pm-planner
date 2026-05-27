import { useCallback, useEffect, useMemo, useState } from 'react'
import type { PlannerPayload, PlannerSummary } from '../types'
import * as backend from '../services/backend'
import { formatDurationSecs, localDateYYYYMMDD } from '../util/timeFormat'
import {
  adjustPlannerTime,
  enforcePlannerOrder,
  type PlannerTimeField,
  type PlannerTimes,
} from '../util/plannerTimes'
import PlannerDatePicker from '../components/PlannerDatePicker'
import PlannerTimeInput from '../components/PlannerTimeInput'

export default function PlannerPage() {
  const [date, setDate] = useState(localDateYYYYMMDD)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [loaded, setLoaded] = useState<PlannerPayload | null>(null)

  const [in1, setIn1] = useState('')
  const [out1, setOut1] = useState('')
  const [in2, setIn2] = useState('')
  const [summary, setSummary] = useState<PlannerSummary | null>(null)

  const targetSecs = loaded?.targetSecs ?? 0
  const timesDisabled = busy || !loaded

  const applyTimes = useCallback((next: PlannerTimes) => {
    setIn1(next.in1)
    setOut1(next.out1)
    setIn2(next.in2)
  }, [])

  const stepTime = useCallback(
    (field: PlannerTimeField, delta: number) => {
      applyTimes(adjustPlannerTime({ in1, out1, in2 }, field, delta))
    },
    [applyTimes, in1, out1, in2],
  )

  const blurTime = useCallback(
    (field: PlannerTimeField, value: string) => {
      applyTimes(enforcePlannerOrder({ in1, out1, in2, [field]: value }))
    },
    [applyTimes, in1, out1, in2],
  )

  const recalc = useCallback(async () => {
    if (!loaded || !backend.hasWailsRuntime()) return
    try {
      const s = await backend.recalculatePlanner(
        date,
        targetSecs,
        in1,
        out1,
        in2,
      )
      setSummary(s)
    } catch {
      /* ignore debounced noise */
    }
  }, [date, targetSecs, in1, out1, in2, loaded])

  useEffect(() => {
    const h = window.setTimeout(() => {
      void recalc()
    }, 100)
    return () => window.clearTimeout(h)
  }, [recalc])

  const load = async () => {
    setErr(null)
    setBusy(true)
    setLoaded(null)
    setSummary(null)
    try {
      if (!backend.hasWailsRuntime()) {
        setErr('Abra esta interface pelo app desktop (pm-desktop).')
        return
      }
      const payload = await backend.loadPlanner(date)
      setLoaded(payload)
      setIn1(payload.in1)
      setOut1(payload.out1)
      setIn2(payload.in2)
      const sum = await backend.recalculatePlanner(
        date,
        payload.targetSecs,
        payload.in1,
        payload.out1,
        payload.in2,
      )
      setSummary(sum)
    } catch (e) {
      setErr(e instanceof Error ? e.message : String(e))
    } finally {
      setBusy(false)
    }
  }

  const originals = useMemo(
    () => loaded?.originalsLine ?? '(carregue o dia primeiro)',
    [loaded?.originalsLine],
  )

  const preview = useMemo(() => {
    if (!summary && loaded) {
      return {
        out2: loaded.out2,
        first: '—',
        second: '—',
        total: '—',
        extra: '—',
      }
    }
    if (summary) {
      return {
        out2: summary.out2,
        first: formatDurationSecs(summary.firstSpanSecs),
        second: formatDurationSecs(summary.secondSpanSecs),
        total: formatDurationSecs(summary.totalSpanSecs),
        extra: formatDurationSecs(summary.overtimeSecs),
      }
    }
    return null
  }, [summary, loaded])

  return (
    <section className="page">
      <header className="page-header">
        <h1>Planejar jornada</h1>
        <p className="muted">
          Carrega o mesmo dia útil da API que o comando{' '}
          <code className="pill">pm plan</code>.
        </p>
      </header>

      <div className="stack">
        <div className="card">
          <div className="row wrap">
            <label className="field">
              <span>Data</span>
              <PlannerDatePicker
                value={date}
                onChange={setDate}
                disabled={busy}
                autoFocus
              />
            </label>
            <button
              type="button"
              className="btn primary"
              onClick={() => void load()}
              disabled={busy}
            >
              {busy ? 'Carregando…' : 'Carregar dia'}
            </button>
          </div>
        </div>

        {err && (
          <div className="banner error" role="alert">
            {err}
          </div>
        )}

        <div className="grid-2">
          <div className="card stretch">
          <h2>Marcações sugeridas</h2>
          <p className="muted small originals-line">{originals}</p>
          <div className="time-grid">
            <label className="field">
              <span>Entrada 1</span>
              <PlannerTimeInput
                value={in1}
                onChange={setIn1}
                onStep={(delta) => stepTime('in1', delta)}
                onBlurNormalize={() => blurTime('in1', in1)}
                disabled={timesDisabled}
              />
            </label>
            <label className="field">
              <span>Saída 1</span>
              <PlannerTimeInput
                value={out1}
                onChange={setOut1}
                onStep={(delta) => stepTime('out1', delta)}
                onBlurNormalize={() => blurTime('out1', out1)}
                disabled={timesDisabled}
              />
            </label>
            <label className="field">
              <span>Entrada 2</span>
              <PlannerTimeInput
                value={in2}
                onChange={setIn2}
                onStep={(delta) => stepTime('in2', delta)}
                onBlurNormalize={() => blurTime('in2', in2)}
                disabled={timesDisabled}
              />
            </label>
            <label className="field">
              <span>Saída 2 (calc.)</span>
              <input readOnly tabIndex={-1} value={preview?.out2 ?? '—'} />
            </label>
          </div>
        </div>
        <div className="card stretch">
          <h2>Resumo</h2>
          {!loaded ? (
            <p className="muted">Carregue um dia para ver metas e totais.</p>
          ) : (
            <>
              <div className="stat-list">
                <div className="stat">
                  <span className="stat-label">Meta do dia</span>
                  <span className="stat-value">
                    {formatDurationSecs(loaded.targetSecs)}
                  </span>
                </div>
                <div className="stat">
                  <span className="stat-label">1ª jornada</span>
                  <span className="stat-value">{preview?.first ?? '—'}</span>
                </div>
                <div className="stat">
                  <span className="stat-label">2ª jornada</span>
                  <span className="stat-value">{preview?.second ?? '—'}</span>
                </div>
                <div className="stat">
                  <span className="stat-label">Total</span>
                  <span className="stat-value">{preview?.total ?? '—'}</span>
                </div>
                <div className="stat accent">
                  <span className="stat-label">Hora extra</span>
                  <span className="stat-value">{preview?.extra ?? '—'}</span>
                </div>
              </div>
            </>
          )}
        </div>
        </div>
      </div>
    </section>
  )
}
