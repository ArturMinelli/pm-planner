import { useCallback, useEffect, useState } from 'react'
import type { ClockInStatus, RegisterTimeCardResult } from '../../types'
import * as backend from '../../services/backend'
import { resolveClockInLocation } from '../../util/geolocation'
import { Banner, Button, Card, Toast } from '../../components/ui'

type ClockInPhase = 'idle' | 'confirming' | 'locating' | 'registering'

type ClockInCardProps = {
  onRegistered?: (result: RegisterTimeCardResult) => void
}

function formatLastPunch(status: ClockInStatus | null): string {
  if (!status?.hasRecent) {
    return 'Nenhum registro recente carregado.'
  }
  const parts = [`${status.date} às ${status.time}`]
  if (status.address) {
    parts.push(status.address)
  }
  return parts.join(' · ')
}

export default function ClockInCard({ onRegistered }: ClockInCardProps) {
  const hasRuntime = backend.hasWailsRuntime()
  const [status, setStatus] = useState<ClockInStatus | null>(null)
  const [statusError, setStatusError] = useState<string | null>(null)
  const [phase, setPhase] = useState<ClockInPhase>('idle')
  const [actionError, setActionError] = useState<string | null>(null)
  const [success, setSuccess] = useState<RegisterTimeCardResult | null>(null)

  const busy = phase === 'locating' || phase === 'registering'

  const refreshStatus = useCallback(async () => {
    if (!hasRuntime) {
      return
    }
    setStatusError(null)
    try {
      const next = await backend.getClockInStatus()
      setStatus(next)
    } catch (error) {
      setStatusError(
        error instanceof Error ? error.message : 'Falha ao carregar último ponto.',
      )
    }
  }, [hasRuntime])

  useEffect(() => {
    void refreshStatus()
  }, [refreshStatus])

  const resetActionState = () => {
    setPhase('idle')
    setActionError(null)
  }

  const handleRegister = async () => {
    setActionError(null)
    setSuccess(null)
    setPhase('locating')
    try {
      const location = await resolveClockInLocation()
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
      })
      onRegistered?.(result)
      setPhase('idle')
    } catch (error) {
      setActionError(
        error instanceof Error ? error.message : 'Falha ao registrar ponto.',
      )
      setPhase('idle')
    }
  }

  const primaryLabel = (() => {
    switch (phase) {
      case 'confirming':
        return 'Confirmar registro'
      case 'locating':
        return 'Obtendo localização...'
      case 'registering':
        return 'Registrando ponto...'
      default:
        return 'Bater ponto'
    }
  })()

  return (
    <Card
      title="Registrar ponto"
      intro="Registra um novo ponto na PontoMais usando sua localização atual."
    >
      <div className="clockin-card">
        <p className="muted clockin-last">{formatLastPunch(status)}</p>

        {statusError ? <Banner tone="error">{statusError}</Banner> : null}
        {actionError ? <Banner tone="error">{actionError}</Banner> : null}
        {success ? (
          <Toast tone="success" onDismiss={() => setSuccess(null)}>
            Ponto registrado às {success.time} em {success.date}.
          </Toast>
        ) : null}

        <div className="btn-row clockin-actions">
          {phase === 'confirming' ? (
            <Button variant="secondary" onClick={resetActionState} disabled={busy}>
              Cancelar
            </Button>
          ) : null}
          <Button
            variant="primary"
            className="clockin-primary"
            disabled={!hasRuntime || busy}
            aria-busy={busy}
            onClick={() => {
              if (phase === 'confirming') {
                void handleRegister()
                return
              }
              setActionError(null)
              setSuccess(null)
              setPhase('confirming')
            }}
          >
            {primaryLabel}
          </Button>
        </div>

        {phase === 'confirming' ? (
          <p className="hint clockin-confirm">
            Confirme para registrar um ponto real na sua conta PontoMais.
          </p>
        ) : null}

        {!hasRuntime ? (
          <p className="hint">Abra pelo app desktop para bater ponto.</p>
        ) : null}
      </div>
    </Card>
  )
}
