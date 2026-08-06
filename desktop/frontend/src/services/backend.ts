import type {
  AuthResult,
  NotificationPermissionStatus,
  PlannerSummary,
  PmConfig,
  PlannerAnchorsConfig,
  PlannerPayload,
  RecalculateRequest,
  ReminderSettings,
  ReminderStatus,
  UpdateResult,
  UpdateStatus,
} from '../types'
import {
  builtinPlannerAnchors,
  DEFAULT_BALANCE_CREDIT_MULTIPLIER,
  DEFAULT_MAX_DAILY_EXTRA_MINUTES,
} from '../util/plannerDefaults'
import { defaultReminderSettings, normalizeReminderSettings } from '../util/reminderSettings'
import { browserUpdateStatus } from '../util/updateStatus'

interface GoPmApp {
  GetConfig(): Promise<PmConfig>
  SaveConfig(c: PmConfig): Promise<void>
  TestAuth(login: string, password: string): Promise<AuthResult>
  LoadPlanner(date: string): Promise<PlannerPayload>
  RecalculateDay(request: RecalculateRequest): Promise<PlannerSummary>
  GetReminderStatus(): Promise<ReminderStatus>
  SaveReminderSettings(settings: ReminderSettings): Promise<void>
  RequestNotificationPermission(): Promise<NotificationPermissionStatus>
  CheckForUpdate(): Promise<UpdateStatus>
  StartUpdate(): Promise<void>
  ConsumeUpdateResult(): Promise<UpdateResult | null>
}

function goApp(): GoPmApp | undefined {
  return (globalThis as unknown as { go?: { main?: { App: GoPmApp } } }).go
    ?.main?.App
}

export function hasWailsRuntime(): boolean {
  return !!goApp()
}

function requireRuntime(): GoPmApp {
  const app = goApp()
  if (!app) throw new Error('errors.desktopShellRequired')
  return app
}

/** Safe for browser/dev preview — returns mock mode flag */
function normalizePlanner(
  planner?: PlannerAnchorsConfig,
): PlannerAnchorsConfig | undefined {
  if (!planner) return undefined
  const builtins = builtinPlannerAnchors()
  return {
    in1: planner.in1?.trim() || builtins.in1,
    out1: planner.out1?.trim() || builtins.out1,
    in2: planner.in2?.trim() || builtins.in2,
    out2: planner.out2?.trim() || builtins.out2,
  }
}

export async function getConfig(): Promise<PmConfig> {
  if (!hasWailsRuntime()) {
    return {
      login: '',
      password: '',
      max_daily_extra_minutes: DEFAULT_MAX_DAILY_EXTRA_MINUTES,
      balance_credit_multiplier: DEFAULT_BALANCE_CREDIT_MULTIPLIER,
      planner: builtinPlannerAnchors(),
      reminders: defaultReminderSettings(),
    }
  }
  const config = await requireRuntime().GetConfig()
  return {
    ...config,
    planner: normalizePlanner(config.planner) ?? builtinPlannerAnchors(),
    reminders: normalizeReminderSettings(config.reminders),
  }
}

export async function saveConfig(config: PmConfig): Promise<void> {
  const payload: PmConfig = {
    login: config.login,
    password: config.password,
    locale: config.locale,
  }
  if (
    typeof config.max_daily_extra_minutes === 'number' &&
    Number.isFinite(config.max_daily_extra_minutes) &&
    config.max_daily_extra_minutes > 0
  ) {
    payload.max_daily_extra_minutes = Math.trunc(config.max_daily_extra_minutes)
  }
  if (
    typeof config.balance_credit_multiplier === 'number' &&
    Number.isFinite(config.balance_credit_multiplier) &&
    config.balance_credit_multiplier > 0
  ) {
    payload.balance_credit_multiplier = config.balance_credit_multiplier
  }
  if (config.planner) {
    payload.planner = {
      in1: config.planner.in1,
      out1: config.planner.out1,
      in2: config.planner.in2,
      out2: config.planner.out2,
    }
  }
  if (config.reminders) {
    payload.reminders = normalizeReminderSettings(config.reminders)
  }
  return requireRuntime().SaveConfig(payload)
}

/** Merges patch into the on-disk config, then saves. */
export async function mergeAndSave(patch: Partial<PmConfig>): Promise<void> {
  const current = await getConfig()
  await saveConfig({ ...current, ...patch })
}

export async function testAuth(
  login: string,
  password: string,
): Promise<AuthResult> {
  return requireRuntime().TestAuth(login, password)
}

export async function loadPlanner(date: string): Promise<PlannerPayload> {
  return requireRuntime().LoadPlanner(date)
}

export async function recalculateDay(request: RecalculateRequest): Promise<PlannerSummary> {
  return requireRuntime().RecalculateDay(request)
}

export async function getReminderStatus(): Promise<ReminderStatus> {
  if (!hasWailsRuntime()) {
    return {
      settings: defaultReminderSettings(),
      autostartEnabled: false,
      daemonRunning: false,
      notificationAvailable: false,
      notificationAuthorized: false,
      notificationStatusDetail: 'settings.reminders.browserDetail',
    }
  }
  const status = await requireRuntime().GetReminderStatus()
  return {
    ...status,
    settings: normalizeReminderSettings(status.settings),
  }
}

export async function saveReminderSettings(
  settings: ReminderSettings,
): Promise<void> {
  return requireRuntime().SaveReminderSettings(
    normalizeReminderSettings(settings),
  )
}

export async function requestNotificationPermission(): Promise<NotificationPermissionStatus> {
  return requireRuntime().RequestNotificationPermission()
}

export async function checkForUpdate(): Promise<UpdateStatus> {
  if (!hasWailsRuntime()) return browserUpdateStatus()
  const status = await requireRuntime().CheckForUpdate()
  // Go marshals an empty blocker slice as null.
  return { ...status, blockers: status.blockers ?? [] }
}

/** Closes the app: the update runs detached and reopens it when it finishes. */
export async function startUpdate(): Promise<void> {
  return requireRuntime().StartUpdate()
}

/** Returns the outcome of an update that ran while the app was closed, once. */
export async function consumeUpdateResult(): Promise<UpdateResult | null> {
  if (!hasWailsRuntime()) return null
  return (await requireRuntime().ConsumeUpdateResult()) ?? null
}
