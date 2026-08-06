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

    expect(html).toContain('Daily target')
    expect(html).toContain('1 journey')
    expect(html).toContain('2 journey')
    expect(html).toContain('Overtime')
    expect(html).not.toContain('Ajuste do Banco')
    expect(html).not.toContain('Meta Planejada')
  })

  it('shows placeholder when no data is loaded', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel loaded={null} summary={null} />,
    )

    expect(html).toContain('Select a date')
    expect(html).not.toContain('Daily target')
  })

  it('shows skeleton placeholders while loading', () => {
    const html = renderToStaticMarkup(
      <PlannerSummaryPanel loading loaded={null} summary={null} />,
    )

    expect(html).toContain('skeleton')
    expect(html).toContain('Loading summary')
    expect(html).not.toContain('Select a date')
    expect(html).not.toContain('Daily target')
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

    expect(html).toContain('1 journey')
    expect(html).toContain('2 journey')
    expect(html).toContain('3 journey')
  })
})
