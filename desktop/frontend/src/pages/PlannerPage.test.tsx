/**
 * @vitest-environment jsdom
 */
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { PmConfig, PlannerPayload, PlannerSummary } from '../types'
import * as backend from '../services/backend'
import PlannerPage from './PlannerPage'
import { builtinPlannerAnchors } from '../util/plannerDefaults'

vi.mock('../services/backend', () => ({
  hasWailsRuntime: vi.fn(),
  getConfig: vi.fn(),
  loadPlanner: vi.fn(),
  recalculateDay: vi.fn(),
}))

vi.mock('../util/timeFormat', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../util/timeFormat')>()
  return {
    ...actual,
    localDateYYYYMMDD: () => '2026-06-09',
  }
})

const mockedBackend = vi.mocked(backend)

function defaultConfig(): PmConfig {
  return {
    login: '',
    password: '',
    planner: builtinPlannerAnchors(),
  }
}

function setInputValue(input: HTMLInputElement, value: string) {
  const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
    window.HTMLInputElement.prototype,
    'value',
  )?.set
  nativeInputValueSetter?.call(input, value)
  input.dispatchEvent(new Event('input', { bubbles: true }))
}

function samplePayload(overrides: Partial<PlannerPayload> = {}): PlannerPayload {
  return {
    date: '2026-06-09',
    baseTargetSecs: 8.5 * 60 * 60,
    originalTimes: [],
    journeys: [
      {
        entry: { time: '08:00', registered: false },
        exit: { time: '12:00', registered: false },
      },
      {
        entry: { time: '13:30', registered: false },
        exit: { time: '18:00', registered: false },
      },
    ],
    solvedSlot: { journeyIndex: 1, isEntry: false },
    originalsLine: '',
    ...overrides,
  }
}

function sampleSummary(): PlannerSummary {
  return {
    journeys: samplePayload().journeys,
    solvedSlot: { journeyIndex: 1, isEntry: false },
    journeySpanSecs: [4 * 60 * 60, 4.5 * 60 * 60],
    totalSpanSecs: 8.5 * 60 * 60,
    overtimeSecs: 0,
  }
}

async function flushPromises() {
  await Promise.resolve()
  await Promise.resolve()
  await Promise.resolve()
}

describe('PlannerPage', () => {
  let container: HTMLDivElement
  let root: Root

  beforeEach(() => {
    vi.clearAllMocks()
    mockedBackend.hasWailsRuntime.mockReturnValue(true)
    mockedBackend.getConfig.mockResolvedValue(defaultConfig())
    mockedBackend.loadPlanner.mockImplementation(async (date) =>
      samplePayload({ date }),
    )
    mockedBackend.recalculateDay.mockResolvedValue(sampleSummary())

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
      root.render(<PlannerPage />)
      await flushPromises()
    })
  }

  it('auto-loads the selected date on mount', async () => {
    await renderPage()

    expect(mockedBackend.loadPlanner).toHaveBeenCalledTimes(1)
    expect(mockedBackend.loadPlanner).toHaveBeenCalledWith('2026-06-09')
    expect(mockedBackend.getConfig).toHaveBeenCalledTimes(1)
    expect(mockedBackend.recalculateDay).toHaveBeenCalledTimes(1)
    expect(container.textContent).toContain('Meta do dia')
    expect(container.textContent).toContain('Entrada 1')
    expect(container.textContent).not.toContain('Carregar Dia')
  })

  it('re-fetches when the user changes the date', async () => {
    await renderPage()

    mockedBackend.loadPlanner.mockClear()
    mockedBackend.getConfig.mockClear()
    mockedBackend.recalculateDay.mockClear()

    await act(async () => {
      document.getElementById('planner-date')?.click()
      await flushPromises()
    })

    const dayFifteen = Array.from(document.querySelectorAll('.calendar-day')).find(
      (element) =>
        element.textContent === '15' && !element.classList.contains('outside'),
    )

    await act(async () => {
      dayFifteen?.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      await flushPromises()
    })

    expect(mockedBackend.loadPlanner).toHaveBeenCalledTimes(1)
    expect(mockedBackend.loadPlanner).toHaveBeenCalledWith('2026-06-15')
    expect(mockedBackend.getConfig).toHaveBeenCalledTimes(1)
    expect(mockedBackend.recalculateDay).toHaveBeenCalledTimes(1)
  })

  it('uses configured anchors for unregistered journey fields', async () => {
    mockedBackend.getConfig.mockResolvedValue({
      login: '',
      password: '',
      planner: {
        in1: '09:00',
        out1: '12:30',
        in2: '14:00',
        out2: '19:00',
      },
    })
    mockedBackend.loadPlanner.mockResolvedValue(
      samplePayload({
        journeys: [
          {
            entry: { time: '08:00', registered: false },
            exit: { time: '12:00', registered: false },
          },
          {
            entry: { time: '13:30', registered: false },
            exit: { time: '18:00', registered: false },
          },
        ],
        solvedSlot: { journeyIndex: 1, isEntry: false },
      }),
    )

    await renderPage()

    const journeyInputs = container.querySelectorAll('.journey-list input[type="text"]')
    expect(journeyInputs).toHaveLength(3)
    expect((journeyInputs[0] as HTMLInputElement).value).toBe('09:00')
    expect((journeyInputs[1] as HTMLInputElement).value).toBe('12:30')
    expect((journeyInputs[2] as HTMLInputElement).value).toBe('14:00')

    await act(async () => {
      setInputValue(journeyInputs[0] as HTMLInputElement, '0915')
      await flushPromises()
    })

    const updatedInputs = container.querySelectorAll('.journey-list input[type="text"]')
    expect((updatedInputs[0] as HTMLInputElement).value).toBe('09:15')
    expect((updatedInputs[1] as HTMLInputElement).value).toBe('12:30')
    expect((updatedInputs[2] as HTMLInputElement).value).toBe('14:00')
  })

  it('shows a warning banner when the payload includes loadWarning', async () => {
    mockedBackend.loadPlanner.mockResolvedValue(
      samplePayload({
        loadWarning: { key: 'errors.planner.load_fallback' },
      }),
    )

    await renderPage()

    expect(container.textContent).toContain('Não foi possível carregar o dia')
    expect(container.querySelector('.banner.error')).toBeNull()
    expect(container.textContent).toContain('Entrada 1')
  })

  it('loads planner data when Wails is unavailable (HTTP transport)', async () => {
    mockedBackend.hasWailsRuntime.mockReturnValue(false)

    await renderPage()

    expect(mockedBackend.loadPlanner).toHaveBeenCalledTimes(1)
    expect(mockedBackend.loadPlanner).toHaveBeenCalledWith('2026-06-09')
    expect(container.textContent).not.toContain('Browser mode')
  })

  it('renders journey list and summary inside a two-column grid', async () => {
    await renderPage()

    const grid = container.querySelector('.grid-2')
    expect(grid).not.toBeNull()
    expect(grid?.children.length).toBe(2)
    expect(grid?.textContent).toContain('Entrada 1')
    expect(grid?.textContent).toContain('Meta do dia')
  })
})
