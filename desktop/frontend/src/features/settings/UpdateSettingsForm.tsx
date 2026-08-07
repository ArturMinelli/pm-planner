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

      {!showSkeleton && status ? (
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
        <p className="hint">{t('settings.updates.restartHint')}</p>
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
