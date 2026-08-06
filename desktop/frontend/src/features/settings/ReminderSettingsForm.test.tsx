import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import ReminderSettingsForm from './ReminderSettingsForm'

describe('ReminderSettingsForm', () => {
  it('shows custom reminder leads without notification status controls', () => {
    const html = renderToStaticMarkup(
      <ReminderSettingsForm
        settings={{
          enabled: true,
          autostart: true,
          lead_minutes: [120, 15],
          native_notifications: false,
        }}
        busy={false}
        canUseRuntime
        onSettingsChange={() => undefined}
        onSave={() => undefined}
      />,
    )

    expect(html).toContain('Notify before')
    expect(html).toContain('2 hours')
    expect(html).toContain('15 minutes')
    expect(html).toContain('Add')
    expect(html).not.toContain('Notificação nativa')
    expect(html).not.toContain('Permitir Notificações')
    expect(html).not.toContain('Daemon ativo')
    expect(html).not.toContain('Autostart ativo')
  })
})
