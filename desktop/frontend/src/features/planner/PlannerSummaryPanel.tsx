import { Card, StatRow } from '../../components/ui'
import type { PlannerPayload, PlannerSummary } from '../../types'
import {
  formatDurationMinutes,
  formatDurationSecs,
  formatSignedRoundedMinutes,
} from '../../util/timeFormat'

type PlannerSummaryPanelProps = {
  loading?: boolean
  loaded: PlannerPayload | null
  summary: PlannerSummary | null
}

function portugueseOrdinal(position: number): string {
  return `${position}ª`
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
            <StatRow
              label="Banco de Horas"
              value={formatSignedRoundedMinutes(balance.balanceSecs)}
            />
          ) : null}
          {hasAlternative ? (
            <StatRow
              label="Horário alternativo"
              value={summary.alternativeTime ?? ''}
              accent
            />
          ) : null}
          {balanceError ? (
            <p className="muted small summary-note">Saldo indisponível; horário normal mantido.</p>
          ) : null}
          {balance?.appliesToday && balance.targetAdjustmentSecs !== 0 ? (
            <p className="muted small summary-note">
              {balance.targetAdjustmentSecs < 0
                ? `Usa ${formatDurationMinutes(-balance.targetAdjustmentSecs)} do banco para sair mais cedo.`
                : `Trabalhe ${formatDurationMinutes(balance.targetAdjustmentSecs)} a mais para gerar crédito.`}
            </p>
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
