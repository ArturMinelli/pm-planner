import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import PlannerDefaultsForm from './PlannerDefaultsForm'

describe('PlannerDefaultsForm', () => {
  it('shows the configurable daily extra-work cap', () => {
    const html = renderToStaticMarkup(
      <PlannerDefaultsForm
        anchors={{ in1: '08:00', out1: '12:00', in2: '13:30', out2: '18:00' }}
        maxDailyExtraHours="1.5"
        busy={false}
        onAnchorsChange={() => undefined}
        onMaxDailyExtraHoursChange={() => undefined}
        onSave={() => undefined}
      />,
    )

    expect(html).toContain('Limite Diário de Trabalho Extra')
    expect(html).toContain('name="max_daily_extra_hours"')
    expect(html).toContain('value="1.5"')
    expect(html).toContain('Salvar Planner')
  })
})
