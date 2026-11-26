import { useState } from 'react'
import { useTranslation } from 'react-i18next'
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
  onSettingsChange: (settings: ReminderSettings) => void
  onSave: () => void
}

export default function ReminderSettingsForm({
  settings,
  busy,
  onSettingsChange,
  onSave,
}: ReminderSettingsFormProps) {
  const { t } = useTranslation()
  const [leadAmount, setLeadAmount] = useState('')
  const normalized = normalizeReminderSettings(settings)
  const leadMinutes = normalized.lead_minutes ?? []
  const parsedLeadAmount = Number(leadAmount)
  const candidateLeadMinutes =
    Number.isInteger(parsedLeadAmount) && parsedLeadAmount > 0 ? parsedLeadAmount : null
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
    <Card collapsible title={t('settings.reminders.title')}>
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          onSave()
        }}
      >
        <Field
          id="settings-reminder-lead-amount"
          label={t('settings.reminders.notifyBefore')}
          hint={t('settings.reminders.notifyBeforeHint')}
        >
          <div className="reminder-lead-builder">
            <div className="reminder-lead-add">
              <input
                id="settings-reminder-lead-amount"
                name="reminder_lead_amount"
                type="number"
                min="1"
                max={MAX_REMINDER_LEAD_MINUTES}
                step="1"
                inputMode="numeric"
                placeholder={t('settings.reminders.amountPlaceholderMinutes')}
                value={leadAmount}
                onChange={(event) => setLeadAmount(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key !== 'Enter') return
                  event.preventDefault()
                  addLead()
                }}
              />
              <Button
                type="button"
                variant="secondary"
                disabled={!canAddLead}
                onClick={addLead}
              >
                {t('common.add')}
              </Button>
            </div>
            <div className="reminder-lead-list" aria-label={t('settings.reminders.configuredNotices')}>
              {leadMinutes.map((lead) => (
                <span className="reminder-lead-chip" key={lead}>
                  {formatReminderLead(t, lead)}
                  <button
                    type="button"
                    disabled={leadMinutes.length <= 1}
                    aria-label={t('settings.reminders.removeNotice', {
                      lead: formatReminderLead(t, lead),
                    })}
                    title={
                      leadMinutes.length <= 1
                        ? t('settings.reminders.removeNoticeDisabled')
                        : t('settings.reminders.removeNoticeTitle')
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
            disabled={busy}
            aria-busy={busy}
          >
            {busy ? t('common.saving') : t('settings.reminders.save')}
          </Button>
        </div>
      </form>
    </Card>
  )
}
