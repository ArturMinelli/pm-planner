import { describe, expect, it } from 'vitest'
import { calculatePlannerSummary, DEFAULT_TARGET_SECS } from './plannerSummary'

describe('calculatePlannerSummary', () => {
  it('calculates a normal 8h30 workday', () => {
    expect(
      calculatePlannerSummary({
        baseTargetSecs: DEFAULT_TARGET_SECS,
        targetSecs: DEFAULT_TARGET_SECS,
        in1: '08:00',
        out1: '12:00',
        in2: '13:30',
      }),
    ).toEqual({
      out2: '18:00',
      firstSpanSecs: 4 * 60 * 60,
      secondSpanSecs: 4 * 60 * 60 + 30 * 60,
      totalSpanSecs: DEFAULT_TARGET_SECS,
      overtimeSecs: 0,
    })
  })

  it('uses zero remaining time when the first span is longer than the target', () => {
    const summary = calculatePlannerSummary({
      baseTargetSecs: 60 * 60,
      targetSecs: 60 * 60,
      in1: '08:00',
      out1: '10:00',
      in2: '10:15',
    })

    expect(summary.out2).toBe('10:15')
    expect(summary.secondSpanSecs).toBe(0)
    expect(summary.overtimeSecs).toBe(60 * 60)
  })

  it('treats invalid or missing clock strings as zero-duration spans', () => {
    const summary = calculatePlannerSummary({
      baseTargetSecs: DEFAULT_TARGET_SECS,
      targetSecs: DEFAULT_TARGET_SECS,
      in1: 'nope',
      out1: '12:00',
      in2: '',
    })

    expect(summary.out2).toBe('--:--')
    expect(summary.firstSpanSecs).toBe(0)
    expect(summary.secondSpanSecs).toBe(0)
    expect(summary.totalSpanSecs).toBe(0)
  })

  it('treats reversed clock spans as zero duration', () => {
    const summary = calculatePlannerSummary({
      baseTargetSecs: DEFAULT_TARGET_SECS,
      targetSecs: DEFAULT_TARGET_SECS,
      in1: '12:00',
      out1: '08:00',
      in2: '13:30',
    })

    expect(summary.firstSpanSecs).toBe(0)
    expect(summary.out2).toBe('22:00')
  })

  it('accepts a zero adjusted target', () => {
    expect(
      calculatePlannerSummary({
        baseTargetSecs: DEFAULT_TARGET_SECS,
        targetSecs: 0,
        in1: '08:00',
        out1: '12:00',
        in2: '13:30',
      }).out2,
    ).toBe('13:30')
  })

  it('calculates overtime against the contractual target', () => {
    const summary = calculatePlannerSummary({
      baseTargetSecs: DEFAULT_TARGET_SECS,
      targetSecs: DEFAULT_TARGET_SECS + 60 * 60,
      in1: '08:00',
      out1: '12:00',
      in2: '13:30',
    })

    expect(summary.out2).toBe('19:00')
    expect(summary.overtimeSecs).toBe(60 * 60)
  })
})
