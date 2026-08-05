import { useId } from 'react'
import type { BalanceAdjustment, Journey, SolvedSlot } from '../../types'
import { isSlotSolved } from '../../util/plannerSuggestions'
import PlannerSlotField from './PlannerSlotField'

type PlannerJourneyGroupProps = {
  journey: Journey
  journeyIndex: number
  solvedSlot: SolvedSlot
  balance?: BalanceAdjustment
  balanceError?: string
  alternativeTime?: string
  disabled?: boolean
  onUpdateTime: (journeyIndex: number, isEntry: boolean, time: string) => void
  onRemove: (journeyIndex: number) => void
  onToggleSolved: (journeyIndex: number, isEntry: boolean) => void
}

function portugueseOrdinal(position: number): string {
  return `${position}ª`
}

export default function PlannerJourneyGroup({
  journey,
  journeyIndex,
  solvedSlot,
  balance,
  balanceError,
  alternativeTime,
  disabled,
  onUpdateTime,
  onRemove,
  onToggleSolved,
}: PlannerJourneyGroupProps) {
  const entryInputId = useId()
  const exitInputId = useId()
  const ordinal = portugueseOrdinal(journeyIndex + 1)
  const isRegistered = journey.entry.registered || journey.exit.registered
  const entrySolved = isSlotSolved(solvedSlot, journeyIndex, true)
  const exitSolved = isSlotSolved(solvedSlot, journeyIndex, false)

  return (
    <div className="journey-group">
      <div className="journey-group-header">
        <span className="journey-group-title">{ordinal} Jornada</span>
        <button
          type="button"
          className="btn secondary journey-remove-btn"
          onClick={() => onRemove(journeyIndex)}
          disabled={disabled || isRegistered}
          title={isRegistered ? 'Não é possível remover uma jornada com marcações registradas' : undefined}
          aria-label={`Remover ${ordinal} Jornada`}
        >
          ✕
        </button>
      </div>

      <div className="journey-group-fields">
        <PlannerSlotField
          slot={journey.entry}
          isSolved={entrySolved}
          fieldId={entryInputId}
          label={`Entrada ${journeyIndex + 1}`}
          balance={entrySolved ? balance : undefined}
          balanceError={entrySolved ? balanceError : undefined}
          alternativeTime={entrySolved ? alternativeTime : undefined}
          disabled={disabled}
          onTimeChange={(time) => onUpdateTime(journeyIndex, true, time)}
          onToggleSolved={() => onToggleSolved(journeyIndex, true)}
        />

        <PlannerSlotField
          slot={journey.exit}
          isSolved={exitSolved}
          fieldId={exitInputId}
          label={`Saída ${journeyIndex + 1}`}
          balance={exitSolved ? balance : undefined}
          balanceError={exitSolved ? balanceError : undefined}
          alternativeTime={exitSolved ? alternativeTime : undefined}
          disabled={disabled}
          onTimeChange={(time) => onUpdateTime(journeyIndex, false, time)}
          onToggleSolved={() => onToggleSolved(journeyIndex, false)}
        />
      </div>
    </div>
  )
}
