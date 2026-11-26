import { useTranslation } from 'react-i18next'
import type { UpdateStatus } from '../../types'
import { Banner, Button, Card } from '../../components/ui'
import { messageKey, translateMessage } from '../../i18n/translateMessage'
import {
  describeUpdateBadge,
  formatInstalledVersion,
} from '../../util/updateStatus'

type UpdateSettingsFormProps = {
  status: UpdateStatus | null
  checking: boolean
  checkError: string | null
  updating: boolean
  canApplyUpdate: boolean
  onUpdate: () => void
}

function UpdateStatusSkeleton({ label }: { label: string }) {
  return (
    <div className="skeleton-stat-list" aria-busy="true" aria-label={label}>
      <div className="skeleton skeleton-stat-row" aria-hidden="true" />
    </div>
  )
}

export default function UpdateSettingsForm({
  status,
  checking,
  checkError,
  updating,
  canApplyUpdate,
  onUpdate,
}: UpdateSettingsFormProps) {
  const { t } = useTranslation()
  const blockers = status?.blockers ?? []
  const showUpdateButton = !!status?.updateAvailable && canApplyUpdate
  const canUpdate = showUpdateButton && !updating
  const badge = status ? describeUpdateBadge(t, status) : null
  const showSkeleton = checking && !status

  return (
    <Card
      title={t('settings.updates.title')}
      intro={t('settings.updates.intro')}
    >
      {showSkeleton ? (
        <UpdateStatusSkeleton label={t('settings.updates.checking')} />
      ) : null}

      {/* Without git the version is always "unknown", so the pill and badge would only add noise. */}
      {!showSkeleton && status?.isGit ? (
        <div className="update-status-row" aria-live="polite">
          <code className="version-pill" translate="no">
            {formatInstalledVersion(t, status)}
          </code>
          {badge ? (
            <span className={`status-badge ${badge.tone}`}>{badge.label}</span>
          ) : null}
        </div>
      ) : null}

      {checkError ? (
        <Banner tone="error">{checkError}</Banner>
      ) : null}

      {blockers.map((blocker) => (
        <Banner tone="error" key={messageKey(blocker)}>
          {translateMessage(t, blocker)}
        </Banner>
      ))}

      {showUpdateButton ? (
        <p className="update-note" role="note">
          <svg
            className="update-note-icon"
            width="16"
            height="16"
            viewBox="0 0 16 16"
            aria-hidden="true"
            focusable="false"
          >
            <circle cx="8" cy="8" r="6.25" fill="none" stroke="currentColor" strokeWidth="1.4" />
            <path
              d="M8 7.2v3.6M8 5.1v.01"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
            />
          </svg>
          <span>{t('settings.updates.restartHint')}</span>
        </p>
      ) : null}

      {showUpdateButton ? (
        <div className="btn-row">
          <Button
            variant="primary"
            disabled={!canUpdate}
            aria-busy={updating}
            onClick={onUpdate}
          >
            {updating ? t('settings.updates.updating') : t('settings.updates.updateNow')}
          </Button>
        </div>
      ) : null}
    </Card>
  )
}
