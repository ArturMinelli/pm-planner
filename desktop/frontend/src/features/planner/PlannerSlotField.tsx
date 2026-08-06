import { useId, useState, type MouseEvent } from 'react'
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
  const [inputFocused, setInputFocused] = useState(false)

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
  const calculateHint = 'Clique para calcular pela meta do dia'
  const registeredHint = 'Marcação registrada no PontoMais'

  function handleTimeWrapClick(event: MouseEvent<HTMLDivElement>) {
    if (isSolved || !canCalculate || disabled || inputFocused) return
    if (event.target instanceof HTMLInputElement) return
    onToggleSolved()
  }

  return (
    <div className="field slot-field">
      <div className="slot-field-header">
        <label htmlFor={fieldId}>{label}</label>
        {showBalanceBadge ? (
          <div className="slot-status-badges">
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
          </div>
        ) : null}
      </div>
      <div className="clockout-values">
        <div className="clockout-values-primary">
          {isSolved ? (
            <div className="calculada-timebox">
              <button
                type="button"
                id={fieldId}
                className="calculada-time"
                aria-label={`${label} calculada: ${slot.time}. Clique para desativar o cálculo automático`}
                onClick={onToggleSolved}
                disabled={disabled || !canCalculate}
                title={
                  !canCalculate
                    ? 'Marcações registradas não podem ser calculadas'
                    : 'Desativar cálculo automático'
                }
              >
                {slot.time}
              </button>
              {hasAlternative ? (
                <output
                  className={`calculada-time-alt alternative-clockout ${tone}`}
                  aria-label={`${label} alternativa: ${alternativeTime}`}
                >
                  {alternativeTime}
                </output>
              ) : null}
            </div>
          ) : (
            <div
              className={`clockout-time-wrap${slot.registered ? ' registered' : ''}${canCalculate && !disabled ? ' calculable' : ''}`}
              onClick={handleTimeWrapClick}
              title={
                slot.registered
                  ? registeredHint
                  : canCalculate && !disabled
                    ? calculateHint
                    : undefined
              }
              aria-label={slot.registered ? `${label}, ${registeredHint}` : undefined}
            >
              <PlannerTimeInput
                id={fieldId}
                name={fieldId}
                value={slot.time}
                onChange={onTimeChange}
                disabled={disabled}
                onFocus={() => setInputFocused(true)}
                onBlur={() => setInputFocused(false)}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
