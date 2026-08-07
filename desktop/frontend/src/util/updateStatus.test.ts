import { describe, expect, it } from 'vitest'
import i18n from '../i18n'
import type { UpdateStatus } from '../types'
import {
  describeUpdateAvailability,
  describeUpdateBadge,
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
  const translate = i18n.t.bind(i18n)

  it('shows the short SHA and its date', () => {
    expect(formatInstalledVersion(translate, status())).toBe('a1b2c3d (03/08/2026)')
  })

  it('falls back to the SHA alone when the date is missing', () => {
    expect(formatInstalledVersion(translate, status({ commitDate: '' }))).toBe('a1b2c3d')
  })

  it('marks tarball installs as unknown', () => {
    expect(formatInstalledVersion(translate, status({ isGit: false }))).toContain('git')
  })

  it('prompts for a check before anything is known', () => {
    expect(formatInstalledVersion(translate, null)).toBe('Verifique para descobrir')
  })
})

describe('describeUpdateBadge', () => {
  const translate = i18n.t.bind(i18n)

  it('reports being up to date with a success tone', () => {
    expect(describeUpdateBadge(translate, status())).toEqual({
      label: 'PM Planner está atualizado.',
      tone: 'success',
    })
  })

  it('uses the singular for a single commit with a warning tone', () => {
    expect(describeUpdateBadge(translate, status({ behind: 1 }))).toEqual({
      label: '1 atualização disponível.',
      tone: 'warning',
    })
  })

  it('uses the plural beyond one commit with a warning tone', () => {
    expect(describeUpdateBadge(translate, status({ behind: 5 }))).toEqual({
      label: '5 atualizações disponíveis.',
      tone: 'warning',
    })
  })

  it('explains that tarball installs cannot be compared with a muted tone', () => {
    expect(describeUpdateBadge(translate, status({ isGit: false }))).toEqual({
      label: expect.stringContaining('Não é possível comparar versões'),
      tone: 'muted',
    })
  })
})

describe('describeUpdateAvailability', () => {
  const translate = i18n.t.bind(i18n)

  it('reports being up to date', () => {
    expect(describeUpdateAvailability(translate, status())).toBe('PM Planner está atualizado.')
  })

  it('uses the singular for a single commit', () => {
    expect(describeUpdateAvailability(translate, status({ behind: 1 }))).toBe('1 atualização disponível.')
  })

  it('uses the plural beyond one commit', () => {
    expect(describeUpdateAvailability(translate, status({ behind: 5 }))).toBe('5 atualizações disponíveis.')
  })

  it('explains that tarball installs cannot be compared', () => {
    expect(describeUpdateAvailability(translate, status({ isGit: false }))).toContain(
      'Não é possível comparar versões',
    )
  })
})
