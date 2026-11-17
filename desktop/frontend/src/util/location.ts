export interface ClockInLocation {
  latitude: number
  longitude: number
  accuracy: number
  address?: string
  source: 'detected' | 'default'
}

export const DEFAULT_LOCATION: ClockInLocation = {
  latitude: -27.6017026,
  longitude: -48.6144075,
  accuracy: 50,
  address: 'Av. Presidente Kennedy, 861 - 1º Andar - Campinas, São José - SC, 88102-401',
  source: 'default',
}

// Never rejects: any failure (unsupported, denied, error, timeout) falls back to DEFAULT_LOCATION.
export function detectLocation(timeoutMs = 3000): Promise<ClockInLocation> {
  if (typeof navigator === 'undefined' || !navigator.geolocation) {
    return Promise.resolve(DEFAULT_LOCATION)
  }

  const locate = new Promise<ClockInLocation>((resolve) => {
    try {
      navigator.geolocation.getCurrentPosition(
        ({ coords }) =>
          resolve({
            latitude: coords.latitude,
            longitude: coords.longitude,
            accuracy: coords.accuracy,
            source: 'detected',
          }),
        () => resolve(DEFAULT_LOCATION),
        { timeout: timeoutMs },
      )
    } catch {
      resolve(DEFAULT_LOCATION)
    }
  })

  // Own timer: some webviews ignore the geolocation timeout option.
  let timer: ReturnType<typeof setTimeout> | undefined
  const timeout = new Promise<ClockInLocation>((resolve) => {
    timer = setTimeout(() => resolve(DEFAULT_LOCATION), timeoutMs)
  })

  return Promise.race([locate, timeout]).finally(() => clearTimeout(timer))
}
