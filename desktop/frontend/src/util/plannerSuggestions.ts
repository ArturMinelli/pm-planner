import type { Journey, SolvedSlot } from '../types'
import { addMinutes, parseHHMM } from './plannerTimes'

export const BUILTIN_ANCHORS = ['08:00', '12:00', '13:30', '18:00'] as const
export const DEFAULT_BREAK_MINUTES = 60

export const NO_SOLVED_SLOT: SolvedSlot = { journeyIndex: -1, isEntry: false }

export type SlotRef = {
  journeyIndex: number
  isEntry: boolean
}

export type InstantSuggestionOptions = {
  baseAnchors?: string[]
  breakMinutes?: number
  preserveSlot?: SlotRef
}

export function punchJourneyFloor(punchCount: number): number {
  return Math.max(2, Math.ceil(punchCount / 2))
}

export function extraJourneyCount(journeyCount: number, punchCount: number): number {
  return Math.max(0, journeyCount - punchJourneyFloor(punchCount))
}

export function canRemoveJourney(
  journeyCount: number,
  punchCount: number,
  journeyRegistered: boolean,
): boolean {
  if (journeyRegistered) return false
  return journeyCount > punchJourneyFloor(punchCount)
}

export function sortPunches(punches: string[]): string[] {
  return [...punches].sort()
}

export function assignStampsToPlannerSlots(stamps: string[], slotCount: number): string[] {
  const assigned = Array.from({ length: slotCount }, () => '')
  const limit = Math.min(stamps.length, slotCount)
  for (let index = 0; index < limit; index++) {
    assigned[index] = stamps[index]
  }
  return assigned
}

export function replacePunchAtSlot(punches: string[], slotIndex: number, time: string): string[] {
  const next = sortPunches(punches)
  if (slotIndex < 0 || slotIndex >= next.length) return next
  if (parseHHMM(time) === null) return next
  next[slotIndex] = time
  return sortPunches(next)
}

export function journeysFromPunches(
  punches: string[],
  extraCount: number,
  targetSecs: number,
  options: InstantSuggestionOptions = {},
): { journeys: Journey[]; solvedSlot: SolvedSlot } {
  const valid = punches.filter((punch) => parseHHMM(punch) !== null)
  const sorted = sortPunches(valid)
  const journeyCount = punchJourneyFloor(sorted.length) + Math.max(0, extraCount)
  const slotCount = journeyCount * 2
  const assigned = assignStampsToPlannerSlots(sorted, slotCount)
  const baseAnchors = options.baseAnchors ?? [...BUILTIN_ANCHORS]
  const breakMinutes = options.breakMinutes ?? DEFAULT_BREAK_MINUTES
  const anchors = resolveAnchors(baseAnchors, slotCount, breakMinutes, targetSecs)

  const journeys: Journey[] = []
  for (let journeyIndex = 0; journeyIndex < journeyCount; journeyIndex++) {
    const entryTime = assigned[journeyIndex * 2]
    const exitTime = assigned[journeyIndex * 2 + 1]
    journeys.push({
      entry: {
        time: entryTime || anchors[journeyIndex * 2] || '',
        registered: Boolean(entryTime),
      },
      exit: {
        time: exitTime,
        registered: Boolean(exitTime),
      },
    })
  }

  return applyInstantSuggestions(journeys, defaultSolvedSlot(journeys), targetSecs, options)
}

export function isSolvedSlotValid(slot: SolvedSlot): boolean {
  return slot.journeyIndex >= 0
}

export function isSlotSolved(slot: SolvedSlot, journeyIndex: number, isEntry: boolean): boolean {
  return slot.journeyIndex === journeyIndex && slot.isEntry === isEntry
}

export function defaultSolvedSlot(journeys: Journey[]): SolvedSlot {
  let lastSlot = NO_SOLVED_SLOT
  for (let journeyIndex = 0; journeyIndex < journeys.length; journeyIndex++) {
    const journey = journeys[journeyIndex]
    if (!journey.entry.registered) {
      lastSlot = { journeyIndex, isEntry: true }
    }
    if (!journey.exit.registered) {
      lastSlot = { journeyIndex, isEntry: false }
    }
  }
  return lastSlot
}

function parseBaseJourneySpanMinutes(base: string[]): number {
  let totalMinutes = 0
  for (let entryIndex = 0; entryIndex + 1 < base.length; entryIndex += 2) {
    const entryMinutes = parseHHMM(base[entryIndex])
    const exitMinutes = parseHHMM(base[entryIndex + 1])
    if (entryMinutes === null || exitMinutes === null) {
      continue
    }
    const spanMinutes = exitMinutes - entryMinutes
    if (spanMinutes > 0) {
      totalMinutes += spanMinutes
    }
  }
  return totalMinutes
}

export function resolveAnchors(
  base: string[],
  slotCount: number,
  breakMinutes: number,
  targetSecs: number,
): string[] {
  const result = new Array<string>(slotCount)
  const baseJourneyCount = Math.floor(base.length / 2)
  const extraJourneyCount = slotCount / 2 - baseJourneyCount
  const baseSpanMinutes = parseBaseJourneySpanMinutes(base)
  let remainingMinutes = Math.round(targetSecs / 60) - baseSpanMinutes
  if (remainingMinutes < 0) {
    remainingMinutes = 0
  }

  const remainingShareMinutes =
    extraJourneyCount > 0 ? Math.floor(remainingMinutes / extraJourneyCount) : 0

  for (let slotIndex = 0; slotIndex < slotCount; slotIndex++) {
    if (slotIndex < base.length) {
      result[slotIndex] = base[slotIndex]
      continue
    }
    if (slotIndex % 2 === 0) {
      result[slotIndex] = addMinutes(result[slotIndex - 1], breakMinutes) ?? ''
      continue
    }
    result[slotIndex] = addMinutes(result[slotIndex - 1], remainingShareMinutes) ?? ''
  }

  return result
}

function shouldApplyAnchorSuggestion(
  currentTime: string,
  resolvedAnchor: string,
  baseAnchors: string[],
  slotIndex: number,
): boolean {
  if (currentTime === '' || currentTime === resolvedAnchor) {
    return true
  }

  const configuredBaseAnchor = baseAnchors[slotIndex]
  if (configuredBaseAnchor && currentTime === configuredBaseAnchor) {
    return true
  }

  const builtinBaseAnchor = BUILTIN_ANCHORS[slotIndex]
  if (builtinBaseAnchor && currentTime === builtinBaseAnchor) {
    return true
  }

  return false
}

export function applyAnchorSuggestions(
  journeys: Journey[],
  baseAnchors: string[],
  targetSecs: number,
  breakMinutes: number,
  solvedSlot: SolvedSlot,
  preserveSlot?: SlotRef,
): Journey[] {
  const slotCount = journeys.length * 2
  const anchors = resolveAnchors(baseAnchors, slotCount, breakMinutes, targetSecs)
  const result = journeys.map((journey) => ({
    entry: { ...journey.entry },
    exit: { ...journey.exit },
  }))

  for (let journeyIndex = 0; journeyIndex < result.length; journeyIndex++) {
    const entrySlotIndex = journeyIndex * 2
    const exitSlotIndex = journeyIndex * 2 + 1
    const preserveEntry =
      preserveSlot?.journeyIndex === journeyIndex && preserveSlot.isEntry
    const preserveExit =
      preserveSlot?.journeyIndex === journeyIndex && !preserveSlot.isEntry

    const entryTime = result[journeyIndex].entry.time
    const entryAnchor = anchors[entrySlotIndex]
    if (
      !result[journeyIndex].entry.registered &&
      !isSlotSolved(solvedSlot, journeyIndex, true) &&
      !preserveEntry &&
      entryAnchor &&
      shouldApplyAnchorSuggestion(entryTime, entryAnchor, baseAnchors, entrySlotIndex)
    ) {
      result[journeyIndex].entry.time = entryAnchor
    }

    const exitTime = result[journeyIndex].exit.time
    const exitAnchor = anchors[exitSlotIndex]
    if (
      !result[journeyIndex].exit.registered &&
      !isSlotSolved(solvedSlot, journeyIndex, false) &&
      !preserveExit &&
      exitAnchor &&
      shouldApplyAnchorSuggestion(exitTime, exitAnchor, baseAnchors, exitSlotIndex)
    ) {
      result[journeyIndex].exit.time = exitAnchor
    }
  }

  return result
}

function journeySpanSeconds(journey: Journey): number {
  const entryMinutes = parseHHMM(journey.entry.time)
  const exitMinutes = parseHHMM(journey.exit.time)
  if (entryMinutes === null || exitMinutes === null) {
    return 0
  }
  const spanMinutes = exitMinutes - entryMinutes
  if (spanMinutes <= 0) {
    return 0
  }
  return spanMinutes * 60
}

export function solveSlot(
  journeys: Journey[],
  solvedSlot: SolvedSlot,
  targetSecs: number,
): Journey[] {
  if (!isSolvedSlotValid(solvedSlot) || solvedSlot.journeyIndex >= journeys.length) {
    return journeys
  }

  const solvedJourney = journeys[solvedSlot.journeyIndex]
  if (solvedSlot.isEntry && solvedJourney.entry.registered) {
    return journeys
  }
  if (!solvedSlot.isEntry && solvedJourney.exit.registered) {
    return journeys
  }

  let fixedSpanSecs = 0
  for (let journeyIndex = 0; journeyIndex < journeys.length; journeyIndex++) {
    if (journeyIndex === solvedSlot.journeyIndex) {
      continue
    }
    fixedSpanSecs += journeySpanSeconds(journeys[journeyIndex])
  }

  let remainingSecs = targetSecs - fixedSpanSecs
  if (remainingSecs < 0) {
    remainingSecs = 0
  }

  const result = journeys.map((journey) => ({
    entry: { ...journey.entry },
    exit: { ...journey.exit },
  }))

  if (solvedSlot.isEntry) {
    const exitMinutes = parseHHMM(result[solvedSlot.journeyIndex].exit.time)
    if (exitMinutes === null) {
      return journeys
    }
    const entryMinutes = exitMinutes - Math.round(remainingSecs / 60)
    const clampedEntry = Math.max(0, Math.min(exitMinutes, entryMinutes))
    result[solvedSlot.journeyIndex].entry = {
      time: formatFromMinutes(clampedEntry),
      registered: false,
    }
    return result
  }

  const entryMinutes = parseHHMM(result[solvedSlot.journeyIndex].entry.time)
  if (entryMinutes === null) {
    return journeys
  }
  const exitMinutes = entryMinutes + Math.round(remainingSecs / 60)
  result[solvedSlot.journeyIndex].exit = {
    time: formatFromMinutes(exitMinutes),
    registered: false,
  }
  return result
}

function formatFromMinutes(totalMinutes: number): string {
  const hours = Math.floor(totalMinutes / 60) % 24
  const minutes = totalMinutes % 60
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
}

export function applyInstantSuggestions(
  journeys: Journey[],
  solvedSlot: SolvedSlot,
  targetSecs: number,
  options: InstantSuggestionOptions = {},
): { journeys: Journey[]; solvedSlot: SolvedSlot } {
  const baseAnchors = options.baseAnchors ?? [...BUILTIN_ANCHORS]
  const breakMinutes = options.breakMinutes ?? DEFAULT_BREAK_MINUTES
  const activeSlot = isSolvedSlotValid(solvedSlot) ? solvedSlot : defaultSolvedSlot(journeys)
  const anchored = applyAnchorSuggestions(
    journeys,
    baseAnchors,
    targetSecs,
    breakMinutes,
    activeSlot,
    options.preserveSlot,
  )
  const solvedJourneys = solveSlot(anchored, activeSlot, targetSecs)
  return { journeys: solvedJourneys, solvedSlot: activeSlot }
}

export function seedEntryAfterExit(previousExitTime: string, breakMinutes: number): string {
  return addMinutes(previousExitTime, breakMinutes) ?? ''
}
