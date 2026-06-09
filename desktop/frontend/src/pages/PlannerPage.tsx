import { useCallback, useEffect, useMemo, useState } from 'react'
import type { BalancePayload, PlannerPayload } from '../types'
import * as backend from '../services/backend'
import { localDateYYYYMMDD } from '../util/timeFormat'
import {
  adjustPlannerTime,
  enforcePlannerOrder,
  type PlannerTimeField,
  type PlannerTimes,
} from '../util/plannerTimes'
import { Banner, Page, PageHeader, Stack } from '../components/ui'
import PlannerLoadCard from '../features/planner/PlannerLoadCard'
import PlannerBalanceCard from '../features/planner/PlannerBalanceCard'
import PlannerSummaryPanel from '../features/planner/PlannerSummaryPanel'
import PlannerTimeFields from '../features/planner/PlannerTimeFields'
import { calculatePlannerSummary } from '../features/planner/plannerSummary'

export default function PlannerPage() {
  const [date, setDate] = useState(localDateYYYYMMDD)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)
  const [loaded, setLoaded] = useState<PlannerPayload | null>(null)
  const [balance, setBalance] = useState<BalancePayload | null>(null)
  const [balanceBusy, setBalanceBusy] = useState(backend.hasWailsRuntime)
  const [balanceErr, setBalanceErr] = useState<string | null>(null)

  const [in1, setIn1] = useState('')
  const [out1, setOut1] = useState('')
  const [in2, setIn2] = useState('')

  const targetSecs = loaded?.targetSecs ?? 0
  const baseTargetSecs = loaded?.baseTargetSecs ?? targetSecs
  const timesDisabled = busy || !loaded
  const hasRuntime = backend.hasWailsRuntime()
  const times = useMemo<PlannerTimes>(() => ({ in1, out1, in2 }), [in1, out1, in2])

  const summary = useMemo(
    () =>
      loaded
        ? calculatePlannerSummary({
            targetSecs,
            baseTargetSecs,
            in1,
            out1,
            in2,
          })
        : null,
    [loaded, baseTargetSecs, targetSecs, in1, out1, in2],
  )

  const refreshBalance = useCallback(async () => {
    if (!backend.hasWailsRuntime()) return
    setBalanceBusy(true)
    setBalanceErr(null)
    try {
      setBalance(await backend.loadBalance())
    } catch (e) {
      setBalanceErr(e instanceof Error ? e.message : String(e))
    } finally {
      setBalanceBusy(false)
    }
  }, [])

  useEffect(() => {
    if (!backend.hasWailsRuntime()) return
    backend
      .loadBalance()
      .then((nextBalance) => {
        setBalance(nextBalance)
        setBalanceErr(null)
      })
      .catch((e: unknown) => {
        setBalanceErr(e instanceof Error ? e.message : String(e))
      })
      .finally(() => {
        setBalanceBusy(false)
      })
  }, [])

  const applyTimes = useCallback((next: PlannerTimes) => {
    setIn1(next.in1)
    setOut1(next.out1)
    setIn2(next.in2)
  }, [])

  const changeTime = useCallback(
    (field: PlannerTimeField, value: string) => {
      applyTimes({ in1, out1, in2, [field]: value })
    },
    [applyTimes, in1, out1, in2],
  )

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

  const load = async () => {
    setErr(null)
    setBusy(true)
    setLoaded(null)
    try {
      if (!backend.hasWailsRuntime()) {
        setErr('Abra esta interface pelo app desktop pm-desktop para carregar dados.')
        return
      }
      const payload = await backend.loadPlanner(date)
      setLoaded(payload)
      if (payload.balance) {
        setBalance((current) => ({
          employeeId: current?.employeeId ?? '',
          balanceSecs: payload.balance?.balanceSecs ?? 0,
          updatedAt: payload.balanceUpdatedAt,
        }))
        setBalanceErr(null)
      } else if (payload.balanceError) {
        setBalanceErr(payload.balanceError)
      }
      setIn1(payload.in1)
      setOut1(payload.out1)
      setIn2(payload.in2)
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

        <PlannerLoadCard
          date={date}
          busy={busy}
          onDateChange={setDate}
          onLoad={() => void load()}
        />

        {err ? <Banner tone="error">{err}</Banner> : null}

        <PlannerBalanceCard
          balance={balance}
          adjustment={loaded?.balance}
          busy={balanceBusy}
          canRefresh={hasRuntime}
          error={balanceErr}
          onRefresh={() => void refreshBalance()}
        />

        <div className="grid-2">
          <PlannerTimeFields
            originals={originals}
            times={times}
            disabled={timesDisabled}
            out2={summary?.out2 ?? loaded?.out2 ?? '—'}
            onTimeChange={changeTime}
            onStep={stepTime}
            onBlurNormalize={(field) => blurTime(field, times[field])}
          />
          <PlannerSummaryPanel loaded={loaded} summary={summary} />
        </div>
      </Stack>
    </Page>
  )
}
