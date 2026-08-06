import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import PlannerDefaultsForm from './PlannerDefaultsForm'

const defaultProps = {
  anchors: { in1: '08:00', out1: '12:00', in2: '13:30', out2: '18:00' },
  maxDailyExtraHours: '3',
  balanceCreditMultiplier: '1.5',
  busy: false,
  onAnchorsChange: () => undefined,
  onMaxDailyExtraHoursChange: () => undefined,
  onBalanceCreditMultiplierChange: () => undefined,
  onSave: () => undefined,
}

describe('PlannerDefaultsForm', () => {
  it('shows the configurable daily extra-work cap', () => {
    const html = renderToStaticMarkup(<PlannerDefaultsForm {...defaultProps} />)

    expect(html).toContain('Daily extra-work limit')
    expect(html).toContain('name="max_daily_extra_hours"')
    expect(html).toContain('value="3"')
    expect(html).toContain('Save planner')
  })

  it('shows the configurable balance credit multiplier field', () => {
    const html = renderToStaticMarkup(<PlannerDefaultsForm {...defaultProps} />)

    expect(html).toContain('Time bank credit multiplier')
    expect(html).toContain('name="balance_credit_multiplier"')
    expect(html).toContain('value="1.5"')
  })

  it('renders a custom multiplier value', () => {
    const html = renderToStaticMarkup(
      <PlannerDefaultsForm {...defaultProps} balanceCreditMultiplier="2" />,
    )

    expect(html).toContain('value="2"')
  })
})
