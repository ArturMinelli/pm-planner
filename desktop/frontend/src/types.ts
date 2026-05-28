/** Mirrors pm-cli pkg/config.PlannerAnchors */
export interface PlannerAnchorsConfig {
  in1?: string
  out1?: string
  in2?: string
  out2?: string
}

/** Mirrors pm-cli pkg/config.File */
export interface PmConfig {
  email: string
  password: string
  cache_ttl_hours?: number
  planner?: PlannerAnchorsConfig
}

/** Mirrors pkg/app payloads */
export interface PlannerPayload {
  date: string
  targetSecs: number
  originalTimes: string[]
  in1: string
  out1: string
  in2: string
  out2: string
  originalsLine: string
}

export interface PlannerSummary {
  out2: string
  firstSpanSecs: number
  secondSpanSecs: number
  totalSpanSecs: number
  overtimeSecs: number
}
