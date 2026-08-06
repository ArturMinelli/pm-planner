import type { TFunction } from 'i18next'
import type { UpdateStatus } from '../types'

const isoDateParts = /^(\d{4})-(\d{2})-(\d{2})/

/** Reads the calendar date out of a commit timestamp, ignoring its time zone. */
export function formatCommitDate(commitDate: string): string {
  const parts = isoDateParts.exec(commitDate.trim())
  if (!parts) return ''
  return `${parts[3]}/${parts[2]}/${parts[1]}`
}

export function formatInstalledVersion(
  translate: TFunction,
  status: UpdateStatus | null,
): string {
  if (!status) return translate('settings.updates.unknownVersion')
  if (!status.isGit) return translate('settings.updates.unknownGitless')
  if (!status.commitSha) return translate('settings.updates.unknown')

  const date = formatCommitDate(status.commitDate)
  return date ? `${status.commitSha} (${date})` : status.commitSha
}

/** Summarizes a check result. Only meaningful once blockers have been ruled out. */
export function describeUpdateAvailability(
  translate: TFunction,
  status: UpdateStatus,
): string {
  if (!status.isGit) {
    return translate('settings.updates.cannotCompare')
  }
  if (status.behind === 1) return translate('settings.updates.oneAvailable')
  if (status.behind > 1) {
    return translate('settings.updates.manyAvailable', { count: status.behind })
  }
  return translate('settings.updates.upToDate')
}
