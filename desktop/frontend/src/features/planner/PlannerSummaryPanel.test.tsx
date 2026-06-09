import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import PlannerSummaryPanel from './PlannerSummaryPanel'

describe('PlannerSummaryPanel', () => {
  it('shows contractual target, balance adjustment, and planned target separately', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel
        loaded={{
          date: '2026-06-09',
          baseTargetSecs: 8.5 * 60 * 60,
          targetSecs: 9.5 * 60 * 60,
          originalTimes: [],
          in1: '08:00',
          out1: '12:00',
          in2: '13:30',
          out2: '19:00',
          originalsLine: '(nenhum)',
        }}
        summary={{
          out2: '19:00',
          firstSpanSecs: 4 * 60 * 60,
          secondSpanSecs: 5.5 * 60 * 60,
          totalSpanSecs: 9.5 * 60 * 60,
          overtimeSecs: 60 * 60,
        }}
      />,
    )

    expect(html).toContain('Meta Contratual')
    expect(html).toContain('Ajuste do Banco')
    expect(html).toContain('+01:00:00')
    expect(html).toContain('Meta Planejada')
    expect(html).toContain('09:30:00')
  })
})
