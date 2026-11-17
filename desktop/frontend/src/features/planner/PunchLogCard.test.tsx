/**
 * @vitest-environment jsdom
 */
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { PunchLogState, PunchRecord } from '../../types'
import PunchLogCard, { UnconfirmedWarning } from './PunchLogCard'

const rec = (over: Partial<PunchRecord>): PunchRecord => ({
  localId: 'a', date: '2026-11-17', time: '10:01:00', createdAt: '2026-11-17T10:01:00Z',
  source: 'app', status: 'synced', latitude: 0, longitude: 0, checks: 0, ...over,
})
const log = (over: Partial<PunchLogState> = {}): PunchLogState => ({
  date: '2026-11-17', punches: [], pendingCount: 0, unconfirmedCount: 0, ...over,
})

describe('punch log UI', () => {
  let container: HTMLDivElement
  let root: Root
  const onSync = vi.fn()
  const render = (el: React.ReactElement) => act(() => root.render(el))

  beforeEach(() => {
    onSync.mockClear()
    container = document.createElement('div')
    document.body.appendChild(container)
    root = createRoot(container)
  })
  afterEach(() => {
    act(() => root.unmount())
    container.remove()
  })

  it('shows the warning only with unconfirmed records and Verificar agora syncs', () => {
    const state = log({ unconfirmedCount: 1, punches: [rec({ status: 'unconfirmed' })] })
    render(<UnconfirmedWarning state={state} syncing={false} onSync={onSync} />)
    const alert = container.querySelector('[role="alert"]')
    expect(alert?.textContent).toContain('O ponto das 10:01 não foi confirmado')
    const btn = [...container.querySelectorAll('button')].find((b) => b.textContent === 'Verificar agora')!
    act(() => btn.click())
    expect(onSync).toHaveBeenCalledTimes(1)

    render(<UnconfirmedWarning state={state} syncing onSync={onSync} />)
    expect(container.querySelector('button')?.disabled).toBe(true)

    render(<UnconfirmedWarning state={log()} syncing={false} onSync={onSync} />)
    expect(container.querySelector('[role="alert"]')).toBeNull()
  })

  it('falls back to a count when the visible day holds no unconfirmed record', () => {
    render(<UnconfirmedWarning state={log({ unconfirmedCount: 2 })} syncing={false} onSync={onSync} />)
    expect(container.textContent).toContain('Há 2 pontos não confirmados')
  })

  it('renders every status, origin, disabled and the footer', () => {
    const state = log({
      lastSyncAt: '2026-11-17T12:30:00',
      punches: [
        rec({ localId: '1', status: 'synced', source: 'remote' }),
        rec({ localId: '2', status: 'pending' }),
        rec({ localId: '3', status: 'unconfirmed' }),
        rec({ localId: '4', status: 'failed' }),
        rec({ localId: '5', disabled: true }),
      ],
    })
    render(<PunchLogCard state={state} syncing={false} isToday onSync={onSync} />)
    const text = container.textContent ?? ''
    for (const s of ['Sincronizado', 'Aguardando confirmação', 'Não confirmado', 'Falhou', 'PontoMais', 'Pelo app', 'Desconsiderado', 'Última sincronização às 12:30']) {
      expect(text).toContain(s)
    }
    expect(container.querySelectorAll('li').length).toBe(5)
    expect(container.querySelector('[data-disabled="true"]')).not.toBeNull()
    act(() => container.querySelector<HTMLButtonElement>('button[aria-label="Sincronizar"]')!.click())
    expect(onSync).toHaveBeenCalled()
  })

  it('shows the empty state and the sync error notice', () => {
    render(<PunchLogCard state={log({ lastSyncError: 'x' })} syncing={false} isToday onSync={onSync} />)
    expect(container.textContent).toContain('Nenhum ponto registrado')
    expect(container.textContent).toContain('Falha ao sincronizar. Tentando novamente.')
  })
})
