import { describe, expect, it } from 'vitest'
import type { UpdateStatus } from '../types'
import {
  describeUpdateAvailability,
  formatCommitDate,
  formatInstalledVersion,
} from './updateStatus'

function status(overrides: Partial<UpdateStatus> = {}): UpdateStatus {
  return {
    root: '/home/user/.local/share/pm-planner',
    isGit: true,
    commitSha: 'a1b2c3d',
    commitDate: '2026-08-03T15:57:02-03:00',
    behind: 0,
    dirty: false,
    blockers: [],
    updateAvailable: false,
    ...overrides,
  }
}

describe('formatCommitDate', () => {
  it('keeps the commit calendar date regardless of the local time zone', () => {
    expect(formatCommitDate('2026-08-03T23:57:02-03:00')).toBe('03/08/2026')
    expect(formatCommitDate('2026-01-09T00:10:00+09:00')).toBe('09/01/2026')
  })

  it('returns an empty string for unparseable input', () => {
    expect(formatCommitDate('')).toBe('')
    expect(formatCommitDate('ontem')).toBe('')
  })
})

describe('formatInstalledVersion', () => {
  it('shows the short SHA and its date', () => {
    expect(formatInstalledVersion(status())).toBe('a1b2c3d (03/08/2026)')
  })

  it('falls back to the SHA alone when the date is missing', () => {
    expect(formatInstalledVersion(status({ commitDate: '' }))).toBe('a1b2c3d')
  })

  it('marks tarball installs as unknown', () => {
    expect(formatInstalledVersion(status({ isGit: false }))).toContain(
      'sem git',
    )
  })

  it('prompts for a check before anything is known', () => {
    expect(formatInstalledVersion(null)).toBe('Verifique para descobrir')
  })
})

describe('describeUpdateAvailability', () => {
  it('reports being up to date', () => {
    expect(describeUpdateAvailability(status())).toBe(
      'PM Planner está atualizado.',
    )
  })

  it('uses the singular for a single commit', () => {
    expect(describeUpdateAvailability(status({ behind: 1 }))).toBe(
      '1 atualização disponível.',
    )
  })

  it('uses the plural beyond one commit', () => {
    expect(describeUpdateAvailability(status({ behind: 5 }))).toBe(
      '5 atualizações disponíveis.',
    )
  })

  it('explains that tarball installs cannot be compared', () => {
    expect(describeUpdateAvailability(status({ isGit: false }))).toContain(
      'Não é possível comparar versões',
    )
  })
})
