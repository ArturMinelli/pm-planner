import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import PlannerSummaryPanel from './PlannerSummaryPanel'

describe('PlannerSummaryPanel', () => {
  it('shows only normal-day summary rows', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel
        loaded={{
          date: '2026-06-09',
          baseTargetSecs: 8.5 * 60 * 60,
          originalTimes: [],
          in1: '08:00',
          out1: '12:00',
          in2: '13:30',
          out2: '18:00',
          originalsLine: '(nenhum)',
        }}
        summary={{
          out2: '18:00',
          alternativeOut2: '19:00',
          firstSpanSecs: 4 * 60 * 60,
          secondSpanSecs: 4.5 * 60 * 60,
          totalSpanSecs: 8.5 * 60 * 60,
          overtimeSecs: 0,
        }}
      />,
    )

    expect(html).toContain('Meta do Dia')
    expect(html).toContain('1ª Jornada')
    expect(html).toContain('2ª Jornada')
    expect(html).toContain('Hora Extra')
    expect(html).not.toContain('Ajuste do Banco')
    expect(html).not.toContain('Meta Planejada')
  })
})
