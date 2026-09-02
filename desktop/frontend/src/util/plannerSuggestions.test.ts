import { describe, expect, it } from 'vitest'
import type { Journey } from '../types'
import {
  applyInstantSuggestions,
  assignStampsToPlannerSlots,
  BUILTIN_ANCHORS,
  canRemoveJourney,
  defaultSolvedSlot,
  journeysFromPunches,
  punchJourneyFloor,
  replacePunchAtSlot,
  resolveAnchors,
  solveSlot,
  sortPunches,
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

  it('preserves other user-edited unregistered slots on an empty day', () => {
    const emptyDay = [
      journey('08:00', '12:00', false, false),
      journey('13:30', '18:00', false, false),
    ]
    const solvedSlot = defaultSolvedSlot(emptyDay)

    const { journeys: result } = applyInstantSuggestions(
      [
        journey('08:00', '11:30', false, false),
        journey('14:00', '18:00', false, false),
      ],
      solvedSlot,
      TARGET_SECS,
      { preserveSlot: { journeyIndex: 1, isEntry: true } },
    )

    expect(result[0].exit.time).toBe('11:30')
    expect(result[1].entry.time).toBe('14:00')
  })
})

describe('assignStampsToPlannerSlots', () => {
  it('fills slots in punch order so 13:18 is Saída 1', () => {
    expect(assignStampsToPlannerSlots(['09:02', '13:18', '15:45'], 4)).toEqual([
      '09:02',
      '13:18',
      '15:45',
      '',
    ])
  })

  it('puts a lone punch on the first slot', () => {
    expect(assignStampsToPlannerSlots(['13:45'], 4)).toEqual(['13:45', '', '', ''])
  })
})

describe('punchJourneyFloor', () => {
  it('keeps a minimum of two journeys', () => {
    expect(punchJourneyFloor(0)).toBe(2)
    expect(punchJourneyFloor(1)).toBe(2)
    expect(punchJourneyFloor(3)).toBe(2)
  })

  it('opens a third journey for a fifth punch', () => {
    expect(punchJourneyFloor(5)).toBe(3)
  })
})

describe('sortPunches', () => {
  it('orders times chronologically', () => {
    expect(sortPunches(['15:45', '09:02', '13:18'])).toEqual(['09:02', '13:18', '15:45'])
  })
})

describe('replacePunchAtSlot', () => {
  it('sorts after an edit so times can jump slots', () => {
    expect(replacePunchAtSlot(['09:02', '13:18', '15:45'], 1, '16:00')).toEqual([
      '09:02',
      '15:45',
      '16:00',
    ])
  })
})

describe('canRemoveJourney', () => {
  it('blocks remove at the punch floor even for empty journeys', () => {
    expect(canRemoveJourney(2, 1, false)).toBe(false)
  })

  it('blocks remove when the journey has a registered punch', () => {
    expect(canRemoveJourney(3, 3, true)).toBe(false)
  })

  it('allows remove for an extra empty journey', () => {
    expect(canRemoveJourney(3, 3, false)).toBe(true)
  })
})

describe('journeysFromPunches', () => {
  it('assigns three punches sequentially and solves Saída 2', () => {
    const { journeys, solvedSlot } = journeysFromPunches(
      ['09:02', '13:18', '15:45'],
      0,
      TARGET_SECS,
    )

    expect(journeys).toHaveLength(2)
    expect(journeys[0].entry).toEqual({ time: '09:02', registered: true })
    expect(journeys[0].exit).toEqual({ time: '13:18', registered: true })
    expect(journeys[1].entry).toEqual({ time: '15:45', registered: true })
    expect(journeys[1].exit.registered).toBe(false)
    expect(journeys[1].exit.time).toBe('19:59')
    expect(solvedSlot).toEqual({ journeyIndex: 1, isEntry: false })
  })

  it('keeps extra empty journeys above the punch floor', () => {
    const { journeys } = journeysFromPunches(['09:02', '13:18'], 1, TARGET_SECS)
    expect(journeys).toHaveLength(3)
    expect(journeys[2].entry.registered).toBe(false)
    expect(journeys[2].exit.registered).toBe(false)
  })
})
