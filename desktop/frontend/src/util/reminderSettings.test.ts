import { describe, expect, it } from 'vitest'
import i18n from '../i18n'
import {
  defaultReminderSettings,
  formatReminderLead,
  normalizeReminderSettings,
} from './reminderSettings'

describe('reminderSettings', () => {
  it('uses default leads with native notifications enabled', () => {
    expect(defaultReminderSettings()).toEqual({
      lead_minutes: [15, 5],
      native_notifications: true,
    })
  })

  it('normalizes lead minutes', () => {
    expect(
      normalizeReminderSettings({
        lead_minutes: [5, 15, 5, -1, 241],
        native_notifications: false,
      }),
    ).toMatchObject({
      lead_minutes: [15, 5],
      native_notifications: true,
    })
  })

  it('formats lead times for display', () => {
    const translate = i18n.t.bind(i18n)
    expect(formatReminderLead(translate, 15)).toBe('15 minutos')
    expect(formatReminderLead(translate, 60)).toBe('1 hora')
    expect(formatReminderLead(translate, 120)).toBe('2 horas')
  })
})
