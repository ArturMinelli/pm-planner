import type { ReminderAnimation, ReminderSettings } from '../types'

export const DEFAULT_REMINDER_LEADS = [15, 5] as const
export const REMINDER_ANIMATIONS: ReminderAnimation[] = ['train', 'rocket', 'bell']

export function defaultReminderSettings(): ReminderSettings {
  return {
    enabled: false,
    autostart: false,
    lead_minutes: [...DEFAULT_REMINDER_LEADS],
    animation: 'train',
    native_notifications: true,
    popup_notifications: true,
  }
}

export function normalizeReminderSettings(
  settings?: Partial<ReminderSettings>,
): ReminderSettings {
  const defaults = defaultReminderSettings()
  const leadMinutes =
    settings?.lead_minutes?.filter((value) => Number.isFinite(value) && value > 0)
  const animation = REMINDER_ANIMATIONS.includes(
    settings?.animation as ReminderAnimation,
  )
    ? settings?.animation
    : defaults.animation

  const enabled = Boolean(settings?.enabled)

  return {
    enabled,
    autostart: enabled && Boolean(settings?.autostart),
    lead_minutes: leadMinutes && leadMinutes.length > 0
      ? [...new Set(leadMinutes.map((value) => Math.trunc(value)))].sort(
          (a, b) => b - a,
        )
      : defaults.lead_minutes,
    animation,
    native_notifications:
      settings?.native_notifications ?? defaults.native_notifications,
    popup_notifications:
      settings?.popup_notifications ?? defaults.popup_notifications,
  }
}

export function reminderLeadPreset(leads: number[]): '15-5' | '10-2' | '5' {
  const normalized = normalizeReminderSettings({ lead_minutes: leads }).lead_minutes
  const key = normalized?.join('-')
  if (key === '10-2') return '10-2'
  if (key === '5') return '5'
  return '15-5'
}
