import { useCallback, useEffect, useState } from 'react'
import type { ClockInStatus, RegisterTimeCardResult } from '../../types'
import * as backend from '../../services/backend'
import {
  resolveClockInLocation,
  type GeolocationResult,
} from '../../util/geolocation'
import { formatError } from '../../util/formatError'
import { Banner, Button, Card, Toast } from '../../components/ui'

type ClockInPhase = 'idle' | 'confirming' | 'locating' | 'registering'

type ClockInCardProps = {
  embedded?: boolean
  onRegistered?: (result: RegisterTimeCardResult) => void
}

function formatPunchDate(date: string): string {
  const [year, month, day] = date.split('-')
  if (!year || !month || !day) {
    return date
  }
  return `${day}/${month}`
}

function hasFallbackLocation(status: ClockInStatus | null): status is ClockInStatus & {
  latitude: number
  longitude: number
} {
  return (
    status != null &&
    typeof status.latitude === 'number' &&
    typeof status.longitude === 'number'
  )
}

function fallbackLocation(status: ClockInStatus): GeolocationResult {
  return {
    latitude: status.latitude!,
    longitude: status.longitude!,
    accuracy: status.accuracy ?? 1100,
    address:
      status.address?.trim() ||
      `${status.latitude!.toFixed(6)}, ${status.longitude!.toFixed(6)}`,
  }
}

function LastPunchMeta({ status }: { status: ClockInStatus | null }) {
  if (!status?.hasRecent) {
    return <p className="clockin-meta-value muted">Nenhum registro recente</p>
  }

  return (
    <>
      <p className="clockin-meta-value">
        {status.time}
        <span className="clockin-meta-sep">·</span>
        {formatPunchDate(status.date)}
      </p>
      {status.address ? (
        <p className="clockin-meta-sub" title={status.address}>
          {status.address}
        </p>
      ) : null}
    </>
  )
}

export default function ClockInCard({
  embedded = false,
  onRegistered,
}: ClockInCardProps) {
  const hasRuntime = backend.hasWailsRuntime()
  const [status, setStatus] = useState<ClockInStatus | null>(null)
  const [statusError, setStatusError] = useState<string | null>(null)
  const [phase, setPhase] = useState<ClockInPhase>('idle')
  const [actionError, setActionError] = useState<string | null>(null)
  const [offerFallback, setOfferFallback] = useState(false)
  const [success, setSuccess] = useState<RegisterTimeCardResult | null>(null)

  const busy = phase === 'locating' || phase === 'registering'
  const confirming = phase === 'confirming'

  const refreshStatus = useCallback(async () => {
    if (!hasRuntime) {
      return
    }
    setStatusError(null)
    try {
      const next = await backend.getClockInStatus()
      setStatus(next)
    } catch (error) {
      setStatusError(formatError(error, 'Falha ao carregar último ponto.'))
    }
  }, [hasRuntime])

  useEffect(() => {
    void refreshStatus()
  }, [refreshStatus])

  const resetActionState = () => {
    setPhase('idle')
    setActionError(null)
    setOfferFallback(false)
  }

  const submitRegister = async (location: GeolocationResult) => {
    setPhase('registering')
    const result = await backend.registerTimeCard(
      location.latitude,
      location.longitude,
      location.accuracy,
      location.address,
    )
    setSuccess(result)
    setStatus({
      date: result.date,
      time: result.time,
      address: result.address,
      hasRecent: true,
      latitude: location.latitude,
      longitude: location.longitude,
      accuracy: location.accuracy,
    })
    onRegistered?.(result)
    setPhase('idle')
    setOfferFallback(false)
  }

  const handleRegister = async () => {
    setActionError(null)
    setSuccess(null)
    setOfferFallback(false)

    let location: GeolocationResult
    try {
      setPhase('locating')
      location = await resolveClockInLocation()
    } catch (error) {
      if (hasFallbackLocation(status)) {
        setOfferFallback(true)
        setActionError(
          `${formatError(error, 'Não foi possível obter a localização.')} Você pode usar o local do último ponto registrado.`,
        )
      } else {
        setActionError(formatError(error, 'Não foi possível obter a localização.'))
      }
      setPhase('idle')
      return
    }

    try {
      await submitRegister(location)
    } catch (error) {
      setActionError(formatError(error, 'Falha ao registrar ponto.'))
      setPhase('idle')
    }
  }

  const handleFallbackRegister = async () => {
    if (!hasFallbackLocation(status)) {
      return
    }
    setActionError(null)
    setSuccess(null)
    setOfferFallback(false)
    setPhase('registering')
    try {
      await submitRegister(fallbackLocation(status))
    } catch (error) {
      setActionError(formatError(error, 'Falha ao registrar ponto.'))
      setPhase('idle')
    }
  }

  const primaryLabel = (() => {
    switch (phase) {
      case 'confirming':
        return 'Confirmar'
      case 'locating':
        return 'Obtendo localização...'
      case 'registering':
        return 'Registrando...'
      default:
        return 'Bater ponto'
    }
  })()

  const panel = (
    <div className={embedded ? 'clockin-panel' : 'clockin-card'}>
      <div className="clockin-meta">
        <span className="clockin-meta-label">Último ponto</span>
        <LastPunchMeta status={status} />
      </div>

      {statusError ? <Banner tone="error">{statusError}</Banner> : null}
      {actionError ? <Banner tone="error">{actionError}</Banner> : null}
      {success ? (
        <Toast tone="success" onDismiss={() => setSuccess(null)}>
          Registrado às {success.time}.
        </Toast>
      ) : null}

      {confirming ? (
        <p className="hint clockin-confirm">
          Isso registra um ponto real na PontoMais.
        </p>
      ) : null}

      <div className="clockin-actions">
        {confirming ? (
          <Button variant="secondary" onClick={resetActionState} disabled={busy}>
            Cancelar
          </Button>
        ) : null}
        {offerFallback && hasFallbackLocation(status) ? (
          <>
            <Button variant="secondary" onClick={resetActionState} disabled={busy}>
              Cancelar
            </Button>
            <Button
              variant="primary"
              disabled={busy}
              onClick={() => void handleFallbackRegister()}
            >
              Usar último local
            </Button>
          </>
        ) : null}
        {!offerFallback ? (
          <Button
            variant="primary"
            className="clockin-primary"
            disabled={!hasRuntime || busy}
            aria-busy={busy}
            onClick={() => {
              if (confirming) {
                void handleRegister()
                return
              }
              setActionError(null)
              setSuccess(null)
              setOfferFallback(false)
              setPhase('confirming')
            }}
          >
            {primaryLabel}
          </Button>
        ) : null}
      </div>

      {!hasRuntime ? (
        <p className="hint clockin-runtime-hint">Disponível no app desktop.</p>
      ) : null}
    </div>
  )

  if (embedded) {
    return panel
  }

  return (
    <Card title="Registrar ponto" intro="Usa a localização do sistema.">
      {panel}
    </Card>
  )
}
