import type { PlannerSummary } from '../../types'
import { parseHHMM } from '../../util/plannerTimes'

export const DEFAULT_TARGET_SECS = 8 * 60 * 60 + 30 * 60

export type PlannerSummaryInput = {
  targetSecs: number
  in1: string
  out1: string
  in2: string
}

function normalizedTargetSecs(targetSecs: number): number {
  return Number.isFinite(targetSecs) && targetSecs > 0
    ? Math.floor(targetSecs)
    : DEFAULT_TARGET_SECS
}

function spanSecs(start: string, end: string): number {
  const startMin = parseHHMM(start)
  const endMin = parseHHMM(end)
  if (startMin === null || endMin === null || endMin < startMin) return 0
  return (endMin - startMin) * 60
}

function formatClock(totalMinutes: number): string {
  const wrapped = ((Math.floor(totalMinutes) % 1440) + 1440) % 1440
  const hours = Math.floor(wrapped / 60)
  const minutes = wrapped % 60
  return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`
}

function computeOut2(in1: string, out1: string, in2: string, targetSecs: number): string {
  const secondStart = parseHHMM(in2)
  if (secondStart === null) return '--:--'

  const firstSpanSecs = spanSecs(in1, out1)
  const remainingSecs = Math.max(normalizedTargetSecs(targetSecs) - firstSpanSecs, 0)
  return formatClock(secondStart + Math.floor(remainingSecs / 60))
}

export function calculatePlannerSummary({
  targetSecs,
  in1,
  out1,
  in2,
}: PlannerSummaryInput): PlannerSummary {
  const target = normalizedTargetSecs(targetSecs)
  const out2 = computeOut2(in1, out1, in2, target)
  const firstSpanSecs = spanSecs(in1, out1)
  const secondSpanSecs = spanSecs(in2, out2)
  const totalSpanSecs = firstSpanSecs + secondSpanSecs

  return {
    out2,
    firstSpanSecs,
    secondSpanSecs,
    totalSpanSecs,
    overtimeSecs: Math.max(totalSpanSecs - target, 0),
  }
}
