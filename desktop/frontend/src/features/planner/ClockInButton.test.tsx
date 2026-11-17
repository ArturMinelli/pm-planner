/**
 * @vitest-environment jsdom
 */
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import * as backend from '../../services/backend'
import { detectLocation } from '../../util/location'
import type { ClockInResult } from '../../types'
import ClockInButton from './ClockInButton'

vi.mock('../../services/backend', () => ({
  clockIn: vi.fn(),
}))

vi.mock('../../util/location', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../util/location')>()
  return { ...actual, detectLocation: vi.fn() }
})

const mockedBackend = vi.mocked(backend)
const mockedDetect = vi.mocked(detectLocation)

const location = { latitude: 1, longitude: 2, accuracy: 5, source: 'default' as const }

function flush() {
  return new Promise<void>((resolve) => setTimeout(resolve, 0))
}

describe('ClockInButton', () => {
  let container: HTMLDivElement
  let root: Root
  const onRegistered = vi.fn()

  function button(): HTMLButtonElement {
    const found = [...container.querySelectorAll('button')].find(
      (b) => b.textContent === 'Bater ponto',
    )
    if (!found) throw new Error('clock-in button missing')
    return found
  }

  function buttonLabelled(label: string): HTMLButtonElement {
    const found = [...container.querySelectorAll('button')].find(
      (b) => b.textContent === label,
    )
    if (!found) throw new Error(`button "${label}" missing`)
    return found
  }

  function status(): HTMLElement {
    const el = container.querySelector<HTMLElement>('[role="status"]')
    if (!el) throw new Error('status region missing')
    return el
  }

  async function click(el: HTMLButtonElement) {
    await act(async () => {
      el.click()
      await flush()
    })
  }

  beforeEach(() => {
    vi.clearAllMocks()
    mockedDetect.mockResolvedValue(location)
    container = document.createElement('div')
    document.body.appendChild(container)
    root = createRoot(container)
    act(() => root.render(<ClockInButton onRegistered={onRegistered} />))
  })

  afterEach(() => {
    act(() => root.unmount())
    container.remove()
  })

  it('shows the time and default location label after a registered punch', async () => {
    const ok: ClockInResult = {
      status: 'registered',
      registered: true,
      punch: { id: 1, date: '2026-11-17', time: '08:00:00', address: '', latitude: 1, longitude: 2, disabled: false },
      locationSource: 'default',
    }
    mockedBackend.clockIn.mockResolvedValue(ok)

    await click(button())

    expect(status().textContent).toBe('Ponto registrado às 08:00:00 · Localização padrão (escritório)')
    expect(status().className).toBe('banner success')
    expect(onRegistered).toHaveBeenCalledTimes(1)
  })

  it('disables the button with aria-busy while locating', async () => {
    let finishLocating: (v: typeof location) => void = () => undefined
    mockedDetect.mockReturnValue(new Promise((resolve) => { finishLocating = resolve }))
    mockedBackend.clockIn.mockResolvedValue({ status: 'registered', registered: true })

    await act(async () => {
      button().click()
      await flush()
    })

    expect(button().disabled).toBe(true)
    expect(button().getAttribute('aria-busy')).toBe('true')
    expect(status().textContent).toBe('Obtendo localização…')
    expect(container.querySelector('.clockin-spinner')?.getAttribute('aria-hidden')).toBe('true')

    await act(async () => {
      finishLocating(location)
      await flush()
    })
    expect(button().disabled).toBe(false)
  })

  it('asks to confirm a too_close punch and resends with confirmNear', async () => {
    mockedBackend.clockIn
      .mockResolvedValueOnce({ status: 'failed', registered: false, code: 'too_close' })
      .mockResolvedValueOnce({ status: 'registered', registered: true })

    await click(button())
    expect(status().textContent).toBe('Você já registrou um ponto há menos de 1 minuto. Registrar mesmo assim?')
    expect(onRegistered).not.toHaveBeenCalled()
    expect(container.textContent).not.toContain('Tentar novamente')

    await click(buttonLabelled('Registrar mesmo assim'))

    expect(mockedBackend.clockIn).toHaveBeenLastCalledWith({ location, confirmNear: true })
    expect(onRegistered).toHaveBeenCalledTimes(1)
  })

  it('offers retry on a failed attempt and shows the offline message', async () => {
    mockedBackend.clockIn.mockResolvedValue({ status: 'failed', registered: false, code: 'offline' })

    await click(button())

    expect(status().textContent).toBe('Sem conexão. O ponto NÃO foi registrado.')
    expect(status().className).toBe('banner error')
    expect(container.textContent).toContain('Tentar novamente')
  })

  it('shows the session expired message', async () => {
    mockedBackend.clockIn.mockResolvedValue({ status: 'failed', registered: false, code: 'session_expired' })

    await click(button())

    expect(status().textContent).toBe('Sessão expirada. Entre novamente.')
  })

  it('shows the uncertain message without a retry button', async () => {
    mockedBackend.clockIn.mockResolvedValue({ status: 'uncertain', registered: null, code: 'uncertain' })

    await click(button())

    expect(status().textContent).toBe('Não foi possível confirmar se o ponto foi registrado. Verificando…')
    expect(status().className).toBe('banner warning')
    expect(container.querySelector('.clockin-spinner')).toBeNull()
    expect(container.textContent).not.toContain('Tentar novamente')
    expect(onRegistered).not.toHaveBeenCalled()
  })
})
