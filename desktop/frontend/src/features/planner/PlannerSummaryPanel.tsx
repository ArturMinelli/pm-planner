import { useTranslation } from 'react-i18next'
import { Card, StatRow } from '../../components/ui'
import type { PlannerPayload, PlannerSummary } from '../../types'
import {
  formatDurationMinutes,
  formatDurationSecs,
  formatSignedRoundedMinutes,
  formatTargetHours,
} from '../../util/timeFormat'

type PlannerSummaryPanelProps = {
  loading?: boolean
  loaded: PlannerPayload | null
  summary: PlannerSummary | null
}

function SummarySkeleton({ label }: { label: string }) {
  return (
    <div className="stat-list skeleton-stat-list" aria-busy="true" aria-label={label}>
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
  const { t } = useTranslation()
  const balance = loaded?.balance
  const balanceError = loaded?.balanceError
  const hasAlternative =
    Boolean(summary?.alternativeTime) &&
    balance?.appliesToday &&
    balance.targetAdjustmentSecs !== 0

  const formatJourneyLabel = (position: number) => {
    const ordinal = `${position}ª`
    return t('planner.journeyOrdinal', { ordinal })
  }

  return (
    <Card title={t('planner.summary')} stretch>
      {loading ? (
        <SummarySkeleton label={t('planner.loadingSummary')} />
      ) : !loaded || !summary ? (
        <p className="muted">{t('planner.selectDateHint')}</p>
      ) : (
        <div className="stat-list" aria-live="polite">
          <StatRow
            label={t('planner.dailyTarget')}
            value={formatDurationSecs(loaded.baseTargetSecs)}
          />
          {summary.journeySpanSecs.map((spanSecs, index) => (
            <StatRow
              key={index}
              label={formatJourneyLabel(index + 1)}
              value={formatDurationSecs(spanSecs)}
            />
          ))}
          <StatRow
            label={t('common.total')}
            value={formatDurationSecs(summary.totalSpanSecs)}
            emphasis
          />
          <StatRow
            label={t('planner.overtime')}
            value={formatDurationSecs(summary.overtimeSecs)}
            accent
          />
          {balance && balance.appliesToday ? (
            <StatRow
              label={t('planner.timeBank')}
              value={formatSignedRoundedMinutes(balance.balanceSecs)}
            />
          ) : null}
          {hasAlternative ? (
            <StatRow
              label={t('planner.alternativeTime')}
              value={summary.alternativeTime ?? ''}
              accent
            />
          ) : null}
          {balanceError ? (
            <p className="summary-callout" data-tone="neutral" role="note">
              {t('planner.balanceUnavailableNote')}
            </p>
          ) : null}
          {balance?.appliesToday && balance.targetAdjustmentSecs !== 0 ? (
            <p
              className="summary-callout"
              data-tone={balance.targetAdjustmentSecs < 0 ? 'positive' : 'debt'}
              role="note"
            >
              {balance.targetAdjustmentSecs < 0
                ? t('planner.useBankEarly', {
                    duration: formatDurationMinutes(-balance.targetAdjustmentSecs),
                  })
                : t('planner.owingBank', {
                    debt: formatDurationMinutes(balance.balanceSecs),
                    duration: formatDurationMinutes(balance.targetAdjustmentSecs),
                    target: formatTargetHours(loaded.baseTargetSecs),
                  })}
            </p>
          ) : null}
        </div>
      )}
    </Card>
  )
}
