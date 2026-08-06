import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it, vi } from 'vitest'
import type { BalanceAdjustment, ClockSlot } from '../../types'
import PlannerSlotField from './PlannerSlotField'

const noop = vi.fn()

function slot(overrides: Partial<ClockSlot> = {}): ClockSlot {
  return { time: '18:00', registered: false, ...overrides }
}

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

describe('PlannerSlotField', () => {
  it('shows editable input in a calculable wrap when not solved', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={false}
        fieldId="exit-0"
        label="Saída 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('Saída 1')
    expect(html).toContain('value="18:00"')
    expect(html).toContain('clockout-time-wrap calculable')
    expect(html).toContain('Clique para calcular pela meta do dia')
    expect(html).not.toContain('calculada-time')
    expect(html).not.toContain('calculator-btn')
    expect(html).not.toContain('calculada-badge')
  })

  it('shows calculada time as a toggle button when solved', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ time: '18:00' })}
        isSolved={true}
        fieldId="exit-0"
        label="Saída 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('calculada-time')
    expect(html).toContain('18:00')
    expect(html).toMatch(/<button[^>]*class="calculada-time"/)
    expect(html).toContain('Desativar cálculo automático')
    expect(html).not.toContain('calculator-btn')
    expect(html).not.toContain('calculada-badge')
    expect(html).not.toContain('slot-solved')
  })

  it('shows balance badge and alternative time inline inside the calculada box', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={true}
        fieldId="exit-0"
        label="Saída 1"
        balance={adjustment()}
        alternativeTime="21:00"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('Banco -08:00')
    expect(html).toContain('Saída 1 alternativa: 21:00')
    expect(html).toContain('1.5x')
    expect(html).toContain('Limite diário de 03:00 aplicado')
    expect(html).toMatch(
      /calculada-timebox[\s\S]*calculada-time[\s\S]*calculada-time-alt alternative-clockout/,
    )
  })

  it('shows editable input for registered slots that are not solved', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ registered: true })}
        isSolved={false}
        fieldId="exit-0"
        label="Saída 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('value="18:00"')
    expect(html).toContain('clockout-time-wrap registered')
    expect(html).toContain('Marcação registrada no PontoMais')
    expect(html).not.toContain('registered-badge')
    expect(html).not.toContain('>registrada<')
    expect(html).not.toContain('calculada-time')
    expect(html).not.toContain('clockout-time-wrap calculable')
  })

  it('outlines registered slots without a text badge', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ registered: true })}
        isSolved={false}
        fieldId="exit-0"
        label="Saída 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('clockout-time-wrap registered')
    expect(html).not.toContain('registered-badge')
    expect(html).not.toContain('>registrada<')
  })

  it('structures label and badge rows for responsive wrapping', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ registered: true })}
        isSolved={true}
        fieldId="exit-0"
        label="Saída 1"
        balance={adjustment()}
        alternativeTime="21:00"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('slot-field-header')
    expect(html).toContain('slot-status-badges')
    expect(html).toContain('clockout-values')
    expect(html).toContain('clockout-values-primary')
    expect(html).toMatch(/<label[^>]*>Saída 1<\/label>/)
    expect(html).toMatch(
      /slot-field-header[\s\S]*<label[^>]*>Saída 1<\/label>[\s\S]*slot-status-badges[\s\S]*balance-badge/,
    )
    expect(html).not.toContain('calculada-badge')
    expect(html).toMatch(
      /slot-field-header[\s\S]*balance-badge[\s\S]*clockout-tooltip/,
    )
  })

  it('keeps entry and exit labels as single uninterrupted text nodes', () => {
    const entryHtml = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={false}
        fieldId="entry-0"
        label="Entrada 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )
    const exitHtml = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={true}
        fieldId="exit-0"
        label="Saída 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(entryHtml).toMatch(/<label[^>]*>Entrada 1<\/label>/)
    expect(exitHtml).toMatch(/<label[^>]*>Saída 1<\/label>/)
    expect(entryHtml).toContain('slot-field')
    expect(exitHtml).toContain('calculada-time')
    expect(exitHtml).not.toContain('calculada-badge')
  })

  it('keeps multi-digit exit labels intact beside status badge markup', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ time: '18:00' })}
        isSolved={true}
        fieldId="exit-1"
        label="Saída 2"
        balance={adjustment()}
        alternativeTime="19:43"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toMatch(/<label[^>]*for="exit-1"[^>]*>Saída 2<\/label>/)
    expect(html).toMatch(
      /slot-field-header[\s\S]*<label[^>]*>Saída 2<\/label>[\s\S]*slot-status-badges[\s\S]*balance-badge/,
    )
    expect(html).not.toMatch(/<label[^>]*>Saída<\/label>/)
    expect(html).not.toMatch(/<label[^>]*>2<\/label>/)
  })
})
