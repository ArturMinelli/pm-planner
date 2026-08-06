/**
 * @vitest-environment jsdom
 */
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ConfigProvider, useConfig } from '../context/ConfigContext'
import type { PmConfig } from '../types'
import * as backend from '../services/backend'
import { builtinPlannerAnchors } from '../util/plannerDefaults'

vi.mock('../services/backend', () => ({
  getConfig: vi.fn(),
  saveConfig: vi.fn(),
  saveReminderSettings: vi.fn(),
}))

vi.mock('../services/browserReminders', () => ({
  notifyConfigChanged: vi.fn(),
}))

const mockedBackend = vi.mocked(backend)

function defaultConfig(): PmConfig {
  return {
    login: 'user@example.com',
    password: 'secret',
    planner: builtinPlannerAnchors(),
    max_daily_extra_minutes: 180,
    balance_credit_multiplier: 1.5,
  }
}

async function flushPromises() {
  await Promise.resolve()
  await Promise.resolve()
  await Promise.resolve()
}

describe('ConfigContext', () => {
  let container: HTMLDivElement
  let root: Root
  let latestRevision = 0
  let latestConfig: PmConfig | null = null
  let mergeAndSaveRef: ((patch: Partial<PmConfig>) => Promise<void>) | null = null
  let saveReminderSettingsRef:
    | ((settings: PmConfig['reminders'] & object) => Promise<void>)
    | null = null

  function ConfigProbe() {
    const { config, configRevision, mergeAndSave, saveReminderSettings } = useConfig()
    latestRevision = configRevision
    latestConfig = config
    mergeAndSaveRef = mergeAndSave
    saveReminderSettingsRef = saveReminderSettings
    return null
  }

  beforeEach(() => {
    vi.clearAllMocks()
    latestRevision = 0
    latestConfig = null
    mergeAndSaveRef = null
    saveReminderSettingsRef = null
    mockedBackend.getConfig.mockResolvedValue(defaultConfig())
    mockedBackend.saveConfig.mockResolvedValue(undefined)
    mockedBackend.saveReminderSettings.mockResolvedValue(undefined)

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

  async function renderProvider() {
    await act(async () => {
      root.render(
        <ConfigProvider>
          <ConfigProbe />
        </ConfigProvider>,
      )
      await flushPromises()
    })
  }

  it('loads config on mount and sets revision to 1', async () => {
    await renderProvider()

    expect(mockedBackend.getConfig).toHaveBeenCalledTimes(1)
    expect(latestRevision).toBe(1)
    expect(latestConfig?.login).toBe('user@example.com')
  })

  it('mergeAndSave updates config and bumps revision', async () => {
    await renderProvider()

    mockedBackend.getConfig.mockClear()

    await act(async () => {
      await mergeAndSaveRef?.({
        planner: {
          in1: '09:30',
          out1: '12:30',
          in2: '14:00',
          out2: '19:00',
        },
      })
      await flushPromises()
    })

    expect(mockedBackend.saveConfig).toHaveBeenCalledTimes(1)
    expect(latestRevision).toBe(2)
    expect(latestConfig?.planner?.in1).toBe('09:30')
  })

  it('saveReminderSettings reloads config and bumps revision', async () => {
    await renderProvider()

    mockedBackend.getConfig.mockResolvedValue({
      ...defaultConfig(),
      reminders: {
        enabled: true,
        autostart: false,
        lead_minutes: [10],
      },
    })

    await act(async () => {
      await saveReminderSettingsRef?.({
        enabled: true,
        autostart: false,
        lead_minutes: [10],
      })
      await flushPromises()
    })

    expect(mockedBackend.saveReminderSettings).toHaveBeenCalledTimes(1)
    expect(mockedBackend.getConfig).toHaveBeenCalledTimes(2)
    expect(latestRevision).toBe(2)
    expect(latestConfig?.reminders?.lead_minutes).toEqual([10])
  })
})
