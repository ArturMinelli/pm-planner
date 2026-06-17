/** Mirrors pm-cli pkg/config.PlannerAnchors */
export interface PlannerAnchorsConfig {
  in1?: string
  out1?: string
  in2?: string
  out2?: string
}

/** Mirrors pm-cli pkg/config.File */
export interface PmConfig {
  login: string
  password: string
  cache_ttl_hours?: number
  max_daily_extra_minutes?: number
  planner?: PlannerAnchorsConfig
  reminders?: ReminderSettings
}

/** Mirrors pkg/app payloads */
export interface PlannerPayload {
  date: string
  baseTargetSecs: number
  balance?: BalanceAdjustment
  balanceError?: string
  originalTimes: string[]
  in1: string
  out1: string
  in2: string
  out2: string
  originalsLine: string
}

export interface PlannerSummary {
  out2: string
  alternativeOut2?: string
  firstSpanSecs: number
  secondSpanSecs: number
  totalSpanSecs: number
  overtimeSecs: number
}

export interface BalanceAdjustment {
  balanceSecs: number
  appliesToday: boolean
  targetAdjustmentSecs: number
  adjustedTargetSecs: number
  estimatedBalanceChangeSecs: number
  remainingBalanceSecs: number
  multiplier: number
  capped: boolean
}

export interface ReminderSettings {
  enabled: boolean
  autostart: boolean
  lead_minutes?: number[]
  native_notifications?: boolean
}

export interface ReminderStatus {
  settings: ReminderSettings
  autostartEnabled: boolean
  daemonRunning: boolean
  notificationAvailable: boolean
  notificationAuthorized: boolean
  notificationStatusDetail?: string
}

export interface NotificationPermissionStatus {
  available: boolean
  authorized: boolean
  detail?: string
}

export interface ClockInStatus {
  date: string
  time: string
  address?: string
  hasRecent: boolean
}

export interface RegisterTimeCardResult {
  time: string
  date: string
  address: string
}
