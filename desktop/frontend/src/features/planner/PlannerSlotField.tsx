import { useId, useState } from 'react'
import PlannerTimeInput from '../../components/PlannerTimeInput'
import type { BalanceAdjustment, ClockSlot } from '../../types'
import {
  formatDurationMinutes,
  formatSignedRoundedMinutes,
} from '../../util/timeFormat'

type PlannerSlotFieldProps = {
  slot: ClockSlot
  isSolved: boolean
  fieldId: string
  label: string
  balance?: BalanceAdjustment
  balanceError?: string
  alternativeTime?: string
  disabled?: boolean
  onTimeChange: (time: string) => void
  onToggleSolved: () => void
}

function tooltipCopy(balance: BalanceAdjustment): string {
  if (balance.targetAdjustmentSecs < 0) {
    const used = formatDurationMinutes(-balance.targetAdjustmentSecs)
    return `Usa ${used} do banco para sair ${used} mais cedo. Saldo estimado: ${formatSignedRoundedMinutes(balance.remainingBalanceSecs)}.`
  }

  const cappedNote = balance.capped
    ? ` Limite diário de ${formatDurationMinutes(balance.targetAdjustmentSecs)} aplicado.`
    : ''
  return `Trabalhe ${formatDurationMinutes(balance.targetAdjustmentSecs)} a mais para gerar ${formatSignedRoundedMinutes(balance.estimatedBalanceChangeSecs)} de crédito a ${balance.multiplier.toFixed(1)}x. Saldo estimado: ${formatSignedRoundedMinutes(balance.remainingBalanceSecs)}.${cappedNote}`
}

export default function PlannerSlotField({
  slot,
  isSolved,
  fieldId,
  label,
  balance,
  balanceError,
  alternativeTime,
  disabled,
  onTimeChange,
  onToggleSolved,
}: PlannerSlotFieldProps) {
  const tooltipId = useId()
  const [tooltipOpen, setTooltipOpen] = useState(false)

  const hasAlternative = Boolean(
    balance?.appliesToday &&
      balance.targetAdjustmentSecs !== 0 &&
      alternativeTime,
  )
  const unavailable = !balance && Boolean(balanceError)
  const tone = balance && balance.targetAdjustmentSecs < 0 ? 'positive' : 'debt'
  const tooltipText = balance
    ? tooltipCopy(balance)
    : 'O horário normal continua válido. Recarregue o dia para tentar consultar o saldo novamente.'

  const showBalanceBadge = isSolved && (hasAlternative || unavailable)
  const canCalculate = !slot.registered

  return (
    <div className={`field slot-field ${isSolved ? 'slot-solved' : ''}`}>
      <div className="clockout-label-row">
        <label htmlFor={fieldId}>{label}</label>
        <div className="exit-label-actions">
          {slot.registered ? (
            <span className="registered-badge">registrada</span>
          ) : null}
          {isSolved ? (
            <span className="calculada-badge">calculada</span>
          ) : null}
          {showBalanceBadge ? (
            <span className="clockout-tooltip-wrap">
              <button
                type="button"
                className={`balance-badge ${unavailable ? 'unavailable' : tone}`}
                aria-describedby={tooltipId}
                aria-expanded={tooltipOpen}
                onClick={() => setTooltipOpen((open) => !open)}
                onKeyDown={(event) => {
                  if (event.key === 'Escape') setTooltipOpen(false)
                }}
                onBlur={() => setTooltipOpen(false)}
              >
                {unavailable
                  ? 'Saldo indisponível'
                  : `Banco ${formatSignedRoundedMinutes(balance?.balanceSecs ?? 0)}`}
              </button>
              <span
                id={tooltipId}
                role="tooltip"
                className={`clockout-tooltip ${tooltipOpen ? 'open' : ''}`}
              >
                {tooltipText}
              </span>
            </span>
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
          onClick={onToggleSolved}
          disabled={disabled || !canCalculate}
          title={!canCalculate ? 'Marcações registradas não podem ser calculadas' : undefined}
        >
          ⊕
        </button>
        {hasAlternative ? (
          <output
            className={`alternative-clockout ${tone}`}
            aria-label={`${label} alternativa: ${alternativeTime}`}
          >
            {alternativeTime}
          </output>
        ) : null}
      </div>
    </div>
  )
}
