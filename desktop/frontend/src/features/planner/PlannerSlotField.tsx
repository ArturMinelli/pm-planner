import { useId } from 'react'
import PlannerTimeInput from '../../components/PlannerTimeInput'
import type { ClockSlot } from '../../types'

type PlannerSlotFieldProps = {
  slot: ClockSlot
  isSolved: boolean
  fieldId: string
  label: string
  disabled?: boolean
  onTimeChange: (time: string) => void
  onToggleSolved: () => void
}

const REGISTERED_CALCULATOR_HINT =
  'Marcações registradas não podem ser calculadas'

export default function PlannerSlotField({
  slot,
  isSolved,
  fieldId,
  label,
  disabled,
  onTimeChange,
  onToggleSolved,
}: PlannerSlotFieldProps) {
  const registeredHintId = useId()
  const canCalculate = !slot.registered
  const fieldClass = [
    'field',
    'slot-field',
    isSolved ? 'slot-solved' : '',
    slot.registered ? 'slot-registered' : '',
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div className={fieldClass}>
      <div className="clockout-label-row">
        <label htmlFor={fieldId}>{label}</label>
        <div className="exit-label-actions">
          {isSolved ? (
            <span className="calculada-badge">calculada</span>
          ) : null}
        </div>
      </div>
      <div className="clockout-values">
        {isSolved ? (
          <output
            id={fieldId}
            className="calculada-time"
            aria-label={`${label} calculada: ${slot.time}`}
          >
            {slot.time}
          </output>
        ) : (
          <PlannerTimeInput
            id={fieldId}
            name={fieldId}
            value={slot.time}
            onChange={onTimeChange}
            disabled={disabled}
          />
        )}
        <button
          type="button"
          className={`calculator-btn ${isSolved ? 'active' : ''}`}
          aria-label={
            isSolved
              ? 'Desativar cálculo automático'
              : 'Calcular este horário pela meta do dia'
          }
          aria-describedby={!canCalculate ? registeredHintId : undefined}
          onClick={onToggleSolved}
          disabled={disabled || !canCalculate}
          title={!canCalculate ? REGISTERED_CALCULATOR_HINT : undefined}
        >
          ⊕
        </button>
        {!canCalculate ? (
          <span id={registeredHintId} className="sr-only">
            {REGISTERED_CALCULATOR_HINT}
          </span>
        ) : null}
      </div>
    </div>
  )
}
