import { useTranslation } from 'react-i18next'
import { Card } from '../../components/ui'
import type { BalanceAdjustment, Journey, PlannerSummary, SolvedSlot } from '../../types'
import { messageKey, translateMessage } from '../../i18n/translateMessage'
import { canRemoveJourney } from '../../util/plannerSuggestions'
import PlannerJourneyGroup from './PlannerJourneyGroup'

type PlannerJourneyListProps = {
  loading?: boolean
  journeys: Journey[]
  solvedSlot: SolvedSlot
  balance?: BalanceAdjustment
  summary?: PlannerSummary | null
  punches?: string[]
  originalsLine?: string
  disabled?: boolean
  onAddJourney: () => void
  onRemoveJourney: (journeyIndex: number) => void
  onUpdateJourney: (journeyIndex: number, isEntry: boolean, time: string) => void
  onToggleSolved: (journeyIndex: number, isEntry: boolean) => void
}

function JourneyListSkeleton({ label }: { label: string }) {
  return (
    <>
      <div className="skeleton skeleton-line" style={{ width: '42%' }} aria-hidden="true" />
      <div className="journey-list" aria-busy="true" aria-label={label}>
        <div className="skeleton skeleton-block journey-skeleton" aria-hidden="true" />
        <div className="skeleton skeleton-block journey-skeleton" aria-hidden="true" />
      </div>
    </>
  )
}

export default function PlannerJourneyList({
  loading = false,
  journeys,
  solvedSlot,
  balance,
  summary,
  punches = [],
  originalsLine,
  disabled,
  onAddJourney,
  onRemoveJourney,
  onUpdateJourney,
  onToggleSolved,
}: PlannerJourneyListProps) {
  const { t } = useTranslation()
  const alternativeTime = summary?.alternativeTime

  return (
    <Card title={t('planner.dayTimes')} stretch>
      {loading ? (
        <JourneyListSkeleton label={t('planner.loadingJourneys')} />
      ) : (
        <>
          <p className="muted small originals-line">{originalsLine ?? t('planner.none')}</p>

          {summary?.warnings?.length ? (
            <ul className="planner-warnings" aria-live="polite">
              {summary.warnings.map((warning) => (
                <li key={messageKey(warning)} className="planner-warning">
                  {translateMessage(t, warning)}
                </li>
              ))}
            </ul>
          ) : null}

          <div className="journey-list">
            {journeys.map((journey, index) => {
              const registered = journey.entry.registered || journey.exit.registered
              return (
                <PlannerJourneyGroup
                  key={index}
                  journey={journey}
                  journeyIndex={index}
                  solvedSlot={solvedSlot}
                  balance={balance}
                  alternativeTime={alternativeTime}
                  disabled={disabled}
                  canRemove={canRemoveJourney(journeys.length, punches.length, registered)}
                  onUpdateTime={onUpdateJourney}
                  onRemove={onRemoveJourney}
                  onToggleSolved={onToggleSolved}
                />
              )
            })}
          </div>
        </>
      )}

      <div className="btn-row">
        <button
          type="button"
          className="btn secondary"
          onClick={onAddJourney}
          disabled={disabled}
        >
          {t('planner.addJourney')}
        </button>
      </div>
    </Card>
  )
}
