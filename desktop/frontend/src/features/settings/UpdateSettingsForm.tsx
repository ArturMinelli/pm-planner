import { useTranslation } from 'react-i18next'
import type { UpdateStatus } from '../../types'
import { Banner, Button, Card, StatRow } from '../../components/ui'
import { messageKey, translateMessage } from '../../i18n/translateMessage'
import {
  describeUpdateAvailability,
  formatInstalledVersion,
} from '../../util/updateStatus'

type UpdateSettingsFormProps = {
  status: UpdateStatus | null
  checking: boolean
  updating: boolean
  canUseRuntime: boolean
  onCheck: () => void
  onUpdate: () => void
}

export default function UpdateSettingsForm({
  status,
  checking,
  updating,
  canUseRuntime,
  onCheck,
  onUpdate,
}: UpdateSettingsFormProps) {
  const { t } = useTranslation()
  const blockers = status?.blockers ?? []
  const canUpdate = !!status?.updateAvailable && !updating && canUseRuntime

  return (
    <Card
      title={t('settings.updates.title')}
      intro={t('settings.updates.intro')}
    >
      <div className="stat-list">
        <StatRow
          label={t('settings.updates.installedVersion')}
          value={formatInstalledVersion(t, status)}
        />
      </div>

      {blockers.map((blocker) => (
        <Banner tone="error" key={messageKey(blocker)}>
          {translateMessage(t, blocker)}
        </Banner>
      ))}

      {status && blockers.length === 0 ? (
        <p className="muted">{describeUpdateAvailability(t, status)}</p>
      ) : null}

      {canUpdate ? (
        <p className="hint">{t('settings.updates.restartHint')}</p>
      ) : null}

      <div className="btn-row">
        <Button
          variant="secondary"
          disabled={checking || updating || !canUseRuntime}
          aria-busy={checking}
          onClick={onCheck}
        >
          {checking ? t('settings.updates.checking') : t('settings.updates.check')}
        </Button>
        <Button
          variant="primary"
          disabled={!canUpdate}
          aria-busy={updating}
          onClick={onUpdate}
        >
          {updating ? t('settings.updates.updating') : t('settings.updates.updateNow')}
        </Button>
      </div>
    </Card>
  )
}
