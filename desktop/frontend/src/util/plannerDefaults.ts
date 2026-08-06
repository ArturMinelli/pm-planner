import type { PlannerAnchorsConfig } from '../types'
import { parseHHMM } from './plannerTimes'

/** Must match plan.BuiltinAnchors() in pkg/plan/assign.go */
export const BUILTIN_PLANNER_ANCHORS = {
  in1: '08:00',
  out1: '12:00',
  in2: '13:30',
  out2: '18:00',
} as const

export const DEFAULT_MAX_DAILY_EXTRA_MINUTES = 180
export const DEFAULT_BALANCE_CREDIT_MULTIPLIER = 1.5

export type PlannerAnchorTimes = {
  in1: string
  out1: string
  in2: string
  out2: string
}

export function builtinPlannerAnchors(): PlannerAnchorTimes {
  return { ...BUILTIN_PLANNER_ANCHORS }
}

/** Mirrors pkg/config/planner.go mergePlannerAnchors(strict=false). */
export function resolvePlannerAnchorsArray(
  planner?: PlannerAnchorsConfig,
): [string, string, string, string] {
  const builtins = builtinPlannerAnchors()
  const base: [string, string, string, string] = [
    builtins.in1,
    builtins.out1,
    builtins.in2,
    builtins.out2,
  ]
  if (!planner) {
    return base
  }

  const fields = [planner.in1, planner.out1, planner.in2, planner.out2] as const
  return base.map((fallback, index) => {
    const value = fields[index]?.trim() ?? ''
    if (!value || parseHHMM(value) === null) {
      return fallback
    }
    return value
  }) as [string, string, string, string]
}
