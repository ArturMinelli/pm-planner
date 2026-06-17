export type GeolocationResult = {
  latitude: number
  longitude: number
  accuracy: number
  address: string
}

const GEOLOCATION_TIMEOUT_MS = 15_000

function geolocationErrorMessage(code: number): string {
  switch (code) {
    case 1:
      return 'Permissão de localização negada. Habilite o acesso à localização para bater ponto.'
    case 2:
      return 'Não foi possível obter a localização atual.'
    case 3:
      return 'A solicitação de localização expirou. Tente novamente.'
    default:
      return 'Não foi possível obter a localização.'
  }
}

export function getCurrentPosition(): Promise<GeolocationPosition> {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error('Geolocalização não é suportada neste ambiente.'))
      return
    }
    navigator.geolocation.getCurrentPosition(resolve, reject, {
      enableHighAccuracy: true,
      timeout: GEOLOCATION_TIMEOUT_MS,
      maximumAge: 0,
    })
  })
}

export async function reverseGeocode(
  latitude: number,
  longitude: number,
): Promise<string> {
  const params = new URLSearchParams({
    format: 'jsonv2',
    lat: String(latitude),
    lon: String(longitude),
    'accept-language': 'pt-BR',
  })
  const response = await fetch(
    `https://nominatim.openstreetmap.org/reverse?${params.toString()}`,
    {
      headers: {
        Accept: 'application/json',
      },
    },
  )
  if (!response.ok) {
    throw new Error('Não foi possível converter a localização em endereço.')
  }
  const data = (await response.json()) as { display_name?: string }
  const address = data.display_name?.trim()
  if (!address) {
    throw new Error('Endereço não encontrado para a localização atual.')
  }
  return address
}

export async function resolveClockInLocation(): Promise<GeolocationResult> {
  let position: GeolocationPosition
  try {
    position = await getCurrentPosition()
  } catch (error) {
    if (error instanceof GeolocationPositionError) {
      throw new Error(geolocationErrorMessage(error.code))
    }
    throw error instanceof Error
      ? error
      : new Error('Não foi possível obter a localização.')
  }

  const { latitude, longitude, accuracy } = position.coords
  let address = ''
  try {
    address = await reverseGeocode(latitude, longitude)
  } catch {
    address = `${latitude.toFixed(6)}, ${longitude.toFixed(6)}`
  }

  return { latitude, longitude, accuracy, address }
}
