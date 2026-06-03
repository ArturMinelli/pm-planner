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
import {
  Banner,
  Button,
  Card,
  Field,
  Page,
  PageHeader,
  Stack,
  StatRow,
} from '../components/ui'

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
  const hasRuntime = backend.hasWailsRuntime()

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
        setErr('Abra esta interface pelo app desktop pm-desktop para carregar dados.')
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
    <Page>
      <PageHeader
        title="Planejar Jornada"
        description={
          <>
            Carrega o mesmo dia útil da API que o comando{' '}
            <code className="pill" translate="no">
              pm plan
            </code>
            .
          </>
        }
      />

      <Stack>
        {!hasRuntime ? (
          <Banner>
            Modo navegador: use o app desktop para carregar dados reais da API.
          </Banner>
        ) : null}

        <Card>
          <div className="row wrap">
            <Field id="planner-date" label="Data">
              <PlannerDatePicker
                id="planner-date"
                value={date}
                onChange={setDate}
                disabled={busy}
                autoFocus
              />
            </Field>
            <Button
              variant="primary"
              onClick={() => void load()}
              disabled={busy}
              aria-busy={busy}
            >
              {busy ? 'Carregando…' : 'Carregar Dia'}
            </Button>
          </div>
        </Card>

        {err ? <Banner tone="error">{err}</Banner> : null}

        <div className="grid-2">
          <Card title="Marcações Sugeridas" stretch>
            <p className="muted small originals-line">{originals}</p>
            <div className="time-grid">
              <Field id="planner-in1" label="Entrada 1">
              <PlannerTimeInput
                id="planner-in1"
                name="planner-in1"
                value={in1}
                onChange={setIn1}
                onStep={(delta) => stepTime('in1', delta)}
                onBlurNormalize={() => blurTime('in1', in1)}
                disabled={timesDisabled}
              />
              </Field>
              <Field id="planner-out1" label="Saída 1">
              <PlannerTimeInput
                id="planner-out1"
                name="planner-out1"
                value={out1}
                onChange={setOut1}
                onStep={(delta) => stepTime('out1', delta)}
                onBlurNormalize={() => blurTime('out1', out1)}
                disabled={timesDisabled}
              />
              </Field>
              <Field id="planner-in2" label="Entrada 2">
              <PlannerTimeInput
                id="planner-in2"
                name="planner-in2"
                value={in2}
                onChange={setIn2}
                onStep={(delta) => stepTime('in2', delta)}
                onBlurNormalize={() => blurTime('in2', in2)}
                disabled={timesDisabled}
              />
              </Field>
              <Field id="planner-out2" label="Saída 2 Calculada">
                <input
                  id="planner-out2"
                  name="planner-out2"
                  readOnly
                  tabIndex={-1}
                  value={preview?.out2 ?? '—'}
                />
              </Field>
            </div>
          </Card>

          <Card title="Resumo" stretch>
          {!loaded ? (
            <p className="muted">Carregue um dia para ver metas e totais.</p>
          ) : (
            <div className="stat-list" aria-live="polite">
              <StatRow
                label="Meta do Dia"
                value={formatDurationSecs(loaded.targetSecs)}
              />
              <StatRow label="1ª Jornada" value={preview?.first ?? '—'} />
              <StatRow label="2ª Jornada" value={preview?.second ?? '—'} />
              <StatRow label="Total" value={preview?.total ?? '—'} />
              <StatRow
                label="Hora Extra"
                value={preview?.extra ?? '—'}
                accent
              />
            </div>
          )}
          </Card>
        </div>
      </Stack>
    </Page>
  )
}
