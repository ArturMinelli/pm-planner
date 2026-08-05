import type { UpdateStatus } from '../types'

const isoDateParts = /^(\d{4})-(\d{2})-(\d{2})/

export function browserUpdateStatus(): UpdateStatus {
  return {
    root: '',
    isGit: false,
    commitSha: '',
    commitDate: '',
    behind: 0,
    dirty: false,
    blockers: ['Atualizações só funcionam dentro do app desktop.'],
    updateAvailable: false,
  }
}

/** Reads the calendar date out of a commit timestamp, ignoring its time zone. */
export function formatCommitDate(commitDate: string): string {
  const parts = isoDateParts.exec(commitDate.trim())
  if (!parts) return ''
  return `${parts[3]}/${parts[2]}/${parts[1]}`
}

export function formatInstalledVersion(status: UpdateStatus | null): string {
  if (!status) return 'Verifique para descobrir'
  if (!status.isGit) return 'Desconhecida (instalação sem git)'
  if (!status.commitSha) return 'Desconhecida'

  const date = formatCommitDate(status.commitDate)
  return date ? `${status.commitSha} (${date})` : status.commitSha
}

/** Summarizes a check result. Only meaningful once blockers have been ruled out. */
export function describeUpdateAvailability(status: UpdateStatus): string {
  if (!status.isGit) {
    return 'Não é possível comparar versões. Atualizar reinstala a partir do código mais recente.'
  }
  if (status.behind === 1) return '1 atualização disponível.'
  if (status.behind > 1) return `${status.behind} atualizações disponíveis.`
  return 'PM Planner está atualizado.'
}
