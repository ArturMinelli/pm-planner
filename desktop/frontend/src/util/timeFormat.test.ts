import { describe, expect, it } from 'vitest'
import { formatSignedDurationSecs } from './timeFormat'

describe('formatSignedDurationSecs', () => {
  it('formats positive, negative, and zero durations', () => {
    expect(formatSignedDurationSecs(3661)).toBe('+01:01:01')
    expect(formatSignedDurationSecs(-3661)).toBe('-01:01:01')
    expect(formatSignedDurationSecs(0)).toBe('00:00:00')
  })
})
