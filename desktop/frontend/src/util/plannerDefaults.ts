/** Must match plan.BuiltinAnchors() in pkg/plan/assign.go */
export const BUILTIN_PLANNER_ANCHORS = {
  in1: '08:00',
  out1: '12:00',
  in2: '13:30',
  out2: '18:00',
} as const

export type PlannerAnchorTimes = {
  in1: string
  out1: string
  in2: string
  out2: string
}

export function builtinPlannerAnchors(): PlannerAnchorTimes {
  return { ...BUILTIN_PLANNER_ANCHORS }
}
