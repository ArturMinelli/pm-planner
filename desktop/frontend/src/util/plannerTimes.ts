import type { Journey } from '../types'

export const STEP_MINUTES = 15
export const MIN_GAP_MINUTES = 15
export const DAY_START = 0
export const DAY_END = 23 * 60 + 45

export type PlannerTimes = { in1: string; out1: string; in2: string }
export type PlannerTimeField = keyof PlannerTimes

export type PlannerAnchorTimes = {
  in1: string
  out1: string
  in2: string
  out2: string
}

const hhmmRe = /^(\d{2}):(\d{2})$/

export const ANCHOR_FIELD_KEYS = ['in1', 'out1', 'in2', 'out2'] as const
export const ANCHOR_LABEL_KEYS = [
  'planner.entry',
  'planner.exit',
  'planner.entry',
  'planner.exit',
] as const

export function anchorLabelKey(index: number): string {
  return ANCHOR_LABEL_KEYS[index]
}

export function anchorLabelParams(index: number): Record<string, string> {
  const number = index < 2 ? '1' : '2'
  return { number }
}

export function parseHHMM(value: string): number | null {
  const match = hhmmRe.exec(value.trim())
  if (!match) return null
  const hours = Number(match[1])
  const minutes = Number(match[2])
  if (hours < 0 || hours > 23 || minutes < 0 || minutes > 59) return null
  return hours * 60 + minutes
}

export function formatMinutes(minutes: number): string {
  const clamped = Math.max(DAY_START, Math.min(DAY_END, minutes))
  const hours = Math.floor(clamped / 60)
  const mins = clamped % 60
  return `${String(hours).padStart(2, '0')}:${String(mins).padStart(2, '0')}`
}

export function addMinutes(hhmm: string, delta: number): string | null {
  const base = parseHHMM(hhmm)
  if (base === null) return null
  return formatMinutes(base + delta)
}

function clampMinutes(minutes: number): number {
  return Math.max(DAY_START, Math.min(DAY_END, minutes))
}

function forwardCascade(
  in1: number,
  out1: number,
  in2: number,
): [number, number, number] {
  let firstIn = in1
  let firstOut = out1
  let secondIn = in2

  if (firstOut < firstIn + MIN_GAP_MINUTES) firstOut = firstIn + MIN_GAP_MINUTES
  if (secondIn < firstOut + MIN_GAP_MINUTES) secondIn = firstOut + MIN_GAP_MINUTES

  if (secondIn > DAY_END) {
    secondIn = DAY_END
    if (firstOut > secondIn - MIN_GAP_MINUTES) firstOut = secondIn - MIN_GAP_MINUTES
    if (firstIn > firstOut - MIN_GAP_MINUTES) firstIn = firstOut - MIN_GAP_MINUTES
    firstIn = clampMinutes(firstIn)
    firstOut = clampMinutes(firstOut)
    if (firstOut < firstIn + MIN_GAP_MINUTES) firstOut = Math.min(DAY_END, firstIn + MIN_GAP_MINUTES)
    if (secondIn < firstOut + MIN_GAP_MINUTES) secondIn = Math.min(DAY_END, firstOut + MIN_GAP_MINUTES)
  }

  return [firstIn, firstOut, secondIn]
}

export function enforcePlannerOrder(times: PlannerTimes): PlannerTimes {
  const in1 = parseHHMM(times.in1)
  const out1 = parseHHMM(times.out1)
  const in2 = parseHHMM(times.in2)

  if (in1 === null || out1 === null || in2 === null) return times

  const [firstIn, firstOut, secondIn] = forwardCascade(in1, out1, in2)
  return {
    in1: formatMinutes(firstIn),
    out1: formatMinutes(firstOut),
    in2: formatMinutes(secondIn),
  }
}

export function adjustPlannerTime(
  times: PlannerTimes,
  field: PlannerTimeField,
  deltaMinutes: number,
): PlannerTimes {
  const current = times[field]
  const next = addMinutes(current, deltaMinutes)
  if (next === null) return times

  const updated = { ...times, [field]: next }
  return enforcePlannerOrder(updated)
}

function forwardCascadeFour(
  in1: number,
  out1: number,
  in2: number,
  out2: number,
): [number, number, number, number] {
  let firstIn = in1
  let firstOut = out1
  let secondIn = in2
  let secondOut = out2

  if (firstOut < firstIn + MIN_GAP_MINUTES) firstOut = firstIn + MIN_GAP_MINUTES
  if (secondIn < firstOut + MIN_GAP_MINUTES) secondIn = firstOut + MIN_GAP_MINUTES
  if (secondOut < secondIn + MIN_GAP_MINUTES) secondOut = secondIn + MIN_GAP_MINUTES

  if (secondOut > DAY_END) {
    secondOut = DAY_END
    if (secondIn > secondOut - MIN_GAP_MINUTES) secondIn = secondOut - MIN_GAP_MINUTES
    if (firstOut > secondIn - MIN_GAP_MINUTES) firstOut = secondIn - MIN_GAP_MINUTES
    if (firstIn > firstOut - MIN_GAP_MINUTES) firstIn = firstOut - MIN_GAP_MINUTES
    firstIn = clampMinutes(firstIn)
    firstOut = clampMinutes(firstOut)
    secondIn = clampMinutes(secondIn)
    if (firstOut < firstIn + MIN_GAP_MINUTES) firstOut = Math.min(DAY_END, firstIn + MIN_GAP_MINUTES)
    if (secondIn < firstOut + MIN_GAP_MINUTES) secondIn = Math.min(DAY_END, firstOut + MIN_GAP_MINUTES)
    if (secondOut < secondIn + MIN_GAP_MINUTES) secondOut = Math.min(DAY_END, secondIn + MIN_GAP_MINUTES)
  }

  return [firstIn, firstOut, secondIn, secondOut]
}

export function enforceAnchorOrder(times: PlannerAnchorTimes): PlannerAnchorTimes {
  const in1 = parseHHMM(times.in1)
  const out1 = parseHHMM(times.out1)
  const in2 = parseHHMM(times.in2)
  const out2 = parseHHMM(times.out2)

  if (in1 === null || out1 === null || in2 === null || out2 === null) return times

  const [firstIn, firstOut, secondIn, secondOut] = forwardCascadeFour(in1, out1, in2, out2)
  return {
    in1: formatMinutes(firstIn),
    out1: formatMinutes(firstOut),
    in2: formatMinutes(secondIn),
    out2: formatMinutes(secondOut),
  }
}

export function adjustPlannerAnchor(
  times: PlannerAnchorTimes,
  field: keyof PlannerAnchorTimes,
  deltaMinutes: number,
): PlannerAnchorTimes {
  const next = addMinutes(times[field], deltaMinutes)
  if (next === null) return times
  return enforceAnchorOrder({ ...times, [field]: next })
}

export function enforceJourneyOrder(
  journeys: Journey[],
  _changedIndex: number,
  _isEntry: boolean,
): Journey[] {
  const result = journeys.map((journey) => ({ ...journey }))

  for (let journeyIndex = 0; journeyIndex < result.length; journeyIndex++) {
    const entryMinutes = parseHHMM(result[journeyIndex].entry.time)
    const exitMinutes = parseHHMM(result[journeyIndex].exit.time)

    if (entryMinutes !== null && exitMinutes !== null && exitMinutes < entryMinutes + MIN_GAP_MINUTES) {
      result[journeyIndex] = {
        ...result[journeyIndex],
        exit: {
          ...result[journeyIndex].exit,
          time: formatMinutes(entryMinutes + MIN_GAP_MINUTES),
        },
      }
    }
  }

  for (let journeyIndex = 1; journeyIndex < result.length; journeyIndex++) {
    const previousExitMinutes = parseHHMM(result[journeyIndex - 1].exit.time)
    const currentEntryMinutes = parseHHMM(result[journeyIndex].entry.time)

    if (
      previousExitMinutes !== null &&
      currentEntryMinutes !== null &&
      currentEntryMinutes < previousExitMinutes + MIN_GAP_MINUTES
    ) {
      result[journeyIndex] = {
        ...result[journeyIndex],
        entry: {
          ...result[journeyIndex].entry,
          time: formatMinutes(previousExitMinutes + MIN_GAP_MINUTES),
        },
      }
    }
  }

  return result
}

export type PlannerAnchorValidationError =
  | { key: 'errors.validation.invalidAnchorTime'; params: { label: string } }
  | { key: 'errors.validation.anchorsNotIncreasing' }
  | { key: 'errors.validation.anchorsMinGap'; params: { minutes: string } }

export function validatePlannerAnchors(times: PlannerAnchorTimes): PlannerAnchorValidationError | null {
  const keys = ANCHOR_FIELD_KEYS
  const minutes: number[] = []

  for (let index = 0; index < keys.length; index++) {
    const parsed = parseHHMM(times[keys[index]])
    if (parsed === null) {
      return {
        key: 'errors.validation.invalidAnchorTime',
        params: { label: `${index + 1}` },
      }
    }
    minutes.push(parsed)
  }

  for (let index = 1; index < minutes.length; index++) {
    if (minutes[index] <= minutes[index - 1]) {
      return { key: 'errors.validation.anchorsNotIncreasing' }
    }
    if (minutes[index] - minutes[index - 1] < MIN_GAP_MINUTES) {
      return {
        key: 'errors.validation.anchorsMinGap',
        params: { minutes: String(MIN_GAP_MINUTES) },
      }
    }
  }

  return null
}
