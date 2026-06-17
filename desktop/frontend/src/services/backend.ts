import type {
  ClockInStatus,
  NotificationPermissionStatus,
  PmConfig,
  PlannerAnchorsConfig,
  PlannerPayload,
  RegisterTimeCardResult,
  ReminderSettings,
  ReminderStatus,
} from '../types'
import {
  builtinPlannerAnchors,
  DEFAULT_MAX_DAILY_EXTRA_MINUTES,
} from '../util/plannerDefaults'
import { defaultReminderSettings, normalizeReminderSettings } from '../util/reminderSettings'

interface GoPmApp {
  GetConfig(): Promise<PmConfig>
  SaveConfig(c: PmConfig): Promise<void>
  TestAuth(login: string, password: string): Promise<string>
  LoadPlanner(date: string): Promise<PlannerPayload>
  GetClockInStatus(): Promise<ClockInStatus>
  RegisterTimeCard(
    latitude: number,
    longitude: number,
    accuracy: number,
    address: string,
  ): Promise<RegisterTimeCardResult>
  GetReminderStatus(): Promise<ReminderStatus>
  SaveReminderSettings(settings: ReminderSettings): Promise<void>
  RequestNotificationPermission(): Promise<NotificationPermissionStatus>
  SendTestReminder(): Promise<void>
}

function goApp(): GoPmApp | undefined {
  return (globalThis as unknown as { go?: { main?: { App: GoPmApp } } }).go
    ?.main?.App
}

export function hasWailsRuntime(): boolean {
  return !!goApp()
}

function requireRuntime(): GoPmApp {
  const a = goApp()
  if (!a) throw new Error('This page must run inside the PM desktop shell.')
  return a
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
      cache_ttl_hours: undefined,
      max_daily_extra_minutes: DEFAULT_MAX_DAILY_EXTRA_MINUTES,
      planner: builtinPlannerAnchors(),
      reminders: defaultReminderSettings(),
    }
  }
  const c = await requireRuntime().GetConfig()
  return {
    ...c,
    planner: normalizePlanner(c.planner) ?? builtinPlannerAnchors(),
    reminders: normalizeReminderSettings(c.reminders),
  }
}

export async function saveConfig(c: PmConfig): Promise<void> {
  const payload: PmConfig = {
    login: c.login,
    password: c.password,
  }
  if (
    typeof c.cache_ttl_hours === 'number' &&
    Number.isFinite(c.cache_ttl_hours) &&
    c.cache_ttl_hours > 0
  ) {
    payload.cache_ttl_hours = Math.trunc(c.cache_ttl_hours)
  }
  if (
    typeof c.max_daily_extra_minutes === 'number' &&
    Number.isFinite(c.max_daily_extra_minutes) &&
    c.max_daily_extra_minutes > 0
  ) {
    payload.max_daily_extra_minutes = Math.trunc(c.max_daily_extra_minutes)
  }
  if (c.planner) {
    payload.planner = {
      in1: c.planner.in1,
      out1: c.planner.out1,
      in2: c.planner.in2,
      out2: c.planner.out2,
    }
  }
  if (c.reminders) {
    payload.reminders = normalizeReminderSettings(c.reminders)
  }
  return requireRuntime().SaveConfig(payload)
}

/** Merges patch into the on-disk config, then saves. */
export async function mergeAndSave(patch: Partial<PmConfig>): Promise<void> {
  const current = await getConfig()
  await saveConfig({ ...current, ...patch })
}

/** Empty string means success */
export async function testAuth(
  login: string,
  password: string,
): Promise<string> {
  return requireRuntime().TestAuth(login, password)
}

export async function loadPlanner(date: string): Promise<PlannerPayload> {
  return requireRuntime().LoadPlanner(date)
}

export async function getClockInStatus(): Promise<ClockInStatus> {
  return requireRuntime().GetClockInStatus()
}

export async function registerTimeCard(
  latitude: number,
  longitude: number,
  accuracy: number,
  address: string,
): Promise<RegisterTimeCardResult> {
  return requireRuntime().RegisterTimeCard(
    latitude,
    longitude,
    accuracy,
    address,
  )
}

export async function getReminderStatus(): Promise<ReminderStatus> {
  if (!hasWailsRuntime()) {
    return {
      settings: defaultReminderSettings(),
      autostartEnabled: false,
      daemonRunning: false,
      notificationAvailable: false,
      notificationAuthorized: false,
      notificationStatusDetail: 'Modo navegador',
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

export async function sendTestReminder(): Promise<void> {
  return requireRuntime().SendTestReminder()
}
