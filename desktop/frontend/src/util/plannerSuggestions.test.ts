import { describe, expect, it } from 'vitest'
import type { Journey } from '../types'
import {
  applyInstantSuggestions,
  BUILTIN_ANCHORS,
  defaultSolvedSlot,
  resolveAnchors,
  solveSlot,
} from './plannerSuggestions'

const TARGET_SECS = 8 * 60 * 60 + 30 * 60

function journey(
  entryTime: string,
  exitTime: string,
  entryRegistered = false,
  exitRegistered = false,
): Journey {
  return {
    entry: { time: entryTime, registered: entryRegistered },
    exit: { time: exitTime, registered: exitRegistered },
  }
}

describe('resolveAnchors', () => {
  it('returns base anchors for two journeys', () => {
    const anchors = resolveAnchors([...BUILTIN_ANCHORS], 4, 60, TARGET_SECS)
    expect(anchors).toEqual(['08:00', '12:00', '13:30', '18:00'])
  })

  it('derives third journey anchors from break and remaining target', () => {
    const anchors = resolveAnchors([...BUILTIN_ANCHORS], 6, 60, 10 * 60 * 60)
    expect(anchors).toEqual(['08:00', '12:00', '13:30', '18:00', '19:00', '20:30'])
  })
})

describe('solveSlot', () => {
  it('computes exit from remaining target for two journeys', () => {
    const journeys = [
      journey('08:00', '12:00', true, true),
      journey('13:30', '', true, false),
    ]
    const result = solveSlot(journeys, { journeyIndex: 1, isEntry: false }, TARGET_SECS)
    expect(result[1].exit.time).toBe('18:00')
  })

  it('computes entry from remaining target for two journeys', () => {
    const journeys = [
      journey('08:00', '12:00', true, true),
      journey('', '18:00', false, false),
    ]
    const result = solveSlot(journeys, { journeyIndex: 1, isEntry: true }, TARGET_SECS)
    expect(result[1].entry.time).toBe('13:30')
  })

  it('computes exit for three journeys', () => {
    const journeys = [
      journey('08:00', '12:00', true, true),
      journey('13:00', '14:00', true, true),
      journey('15:00', '', true, false),
    ]
    const result = solveSlot(journeys, { journeyIndex: 2, isEntry: false }, 8 * 60 * 60)
    expect(result[2].exit.time).toBe('18:00')
  })

  it('computes entry for three journeys', () => {
    const journeys = [
      journey('08:00', '12:00', true, true),
      journey('13:00', '14:00', true, true),
      journey('', '18:00', false, false),
    ]
    const result = solveSlot(journeys, { journeyIndex: 2, isEntry: true }, 8 * 60 * 60)
    expect(result[2].entry.time).toBe('15:00')
  })
})

describe('defaultSolvedSlot', () => {
  it('picks the last unregistered exit', () => {
    const slot = defaultSolvedSlot([
      journey('08:00', '12:00', true, true),
      journey('13:30', '', true, false),
    ])
    expect(slot).toEqual({ journeyIndex: 1, isEntry: false })
  })

  it('picks the last unregistered entry when it is later', () => {
    const slot = defaultSolvedSlot([
      journey('08:00', '12:00', true, true),
      journey('', '18:00', false, true),
    ])
    expect(slot).toEqual({ journeyIndex: 1, isEntry: true })
  })
})

describe('applyInstantSuggestions', () => {
  it('applies anchors then solves the active exit', () => {
    const { journeys: result, solvedSlot } = applyInstantSuggestions(
      [
        journey('08:00', '12:00', true, true),
        journey('13:30', '', true, false),
      ],
      { journeyIndex: 1, isEntry: false },
      TARGET_SECS,
    )

    expect(solvedSlot).toEqual({ journeyIndex: 1, isEntry: false })
    expect(result[1].exit.time).toBe('18:00')
  })

  it('preserves the edited slot while refreshing other unregistered entries', () => {
    const { journeys: result } = applyInstantSuggestions(
      [
        journey('08:00', '12:00', true, true),
        journey('14:00', '', false, false),
      ],
      { journeyIndex: 1, isEntry: false },
      TARGET_SECS,
      { preserveSlot: { journeyIndex: 1, isEntry: true } },
    )

    expect(result[1].entry.time).toBe('14:00')
    expect(result[1].exit.time).toBe('18:30')
  })
})
