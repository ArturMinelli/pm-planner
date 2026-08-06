import { describe, expect, it } from 'vitest'
import {
  loginCredentialFieldMeta,
  normalizeLoginCredential,
  resolveLoginStrategy,
  validateLoginCredential,
} from './loginCredential'

describe('loginCredential', () => {
  it('detects email strategy', () => {
    expect(resolveLoginStrategy('user@example.com')?.kind).toBe('email')
  })

  it('detects cpf strategy', () => {
    expect(resolveLoginStrategy('529.982.247-25')?.kind).toBe('cpf')
    expect(resolveLoginStrategy('52998224725')?.kind).toBe('cpf')
  })

  it('validates email', () => {
    expect(validateLoginCredential('user@example.com')).toBeNull()
    expect(validateLoginCredential('not-an-email')).toBe('errors.validation.loginKind')
    expect(validateLoginCredential('user@')).toBe('errors.validation.invalidEmail')
  })

  it('validates cpf', () => {
    expect(validateLoginCredential('529.982.247-25')).toBeNull()
    expect(validateLoginCredential('52998224725')).toBeNull()
    expect(validateLoginCredential('123')).toBe('errors.validation.invalidCpf')
    expect(validateLoginCredential('111.111.111-11')).toBe(
      'errors.validation.invalidCpf',
    )
  })

  it('rejects empty and unknown formats', () => {
    expect(validateLoginCredential('')).toBe('errors.validation.loginRequired')
    expect(validateLoginCredential('   ')).toBe('errors.validation.loginRequired')
    expect(validateLoginCredential('abc-def')).toBe('errors.validation.loginKind')
  })

  it('normalizes credentials', () => {
    expect(normalizeLoginCredential('  User@Example.COM ')).toBe(
      'user@example.com',
    )
    expect(normalizeLoginCredential('529.982.247-25')).toBe('52998224725')
  })

  it('provides field metadata', () => {
    expect(loginCredentialFieldMeta('user@example.com')).toEqual({
      inputMode: 'email',
      hintKey: 'errors.validation.detectedEmail',
    })
    expect(loginCredentialFieldMeta('52998224725')).toEqual({
      inputMode: 'numeric',
      hintKey: 'errors.validation.detectedCpf',
    })
    expect(loginCredentialFieldMeta('')).toEqual({ inputMode: 'text' })
  })
})
