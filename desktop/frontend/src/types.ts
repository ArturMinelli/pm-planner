/** Mirrors pm-cli pkg/config.PlannerAnchors */
export interface PlannerAnchorsConfig {
  in1?: string
  out1?: string
  in2?: string
  out2?: string
}

export type AppLocale = 'en' | 'pt-BR'

/** Mirrors pm-cli pkg/config.File */
export interface PmConfig {
  login: string
  password: string
  locale?: AppLocale
  max_daily_extra_minutes?: number
  balance_credit_multiplier?: number
  planner?: PlannerAnchorsConfig
  reminders?: ReminderSettings
}

export interface LocalizedMessage {
  key: string
  params?: Record<string, string>
}

export interface AuthResult {
  ok: boolean
  error?: LocalizedMessage
}

export interface ClockSlot {
  time: string
  registered: boolean
}

export interface SolvedSlot {
  journeyIndex: number
  isEntry: boolean
}

export interface Journey {
  entry: ClockSlot
  exit: ClockSlot
}

/** Mirrors pkg/app payloads */
export interface PlannerPayload {
  date: string
  baseTargetSecs: number
  balance?: BalanceAdjustment
  balanceError?: LocalizedMessage
  originalTimes: string[]
  journeys: Journey[]
  solvedSlot: SolvedSlot
  originalsLine: string
  loadWarning?: LocalizedMessage
}

export interface PlannerSummary {
  journeys: Journey[]
  solvedSlot: SolvedSlot
  journeySpanSecs: number[]
  totalSpanSecs: number
  overtimeSecs: number
  alternativeTime?: string
  warnings?: LocalizedMessage[]
}

export interface RecalculateRequest {
  date: string
  baseTargetSecs: number
  balance?: BalanceAdjustment
  journeys: Journey[]
  solvedSlot: SolvedSlot
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

export interface ScheduledReminder {
  id: string
  date: string
  slotKey: string
  slotLabel: string
  slotTime: string
  leadMinutes: number
  fireAt: string
}

export interface ReminderPlanPayload {
  schedule: ScheduledReminder[]
}

export interface NotificationPermissionStatus {
  available: boolean
  authorized: boolean
  detail?: string
}

/** Mirrors pm-cli pkg/update.Status */
export interface UpdateStatus {
  root: string
  /** False for tarball installs, where no commit comparison is possible. */
  isGit: boolean
  commitSha: string
  commitDate: string
  behind: number
  dirty: boolean
  blockers: LocalizedMessage[]
  updateAvailable: boolean
}

/** Mirrors pm-cli pkg/update.Result */
export interface UpdateResult {
  ok: boolean
  message: LocalizedMessage
  logPath: string
  commit: string
  finishedAt: string
}
