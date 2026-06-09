import type {
  BalancePayload,
  NotificationPermissionStatus,
  PmConfig,
  PlannerAnchorsConfig,
  PlannerPayload,
  ReminderSettings,
  ReminderStatus,
} from '../types'
import { builtinPlannerAnchors } from '../util/plannerDefaults'
import { defaultReminderSettings, normalizeReminderSettings } from '../util/reminderSettings'

interface GoPmApp {
  GetConfig(): Promise<PmConfig>
  SaveConfig(c: PmConfig): Promise<void>
  TestAuth(email: string, password: string): Promise<string>
  LoadPlanner(date: string): Promise<PlannerPayload>
  LoadBalance(): Promise<BalancePayload>
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
      email: '',
      password: '',
      cache_ttl_hours: undefined,
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
    email: c.email,
    password: c.password,
  }
  if (
    typeof c.cache_ttl_hours === 'number' &&
    Number.isFinite(c.cache_ttl_hours) &&
    c.cache_ttl_hours > 0
  ) {
    payload.cache_ttl_hours = Math.trunc(c.cache_ttl_hours)
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
  email: string,
  password: string,
): Promise<string> {
  return requireRuntime().TestAuth(email, password)
}

export async function loadPlanner(date: string): Promise<PlannerPayload> {
  return requireRuntime().LoadPlanner(date)
}

export async function loadBalance(): Promise<BalancePayload> {
  return requireRuntime().LoadBalance()
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
