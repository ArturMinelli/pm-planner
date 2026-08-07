/**
 * @vitest-environment jsdom
 */
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { PmConfig, ReminderStatus } from '../types'
import * as backend from '../services/backend'
import SettingsPage from './SettingsPage'
import { renderWithConfig } from '../test/renderWithConfig'

vi.mock('../services/backend', () => ({
  hasWailsRuntime: vi.fn(),
  usesHttpTransport: vi.fn(),
  getConfig: vi.fn(),
  saveConfig: vi.fn(),
  saveReminderSettings: vi.fn(),
  getReminderStatus: vi.fn(),
  testAuth: vi.fn(),
  checkForUpdate: vi.fn(),
  consumeUpdateResult: vi.fn(),
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
    mockedBackend.usesHttpTransport.mockReturnValue(false)
    mockedBackend.getConfig.mockResolvedValue({
      login: 'user@example.com',
      password: 'secret',
    } satisfies PmConfig)
    mockedBackend.saveConfig.mockResolvedValue(undefined)
    mockedBackend.saveReminderSettings.mockResolvedValue(undefined)
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
    mockedBackend.consumeUpdateResult.mockResolvedValue(null)
    mockedBackend.checkForUpdate.mockResolvedValue({
      root: '/home/user/.local/share/pm-planner',
      isGit: true,
      commitSha: 'a1b2c3d',
      commitDate: '2026-08-03T15:57:02-03:00',
      behind: 0,
      dirty: false,
      blockers: [],
      updateAvailable: false,
    })

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
      root.render(renderWithConfig(<SettingsPage />))
      await flushPromises()
    })
  }

  it('checks for updates automatically on mount', async () => {
    await renderPage()

    expect(mockedBackend.checkForUpdate).toHaveBeenCalledTimes(1)
    expect(container.textContent).toContain('a1b2c3d (03/08/2026)')
    expect(container.textContent).toContain('PM Planner está atualizado.')
  })

  it('stacks Account, Planner, and Reminders sections in single-column order', async () => {
    await renderPage()

    const grid = container.querySelector('.settings-grid')
    expect(grid).not.toBeNull()
    expect(grid?.children.length).toBe(3)

    const sections = Array.from(grid!.children)
    expect(sections[0].querySelector('#settings-login')).not.toBeNull()
    expect(sections[1].querySelector('#settings-anchor-in1')).not.toBeNull()
    expect(sections[2].querySelector('#settings-reminder-lead-amount')).not.toBeNull()
  })

  it('saves account settings through the config context', async () => {
    await renderPage()

    const saveButton = Array.from(container.querySelectorAll('button')).find(
      (button) => button.textContent === 'Salvar conta',
    )
    expect(saveButton).not.toBeNull()

    await act(async () => {
      saveButton?.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      await flushPromises()
    })

    expect(mockedBackend.saveConfig).toHaveBeenCalledTimes(1)
    expect(mockedBackend.saveConfig).toHaveBeenCalledWith(
      expect.objectContaining({
        login: 'user@example.com',
        password: 'secret',
      }),
    )
  })
})
