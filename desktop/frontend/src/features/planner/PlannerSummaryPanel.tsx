import { useTranslation } from 'react-i18next'
import { Card, StatRow } from '../../components/ui'
import type { LocalizedMessage, PlannerPayload, PlannerSummary } from '../../types'
import { messageKey, translateMessage } from '../../i18n/translateMessage'
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
  const { t, i18n } = useTranslation()
  const balance = loaded?.balance
  const balanceError = loaded?.balanceError
  const hasAlternative =
    Boolean(summary?.alternativeTime) &&
    balance?.appliesToday &&
    balance.targetAdjustmentSecs !== 0

  const formatJourneyLabel = (position: number) => {
    const ordinal = i18n.language === 'pt-BR' ? `${position}ª` : String(position)
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
          <StatRow label={t('common.total')} value={formatDurationSecs(summary.totalSpanSecs)} />
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
            <p className="muted small summary-note">{t('planner.balanceUnavailableNote')}</p>
          ) : null}
          {balance?.appliesToday && balance.targetAdjustmentSecs !== 0 ? (
            <p className="muted small summary-note">
              {balance.targetAdjustmentSecs < 0
                ? t('planner.useBankEarly', {
                    duration: formatDurationMinutes(-balance.targetAdjustmentSecs),
                  })
                : t('planner.workExtraCredit', {
                    duration: formatDurationMinutes(balance.targetAdjustmentSecs),
                  })}
            </p>
          ) : null}
          {summary.warnings?.length ? (
            <ul className="planner-warnings summary-warnings">
              {summary.warnings.map((warning: LocalizedMessage) => (
                <li key={messageKey(warning)} className="planner-warning">
                  {translateMessage(t, warning)}
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      )}
    </Card>
  )
}
