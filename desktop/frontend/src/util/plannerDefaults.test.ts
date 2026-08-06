import { describe, expect, it } from 'vitest'
import {
  BUILTIN_PLANNER_ANCHORS,
  resolvePlannerAnchorsArray,
} from './plannerDefaults'

describe('resolvePlannerAnchorsArray', () => {
  it('returns builtins when planner config is missing', () => {
    expect(resolvePlannerAnchorsArray()).toEqual([
      BUILTIN_PLANNER_ANCHORS.in1,
      BUILTIN_PLANNER_ANCHORS.out1,
      BUILTIN_PLANNER_ANCHORS.in2,
      BUILTIN_PLANNER_ANCHORS.out2,
    ])
  })

  it('merges partial overrides per slot', () => {
    expect(
      resolvePlannerAnchorsArray({
        in1: '09:00',
      }),
    ).toEqual(['09:00', '12:00', '13:30', '18:00'])
  })

  it('falls back per slot for invalid HH:MM values', () => {
    expect(
      resolvePlannerAnchorsArray({
        in1: '09:00',
        out1: '99:99',
        in2: '14:00',
        out2: '19:00',
      }),
    ).toEqual(['09:00', '12:00', '14:00', '19:00'])
  })
})
