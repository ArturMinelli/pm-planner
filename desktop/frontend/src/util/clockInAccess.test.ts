import { describe, expect, it } from 'vitest'
import { CLOCK_IN_ALLOWED_LOGIN, canClockIn } from './clockInAccess'

describe('canClockIn', () => {
  it('allows the configured login', () => {
    expect(canClockIn(CLOCK_IN_ALLOWED_LOGIN)).toBe(true)
  })

  it('ignores case and surrounding whitespace', () => {
    expect(canClockIn('  Artur.Peixoto@FiveLabs.com.br \n')).toBe(true)
  })

  it('rejects other users', () => {
    expect(canClockIn('someone.else@fivelabs.com.br')).toBe(false)
    expect(canClockIn('artur.peixoto@example.com')).toBe(false)
  })

  it('rejects empty or missing login', () => {
    expect(canClockIn('')).toBe(false)
    expect(canClockIn('   ')).toBe(false)
    expect(canClockIn(undefined)).toBe(false)
  })
})
