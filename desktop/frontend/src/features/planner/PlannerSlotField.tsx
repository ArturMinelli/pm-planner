import { useId, useState, type KeyboardEvent, type MouseEvent } from 'react'
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

function CalculatorIcon() {
  return (
    <svg
      className="slot-mode-icon"
      width="14"
      height="14"
      viewBox="0 0 14 14"
      aria-hidden="true"
      focusable="false"
    >
      <rect x="1.5" y="1.5" width="11" height="11" rx="1.5" fill="none" stroke="currentColor" strokeWidth="1.2" />
      <rect x="3.25" y="3.25" width="2.2" height="1.6" rx="0.3" fill="currentColor" />
      <rect x="6.4" y="3.25" width="2.2" height="1.6" rx="0.3" fill="currentColor" />
      <rect x="9.55" y="3.25" width="1.25" height="1.6" rx="0.3" fill="currentColor" />
      <rect x="3.25" y="6.1" width="2.2" height="1.6" rx="0.3" fill="currentColor" />
      <rect x="6.4" y="6.1" width="2.2" height="1.6" rx="0.3" fill="currentColor" />
      <rect x="9.55" y="6.1" width="1.25" height="1.6" rx="0.3" fill="currentColor" />
      <rect x="3.25" y="8.95" width="7.55" height="1.9" rx="0.3" fill="currentColor" />
    </svg>
  )
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
  const labelId = `${fieldId}-label`
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
  const canToggleCalculate = canCalculate && !disabled
  const showCalculatorIcon = canToggleCalculate && !isSolved
  const showAutoPill = isSolved && canToggleCalculate
  const labelLinksToInput = !showCalculatorIcon

  const cardTitle = slot.registered
    ? registeredHint
    : isSolved
      ? t('planner.disableAutoCalculate')
      : canToggleCalculate
        ? calculateHint
        : undefined

  const cardAriaLabel = isSolved
    ? t('planner.calculatedAria', { label, time: slot.time })
    : slot.registered
      ? `${label}, ${registeredHint}`
      : canToggleCalculate
        ? `${label}. ${calculateHint}`
        : undefined

  function handleCardClick(event: MouseEvent<HTMLDivElement>) {
    if (disabled) return
    if (event.target instanceof HTMLInputElement) return
    if ((event.target as HTMLElement).closest('.clockout-tooltip-wrap')) return
    if (isSolved) {
      if (!canToggleCalculate) return
      onToggleSolved()
      return
    }
    if (!canToggleCalculate || inputFocused) return
    onToggleSolved()
  }

  function handleCardKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (!isSolved || !canToggleCalculate || disabled) return
    if (event.key !== 'Enter' && event.key !== ' ') return
    event.preventDefault()
    onToggleSolved()
  }

  function handleCalculateButtonClick(event: MouseEvent<HTMLButtonElement>) {
    event.stopPropagation()
    if (!canToggleCalculate || disabled) return
    onToggleSolved()
  }

  const cardClassName = [
    'slot-field-card',
    isSolved && canToggleCalculate && 'calculated',
    !isSolved && canToggleCalculate && 'calculable',
    slot.registered && 'registered',
    canToggleCalculate && 'interactive',
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div className="field slot-field">
      <div
        className={cardClassName}
        onClick={handleCardClick}
        onKeyDown={handleCardKeyDown}
        title={cardTitle}
        aria-label={isSolved ? cardAriaLabel : undefined}
        role={isSolved && canToggleCalculate ? 'button' : undefined}
        tabIndex={isSolved && canToggleCalculate ? 0 : undefined}
      >
        {showAutoPill ? (
          <span className="slot-mode-indicator auto-pill">{t('planner.calculateModeAuto')}</span>
        ) : null}
        <div className="slot-field-header">
          <div className="slot-field-label-row">
            {showCalculatorIcon ? (
              <button
                type="button"
                className="slot-mode-indicator calculator"
                title={calculateHint}
                aria-label={calculateHint}
                onClick={handleCalculateButtonClick}
              >
                <CalculatorIcon />
              </button>
            ) : null}
            {labelLinksToInput ? (
              <label htmlFor={fieldId} className="slot-field-label" id={labelId}>
                {label}
              </label>
            ) : (
              <span className="slot-field-label" id={labelId}>
                {label}
              </span>
            )}
          </div>
          {showBalanceBadge ? (
            <div className="slot-status-badges">
              <span className="clockout-tooltip-wrap">
                <span
                  className={`balance-badge ${unavailable ? 'unavailable' : tone}`}
                  aria-describedby={tooltipId}
                  tabIndex={0}
                  role="note"
                >
                  {unavailable
                    ? t('planner.balanceUnavailable')
                    : t('planner.bankBadge', {
                        balance: formatSignedRoundedMinutes(balance?.balanceSecs ?? 0),
                      })}
                </span>
                <span id={tooltipId} role="tooltip" className="clockout-tooltip">
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
                <span id={fieldId} className="calculada-time" aria-hidden="true">
                  {slot.time}
                </span>
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
              <PlannerTimeInput
                id={fieldId}
                name={fieldId}
                value={slot.time}
                onChange={onTimeChange}
                disabled={disabled}
                className="planner-slot-time-input"
                aria-labelledby={labelId}
                onFocus={() => setInputFocused(true)}
                onBlur={() => setInputFocused(false)}
              />
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
