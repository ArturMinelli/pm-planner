/**
 * @vitest-environment jsdom
 */
import { act, createElement, useEffect } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import * as backend from '../services/backend'
import { detectLocation } from '../util/location'
import type { ClockInResult } from '../types'
import { useClockIn } from './useClockIn'

vi.mock('../services/backend', () => ({
  clockIn: vi.fn(),
}))

vi.mock('../util/location', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../util/location')>()
  return { ...actual, detectLocation: vi.fn() }
})

const mockedBackend = vi.mocked(backend)
const mockedDetect = vi.mocked(detectLocation)

const location = {
  latitude: -27.6,
  longitude: -48.6,
  accuracy: 12,
  source: 'detected' as const,
}

type Api = ReturnType<typeof useClockIn>
const latest: { api: Api | null } = { api: null }

function Harness() {
  const value = useClockIn()
  useEffect(() => {
    latest.api = value
  })
  return null
}

function current(): Api {
  if (!latest.api) throw new Error('hook not mounted')
  return latest.api
}

describe('useClockIn', () => {
  let container: HTMLDivElement
  let root: Root

  beforeEach(() => {
    vi.clearAllMocks()
    latest.api = null
    mockedDetect.mockResolvedValue(location)
    container = document.createElement('div')
    document.body.appendChild(container)
    root = createRoot(container)
    act(() => root.render(createElement(Harness)))
  })

  afterEach(() => {
    act(() => root.unmount())
    container.remove()
  })

  it('goes idle to registered and exposes the detected location', async () => {
    const ok: ClockInResult = {
      status: 'registered',
      registered: true,
      punch: { id: 1, date: '2026-11-17', time: '08:00:00', address: 'x', latitude: 1, longitude: 2, disabled: false },
      locationSource: 'detected',
    }
    mockedBackend.clockIn.mockResolvedValue(ok)

    expect(current().phase).toBe('idle')
    await act(async () => {
      await current().run()
    })

    expect(mockedBackend.clockIn).toHaveBeenCalledWith({ location, confirmNear: false })
    expect(current().phase).toBe('registered')
    expect(current().result).toEqual(ok)
    expect(current().location).toEqual(location)
  })

  it('sends exactly one request when run is invoked twice in the same tick', async () => {
    mockedBackend.clockIn.mockResolvedValue({ status: 'registered', registered: true })

    let second: ClockInResult | undefined = { status: 'failed', registered: false }
    await act(async () => {
      const first = current().run()
      second = await current().run()
      await first
    })

    expect(mockedBackend.clockIn).toHaveBeenCalledTimes(1)
    expect(second).toBeUndefined()
  })

  it('re-runs with confirmNear true after a too_close failure', async () => {
    mockedBackend.clockIn
      .mockResolvedValueOnce({ status: 'failed', registered: false, code: 'too_close' })
      .mockResolvedValueOnce({ status: 'registered', registered: true })

    await act(async () => {
      await current().run()
    })
    expect(current().phase).toBe('failed')
    expect(current().result?.code).toBe('too_close')

    await act(async () => {
      await current().run(true)
    })
    expect(mockedBackend.clockIn).toHaveBeenLastCalledWith({ location, confirmNear: true })
    expect(current().phase).toBe('registered')
  })

  it('maps uncertain outcomes to the uncertain phase', async () => {
    mockedBackend.clockIn.mockResolvedValue({ status: 'uncertain', registered: null, code: 'uncertain' })

    await act(async () => {
      await current().run()
    })

    expect(current().phase).toBe('uncertain')
  })

  it('keeps the failure code the backend returns', async () => {
    mockedBackend.clockIn.mockResolvedValue({ status: 'failed', registered: false, code: 'session_expired' })

    await act(async () => {
      await current().run()
    })

    expect(current().phase).toBe('failed')
    expect(current().result?.code).toBe('session_expired')
  })

  it('treats a thrown backend error as failed with code offline', async () => {
    mockedBackend.clockIn.mockRejectedValue(new TypeError('Failed to fetch'))

    await act(async () => {
      await current().run()
    })

    expect(current().phase).toBe('failed')
    expect(current().result).toEqual({ status: 'failed', registered: false, code: 'offline' })
  })

  it('accepts a new run after the previous one finished', async () => {
    mockedBackend.clockIn.mockRejectedValueOnce(new Error('down'))
    mockedBackend.clockIn.mockResolvedValueOnce({ status: 'registered', registered: true })

    await act(async () => {
      await current().run()
    })
    await act(async () => {
      await current().retry()
    })

    expect(mockedBackend.clockIn).toHaveBeenCalledTimes(2)
    expect(current().phase).toBe('registered')
  })
})
