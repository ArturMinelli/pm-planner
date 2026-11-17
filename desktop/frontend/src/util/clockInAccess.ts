export const CLOCK_IN_ALLOWED_LOGIN = 'artur.peixoto@fivelabs.com.br'

export function canClockIn(login?: string): boolean {
  return (login ?? '').trim().toLowerCase() === CLOCK_IN_ALLOWED_LOGIN
}
