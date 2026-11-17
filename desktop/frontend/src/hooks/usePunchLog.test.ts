/**
 * @vitest-environment jsdom
 */
import { act, createElement, useEffect } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import * as backend from '../services/backend'
import type { PunchLogState } from '../types'
import { usePunchLog } from './usePunchLog'

vi.mock('../services/backend', () => ({
  getPunchLog: vi.fn(),
  syncPunches: vi.fn(),
}))

const mocked = vi.mocked(backend)
type Api = ReturnType<typeof usePunchLog>
const latest: { api: Api | null } = { api: null }

function Harness({ date }: { date: string }) {
  const value = usePunchLog(date)
  useEffect(() => {
    latest.api = value
  })
  return null
}

const log = (over: Partial<PunchLogState> = {}): PunchLogState => ({
  date: '2026-11-17',
  punches: [],
  pendingCount: 0,
  unconfirmedCount: 0,
  ...over,
})

describe('usePunchLog', () => {
  let container: HTMLDivElement
  let root: Root

  const mount = (date = '2026-11-17') => act(async () => root.render(createElement(Harness, { date })))
  const tick = (ms: number) => act(async () => { await vi.advanceTimersByTimeAsync(ms) })

  beforeEach(() => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-11-17T10:00:00'))
    vi.clearAllMocks()
    latest.api = null
    mocked.getPunchLog.mockResolvedValue(log())
    mocked.syncPunches.mockResolvedValue(log())
    container = document.createElement('div')
    document.body.appendChild(container)
    root = createRoot(container)
  })

  afterEach(() => {
    act(() => root.unmount())
    container.remove()
    vi.useRealTimers()
  })

  it('loads and syncs once on mount', async () => {
    await mount()
    expect(mocked.syncPunches).toHaveBeenCalledTimes(1)
    expect(mocked.getPunchLog).toHaveBeenCalledWith('2026-11-17')
    expect(latest.api?.state?.date).toBe('2026-11-17')
    expect(latest.api?.loading).toBe(false)
  })

  it('throttles focus syncs to once per 30 s', async () => {
    await mount()
    await act(async () => { window.dispatchEvent(new Event('focus')) })
    expect(mocked.syncPunches).toHaveBeenCalledTimes(1)
    await tick(31_000)
    await act(async () => { window.dispatchEvent(new Event('focus')) })
    expect(mocked.syncPunches).toHaveBeenCalledTimes(2)
  })

  it('polls every 5 min when idle and every 30 s with pending records', async () => {
    await mount()
    mocked.getPunchLog.mockClear()
    await tick(60_000)
    expect(mocked.getPunchLog).not.toHaveBeenCalled()
    await tick(4 * 60_000)
    expect(mocked.getPunchLog).toHaveBeenCalledTimes(1)

    mocked.getPunchLog.mockResolvedValue(log({ pendingCount: 1 }))
    await act(async () => { await latest.api?.refresh() })
    mocked.getPunchLog.mockClear()
    await tick(90_000)
    expect(mocked.getPunchLog).toHaveBeenCalledTimes(3)
  })

  it('ignores a stale response after the date changes', async () => {
    let resolveOld: (s: PunchLogState) => void = () => undefined
    mocked.getPunchLog.mockImplementationOnce(() => new Promise((r) => { resolveOld = r }))
    await mount('2026-11-16')
    mocked.getPunchLog.mockResolvedValue(log({ date: '2026-11-17' }))
    await mount('2026-11-17')
    await act(async () => resolveOld(log({ date: '2026-11-16' })))
    expect(latest.api?.state?.date).toBe('2026-11-17')
  })

  it('syncNow reports syncing and surfaces sync errors', async () => {
    await mount()
    mocked.syncPunches.mockRejectedValueOnce(new Error('boom'))
    await act(async () => { await latest.api?.syncNow() })
    expect(latest.api?.error).toBe('boom')
    expect(latest.api?.syncing).toBe(false)
  })

  it('removes listeners and timers on unmount', async () => {
    await mount()
    act(() => root.unmount())
    mocked.getPunchLog.mockClear()
    mocked.syncPunches.mockClear()
    await tick(10 * 60_000)
    window.dispatchEvent(new Event('focus'))
    expect(mocked.getPunchLog).not.toHaveBeenCalled()
    expect(mocked.syncPunches).not.toHaveBeenCalled()
    root = createRoot(container)
  })
})
