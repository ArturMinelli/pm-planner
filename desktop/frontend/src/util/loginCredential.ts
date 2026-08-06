export type LoginCredentialKind = 'email' | 'cpf'

type LoginCredentialStrategy = {
  kind: LoginCredentialKind
  detect: (raw: string) => boolean
  normalize: (raw: string) => string
  validate: (normalized: string) => string | null
}

const emailStrategy: LoginCredentialStrategy = {
  kind: 'email',
  detect: (raw) => raw.includes('@'),
  normalize: (raw) => raw.trim().toLowerCase(),
  validate: (normalized) => {
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(normalized)) {
      return 'errors.validation.invalidEmail'
    }
    return null
  },
}

function stripDigits(raw: string): string {
  return raw.replace(/\D/g, '')
}

function isValidCPF(digits: string): boolean {
  if (digits.length !== 11) return false
  if (/^(\d)\1{10}$/.test(digits)) return false

  let sum = 0
  for (let index = 0; index < 9; index++) {
    sum += Number(digits[index]) * (10 - index)
  }
  let remainder = (sum * 10) % 11
  if (remainder === 10) remainder = 0
  if (remainder !== Number(digits[9])) return false

  sum = 0
  for (let index = 0; index < 10; index++) {
    sum += Number(digits[index]) * (11 - index)
  }
  remainder = (sum * 10) % 11
  if (remainder === 10) remainder = 0
  return remainder === Number(digits[10])
}

const cpfStrategy: LoginCredentialStrategy = {
  kind: 'cpf',
  detect: (raw) => /^[\d.\-\s]+$/.test(raw.trim()),
  normalize: (raw) => stripDigits(raw),
  validate: (normalized) => {
    if (!isValidCPF(normalized)) {
      return 'errors.validation.invalidCpf'
    }
    return null
  },
}

const strategies: LoginCredentialStrategy[] = [emailStrategy, cpfStrategy]

export function resolveLoginStrategy(raw: string): LoginCredentialStrategy | null {
  const trimmed = raw.trim()
  if (!trimmed) return null
  return strategies.find((strategy) => strategy.detect(trimmed)) ?? null
}

export function validateLoginCredential(raw: string): string | null {
  const trimmed = raw.trim()
  if (!trimmed) {
    return 'errors.validation.loginRequired'
  }
  const strategy = resolveLoginStrategy(trimmed)
  if (!strategy) {
    return 'errors.validation.loginKind'
  }
  return strategy.validate(strategy.normalize(trimmed))
}

export function normalizeLoginCredential(raw: string): string {
  const strategy = resolveLoginStrategy(raw)
  if (!strategy) return raw.trim()
  return strategy.normalize(raw)
}

export function loginCredentialFieldMeta(raw: string): {
  inputMode: 'email' | 'numeric' | 'text'
  hintKey?: string
} {
  const strategy = resolveLoginStrategy(raw)
  if (!strategy) {
    return { inputMode: 'text' }
  }
  if (strategy.kind === 'email') {
    return { inputMode: 'email', hintKey: 'errors.validation.detectedEmail' }
  }
  return { inputMode: 'numeric', hintKey: 'errors.validation.detectedCpf' }
}
