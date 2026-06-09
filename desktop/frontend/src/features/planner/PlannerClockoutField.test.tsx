import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import type { BalanceAdjustment } from '../../types'
import PlannerClockoutField from './PlannerClockoutField'

function adjustment(patch: Partial<BalanceAdjustment> = {}): BalanceAdjustment {
  return {
    balanceSecs: -8 * 60 * 60,
    appliesToday: true,
    targetAdjustmentSecs: 3 * 60 * 60,
    adjustedTargetSecs: 11.5 * 60 * 60,
    estimatedBalanceChangeSecs: 4.5 * 60 * 60,
    remainingBalanceSecs: -3.5 * 60 * 60,
    multiplier: 1.5,
    capped: true,
    ...patch,
  }
}

describe('PlannerClockoutField', () => {
  it('keeps normal clockout primary and shows concise capped debt details', () => {
    const html = renderToStaticMarkup(
      <PlannerClockoutField
        out2="18:00"
        alternativeOut2="21:00"
        balance={adjustment()}
      />,
    )

    expect(html).toContain('value="18:00"')
    expect(html).toContain('Saída 2 alternativa: 21:00')
    expect(html).toContain('Banco -08:00')
    expect(html).toContain('1.5x')
    expect(html).toContain('Limite diário de 03:00 aplicado')
    expect(html).toContain('aria-describedby=')
    expect(html).toContain('role="tooltip"')
  })

  it('shows a green earlier alternative for positive credit', () => {
    const html = renderToStaticMarkup(
      <PlannerClockoutField
        out2="18:00"
        alternativeOut2="17:00"
        balance={adjustment({
          balanceSecs: 60 * 60,
          targetAdjustmentSecs: -60 * 60,
          adjustedTargetSecs: 7.5 * 60 * 60,
          estimatedBalanceChangeSecs: -60 * 60,
          remainingBalanceSecs: 0,
          capped: false,
        })}
      />,
    )

    expect(html).toContain('alternative-clockout positive')
    expect(html).toContain('Banco +01:00')
    expect(html).toContain('sair 01:00 mais cedo')
  })

  it('stays plain for zero balance and non-today dates', () => {
    const zero = renderToStaticMarkup(
      <PlannerClockoutField out2="18:00" balance={adjustment({ balanceSecs: 0, targetAdjustmentSecs: 0 })} />,
    )
    const nonToday = renderToStaticMarkup(
      <PlannerClockoutField
        out2="18:00"
        alternativeOut2="17:00"
        balance={adjustment({ appliesToday: false })}
      />,
    )

    expect(zero).not.toContain('balance-badge')
    expect(nonToday).not.toContain('balance-badge')
    expect(nonToday).not.toContain('alternative-clockout')
  })

  it('shows a muted accessible unavailable badge without changing normal time', () => {
    const html = renderToStaticMarkup(
      <PlannerClockoutField out2="18:00" balanceError="API indisponível" />,
    )

    expect(html).toContain('value="18:00"')
    expect(html).toContain('Saldo indisponível')
    expect(html).toContain('horário normal continua válido')
    expect(html).toContain('aria-describedby=')
    expect(html).toContain('role="tooltip"')
  })
})
