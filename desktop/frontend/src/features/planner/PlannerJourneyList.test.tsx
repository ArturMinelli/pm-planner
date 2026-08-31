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

    expect(html).toContain('1ª jornada')
    expect(html).toContain('2ª jornada')
    expect(html).toContain('Entrada 1')
    expect(html).toContain('Saída 1')
    expect(html).toContain('Entrada 2')
    expect(html).toContain('Saída 2')
  })

  it('shows the API originals line as read-only reference', () => {
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
    expect(html).toContain('originals-line')
    expect(html).not.toContain('Adicionar marcação')
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

    expect(html).toContain('Adicionar jornada')
  })

  it('marks the solved exit journey with a calculated slot card', () => {
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

    expect(html).toContain('slot-field-card calculated')
    expect(html).toContain('calculada-time')
    expect(html).toContain('Auto')
  })

  it('disables the remove button for registered journeys', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[journey('08:00', '12:00', true), journey('13:30', '18:00')]}
        solvedSlot={{ journeyIndex: 1, isEntry: false }}
        punches={['08:00', '12:00']}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('disabled')
    expect(html).toContain('Não é possível remover uma jornada com marcações registradas')
  })

  it('disables remove below the punch floor even for empty journeys', () => {
    const html = renderToStaticMarkup(
      <PlannerJourneyList
        journeys={[journey('08:00', '12:00'), journey('13:30', '18:00')]}
        solvedSlot={{ journeyIndex: 1, isEntry: false }}
        punches={['08:00']}
        onAddJourney={noop}
        onRemoveJourney={noop}
        onUpdateJourney={noop}
        onToggleSolved={noop}
      />,
    )

    expect(html).toContain('Não é possível remover abaixo do mínimo')
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
    expect(html).toContain('Carregando marcações')
    expect(html).not.toContain('Entrada 1')
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

    expect(html).toContain('Entrada da jornada 2 (13:30) é anterior à saída da jornada 1 (14:00).')
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

    expect(html).toContain('Entrada 1')
    expect(html).not.toContain('skeleton')
    expect(html).not.toMatch(/planner-slot-time-input[^>]*disabled/)
  })
})
