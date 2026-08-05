import type { UpdateStatus } from '../../types'
import { Banner, Button, Card, StatRow } from '../../components/ui'
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
  const blockers = status?.blockers ?? []
  const canUpdate = !!status?.updateAvailable && !updating && canUseRuntime

  return (
    <Card
      title="Atualizações"
      intro="Verifique se há uma versão mais recente e atualize quando quiser."
    >
      <div className="stat-list">
        <StatRow label="Versão instalada" value={formatInstalledVersion(status)} />
      </div>

      {blockers.map((blocker) => (
        <Banner tone="error" key={blocker}>
          {blocker}
        </Banner>
      ))}

      {status && blockers.length === 0 ? (
        <p className="muted">{describeUpdateAvailability(status)}</p>
      ) : null}

      {canUpdate ? (
        <p className="hint">
          O app será fechado e reaberto automaticamente ao final da atualização.
        </p>
      ) : null}

      <div className="btn-row">
        <Button
          variant="secondary"
          disabled={checking || updating || !canUseRuntime}
          aria-busy={checking}
          onClick={onCheck}
        >
          {checking ? 'Verificando...' : 'Verificar Atualizações'}
        </Button>
        <Button
          variant="primary"
          disabled={!canUpdate}
          aria-busy={updating}
          onClick={onUpdate}
        >
          {updating ? 'Atualizando...' : 'Atualizar Agora'}
        </Button>
      </div>
    </Card>
  )
}
