import { useId, useState } from 'react'
import type { BalanceAdjustment } from '../../types'
import {
  formatDurationMinutes,
  formatSignedRoundedMinutes,
} from '../../util/timeFormat'

type PlannerClockoutFieldProps = {
  out2: string
  alternativeOut2?: string
  balance?: BalanceAdjustment
  balanceError?: string
}

function tooltipCopy(balance: BalanceAdjustment): string {
  if (balance.targetAdjustmentSecs < 0) {
    const used = formatDurationMinutes(-balance.targetAdjustmentSecs)
    return `Usa ${used} do banco para sair ${used} mais cedo. Saldo estimado: ${formatSignedRoundedMinutes(balance.remainingBalanceSecs)}.`
  }

  const capped = balance.capped ? ' Limite diário de 03:00 aplicado.' : ''
  return `Trabalhe ${formatDurationMinutes(balance.targetAdjustmentSecs)} a mais para gerar ${formatSignedRoundedMinutes(balance.estimatedBalanceChangeSecs)} de crédito a ${balance.multiplier.toFixed(1)}x. Saldo estimado: ${formatSignedRoundedMinutes(balance.remainingBalanceSecs)}.${capped}`
}

export default function PlannerClockoutField({
  out2,
  alternativeOut2,
  balance,
  balanceError,
}: PlannerClockoutFieldProps) {
  const tooltipId = useId()
  const [tooltipOpen, setTooltipOpen] = useState(false)
  const hasAlternative = Boolean(
    balance?.appliesToday &&
      balance.targetAdjustmentSecs !== 0 &&
      alternativeOut2,
  )
  const unavailable = !balance && Boolean(balanceError)
  const tone =
    balance && balance.targetAdjustmentSecs < 0 ? 'positive' : 'debt'
  const copy = balance
    ? tooltipCopy(balance)
    : 'O horário normal continua válido. Recarregue o dia para tentar consultar o saldo novamente.'

  return (
    <div className="field clockout-field">
      <div className="clockout-label-row">
        <label htmlFor="planner-out2">Saída 2</label>
        {hasAlternative || unavailable ? (
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
              {copy}
            </span>
          </span>
        ) : null}
      </div>
      <div className="clockout-values">
        <input
          id="planner-out2"
          name="planner-out2"
          readOnly
          tabIndex={-1}
          value={out2}
          aria-label="Saída 2 normal"
        />
        {hasAlternative ? (
          <output
            className={`alternative-clockout ${tone}`}
            aria-label={`Saída 2 alternativa: ${alternativeOut2}`}
          >
            {alternativeOut2}
          </output>
        ) : null}
      </div>
    </div>
  )
}
