import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it, vi } from 'vitest'
import type { ClockSlot } from '../../types'
import PlannerSlotField from './PlannerSlotField'

const noop = vi.fn()

function slot(overrides: Partial<ClockSlot> = {}): ClockSlot {
  return { time: '18:00', registered: false, ...overrides }
}

describe('PlannerSlotField', () => {
  it('shows editable input and inactive calculator when not solved', () => {
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
    expect(html).toContain('calculator-btn')
    expect(html).not.toContain('calculator-btn active')
    expect(html).not.toContain('calculada-time')
  })

  it('shows calculada output and active calculator when solved', () => {
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
    expect(html).toContain('calculator-btn active')
    expect(html).toContain('calculada-badge')
  })

  it('does not render balance badge or alternative time on the slot', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot()}
        isSolved={true}
        fieldId="exit-0"
        label="Saída 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).not.toContain('balance-badge')
    expect(html).not.toContain('alternative-clockout')
    expect(html).not.toContain('Banco ')
    expect(html).not.toContain('clockout-tooltip')
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
    expect(html).toContain('slot-registered')
    expect(html).not.toContain('registered-badge')
    expect(html).not.toContain('calculada-time')
  })

  it('marks registered slots with field treatment instead of a badge', () => {
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

    expect(html).toContain('slot-registered')
    expect(html).not.toContain('registered-badge')
    expect(html).not.toMatch(/>registrada</)
    expect(html).toContain('Marcações registradas não podem ser calculadas')
    expect(html).toContain('disabled')
  })

  it('structures label and value rows for responsive wrapping', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ registered: true })}
        isSolved={true}
        fieldId="exit-0"
        label="Saída 1"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('clockout-label-row')
    expect(html).toContain('exit-label-actions')
    expect(html).toContain('clockout-values')
    expect(html).toMatch(/<label[^>]*>Saída 1<\/label>/)
    expect(html).toMatch(
      /exit-label-actions[\s\S]*calculada-badge/,
    )
    expect(html).not.toContain('balance-badge')
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
    expect(exitHtml).toContain('calculada-badge')
  })

  it('keeps multi-digit exit labels intact beside calculada badge markup', () => {
    const html = renderToStaticMarkup(
      <PlannerSlotField
        slot={slot({ time: '18:00' })}
        isSolved={true}
        fieldId="exit-1"
        label="Saída 2"
        onTimeChange={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toMatch(/<label[^>]*for="exit-1"[^>]*>Saída 2<\/label>/)
    expect(html).toMatch(
      /<label[^>]*>Saída 2<\/label>[\s\S]*exit-label-actions[\s\S]*calculada-badge/,
    )
    expect(html).not.toMatch(/<label[^>]*>Saída<\/label>/)
    expect(html).not.toMatch(/<label[^>]*>2<\/label>/)
  })
})
