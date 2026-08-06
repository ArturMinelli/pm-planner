import { useId } from 'react'
import { useTranslation } from 'react-i18next'
import type { BalanceAdjustment, Journey, SolvedSlot } from '../../types'
import { isSlotSolved } from '../../util/plannerSuggestions'
import PlannerSlotField from './PlannerSlotField'

type PlannerJourneyGroupProps = {
  journey: Journey
  journeyIndex: number
  solvedSlot: SolvedSlot
  balance?: BalanceAdjustment
  alternativeTime?: string
  disabled?: boolean
  onUpdateTime: (journeyIndex: number, isEntry: boolean, time: string) => void
  onRemove: (journeyIndex: number) => void
  onToggleSolved: (journeyIndex: number, isEntry: boolean) => void
}

export default function PlannerJourneyGroup({
  journey,
  journeyIndex,
  solvedSlot,
  balance,
  alternativeTime,
  disabled,
  onUpdateTime,
  onRemove,
  onToggleSolved,
}: PlannerJourneyGroupProps) {
  const { t } = useTranslation()
  const entryInputId = useId()
  const exitInputId = useId()
  const position = journeyIndex + 1
  const ordinal = `${position}ª`
  const journeyTitle = t('planner.journeyOrdinal', { ordinal })
  const isRegistered = journey.entry.registered || journey.exit.registered
  const entrySolved = isSlotSolved(solvedSlot, journeyIndex, true)
  const exitSolved = isSlotSolved(solvedSlot, journeyIndex, false)

  return (
    <div className="journey-group">
      <div className="journey-group-header">
        <span className="journey-group-title">{journeyTitle}</span>
        <button
          type="button"
          className="btn secondary journey-remove-btn"
          onClick={() => onRemove(journeyIndex)}
          disabled={disabled || isRegistered}
          title={isRegistered ? t('planner.cannotRemoveRegistered') : undefined}
          aria-label={t('planner.removeJourney', { ordinal })}
        >
          ✕
        </button>
      </div>

      <div className="journey-group-fields">
        <PlannerSlotField
          slot={journey.entry}
          isSolved={entrySolved}
          fieldId={entryInputId}
          label={t('planner.entry', { number: String(position) })}
          balance={entrySolved ? balance : undefined}
          alternativeTime={entrySolved ? alternativeTime : undefined}
          disabled={disabled}
          onTimeChange={(time) => onUpdateTime(journeyIndex, true, time)}
          onToggleSolved={() => onToggleSolved(journeyIndex, true)}
        />

        <span className="journey-slot-connector" aria-hidden="true">→</span>

        <PlannerSlotField
          slot={journey.exit}
          isSolved={exitSolved}
          fieldId={exitInputId}
          label={t('planner.exit', { number: String(position) })}
          balance={exitSolved ? balance : undefined}
          alternativeTime={exitSolved ? alternativeTime : undefined}
          disabled={disabled}
          onTimeChange={(time) => onUpdateTime(journeyIndex, false, time)}
          onToggleSolved={() => onToggleSolved(journeyIndex, false)}
        />
      </div>
    </div>
  )
}
