import { useEffect, useState } from 'react'
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

function useNow() {
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 1000)
    return () => clearInterval(id)
  }, [])
  return now
}

const svgProps = {
  viewBox: '0 0 24 24',
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
  'aria-hidden': true,
}

function PunchIcon() {
  return (
    <svg {...svgProps} className="clockin-icon">
      <circle cx="12" cy="12" r="9" />
      <path d="M12 7v5l3 2" />
    </svg>
  )
}

function PinIcon() {
  return (
    <svg {...svgProps} className="clockin-chip-icon">
      <path d="M12 21s-6-5.6-6-10a6 6 0 0 1 12 0c0 4.4-6 10-6 10z" />
      <circle cx="12" cy="11" r="2" />
    </svg>
  )
}

function AlertIcon() {
  return (
    <svg {...svgProps} className="clockin-icon">
      <path d="M12 3 2.5 20h19L12 3z" />
      <path d="M12 10v4M12 17.5v.01" />
    </svg>
  )
}

function CheckMark() {
  return (
    <svg {...svgProps} className="clockin-check" strokeWidth={2.2}>
      <circle className="clockin-check-ring" cx="12" cy="12" r="10" pathLength={1} />
      <path className="clockin-check-tick" d="M7.5 12.5l3 3 6-6.5" pathLength={1} />
    </svg>
  )
}

export default function ClockInButton({ onRegistered }: ClockInButtonProps) {
  const { t } = useTranslation()
  const { phase, result, run, reset } = useClockIn()
  const now = useNow()
  const clock = now.toLocaleTimeString('pt-BR', { hour12: false })
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

  const step = phase === 'locating' ? 0 : phase === 'sending' ? 1 : phase === 'registered' ? 2 : -1
  const steps = [t('clockIn.stepLocation'), t('clockIn.stepSending'), t('clockIn.stepDone')]
  const warn = phase === 'uncertain' || tooClose || failed

  return (
    <Card title={t('clockIn.title')}>
      <div className="clockin" data-phase={tooClose ? 'too_close' : phase}>
        <div className="clockin-readout" aria-hidden="true">
          {phase === 'registered' ? (
            <>
              <CheckMark />
              <span className="clockin-time clockin-time-done">{result?.punch?.time ?? clock}</span>
              <span className="clockin-chip">
                <PinIcon />
                {result?.locationSource === 'detected'
                  ? t('clockIn.locationDetected')
                  : t('clockIn.locationDefault')}
              </span>
            </>
          ) : (
            <>
              {warn ? <AlertIcon /> : null}
              <span className="clockin-time">{clock}</span>
            </>
          )}
        </div>
        <div className="clockin-stage">
          <Button
            className="clockin-action"
            onClick={() => void submit(false)}
            disabled={busy}
            aria-busy={busy}
          >
            <PunchIcon />
            {t('clockIn.action')}
          </Button>
        </div>
        <ol className="clockin-steps" aria-hidden="true" data-visible={busy || phase === 'registered'}>
          {steps.map((label, i) => (
            <li key={label} className={i < step ? 'done' : i === step ? 'active' : undefined}>
              <span className="clockin-dot" />
              {label}
            </li>
          ))}
        </ol>
      </div>
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
      <div className="btn-row clockin-actions">
      {tooClose ? (
        <>
          <Button className="clockin-btn" onClick={() => void submit(true)}>
            {t('clockIn.confirm')}
          </Button>
          <Button className="clockin-btn" variant="secondary" onClick={reset}>
            {t('clockIn.cancel')}
          </Button>
        </>
      ) : null}
      {failed ? (
        <Button className="clockin-btn" onClick={() => void submit(false)}>
          {t('clockIn.retry')}
        </Button>
      ) : null}
      </div>
    </Card>
  )
}
