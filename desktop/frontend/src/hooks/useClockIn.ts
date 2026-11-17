import { useCallback, useRef, useState } from 'react'
import type { ClockInResult } from '../types'
import * as backend from '../services/backend'
import { detectLocation, type ClockInLocation } from '../util/location'

export type ClockInPhase =
  | 'idle'
  | 'locating'
  | 'sending'
  | 'registered'
  | 'failed'
  | 'uncertain'

export function useClockIn() {
  const [phase, setPhase] = useState<ClockInPhase>('idle')
  const [result, setResult] = useState<ClockInResult | null>(null)
  const [location, setLocation] = useState<ClockInLocation | null>(null)
  // A ref, not state: a second click in the same tick must see the first one.
  const busyRef = useRef(false)

  const run = useCallback(async (confirmNear = false): Promise<ClockInResult | undefined> => {
    if (busyRef.current) return undefined
    busyRef.current = true
    setResult(null)
    setPhase('locating')

    let outcome: ClockInResult
    try {
      const loc = await detectLocation()
      setLocation(loc)
      setPhase('sending')
      outcome = await backend.clockIn({ location: loc, confirmNear })
    } catch {
      outcome = { status: 'failed', registered: false, code: 'offline' }
    }
    busyRef.current = false

    setResult(outcome)
    setPhase(
      outcome.status === 'registered'
        ? 'registered'
        : outcome.status === 'uncertain'
          ? 'uncertain'
          : 'failed',
    )
    return outcome
  }, [])

  const retry = useCallback(() => run(false), [run])

  const reset = useCallback(() => {
    setPhase('idle')
    setResult(null)
  }, [])

  return { phase, result, location, run, retry, reset }
}
