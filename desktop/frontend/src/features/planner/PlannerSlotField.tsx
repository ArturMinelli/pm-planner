import { useId, useState, type MouseEvent } from 'react'
import { useTranslation } from 'react-i18next'
import PlannerTimeInput from '../../components/PlannerTimeInput'
import type { BalanceAdjustment, ClockSlot, LocalizedMessage } from '../../types'
import { translateMessage } from '../../i18n/translateMessage'
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
  balanceError?: LocalizedMessage
  alternativeTime?: string
  disabled?: boolean
  onTimeChange: (time: string) => void
  onToggleSolved: () => void
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
  const { t } = useTranslation()
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
    ? balance.targetAdjustmentSecs < 0
      ? t('planner.balanceTooltipUse', {
          used: formatDurationMinutes(-balance.targetAdjustmentSecs),
          remaining: formatSignedRoundedMinutes(balance.remainingBalanceSecs),
        })
      : `${t('planner.balanceTooltipCredit', {
          extra: formatDurationMinutes(balance.targetAdjustmentSecs),
          credit: formatSignedRoundedMinutes(balance.estimatedBalanceChangeSecs),
          multiplier: balance.multiplier.toFixed(1),
          remaining: formatSignedRoundedMinutes(balance.remainingBalanceSecs),
        })}${balance.capped ? t('planner.balanceTooltipCapped', { cap: formatDurationMinutes(balance.targetAdjustmentSecs) }) : ''}`
    : translateMessage(t, balanceError) ?? t('planner.balanceTooltipUnavailable')

  const showBalanceBadge = isSolved && (hasAlternative || unavailable)
  const canCalculate = !slot.registered
  const calculateHint = t('planner.calculateHint')
  const registeredHint = t('planner.registeredHint')

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
                  ? t('planner.balanceUnavailable')
                  : t('planner.bankBadge', {
                      balance: formatSignedRoundedMinutes(balance?.balanceSecs ?? 0),
                    })}
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
                aria-label={t('planner.calculatedAria', { label, time: slot.time })}
                onClick={onToggleSolved}
                disabled={disabled || !canCalculate}
                title={
                  !canCalculate
                    ? t('planner.registeredCannotCalculate')
                    : t('planner.disableAutoCalculate')
                }
              >
                {slot.time}
              </button>
              {hasAlternative ? (
                <output
                  className={`calculada-time-alt alternative-clockout ${tone}`}
                  aria-label={t('planner.alternativeAria', { label, time: alternativeTime ?? '' })}
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
