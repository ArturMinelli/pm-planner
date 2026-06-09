import type { ReminderSettings } from '../types'

export const DEFAULT_REMINDER_LEADS = [15, 5] as const
export const MAX_REMINDER_LEAD_MINUTES = 4 * 60

export function defaultReminderSettings(): ReminderSettings {
  return {
    enabled: false,
    autostart: false,
    lead_minutes: [...DEFAULT_REMINDER_LEADS],
    native_notifications: true,
  }
}

export function normalizeReminderSettings(
  settings?: Partial<ReminderSettings>,
): ReminderSettings {
  const defaults = defaultReminderSettings()
  const leadMinutes =
    settings?.lead_minutes?.filter(
      (value) =>
        Number.isFinite(value) &&
        value > 0 &&
        value <= MAX_REMINDER_LEAD_MINUTES,
    )

  const enabled = Boolean(settings?.enabled)

  return {
    enabled,
    autostart: enabled && Boolean(settings?.autostart),
    lead_minutes: leadMinutes && leadMinutes.length > 0
      ? [...new Set(leadMinutes.map((value) => Math.trunc(value)))].sort(
          (a, b) => b - a,
        )
      : defaults.lead_minutes,
    native_notifications: true,
  }
}

export function formatReminderLead(minutes: number): string {
  if (minutes % 60 === 0) {
    const hours = minutes / 60
    return `${hours} ${hours === 1 ? 'hora' : 'horas'}`
  }
  return `${minutes} min`
}
