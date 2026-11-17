import { useTranslation } from 'react-i18next'
import type { PunchLogState, PunchRecord } from '../../types'
import { Button, Card } from '../../components/ui'

type Props = {
  state: PunchLogState | null
  syncing: boolean
  isToday: boolean
  onSync: () => void
}

const hhmm = (time: string) => time.slice(0, 5)

const svgProps = {
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
  'aria-hidden': true,
}

export function SyncIcon({ spinning }: { spinning: boolean }) {
  return (
    <svg {...svgProps} className={spinning ? 'punchlog-sync-icon spinning' : 'punchlog-sync-icon'}>
      <path d="M20 11a8 8 0 0 0-14.5-4.5L4 8" />
      <path d="M4 4v4h4" />
      <path d="M4 13a8 8 0 0 0 14.5 4.5L20 16" />
      <path d="M20 20v-4h-4" />
    </svg>
  )
}

function AlertIcon() {
  return (
    <svg {...svgProps} className="punchlog-warning-icon">
      <path d="M12 3 2.5 20h19L12 3z" />
      <path d="M12 10v4M12 17.5v.01" />
    </svg>
  )
}

const STATUS_KEY: Record<PunchRecord['status'], string> = {
  synced: 'punchLog.statusSynced',
  pending: 'punchLog.statusPending',
  unconfirmed: 'punchLog.statusUnconfirmed',
  failed: 'punchLog.statusFailed',
}

export function UnconfirmedWarning({
  state,
  syncing,
  onSync,
}: {
  state: PunchLogState | null
  syncing: boolean
  onSync: () => void
}) {
  const { t } = useTranslation()
  if (!state || state.unconfirmedCount <= 0) return null
  const times = state.punches.filter((p) => p.status === 'unconfirmed').map((p) => hhmm(p.time))
  // Counts are global; the visible day may hold none of the unconfirmed records.
  const message = times.length > 0
    ? t('punchLog.warningTimes', { count: times.length, times: times.join(', ') })
    : t('punchLog.warningCount', { count: state.unconfirmedCount })
  return (
    <div role="alert" className="punchlog-warning">
      <AlertIcon />
      <div className="punchlog-warning-body">
        <strong>{t('punchLog.warningTitle')}</strong>
        <p>{message}</p>
        <p className="punchlog-warning-help">{t('punchLog.warningHelp')}</p>
      </div>
      <Button variant="secondary" onClick={onSync} disabled={syncing} aria-busy={syncing}>
        {syncing ? t('punchLog.verifying') : t('punchLog.verify')}
      </Button>
    </div>
  )
}

export default function PunchLogCard({ state, syncing, isToday, onSync }: Props) {
  const { t } = useTranslation()
  const punches = state?.punches ?? []
  const lastSync = state?.lastSyncAt
    ? new Date(state.lastSyncAt).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
    : null
  return (
    <Card title={isToday ? t('punchLog.title') : t('punchLog.titleOther')}>
      {punches.length === 0 ? (
        <p className="muted punchlog-empty">{t('punchLog.empty')}</p>
      ) : (
        <ul className="punchlog-list" aria-label={t('punchLog.listLabel')}>
          {punches.map((p) => (
            <li key={p.localId} className="punchlog-row" data-disabled={p.disabled ? 'true' : undefined}>
              <span className="punchlog-time">{hhmm(p.time)}</span>
              <span className="punchlog-origin">
                {p.source === 'app' ? t('punchLog.sourceApp') : t('punchLog.sourceRemote')}
                {p.disabled ? ` · ${t('punchLog.disabled')}` : ''}
              </span>
              <span className={`punchlog-badge ${p.status}`}>
                {p.status === 'pending' ? <span className="punchlog-dot" aria-hidden="true" /> : null}
                {t(STATUS_KEY[p.status])}
              </span>
            </li>
          ))}
        </ul>
      )}
      <div className="punchlog-footer">
        <span className="muted punchlog-last">
          {lastSync ? t('punchLog.lastSync', { time: lastSync }) : t('punchLog.neverSynced')}
        </span>
        <button
          type="button"
          className="icon-btn punchlog-sync"
          onClick={onSync}
          disabled={syncing}
          aria-busy={syncing}
          aria-label={t('punchLog.syncAction')}
          title={t('punchLog.syncAction')}
        >
          <SyncIcon spinning={syncing} />
        </button>
      </div>
      {state?.lastSyncError ? (
        <div className="banner warning punchlog-error" role="status">
          {t('punchLog.syncError')}
        </div>
      ) : null}
    </Card>
  )
}
