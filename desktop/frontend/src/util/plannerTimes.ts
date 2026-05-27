export const STEP_MINUTES = 15
export const MIN_GAP_MINUTES = 15
export const DAY_START = 0
export const DAY_END = 23 * 60 + 45

export type PlannerTimes = { in1: string; out1: string; in2: string }
export type PlannerTimeField = keyof PlannerTimes

const hhmmRe = /^(\d{2}):(\d{2})$/

export function parseHHMM(s: string): number | null {
  const m = hhmmRe.exec(s.trim())
  if (!m) return null
  const h = Number(m[1])
  const min = Number(m[2])
  if (h < 0 || h > 23 || min < 0 || min > 59) return null
  return h * 60 + min
}

export function formatMinutes(m: number): string {
  const clamped = Math.max(DAY_START, Math.min(DAY_END, m))
  const h = Math.floor(clamped / 60)
  const min = clamped % 60
  return `${String(h).padStart(2, '0')}:${String(min).padStart(2, '0')}`
}

export function addMinutes(hhmm: string, delta: number): string | null {
  const base = parseHHMM(hhmm)
  if (base === null) return null
  return formatMinutes(base + delta)
}

function clampMinutes(m: number): number {
  return Math.max(DAY_START, Math.min(DAY_END, m))
}

function forwardCascade(
  in1: number,
  out1: number,
  in2: number,
): [number, number, number] {
  let a = in1
  let b = out1
  let c = in2

  if (b < a + MIN_GAP_MINUTES) b = a + MIN_GAP_MINUTES
  if (c < b + MIN_GAP_MINUTES) c = b + MIN_GAP_MINUTES

  if (c > DAY_END) {
    c = DAY_END
    if (b > c - MIN_GAP_MINUTES) b = c - MIN_GAP_MINUTES
    if (a > b - MIN_GAP_MINUTES) a = b - MIN_GAP_MINUTES
    a = clampMinutes(a)
    b = clampMinutes(b)
    if (b < a + MIN_GAP_MINUTES) b = Math.min(DAY_END, a + MIN_GAP_MINUTES)
    if (c < b + MIN_GAP_MINUTES) c = Math.min(DAY_END, b + MIN_GAP_MINUTES)
  }

  return [a, b, c]
}

export function enforcePlannerOrder(times: PlannerTimes): PlannerTimes {
  const in1 = parseHHMM(times.in1)
  const out1 = parseHHMM(times.out1)
  const in2 = parseHHMM(times.in2)

  if (in1 === null || out1 === null || in2 === null) return times

  const [a, b, c] = forwardCascade(in1, out1, in2)
  return {
    in1: formatMinutes(a),
    out1: formatMinutes(b),
    in2: formatMinutes(c),
  }
}

export function adjustPlannerTime(
  times: PlannerTimes,
  field: PlannerTimeField,
  deltaMinutes: number,
): PlannerTimes {
  const current = times[field]
  const next = addMinutes(current, deltaMinutes)
  if (next === null) return times

  const updated = { ...times, [field]: next }
  return enforcePlannerOrder(updated)
}
