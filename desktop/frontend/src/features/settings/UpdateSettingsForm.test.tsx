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
      updating={false}
      canApplyUpdate
      onCheck={() => undefined}
      onUpdate={() => undefined}
      {...props}
    />,
  )
}

describe('UpdateSettingsForm', () => {
  it('asks for a check before anything is known', () => {
    const html = render()

    expect(html).toContain('Verificar atualizações')
    expect(html).toContain('Verifique para descobrir')
    expect(html).toContain('disabled=""')
  })

  it('shows the installed version and offers the update when behind', () => {
    const html = render({ status: status({ behind: 2, updateAvailable: true }) })

    expect(html).toContain('a1b2c3d (03/08/2026)')
    expect(html).toContain('2 atualizações disponíveis.')
    expect(html).toContain('fechado e reaberto automaticamente')
  })

  it('keeps the update button disabled when up to date', () => {
    const html = render({ status: status() })

    expect(html).toContain('PM Planner está atualizado.')
    expect(html).not.toContain('fechado e reaberto automaticamente')
  })

  it('surfaces blockers instead of the availability summary', () => {
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
    expect(html).not.toContain('PM Planner está atualizado.')
  })

  it('reports progress while the update starts', () => {
    const html = render({
      status: status({ behind: 1, updateAvailable: true }),
      updating: true,
    })

    expect(html).toContain('Atualizando...')
    expect(html).toContain('aria-busy="true"')
  })
})
