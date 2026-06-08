import { describe, expect, it } from 'vitest'
import {
  defaultReminderSettings,
  normalizeReminderSettings,
  reminderLeadPreset,
} from './reminderSettings'

describe('reminderSettings', () => {
  it('uses opt-in defaults with both channels enabled', () => {
    expect(defaultReminderSettings()).toEqual({
      enabled: false,
      autostart: false,
      lead_minutes: [15, 5],
      animation: 'train',
      native_notifications: true,
      popup_notifications: true,
    })
  })

  it('normalizes lead minutes and unknown animations', () => {
    expect(
      normalizeReminderSettings({
        enabled: true,
        autostart: true,
        lead_minutes: [5, 15, 5, -1],
        animation: 'unknown' as 'train',
      }),
    ).toMatchObject({
      enabled: true,
      autostart: true,
      lead_minutes: [15, 5],
      animation: 'train',
    })
  })

  it('detects supported lead presets', () => {
    expect(reminderLeadPreset([2, 10])).toBe('10-2')
    expect(reminderLeadPreset([5])).toBe('5')
    expect(reminderLeadPreset([20])).toBe('15-5')
  })
})
