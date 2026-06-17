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
    expect(validateLoginCredential('not-an-email')).toBe('Use um e-mail ou CPF.')
    expect(validateLoginCredential('user@')).toBe('Informe um e-mail válido.')
  })

  it('validates cpf', () => {
    expect(validateLoginCredential('529.982.247-25')).toBeNull()
    expect(validateLoginCredential('52998224725')).toBeNull()
    expect(validateLoginCredential('123')).toBe('Informe um CPF válido.')
    expect(validateLoginCredential('111.111.111-11')).toBe(
      'Informe um CPF válido.',
    )
  })

  it('rejects empty and unknown formats', () => {
    expect(validateLoginCredential('')).toBe(
      'Informe o login (e-mail ou CPF).',
    )
    expect(validateLoginCredential('   ')).toBe(
      'Informe o login (e-mail ou CPF).',
    )
    expect(validateLoginCredential('abc-def')).toBe('Use um e-mail ou CPF.')
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
      hint: 'Detectado: e-mail',
    })
    expect(loginCredentialFieldMeta('52998224725')).toEqual({
      inputMode: 'numeric',
      hint: 'Detectado: CPF',
    })
    expect(loginCredentialFieldMeta('')).toEqual({ inputMode: 'text' })
  })
})
