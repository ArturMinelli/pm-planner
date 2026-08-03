/**
 * @vitest-environment jsdom
 */
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { PmConfig, ReminderStatus } from '../types'
import * as backend from '../services/backend'
import SettingsPage from './SettingsPage'

vi.mock('../services/backend', () => ({
  hasWailsRuntime: vi.fn(),
  getConfig: vi.fn(),
  getReminderStatus: vi.fn(),
}))

const mockedBackend = vi.mocked(backend)

async function flushPromises() {
  await Promise.resolve()
  await Promise.resolve()
  await Promise.resolve()
}

describe('SettingsPage', () => {
  let container: HTMLDivElement
  let root: Root

  beforeEach(() => {
    vi.clearAllMocks()
    mockedBackend.hasWailsRuntime.mockReturnValue(true)
    mockedBackend.getConfig.mockResolvedValue({
      login: 'user@example.com',
      password: 'secret',
    } satisfies PmConfig)
    mockedBackend.getReminderStatus.mockResolvedValue({
      settings: {
        enabled: false,
        autostart: false,
        lead_minutes: [15, 5],
        native_notifications: true,
      },
      autostartEnabled: false,
      daemonRunning: false,
      notificationAvailable: true,
      notificationAuthorized: true,
    } satisfies ReminderStatus)

    container = document.createElement('div')
    document.body.appendChild(container)
    root = createRoot(container)
  })

  afterEach(() => {
    act(() => {
      root.unmount()
    })
    container.remove()
  })

  async function renderPage() {
    await act(async () => {
      root.render(<SettingsPage />)
      await flushPromises()
    })
  }

  it('stacks Account, Planner, and Reminders sections in single-column order', async () => {
    await renderPage()

    const grid = container.querySelector('.settings-grid')
    expect(grid).not.toBeNull()
    expect(grid?.children.length).toBe(3)
    expect(grid?.textContent).toContain('Conta e Sessão')
    expect(grid?.textContent).toContain('Planner')
    expect(grid?.textContent).toContain('Lembretes de Jornada')

    const sections = Array.from(grid!.children)
    expect(sections[0].textContent).toContain('Conta e Sessão')
    expect(sections[1].textContent).toContain('Planner')
    expect(sections[2].textContent).toContain('Lembretes de Jornada')
  })
})
