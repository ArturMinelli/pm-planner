import { Card, StatRow } from '../../components/ui'
import type { PlannerPayload, PlannerSummary } from '../../types'
import { formatDurationSecs } from '../../util/timeFormat'

type PlannerSummaryPanelProps = {
  loaded: PlannerPayload | null
  summary: PlannerSummary | null
}

export default function PlannerSummaryPanel({
  loaded,
  summary,
}: PlannerSummaryPanelProps) {
  return (
    <Card title="Resumo" stretch>
      {!loaded || !summary ? (
        <p className="muted">Carregue um dia para ver metas e totais.</p>
      ) : (
        <div className="stat-list" aria-live="polite">
          <StatRow
            label="Meta do Dia"
            value={formatDurationSecs(loaded.targetSecs)}
          />
          <StatRow
            label="1ª Jornada"
            value={formatDurationSecs(summary.firstSpanSecs)}
          />
          <StatRow
            label="2ª Jornada"
            value={formatDurationSecs(summary.secondSpanSecs)}
          />
          <StatRow label="Total" value={formatDurationSecs(summary.totalSpanSecs)} />
          <StatRow
            label="Hora Extra"
            value={formatDurationSecs(summary.overtimeSecs)}
            accent
          />
        </div>
      )}
    </Card>
  )
}
