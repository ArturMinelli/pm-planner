import { afterEach, describe, expect, it, vi } from 'vitest'
import { getPunchLog, syncPunches } from './backend'

const state = { date: '2026-11-17', punches: [], pendingCount: 0, unconfirmedCount: 0 }

describe('punch log services', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('uses the Wails runtime when present', async () => {
    const App = { GetPunchLog: vi.fn().mockResolvedValue(state), SyncPunches: vi.fn().mockResolvedValue(state) }
    vi.stubGlobal('go', { main: { App } })
    await getPunchLog('2026-11-17')
    await syncPunches()
    expect(App.GetPunchLog).toHaveBeenCalledWith('2026-11-17')
    expect(App.SyncPunches).toHaveBeenCalled()
  })

  it('falls back to HTTP without the runtime', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => state, text: async () => JSON.stringify(state) })
    vi.stubGlobal('fetch', fetchMock)
    await getPunchLog('2026-11-17')
    await syncPunches()
    expect(fetchMock.mock.calls[0][0]).toBe('/api/punches?date=2026-11-17')
    expect(fetchMock.mock.calls[1][0]).toBe('/api/punches/sync')
    expect(fetchMock.mock.calls[1][1]).toMatchObject({ method: 'POST' })
  })
})
