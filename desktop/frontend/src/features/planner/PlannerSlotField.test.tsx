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
  it('shows editable input in a calculable card when not solved', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={false}
        fieldId="exit-0"
        label="Exit 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('Exit 1')
    expect(html).toContain('value="18:00"')
    expect(html).toContain('slot-field-card calculable')
    expect(html).toContain('Clique neste campo para calcular pela meta do dia')
    expect(html).toContain('slot-mode-indicator calculator')
    expect(html).not.toContain('calculada-time')
    expect(html).not.toContain('calculada-badge')
  })

  it('shows calculated time inside an interactive card when solved', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ time: '18:00' })}
        isSolved={true}
        fieldId="exit-0"
        label="Exit 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('slot-field-card calculated')
    expect(html).toContain('calculada-time')
    expect(html).toContain('18:00')
    expect(html).toContain('Auto')
    expect(html).toContain('Desativar cálculo automático')
    expect(html).not.toContain('slot-mode-indicator calculator')
    expect(html).not.toContain('calculada-badge')
    expect(html).not.toContain('slot-solved')
  })

  it('shows alternative time with balance tooltip inside the calculated card', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={true}
        fieldId="exit-0"
        label="Exit 1"
        balance={adjustment()}
        alternativeTime="21:00"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).not.toContain('Banco -08:00')
    expect(html).toContain('Exit 1 alternativa: 21:00')
    expect(html).toContain('1.5x')
    expect(html).toContain('Limite diário de 03:00 aplicado')
    expect(html).toMatch(
      /slot-field-card calculated[\s\S]*calculada-time[\s\S]*calculada-time-alt alternative-clockout[\s\S]*clockout-tooltip/,
    )
  })

  it('shows editable input for registered slots that are not solved', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ registered: true })}
        isSolved={false}
        fieldId="exit-0"
        label="Exit 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('value="18:00"')
    expect(html).toContain('slot-field-card registered')
    expect(html).toContain('Marcação registrada no PontoMais')
    expect(html).not.toContain('registered-badge')
    expect(html).not.toContain('>registrada<')
    expect(html).not.toContain('calculada-time')
    expect(html).not.toContain('slot-field-card calculable')
  })

  it('outlines registered slots without a text badge', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ registered: true })}
        isSolved={false}
        fieldId="exit-0"
        label="Exit 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('slot-field-card registered')
    expect(html).not.toContain('registered-badge')
    expect(html).not.toContain('>registrada<')
  })

  it('structures label and tooltip rows for responsive wrapping', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ registered: true })}
        isSolved={true}
        fieldId="exit-0"
        label="Exit 1"
        balance={adjustment()}
        alternativeTime="21:00"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('slot-field-header')
    expect(html).not.toContain('slot-status-badges')
    expect(html).not.toContain('balance-badge')
    expect(html).toContain('clockout-values')
    expect(html).toContain('clockout-values-primary')
    expect(html).toMatch(/<label[^>]*>Exit 1<\/label>/)
    expect(html).not.toContain('calculada-badge')
    expect(html).toMatch(
      /calculada-time-alt alternative-clockout[\s\S]*clockout-tooltip/,
    )
  })

  it('keeps entry and exit labels as single uninterrupted text nodes', () => {
    const entryHtml = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={false}
        fieldId="entry-0"
        label="Entry 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )
    const exitHtml = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={true}
        fieldId="exit-0"
        label="Exit 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(entryHtml).toMatch(/<span[^>]*>Entry 1<\/span>/)
    expect(exitHtml).toMatch(/<label[^>]*>Exit 1<\/label>/)
    expect(entryHtml).toContain('slot-field')
    expect(exitHtml).toContain('calculada-time')
    expect(exitHtml).not.toContain('calculada-badge')
  })

  it('keeps multi-digit exit labels intact beside alternative time markup', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ time: '18:00' })}
        isSolved={true}
        fieldId="exit-1"
        label="Exit 2"
        balance={adjustment()}
        alternativeTime="19:43"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toMatch(/<label[^>]*for="exit-1"[^>]*>Exit 2<\/label>/)
    expect(html).toMatch(
      /slot-field-header[\s\S]*<label[^>]*>Exit 2<\/label>[\s\S]*calculada-time-alt alternative-clockout/,
    )
    expect(html).not.toMatch(/<label[^>]*>Exit<\/label>/)
    expect(html).not.toMatch(/<label[^>]*>2<\/label>/)
  })
})
