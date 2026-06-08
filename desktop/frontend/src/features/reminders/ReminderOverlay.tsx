import { useEffect, useMemo, useState } from 'react'
import type { OverlayPayload, ReminderAnimation } from '../../types'
import * as overlay from '../../services/overlay'

const ANIMATION_TITLES: Record<ReminderAnimation, string> = {
  train: 'Trem chegando',
  rocket: 'Decolagem',
  bell: 'Sino atento',
}

function formatClock(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--:--'
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export default function ReminderOverlay() {
  const [payload, setPayload] = useState<OverlayPayload | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    overlay
      .getOverlayPayload()
      .then((next) => {
        if (!cancelled) setPayload(next)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e))
      })
    const timer = window.setTimeout(() => {
      void overlay.closeOverlay()
    }, 7200)
    return () => {
      cancelled = true
      window.clearTimeout(timer)
    }
  }, [])

  const animation = payload?.animation ?? 'train'
  const clock = useMemo(() => formatClock(payload?.slotTime ?? ''), [payload])

  return (
    <main className="overlay-root">
      <section className="overlay-card" aria-live="assertive">
        <div className={`overlay-scene ${animation}`} aria-hidden="true">
          {animation === 'train' ? <TrainAnimation /> : null}
          {animation === 'rocket' ? <RocketAnimation /> : null}
          {animation === 'bell' ? <BellAnimation /> : null}
        </div>
        <div className="overlay-copy">
          <p className="overlay-kicker">
            {ANIMATION_TITLES[animation]}
          </p>
          <h1>{payload?.slotLabel ?? 'Lembrete'}</h1>
          <p>
            {error ? error : `Horario recomendado: ${clock}`}
          </p>
        </div>
      </section>
    </main>
  )
}

function TrainAnimation() {
  return (
    <div className="train">
      <span className="train-car lead" />
      <span className="train-car" />
      <span className="train-car tail" />
      <span className="track" />
    </div>
  )
}

function RocketAnimation() {
  return (
    <div className="rocket">
      <span className="rocket-body" />
      <span className="rocket-flame" />
      <span className="rocket-trail" />
    </div>
  )
}

function BellAnimation() {
  return (
    <div className="bell">
      <span className="bell-dome" />
      <span className="bell-clapper" />
      <span className="bell-ring left" />
      <span className="bell-ring right" />
    </div>
  )
}
