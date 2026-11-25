import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import PlannerSummaryPanel from './PlannerSummaryPanel'

describe('PlannerSummaryPanel', () => {
  it('shows journey span rows and totals for a two-journey day', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel
        loaded={{
          date: '2026-06-09',
          baseTargetSecs: 8.5 * 60 * 60,
          originalTimes: [],
          journeys: [
            {
              entry: { time: '08:00', registered: false },
              exit: { time: '12:00', registered: false },
            },
            {
              entry: { time: '13:30', registered: false },
              exit: { time: '18:00', registered: false },
            },
          ],
          solvedSlot: { journeyIndex: 1, isEntry: false },
          originalsLine: '',
        }}
        summary={{
          journeys: [
            {
              entry: { time: '08:00', registered: false },
              exit: { time: '12:00', registered: false },
            },
            {
              entry: { time: '13:30', registered: false },
              exit: { time: '18:00', registered: false },
            },
          ],
          solvedSlot: { journeyIndex: 1, isEntry: false },
          journeySpanSecs: [4 * 60 * 60, 4.5 * 60 * 60],
          totalSpanSecs: 8.5 * 60 * 60,
          overtimeSecs: 0,
        }}
      />,
    )

    expect(html).toContain('Meta do dia')
    expect(html).toContain('1ª jornada')
    expect(html).toContain('2ª jornada')
    expect(html).toContain('Hora extra')
    expect(html).not.toContain('Ajuste do Banco')
    expect(html).not.toContain('Meta Planejada')
  })

  it('shows placeholder when no data is loaded', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel loaded={null} summary={null} />,
    )

    expect(html).toContain('Selecione uma data')
    expect(html).not.toContain('Meta do dia')
  })

  it('shows skeleton placeholders while loading', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel loading loaded={null} summary={null} />,
    )

    expect(html).toContain('skeleton')
    expect(html).toContain('Carregando resumo')
    expect(html).not.toContain('Selecione uma data')
    expect(html).not.toContain('Meta do dia')
  })

  it('shows summary stats after load', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel
        loaded={{
          date: '2026-06-09',
          baseTargetSecs: 8 * 60 * 60,
          originalTimes: [],
          journeys: [
            {
              entry: { time: '08:00', registered: false },
              exit: { time: '12:00', registered: false },
            },
            {
              entry: { time: '13:00', registered: false },
              exit: { time: '15:00', registered: false },
            },
            {
              entry: { time: '16:00', registered: false },
              exit: { time: '18:00', registered: false },
            },
          ],
          solvedSlot: { journeyIndex: 2, isEntry: false },
          originalsLine: '',
        }}
        summary={{
          journeys: [
            {
              entry: { time: '08:00', registered: false },
              exit: { time: '12:00', registered: false },
            },
            {
              entry: { time: '13:00', registered: false },
              exit: { time: '15:00', registered: false },
            },
            {
              entry: { time: '16:00', registered: false },
              exit: { time: '18:00', registered: false },
            },
          ],
          solvedSlot: { journeyIndex: 2, isEntry: false },
          journeySpanSecs: [4 * 60 * 60, 2 * 60 * 60, 2 * 60 * 60],
          totalSpanSecs: 8 * 60 * 60,
          overtimeSecs: 0,
        }}
      />,
    )

    expect(html).toContain('1ª jornada')
    expect(html).toContain('2ª jornada')
    expect(html).toContain('3ª jornada')
  })

  const day = (balance: {
    balanceSecs: number
    targetAdjustmentSecs: number
  }) => {
    const journeys = [
      { entry: { time: '08:00', registered: false }, exit: { time: '12:00', registered: false } },
      { entry: { time: '13:30', registered: false }, exit: { time: '18:00', registered: false } },
    ]
    return renderToStaticMarkup(
      <PlannerSummaryPanel
        loaded={{
          date: '2026-06-09',
          baseTargetSecs: 8.5 * 60 * 60,
          originalTimes: [],
          journeys,
          solvedSlot: { journeyIndex: 1, isEntry: false },
          originalsLine: '',
          balance: {
            appliesToday: true,
            adjustedTargetSecs: 0,
            estimatedBalanceChangeSecs: 0,
            remainingBalanceSecs: 0,
            multiplier: 1.5,
            capped: false,
            ...balance,
          },
        }}
        summary={{
          journeys,
          solvedSlot: { journeyIndex: 1, isEntry: false },
          journeySpanSecs: [4 * 60 * 60, 4.5 * 60 * 60],
          totalSpanSecs: 8.5 * 60 * 60,
          overtimeSecs: 0,
        }}
      />,
    )
  }

  it('says how much of the bank is used to leave earlier, in a soft callout', () => {
    const html = day({ balanceSecs: 4 * 60, targetAdjustmentSecs: -4 * 60 })

    expect(html).toContain('summary-callout')
    expect(html).toContain('data-tone="positive"')
    expect(html).toContain('Use 00:04 do banco para sair mais cedo.')
  })

  it('says how much is still owed and what the daily target is when the bank is negative', () => {
    const html = day({ balanceSecs: -60 * 60, targetAdjustmentSecs: 40 * 60 })

    expect(html).toContain('data-tone="debt"')
    expect(html).toContain(
      'Ainda faltam 01:00 do banco: trabalhe 00:40 a mais, além das 8h30 diárias, para repor.',
    )
  })

  it('shows no callout when the bank does not change the day', () => {
    const html = day({ balanceSecs: 0, targetAdjustmentSecs: 0 })

    expect(html).not.toContain('summary-callout')
  })
})
