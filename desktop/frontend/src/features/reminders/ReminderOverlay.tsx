import { useCallback, useEffect, useMemo, useState } from 'react'
import type { OverlayPayload, ReminderAnimation } from '../../types'
import * as overlay from '../../services/overlay'

type OverlayPhase = 'sweeping' | 'docked'

const SWEEP_MS = 3200
const AUTO_CLOSE_MS = 8000

const ANIMATION_TITLES: Record<ReminderAnimation, string> = {
  train: 'Trem chegando',
  rocket: 'Decolagem',
  bell: 'Sino atento',
}

const ANIMATIONS: ReminderAnimation[] = ['train', 'rocket', 'bell']

function formatClock(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--:--'
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export default function ReminderOverlay({ preview = false }: { preview?: boolean }) {
  const [payload, setPayload] = useState<OverlayPayload | null>(() =>
    preview ? createPreviewPayload() : null,
  )
  const [error, setError] = useState<string | null>(null)
  const [phase, setPhase] = useState<OverlayPhase>('sweeping')

  const close = useCallback(() => {
    void overlay.closeOverlay()
  }, [])

  useEffect(() => {
    let cancelled = false
    if (preview) {
      return
    }
    overlay
      .getOverlayPayload()
      .then((next) => {
        if (!cancelled) setPayload(next)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e))
      })
    return () => {
      cancelled = true
    }
  }, [preview])

  useEffect(() => {
    const dock = () => {
      setPhase('docked')
      void overlay.dockOverlay()
    }
    const timer = window.setTimeout(() => {
      close()
    }, AUTO_CLOSE_MS)
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      dock()
      return () => window.clearTimeout(timer)
    }

    const dockTimer = window.setTimeout(dock, SWEEP_MS)
    return () => {
      window.clearTimeout(dockTimer)
      window.clearTimeout(timer)
    }
  }, [close])

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') close()
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [close])

  const animation = payload?.animation ?? 'train'
  const clock = useMemo(() => formatClock(payload?.slotTime ?? ''), [payload])
  const previewClass = preview ? ' preview' : ''

  return (
    <main className={`overlay-root ${phase} overlay-${animation}${previewClass}`}>
      <section className="overlay-sweep" aria-hidden="true">
        <span className="sweep-line" />
        <div className={`sweep-vehicle vehicle-${animation}`}>
          {animation === 'train' ? <TrainAnimation /> : null}
          {animation === 'rocket' ? <RocketAnimation /> : null}
          {animation === 'bell' ? <BellAnimation /> : null}
        </div>
      </section>
      <section className="overlay-card" aria-live="assertive">
        <div className={`overlay-scene scene-${animation}`} aria-hidden="true">
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
        <button
          type="button"
          className="overlay-dismiss"
          aria-label="Fechar lembrete"
          onClick={close}
        >
          x
        </button>
      </section>
    </main>
  )
}

function createPreviewPayload(): OverlayPayload {
  const params = new URLSearchParams(globalThis.location.search)
  const requestedAnimation = params.get('animation') as ReminderAnimation | null
  const animation = requestedAnimation && ANIMATIONS.includes(requestedAnimation)
    ? requestedAnimation
    : 'train'
  const now = new Date()
  now.setMinutes(now.getMinutes() + 5)
  return {
    id: 'preview',
    date: now.toISOString().slice(0, 10),
    slotKey: 'preview',
    slotLabel: params.get('label') ?? 'Saida 2',
    slotTime: now.toISOString(),
    leadMinutes: 5,
    fireAt: new Date().toISOString(),
    animation,
  }
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
