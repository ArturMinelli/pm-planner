import { useState } from 'react'
import type { ReminderSettings } from '../../types'
import { Button, Card, Field } from '../../components/ui'
import {
  formatReminderLead,
  MAX_REMINDER_LEAD_MINUTES,
  normalizeReminderSettings,
} from '../../util/reminderSettings'

type ReminderSettingsFormProps = {
  settings: ReminderSettings
  busy: boolean
  canUseRuntime: boolean
  onSettingsChange: (settings: ReminderSettings) => void
  onSave: () => void
  onTest: () => void
}

type ReminderLeadUnit = 'minutes' | 'hours'

export default function ReminderSettingsForm({
  settings,
  busy,
  canUseRuntime,
  onSettingsChange,
  onSave,
  onTest,
}: ReminderSettingsFormProps) {
  const [leadAmount, setLeadAmount] = useState('')
  const [leadUnit, setLeadUnit] = useState<ReminderLeadUnit>('minutes')
  const normalized = normalizeReminderSettings(settings)
  const leadMinutes = normalized.lead_minutes ?? []
  const parsedLeadAmount = Number(leadAmount)
  const candidateLeadMinutes =
    Number.isInteger(parsedLeadAmount) && parsedLeadAmount > 0
      ? parsedLeadAmount * (leadUnit === 'hours' ? 60 : 1)
      : null
  const canAddLead =
    candidateLeadMinutes !== null &&
    candidateLeadMinutes <= MAX_REMINDER_LEAD_MINUTES &&
    !leadMinutes.includes(candidateLeadMinutes)

  const update = (patch: Partial<ReminderSettings>) => {
    onSettingsChange(normalizeReminderSettings({ ...normalized, ...patch }))
  }
  const addLead = () => {
    if (!canAddLead || candidateLeadMinutes === null) return
    update({ lead_minutes: [...leadMinutes, candidateLeadMinutes] })
    setLeadAmount('')
  }
  const removeLead = (lead: number) => {
    if (leadMinutes.length <= 1) return
    update({ lead_minutes: leadMinutes.filter((value) => value !== lead) })
  }

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
        </div>

        <Field
          id="settings-reminder-lead-amount"
          label="Avisar antes"
          hint="Adicione quantos avisos quiser, de 1 minuto até 4 horas antes. Mantenha pelo menos um."
        >
          <div className="reminder-lead-builder">
            <div className="reminder-lead-add">
              <input
                id="settings-reminder-lead-amount"
                name="reminder_lead_amount"
                type="number"
                min="1"
                max={leadUnit === 'hours' ? 4 : MAX_REMINDER_LEAD_MINUTES}
                step="1"
                inputMode="numeric"
                placeholder={leadUnit === 'hours' ? 'Ex.: 2' : 'Ex.: 30'}
                value={leadAmount}
                onChange={(event) => setLeadAmount(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key !== 'Enter') return
                  event.preventDefault()
                  addLead()
                }}
              />
              <select
                aria-label="Unidade do aviso"
                value={leadUnit}
                onChange={(event) =>
                  setLeadUnit(event.target.value as ReminderLeadUnit)
                }
              >
                <option value="minutes">minutos</option>
                <option value="hours">horas</option>
              </select>
              <Button
                type="button"
                variant="secondary"
                disabled={!canAddLead}
                onClick={addLead}
              >
                Adicionar
              </Button>
            </div>
            <div className="reminder-lead-list" aria-label="Avisos configurados">
              {leadMinutes.map((lead) => (
                <span className="reminder-lead-chip" key={lead}>
                  {formatReminderLead(lead)}
                  <button
                    type="button"
                    disabled={leadMinutes.length <= 1}
                    aria-label={`Remover aviso de ${formatReminderLead(lead)}`}
                    title={
                      leadMinutes.length <= 1
                        ? 'Adicione outro aviso antes de remover este'
                        : 'Remover aviso'
                    }
                    onClick={() => removeLead(lead)}
                  >
                    x
                  </button>
                </span>
              ))}
            </div>
          </div>
        </Field>

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
            onClick={onTest}
          >
            Testar Lembrete
          </Button>
        </div>
      </form>
    </Card>
  )
}
