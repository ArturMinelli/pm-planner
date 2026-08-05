import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import type { BalanceAdjustment } from '../../types'
import PlannerSummaryPanel from './PlannerSummaryPanel'

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
          originalsLine: '(nenhum)',
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

    expect(html).toContain('Meta do Dia')
    expect(html).toContain('1ª Jornada')
    expect(html).toContain('2ª Jornada')
    expect(html).toContain('Hora Extra')
    expect(html).not.toContain('Ajuste do Banco')
    expect(html).not.toContain('Meta Planejada')
  })

  it('shows placeholder when no data is loaded', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel loaded={null} summary={null} />,
    )

    expect(html).toContain('Selecione uma data')
    expect(html).not.toContain('Meta do Dia')
  })

  it('shows skeleton placeholders while loading', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel loading loaded={null} summary={null} />,
    )

    expect(html).toContain('skeleton')
    expect(html).toContain('Carregando resumo')
    expect(html).not.toContain('Selecione uma data')
    expect(html).not.toContain('Meta do Dia')
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
          originalsLine: '(nenhum)',
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

    expect(html).toContain('1ª Jornada')
    expect(html).toContain('2ª Jornada')
    expect(html).toContain('3ª Jornada')
  })

  it('owns balance explanation and attributes alternative exit to its slot', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel
        loaded={{
          date: '2026-06-09',
          baseTargetSecs: 8.5 * 60 * 60,
          balance: adjustment({
            targetAdjustmentSecs: -45 * 60,
            estimatedBalanceChangeSecs: -45 * 60,
            remainingBalanceSecs: -7.25 * 60 * 60,
            capped: false,
          }),
          originalTimes: [],
          journeys: [],
          solvedSlot: { journeyIndex: 1, isEntry: false },
          originalsLine: '(nenhum)',
        }}
        summary={{
          journeys: [],
          solvedSlot: { journeyIndex: 1, isEntry: false },
          journeySpanSecs: [4 * 60 * 60, 4 * 60 * 60],
          totalSpanSecs: 8 * 60 * 60,
          overtimeSecs: 0,
          alternativeTime: '19:00',
        }}
      />,
    )

    expect(html).toContain('Banco de Horas')
    expect(html).toContain('Horário alternativo (Saída 2)')
    expect(html).toContain('19:00')
    expect(html).toContain('summary-explain-btn')
    expect(html).toContain('role="tooltip"')
    expect(html).toContain('Usa 00:45 do banco')
  })

  it('explains when balance is unavailable', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel
        loaded={{
          date: '2026-06-09',
          baseTargetSecs: 8 * 60 * 60,
          balanceError: 'timeout',
          originalTimes: [],
          journeys: [],
          solvedSlot: { journeyIndex: 0, isEntry: false },
          originalsLine: '(nenhum)',
        }}
        summary={{
          journeys: [],
          solvedSlot: { journeyIndex: 0, isEntry: false },
          journeySpanSecs: [8 * 60 * 60],
          totalSpanSecs: 8 * 60 * 60,
          overtimeSecs: 0,
        }}
      />,
    )

    expect(html).toContain('Banco de Horas')
    expect(html).toContain('Indisponível')
    expect(html).toContain('horário normal continua válido')
  })
})
