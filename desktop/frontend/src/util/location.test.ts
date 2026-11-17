import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { DEFAULT_LOCATION, detectLocation } from './location'

type Geo = Pick<Geolocation, 'getCurrentPosition'>

function stubGeolocation(geolocation: Geo | undefined) {
  vi.stubGlobal('navigator', { geolocation })
}

describe('detectLocation', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.unstubAllGlobals()
  })

  it('returns the detected position when granted', async () => {
    stubGeolocation({
      getCurrentPosition: (success) =>
        success({
          coords: { latitude: -27.1, longitude: -48.2, accuracy: 1100 },
        } as GeolocationPosition),
    })

    await expect(detectLocation()).resolves.toEqual({
      latitude: -27.1,
      longitude: -48.2,
      accuracy: 1100,
      source: 'detected',
    })
  })

  it('falls back to the default when denied', async () => {
    stubGeolocation({
      getCurrentPosition: (_success, error) =>
        error?.({ code: 1, message: 'denied' } as GeolocationPositionError),
    })

    await expect(detectLocation()).resolves.toEqual(DEFAULT_LOCATION)
  })

  it('falls back to the default on timeout', async () => {
    stubGeolocation({ getCurrentPosition: () => {} })

    const promise = detectLocation(3000)
    await vi.advanceTimersByTimeAsync(3000)

    await expect(promise).resolves.toEqual(DEFAULT_LOCATION)
  })

  it('falls back to the default when geolocation is unsupported', async () => {
    stubGeolocation(undefined)

    await expect(detectLocation()).resolves.toEqual(DEFAULT_LOCATION)
  })
})
