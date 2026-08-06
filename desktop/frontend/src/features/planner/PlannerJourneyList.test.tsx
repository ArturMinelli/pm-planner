import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it, vi } from 'vitest'
import type { Journey } from '../../types'
import PlannerJourneyList from './PlannerJourneyList'

const noop = vi.fn()

function journey(
  entryTime: string,
  exitTime: string,
  registered = false,
): Journey {
  return {
    entry: { time: entryTime, registered },
    exit: { time: exitTime, registered },
  }
}

describe('PlannerJourneyList', () => {
  it('renders a journey group for each journey', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[journey('08:00', '12:00'), journey('13:30', '18:00')]}
        solvedSlot={{ journeyIndex: 1, isEntry: false }}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('1 journey')
    expect(html).toContain('2 journey')
    expect(html).toContain('Entry 1')
    expect(html).toContain('Exit 1')
    expect(html).toContain('Entry 2')
    expect(html).toContain('Exit 2')
  })

  it('shows the originals line when provided', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[journey('08:00', '18:00')]}
        solvedSlot={{ journeyIndex: 0, isEntry: false }}
        originalsLine="08:00 — 18:00"
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('08:00 — 18:00')
  })

  it('shows each journey on its own originals line', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[journey('08:00', '12:00'), journey('13:00', '18:00')]}
        solvedSlot={{ journeyIndex: 1, isEntry: false }}
        originalsLine={'08:00 — 12:00\n13:00 — 18:00'}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('08:00 — 12:00')
    expect(html).toContain('13:00 — 18:00')
  })

  it('shows the add journey button', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[]}
        solvedSlot={{ journeyIndex: -1, isEntry: false }}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('Add journey')
  })

  it('marks the solved exit journey with an active calculator button', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[journey('08:00', '12:00'), journey('13:30', '18:00')]}
        solvedSlot={{ journeyIndex: 1, isEntry: false }}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('calculada-time')
    expect(html).toMatch(/<button[^>]*class="calculada-time"/)
  })

  it('disables the remove button for registered journeys', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[journey('08:00', '12:00', true)]}
        solvedSlot={{ journeyIndex: -1, isEntry: false }}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('disabled')
    expect(html).toContain('Cannot remove')
  })

  it('shows skeleton placeholders while loading', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        loading
        disabled
        journeys={[]}
        solvedSlot={{ journeyIndex: -1, isEntry: false }}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('skeleton')
    expect(html).toContain('Loading stamps')
    expect(html).not.toContain('Entry 1')
    expect(html).toContain('disabled')
  })

  it('shows inline ordering warnings when present', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[journey('08:00', '14:00'), journey('13:30', '18:00')]}
        solvedSlot={{ journeyIndex: 1, isEntry: false }}
        summary={{
          journeys: [],
          solvedSlot: { journeyIndex: 1, isEntry: false },
          journeySpanSecs: [],
          totalSpanSecs: 0,
          overtimeSecs: 0,
          warnings: [{
            key: 'errors.planner.journey_entry_before_exit',
            params: {
              journey: '2',
              entry: '13:30',
              prevJourney: '1',
              prevExit: '14:00',
            },
          }],
        }}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('Journey 2 entry 13:30 is before journey 1 exit 14:00')
    expect(html).toContain('planner-warning')
  })

  it('renders editable journey fields after load', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        loading={false}
        disabled={false}
        journeys={[journey('08:00', '12:00')]}
        solvedSlot={{ journeyIndex: 0, isEntry: false }}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('Entry 1')
    expect(html).not.toContain('skeleton')
    expect(html).not.toContain('disabled=""')
  })
})
