import { Card } from '../../components/ui'
import type { BalanceAdjustment, Journey, PlannerSummary, SolvedSlot } from '../../types'
import PlannerJourneyGroup from './PlannerJourneyGroup'

type PlannerJourneyListProps = {
  loading?: boolean
  journeys: Journey[]
  solvedSlot: SolvedSlot
  balance?: BalanceAdjustment
  balanceError?: string
  summary?: PlannerSummary | null
  originalsLine?: string
  disabled?: boolean
  onAddJourney: () => void
  onRemoveJourney: (journeyIndex: number) => void
  onUpdateJourney: (journeyIndex: number, isEntry: boolean, time: string) => void
  onToggleSolved: (journeyIndex: number, isEntry: boolean) => void
}

function JourneyListSkeleton() {
  return (
    <>
      <div className="skeleton skeleton-line" style={{ width: '42%' }} aria-hidden="true" />
      <div className="journey-list" aria-busy="true" aria-label="Carregando marcações">
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
  balanceError,
  summary,
  originalsLine,
  disabled,
  onAddJourney,
  onRemoveJourney,
  onUpdateJourney,
  onToggleSolved,
}: PlannerJourneyListProps) {
  const alternativeTime = summary?.alternativeTime

  return (
    <Card title="Horários do dia" stretch>
      {loading ? (
        <JourneyListSkeleton />
      ) : (
        <>
          {originalsLine ? (
            <p className="muted small originals-line">{originalsLine}</p>
          ) : null}

          {summary?.warnings?.length ? (
            <ul className="planner-warnings" aria-live="polite">
              {summary.warnings.map((warning) => (
                <li key={warning} className="planner-warning">
                  {warning}
                </li>
              ))}
            </ul>
          ) : null}

          <div className="journey-list">
            {journeys.map((journey, index) => (
              <PlannerJourneyGroup
                key={index}
                journey={journey}
                journeyIndex={index}
                solvedSlot={solvedSlot}
                balance={balance}
                balanceError={balanceError}
                alternativeTime={alternativeTime}
                disabled={disabled}
                onUpdateTime={onUpdateJourney}
                onRemove={onRemoveJourney}
                onToggleSolved={onToggleSolved}
              />
            ))}
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
          + Adicionar jornada
        </button>
      </div>
    </Card>
  )
}
