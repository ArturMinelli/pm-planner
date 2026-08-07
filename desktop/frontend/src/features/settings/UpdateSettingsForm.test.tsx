import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import type { UpdateStatus } from '../../types'
import UpdateSettingsForm from './UpdateSettingsForm'

function status(overrides: Partial<UpdateStatus> = {}): UpdateStatus {
  return {
    root: '/home/user/.local/share/pm-planner',
    isGit: true,
    commitSha: 'a1b2c3d',
    commitDate: '2026-08-03T15:57:02-03:00',
    behind: 0,
    dirty: false,
    blockers: [],
    updateAvailable: false,
    ...overrides,
  }
}

function render(props: Partial<Parameters<typeof UpdateSettingsForm>[0]> = {}) {
  return renderToStaticMarkup(
    <UpdateSettingsForm
      status={null}
      checking={false}
      checkError={null}
      updating={false}
      canApplyUpdate
      onUpdate={() => undefined}
      {...props}
    />,
  )
}

describe('UpdateSettingsForm', () => {
  it('shows a skeleton while the initial check is running', () => {
    const html = render({ checking: true })

    expect(html).toContain('aria-busy="true"')
    expect(html).toContain('Verificando...')
    expect(html).not.toContain('Verificar atualizações')
    expect(html).not.toContain('Atualizar agora')
  })

  it('shows the installed version and offers the update when behind', () => {
    const html = render({ status: status({ behind: 2, updateAvailable: true }) })

    expect(html).toContain('a1b2c3d (03/08/2026)')
    expect(html).toContain('2 atualizações disponíveis.')
    expect(html).toContain('Atualizar agora')
    expect(html).toContain('fechado e reaberto automaticamente')
    expect(html).not.toContain('Verificar atualizações')
  })

  it('does not render the update button when up to date', () => {
    const html = render({ status: status() })

    expect(html).toContain('PM Planner está atualizado.')
    expect(html).not.toContain('Atualizar agora')
    expect(html).not.toContain('fechado e reaberto automaticamente')
  })

  it('surfaces blockers instead of the availability badge summary', () => {
    const html = render({
      status: status({
        dirty: true,
        blockers: [{
          key: 'update.blockers.dirty_working_tree',
          params: { root: '/home/user/pm-planner' },
        }],
      }),
    })

    expect(html).toContain('alterações locais não commitadas')
    expect(html).toContain('PM Planner está atualizado.')
  })

  it('reports progress while the update starts', () => {
    const html = render({
      status: status({ behind: 1, updateAvailable: true }),
      updating: true,
    })

    expect(html).toContain('Atualizando...')
    expect(html).toContain('aria-busy="true"')
  })

  it('shows a check error banner when the automatic check fails', () => {
    const html = render({
      checkError: 'Não foi possível verificar atualizações.',
    })

    expect(html).toContain('Não foi possível verificar atualizações.')
    expect(html).toContain('role="alert"')
  })
})
