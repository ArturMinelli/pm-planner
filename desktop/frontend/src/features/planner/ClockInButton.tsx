import { useTranslation } from 'react-i18next'
import type { ClockInResult } from '../../types'
import { useClockIn } from '../../hooks/useClockIn'
import { Button, Card } from '../../components/ui'

type ClockInButtonProps = {
  onRegistered: () => void
}

function failureKey(code: ClockInResult['code'] | undefined): string {
  if (code === 'session_expired') return 'clockIn.sessionExpired'
  if (code === 'offline') return 'clockIn.offline'
  return 'clockIn.notRegistered'
}

export default function ClockInButton({ onRegistered }: ClockInButtonProps) {
  const { t } = useTranslation()
  const { phase, result, run, reset } = useClockIn()
  const busy = phase === 'locating' || phase === 'sending'
  const tooClose = phase === 'failed' && result?.code === 'too_close'
  const failed = phase === 'failed' && !tooClose

  async function submit(confirmNear: boolean) {
    const outcome = await run(confirmNear)
    if (outcome?.status === 'registered') onRegistered()
  }

  let message: string | null = null
  let tone: 'info' | 'success' | 'warning' | 'error' = 'info'
  if (phase === 'locating') message = t('clockIn.locating')
  if (phase === 'sending') message = t('clockIn.sending')
  if (phase === 'uncertain') {
    tone = 'warning'
    message = t('clockIn.uncertain')
  }
  if (phase === 'registered') {
    tone = 'success'
    const place = result?.locationSource === 'detected'
      ? t('clockIn.locationDetected')
      : t('clockIn.locationDefault')
    message = `${t('clockIn.registered', { time: result?.punch?.time ?? '' })} · ${place}`
  }
  if (tooClose) {
    tone = 'warning'
    message = t('clockIn.tooClose')
  }
  if (failed) {
    tone = 'error'
    message = t(failureKey(result?.code))
  }

  return (
    <Card title={t('clockIn.title')}>
      <Button
        className="clockin-action"
        onClick={() => void submit(false)}
        disabled={busy}
        aria-busy={busy}
      >
        {t('clockIn.action')}
      </Button>
      {/* Wrapper reserves the status line's height so states do not shift the card. */}
      <div className="clockin-status">
        {/* Always mounted so screen readers announce changes to it. */}
        <div
          role="status"
          aria-live="polite"
          className={message ? `banner ${tone}` : undefined}
        >
          {busy ? <span className="clockin-spinner" aria-hidden="true" /> : null}
          {message}
          {phase === 'failed' && result?.detail ? (
            <small className="clockin-detail">{result.detail}</small>
          ) : null}
        </div>
      </div>
      {tooClose ? (
        <div className="btn-row">
          <Button className="clockin-btn" onClick={() => void submit(true)}>
            {t('clockIn.confirm')}
          </Button>
          <Button className="clockin-btn" variant="secondary" onClick={reset}>
            {t('clockIn.cancel')}
          </Button>
        </div>
      ) : null}
      {failed ? (
        <div className="btn-row">
          <Button className="clockin-btn" onClick={() => void submit(false)}>
            {t('clockIn.retry')}
          </Button>
        </div>
      ) : null}
    </Card>
  )
}
