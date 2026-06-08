import type { ReminderSettings, ReminderStatus } from '../../types'
import { Button, Card, Field } from '../../components/ui'
import {
  normalizeReminderSettings,
  reminderLeadPreset,
} from '../../util/reminderSettings'

type ReminderSettingsFormProps = {
  settings: ReminderSettings
  status: ReminderStatus | null
  busy: boolean
  canUseRuntime: boolean
  onSettingsChange: (settings: ReminderSettings) => void
  onSave: () => void
  onRequestPermission: () => void
  onTest: () => void
}

const LEAD_PRESETS = [
  { id: '15-5', label: '15 e 5 min', leads: [15, 5] },
  { id: '10-2', label: '10 e 2 min', leads: [10, 2] },
  { id: '5', label: '5 min', leads: [5] },
] as const

export default function ReminderSettingsForm({
  settings,
  status,
  busy,
  canUseRuntime,
  onSettingsChange,
  onSave,
  onRequestPermission,
  onTest,
}: ReminderSettingsFormProps) {
  const normalized = normalizeReminderSettings(settings)
  const leadPreset = reminderLeadPreset(normalized.lead_minutes ?? [])
  const update = (patch: Partial<ReminderSettings>) => {
    onSettingsChange(normalizeReminderSettings({ ...normalized, ...patch }))
  }
  const daemonText =
    status?.daemonRunning ? 'Daemon ativo' : 'Daemon parado'
  const notificationText =
    status?.notificationAuthorized ? 'Notificações autorizadas' : (
      'Notificações pendentes'
    )

  return (
    <Card title="Lembretes de Jornada">
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          onSave()
        }}
      >
        <div className="settings-check-grid">
          <label className="check-row">
            <input
              type="checkbox"
              checked={normalized.enabled}
              onChange={(event) =>
                update({
                  enabled: event.target.checked,
                  autostart: event.target.checked ? normalized.autostart : false,
                })
              }
            />
            <span>Ativar lembretes</span>
          </label>
          <label className="check-row">
            <input
              type="checkbox"
              checked={normalized.autostart}
              disabled={!normalized.enabled}
              onChange={(event) => update({ autostart: event.target.checked })}
            />
            <span>Iniciar com o sistema</span>
          </label>
          <label className="check-row">
            <input
              type="checkbox"
              checked={Boolean(normalized.native_notifications)}
              onChange={(event) =>
                update({ native_notifications: event.target.checked })
              }
            />
            <span>Notificação nativa</span>
          </label>
        </div>

        <Field id="settings-reminder-leads" label="Avisar antes">
          <div
            id="settings-reminder-leads"
            className="segmented"
            role="radiogroup"
          >
            {LEAD_PRESETS.map((preset) => (
              <button
                key={preset.id}
                type="button"
                className={leadPreset === preset.id ? 'selected' : undefined}
                aria-pressed={leadPreset === preset.id}
                onClick={() => update({ lead_minutes: [...preset.leads] })}
              >
                {preset.label}
              </button>
            ))}
          </div>
        </Field>

        <div className="status-strip">
          <span>{daemonText}</span>
          <span>{notificationText}</span>
          {status?.autostartEnabled ? <span>Autostart ativo</span> : null}
        </div>

        <div className="btn-row">
          <Button
            type="submit"
            variant="primary"
            disabled={busy || !canUseRuntime}
            aria-busy={busy}
          >
            {busy ? 'Salvando...' : 'Salvar Lembretes'}
          </Button>
          <Button
            type="button"
            variant="secondary"
            disabled={busy || !canUseRuntime}
            onClick={onRequestPermission}
          >
            Permitir Notificações
          </Button>
          <Button
            type="button"
            variant="secondary"
            disabled={busy || !canUseRuntime}
            onClick={onTest}
          >
            Testar Lembrete
          </Button>
        </div>
      </form>
    </Card>
  )
}
