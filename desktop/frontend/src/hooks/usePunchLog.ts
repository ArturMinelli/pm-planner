import { useCallback, useEffect, useRef, useState } from 'react'
import type { PunchLogState } from '../types'
import * as backend from '../services/backend'

export const FOCUS_SYNC_THROTTLE_MS = 30_000
export const POLL_ACTIVE_MS = 30_000
export const POLL_IDLE_MS = 5 * 60_000

export function usePunchLog(date: string) {
  const [state, setState] = useState<PunchLogState | null>(null)
  const [syncing, setSyncing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const requestRef = useRef(0)
  const dateRef = useRef(date)
  const lastSyncRef = useRef(Number.NEGATIVE_INFINITY)

  const refresh = useCallback(async () => {
    const id = ++requestRef.current
    const wanted = dateRef.current
    try {
      const next = await backend.getPunchLog(wanted)
      if (id !== requestRef.current) return
      setState(next)
      setError(null)
    } catch (e) {
      if (id !== requestRef.current) return
      setError(e instanceof Error ? e.message : String(e))
    }
  }, [])

  const syncNow = useCallback(async () => {
    lastSyncRef.current = Date.now()
    setSyncing(true)
    let syncError: string | null = null
    try {
      await backend.syncPunches()
    } catch (e) {
      syncError = e instanceof Error ? e.message : String(e)
    }
    await refresh()
    if (syncError) setError(syncError)
    setSyncing(false)
  }, [refresh])

  // Reload on date change; the request counter drops responses for the old date.
  useEffect(() => {
    dateRef.current = date
    void refresh()
  }, [date, refresh])

  // On-demand sync once on mount and when the window regains focus.
  useEffect(() => {
    const maybeSync = () => {
      if (document.visibilityState === 'hidden') return
      if (Date.now() - lastSyncRef.current < FOCUS_SYNC_THROTTLE_MS) return
      void syncNow()
    }
    maybeSync()
    window.addEventListener('focus', maybeSync)
    document.addEventListener('visibilitychange', maybeSync)
    return () => {
      window.removeEventListener('focus', maybeSync)
      document.removeEventListener('visibilitychange', maybeSync)
    }
  }, [syncNow])

  const active = !!state && (state.pendingCount > 0 || state.unconfirmedCount > 0)
  useEffect(() => {
    const id = setInterval(() => void refresh(), active ? POLL_ACTIVE_MS : POLL_IDLE_MS)
    return () => clearInterval(id)
  }, [active, refresh])

  const loading = state?.date !== date && !error
  return { state, loading, syncing, error, refresh, syncNow }
}
