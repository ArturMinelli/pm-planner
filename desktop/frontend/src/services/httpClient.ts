import type {
  AuthResult,
  NotificationPermissionStatus,
  PlannerSummary,
  PmConfig,
  PlannerAnchorsConfig,
  PlannerPayload,
  RecalculateRequest,
  ReminderPlanPayload,
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
import { appLocale } from '../i18n'

const API_BASE = '/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  })

  if (!response.ok) {
    let detail = response.statusText
    try {
      const body = (await response.json()) as { error?: string }
      if (body.error) detail = body.error
    } catch {
      // ignore parse errors
    }
    throw new Error(detail)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

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

function normalizeConfig(config: PmConfig): PmConfig {
  return {
    ...config,
    planner: normalizePlanner(config.planner) ?? builtinPlannerAnchors(),
    reminders: normalizeReminderSettings(config.reminders),
  }
}

export async function getConfig(): Promise<PmConfig> {
  const config = await request<PmConfig>('/config')
  return normalizeConfig(config)
}

export async function saveConfig(config: PmConfig): Promise<void> {
  const payload: PmConfig = {
    login: config.login,
    password: config.password,
    locale: appLocale,
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
  await request<void>('/config', {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export async function mergeAndSave(patch: Partial<PmConfig>): Promise<void> {
  const current = await getConfig()
  await saveConfig({ ...current, ...patch })
}

export async function testAuth(login: string, password: string): Promise<AuthResult> {
  return request<AuthResult>('/auth/test', {
    method: 'POST',
    body: JSON.stringify({ login, password }),
  })
}

export async function loadPlanner(date: string): Promise<PlannerPayload> {
  return request<PlannerPayload>(`/planner/${encodeURIComponent(date)}`)
}

export async function recalculateDay(recalculateRequest: RecalculateRequest): Promise<PlannerSummary> {
  return request<PlannerSummary>('/planner/recalculate', {
    method: 'POST',
    body: JSON.stringify(recalculateRequest),
  })
}

export async function getReminderStatus(): Promise<ReminderStatus> {
  const status = await request<ReminderStatus>('/reminders/status')
  return {
    ...status,
    settings: normalizeReminderSettings(status.settings),
  }
}

export async function saveReminderSettings(settings: ReminderSettings): Promise<void> {
  await request<void>('/reminders/settings', {
    method: 'PUT',
    body: JSON.stringify(normalizeReminderSettings(settings)),
  })
}

export async function requestNotificationPermission(): Promise<NotificationPermissionStatus> {
  if (!('Notification' in window)) {
    return { available: false, authorized: false }
  }
  const permission = await Notification.requestPermission()
  return {
    available: true,
    authorized: permission === 'granted',
  }
}

export async function getReminderPlan(date: string): Promise<ReminderPlanPayload> {
  const query = date ? `?date=${encodeURIComponent(date)}` : ''
  return request<ReminderPlanPayload>(`/reminders/plan${query}`)
}

export async function checkForUpdate(): Promise<UpdateStatus> {
  const status = await request<UpdateStatus>('/updates/check')
  return { ...status, blockers: status.blockers ?? [] }
}

export async function startUpdate(): Promise<void> {
  throw new Error('errors.desktopShellRequired')
}

export async function consumeUpdateResult(): Promise<UpdateResult | null> {
  return null
}

export function defaultBrowserConfig(): PmConfig {
  return {
    login: '',
    password: '',
    max_daily_extra_minutes: DEFAULT_MAX_DAILY_EXTRA_MINUTES,
    balance_credit_multiplier: DEFAULT_BALANCE_CREDIT_MULTIPLIER,
    planner: builtinPlannerAnchors(),
    reminders: defaultReminderSettings(),
  }
}
