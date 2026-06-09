import { Button, Card } from '../../components/ui'
import type { BalanceAdjustment, BalancePayload } from '../../types'
import {
  formatDurationSecs,
  formatSignedDurationSecs,
} from '../../util/timeFormat'

type PlannerBalanceCardProps = {
  balance: BalancePayload | null
  adjustment?: BalanceAdjustment
  busy: boolean
  canRefresh?: boolean
  error: string | null
  onRefresh: () => void
}

function balanceLabel(seconds: number): string {
  if (seconds < 0) return 'Você deve horas'
  if (seconds > 0) return 'Você tem horas disponíveis'
  return 'Banco de horas zerado'
}

function formatUpdatedAt(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(date)
}

export default function PlannerBalanceCard({
  balance,
  adjustment,
  busy,
  canRefresh = true,
  error,
  onRefresh,
}: PlannerBalanceCardProps) {
  const balanceSecs = balance?.balanceSecs ?? adjustment?.balanceSecs
  const tone =
    balanceSecs === undefined || balanceSecs === 0
      ? 'neutral'
      : balanceSecs > 0
        ? 'positive'
        : 'negative'

  return (
    <Card title="Banco de Horas">
      <div className="balance-card-header">
        <div>
          {balanceSecs === undefined ? (
            <p className="muted balance-state">
              {busy ? 'Carregando saldo atual...' : 'Saldo indisponível.'}
            </p>
          ) : (
            <>
              <p className={`balance-value ${tone}`}>
                {formatSignedDurationSecs(balanceSecs)}
              </p>
              <p className="muted balance-state">{balanceLabel(balanceSecs)}</p>
            </>
          )}
        </div>
        <Button variant="secondary" onClick={onRefresh} disabled={busy || !canRefresh}>
          {busy ? 'Atualizando...' : 'Atualizar Saldo'}
        </Button>
      </div>

      {error ? <p className="balance-message error">{error}</p> : null}

      {adjustment?.appliesToday ? (
        <div className="balance-guidance">
          {adjustment.targetAdjustmentSecs > 0 ? (
            <p>
              Planejado hoje: <strong>{formatSignedDurationSecs(adjustment.targetAdjustmentSecs)}</strong>{' '}
              de trabalho, gerando crédito estimado de{' '}
              <strong>{formatSignedDurationSecs(adjustment.estimatedBalanceChangeSecs)}</strong>{' '}
              com multiplicador de {adjustment.multiplier.toFixed(1)}x.
            </p>
          ) : adjustment.targetAdjustmentSecs < 0 ? (
            <p>
              Planejado hoje: usar{' '}
              <strong>{formatDurationSecs(-adjustment.targetAdjustmentSecs)}</strong>{' '}
              do banco para encerrar mais cedo.
            </p>
          ) : (
            <p>Nenhum ajuste de banco é necessário para hoje.</p>
          )}
          <p className="muted small">
            Saldo estimado após o plano:{' '}
            {formatSignedDurationSecs(adjustment.remainingBalanceSecs)}.
          </p>
          {adjustment.capped ? (
            <p className="balance-message warning">
              Limite saudável aplicado: no máximo 03:00:00 de trabalho extra hoje.
            </p>
          ) : null}
        </div>
      ) : adjustment ? (
        <p className="muted small balance-guidance">
          O saldo atual é apenas informativo para esta data. Ajustes automáticos são
          aplicados somente ao dia de hoje.
        </p>
      ) : null}

      {balance?.updatedAt ? (
        <p className="muted small balance-updated">
          Última liquidação informada: {formatUpdatedAt(balance.updatedAt)}
        </p>
      ) : null}
    </Card>
  )
}
