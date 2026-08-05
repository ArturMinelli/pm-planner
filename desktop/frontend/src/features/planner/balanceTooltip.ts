import type { BalanceAdjustment } from '../../types'
import {
  formatDurationMinutes,
  formatSignedRoundedMinutes,
} from '../../util/timeFormat'

export function balanceTooltipCopy(balance: BalanceAdjustment): string {
  if (balance.targetAdjustmentSecs < 0) {
    const used = formatDurationMinutes(-balance.targetAdjustmentSecs)
    return `Usa ${used} do banco para sair ${used} mais cedo. Saldo estimado: ${formatSignedRoundedMinutes(balance.remainingBalanceSecs)}.`
  }

  const cappedNote = balance.capped
    ? ` Limite diário de ${formatDurationMinutes(balance.targetAdjustmentSecs)} aplicado.`
    : ''
  return `Trabalhe ${formatDurationMinutes(balance.targetAdjustmentSecs)} a mais para gerar ${formatSignedRoundedMinutes(balance.estimatedBalanceChangeSecs)} de crédito a ${balance.multiplier.toFixed(1)}x. Saldo estimado: ${formatSignedRoundedMinutes(balance.remainingBalanceSecs)}.${cappedNote}`
}

export const BALANCE_UNAVAILABLE_COPY =
  'O horário normal continua válido. Recarregue o dia para tentar consultar o saldo novamente.'
