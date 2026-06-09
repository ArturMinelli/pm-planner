import { describe, expect, it } from 'vitest'
import {
  defaultReminderSettings,
  formatReminderLead,
  normalizeReminderSettings,
} from './reminderSettings'

describe('reminderSettings', () => {
  it('uses opt-in defaults with native notifications enabled', () => {
    expect(defaultReminderSettings()).toEqual({
      enabled: false,
      autostart: false,
      lead_minutes: [15, 5],
      native_notifications: true,
    })
  })

  it('normalizes lead minutes', () => {
    expect(
      normalizeReminderSettings({
        enabled: true,
        autostart: true,
        lead_minutes: [5, 15, 5, -1, 241],
        native_notifications: false,
      }),
    ).toMatchObject({
      enabled: true,
      autostart: true,
      lead_minutes: [15, 5],
      native_notifications: true,
    })
  })

  it('formats lead times for display', () => {
    expect(formatReminderLead(15)).toBe('15 min')
    expect(formatReminderLead(60)).toBe('1 hora')
    expect(formatReminderLead(120)).toBe('2 horas')
  })
})
