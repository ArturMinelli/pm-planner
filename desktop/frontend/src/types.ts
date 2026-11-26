/** Mirrors pm-cli pkg/config.PlannerAnchors */
export interface PlannerAnchorsConfig {
  in1?: string
  out1?: string
  in2?: string
  out2?: string
}

export type AppLocale = 'pt-BR'

/** Mirrors pm-cli pkg/config.File */
export interface ThemePresetConfig {
  label?: string
  colors: Record<string, string>
}

export interface PmConfig {
  login: string
  password: string
  locale?: AppLocale
  balance_credit_multiplier?: number
  planner?: PlannerAnchorsConfig
  reminders?: ReminderSettings
  theme?: Record<string, string>
  default_theme?: string
  theme_name?: string
  themes?: Record<string, ThemePresetConfig>
  /** Built-in or user theme ids removed from the picker. */
  hidden_themes?: string[]
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
  lead_minutes?: number[]
  native_notifications?: boolean
}

export interface ReminderStatus {
  settings: ReminderSettings
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
  localId?: string
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

/** Mirrors specs/001-app-clock-in/contracts/clockin.md */
export interface ClockInRequest {
  location: {
    latitude: number
    longitude: number
    accuracy: number
    address?: string
    source: 'detected' | 'default'
  }
  confirmNear: boolean
}

export type ClockInCode =
  | 'invalid_location'
  | 'too_close'
  | 'session_expired'
  | 'service_rejected'
  | 'offline'
  | 'uncertain'
  | 'recovered'

export interface ClockInPunch {
  id: number
  date: string
  time: string
  address: string
  latitude: number
  longitude: number
  disabled: boolean
}

/** Returned for every outcome, including failures; `registered` is null only when unknown. */
export interface ClockInResult {
  status: 'registered' | 'failed' | 'uncertain'
  registered: boolean | null
  code?: ClockInCode
  punch?: ClockInPunch
  locationSource?: string
  detail?: string
}

export type PunchStatus = 'pending' | 'synced' | 'unconfirmed' | 'failed'

export interface PunchRecord {
  localId: string
  date: string
  time: string
  createdAt: string
  source: 'app' | 'remote'
  status: PunchStatus
  remoteId?: number
  address?: string
  latitude: number
  longitude: number
  locationSource?: string
  disabled?: boolean
  checks: number
  lastCheckedAt?: string
  syncedAt?: string
}

/** Counts are global across days. */
export interface PunchLogState {
  date: string
  punches: PunchRecord[]
  lastSyncAt?: string
  lastSyncError?: string
  pendingCount: number
  unconfirmedCount: number
}
