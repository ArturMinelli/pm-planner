import { describe, expect, it } from 'vitest'
import {
  formatDurationMinutes,
  formatSignedDurationSecs,
  formatSignedRoundedMinutes,
  formatTargetHours,
} from './timeFormat'

describe('formatSignedDurationSecs', () => {
  it('formats positive, negative, and zero durations', () => {
    expect(formatSignedDurationSecs(3661)).toBe('+01:01:01')
    expect(formatSignedDurationSecs(-3661)).toBe('-01:01:01')
    expect(formatSignedDurationSecs(0)).toBe('00:00:00')
  })
})

describe('compact minute formatting', () => {
  it('rounds signed balances to the nearest minute', () => {
    expect(formatSignedRoundedMinutes(-31560)).toBe('-08:46')
    expect(formatSignedRoundedMinutes(3599)).toBe('+01:00')
    expect(formatDurationMinutes(5430)).toBe('01:31')
  })
})

describe('formatTargetHours', () => {
  it('writes the daily target the way people say it', () => {
    expect(formatTargetHours(8.5 * 3600)).toBe('8h30')
    expect(formatTargetHours(8 * 3600)).toBe('8h')
    expect(formatTargetHours(6 * 3600 + 5 * 60)).toBe('6h05')
    expect(formatTargetHours(Number.NaN)).toBe('0h')
  })
})
