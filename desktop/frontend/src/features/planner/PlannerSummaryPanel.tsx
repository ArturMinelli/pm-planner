import { useId, useState, type ReactNode } from 'react'
import { Card, StatRow } from '../../components/ui'
import type { PlannerPayload, PlannerSummary, SolvedSlot } from '../../types'
import {
  formatDurationSecs,
  formatSignedRoundedMinutes,
} from '../../util/timeFormat'
import {
  BALANCE_UNAVAILABLE_COPY,
  balanceTooltipCopy,
} from './balanceTooltip'

type PlannerSummaryPanelProps = {
  loading?: boolean
  loaded: PlannerPayload | null
  summary: PlannerSummary | null
}

function portugueseOrdinal(position: number): string {
  return `${position}ª`
}

function solvedSlotLabel(solvedSlot: SolvedSlot): string {
  const kind = solvedSlot.isEntry ? 'Entrada' : 'Saída'
  return `${kind} ${solvedSlot.journeyIndex + 1}`
}

function SummarySkeleton() {
  return (
    <div className="stat-list skeleton-stat-list" aria-busy="true" aria-label="Carregando resumo">
      <div className="skeleton skeleton-stat-row" aria-hidden="true" />
      <div className="skeleton skeleton-stat-row" aria-hidden="true" />
      <div className="skeleton skeleton-stat-row" aria-hidden="true" />
      <div className="skeleton skeleton-stat-row" aria-hidden="true" />
    </div>
  )
}

function ExplainableStatRow({
  label,
  value,
  accent = false,
  explanation,
  explainLabel,
}: {
  label: string
  value: ReactNode
  accent?: boolean
  explanation: string
  explainLabel: string
}) {
  const tooltipId = useId()
  const [tooltipOpen, setTooltipOpen] = useState(false)

  return (
    <div className={`stat${accent ? ' accent' : ''}`}>
      <span className="stat-label">
        <span className="stat-label-text">{label}</span>
        <span className="clockout-tooltip-wrap summary-tooltip-wrap">
          <button
            type="button"
            className="summary-explain-btn"
            aria-label={explainLabel}
            aria-describedby={tooltipId}
            aria-expanded={tooltipOpen}
            onClick={() => setTooltipOpen((open) => !open)}
            onKeyDown={(event) => {
              if (event.key === 'Escape') setTooltipOpen(false)
            }}
            onBlur={() => setTooltipOpen(false)}
          >
            ?
          </button>
          <span
            id={tooltipId}
            role="tooltip"
            className={`clockout-tooltip ${tooltipOpen ? 'open' : ''}`}
          >
            {explanation}
          </span>
        </span>
      </span>
      <span className="stat-value">{value}</span>
    </div>
  )
}

export default function PlannerSummaryPanel({
  loading = false,
  loaded,
  summary,
}: PlannerSummaryPanelProps) {
  const balance = loaded?.balance
  const balanceError = loaded?.balanceError
  const hasAlternative =
    Boolean(summary?.alternativeTime) &&
    balance?.appliesToday &&
    balance.targetAdjustmentSecs !== 0
  const alternativeSlotLabel =
    summary && summary.solvedSlot.journeyIndex >= 0
      ? solvedSlotLabel(summary.solvedSlot)
      : null

  return (
    <Card title="Resumo" stretch>
      {loading ? (
        <SummarySkeleton />
      ) : !loaded || !summary ? (
        <p className="muted">Selecione uma data para ver metas e totais.</p>
      ) : (
        <div className="stat-list" aria-live="polite">
          <StatRow
            label="Meta do Dia"
            value={formatDurationSecs(loaded.baseTargetSecs)}
          />
          {summary.journeySpanSecs.map((spanSecs, index) => (
            <StatRow
              key={index}
              label={`${portugueseOrdinal(index + 1)} Jornada`}
              value={formatDurationSecs(spanSecs)}
            />
          ))}
          <StatRow label="Total" value={formatDurationSecs(summary.totalSpanSecs)} />
          <StatRow
            label="Hora Extra"
            value={formatDurationSecs(summary.overtimeSecs)}
            accent
          />
          {balance && balance.appliesToday ? (
            <ExplainableStatRow
              label="Banco de Horas"
              value={formatSignedRoundedMinutes(balance.balanceSecs)}
              explanation={balanceTooltipCopy(balance)}
              explainLabel="Explicar ajuste do banco de horas"
            />
          ) : null}
          {hasAlternative ? (
            <ExplainableStatRow
              label={
                alternativeSlotLabel
                  ? `Horário alternativo (${alternativeSlotLabel})`
                  : 'Horário alternativo'
              }
              value={summary.alternativeTime ?? ''}
              accent
              explanation={balance ? balanceTooltipCopy(balance) : ''}
              explainLabel="Explicar horário alternativo"
            />
          ) : null}
          {balanceError ? (
            <ExplainableStatRow
              label="Banco de Horas"
              value="Indisponível"
              explanation={BALANCE_UNAVAILABLE_COPY}
              explainLabel="Explicar saldo indisponível"
            />
          ) : null}
          {summary.warnings?.length ? (
            <ul className="planner-warnings summary-warnings">
              {summary.warnings.map((warning) => (
                <li key={warning} className="planner-warning">
                  {warning}
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      )}
    </Card>
  )
}
