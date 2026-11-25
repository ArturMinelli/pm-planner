/**
 * @vitest-environment jsdom
 */
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ThemeSettingsForm from './ThemeSettingsForm'
import { PRESETS, THEME_KEYS } from '../../util/theme'

function setInputValue(input: HTMLInputElement, value: string) {
  const setter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')?.set
  setter?.call(input, value)
  input.dispatchEvent(new Event('input', { bubbles: true }))
}

describe('ThemeSettingsForm', () => {
  let container: HTMLDivElement
  let root: Root

  beforeEach(() => {
    container = document.createElement('div')
    document.body.appendChild(container)
    root = createRoot(container)
  })

  afterEach(() => {
    act(() => root.unmount())
    container.remove()
  })

  function render(props: Partial<Parameters<typeof ThemeSettingsForm>[0]> = {}) {
    const handlers = { onChange: vi.fn(), onSave: vi.fn(), onReset: vi.fn() }
    act(() => {
      root.render(
        <ThemeSettingsForm colors={{}} busy={false} {...handlers} {...props} />,
      )
    })
    return handlers
  }

  it('renders a labelled row for each theme color', () => {
    render()

    for (const key of THEME_KEYS) {
      const picker = container.querySelector<HTMLInputElement>(`input[type="color"][aria-labelledby="theme-name-${key}"]`)
      const hex = container.querySelector<HTMLInputElement>(`#theme-hex-${key}`)
      expect(picker, `color picker for ${key}`).not.toBeNull()
      expect(hex, `hex input for ${key}`).not.toBeNull()
    }
    expect(container.querySelectorAll('input[type="color"]')).toHaveLength(THEME_KEYS.length)
  })

  it('calls onChange with the preset colors when a preset is clicked', () => {
    const { onChange } = render()
    const light = PRESETS.find((preset) => preset.id === 'light')
    const button = [...container.querySelectorAll('button')].find(
      (node) => node.textContent === 'Claro',
    )
    expect(button).toBeTruthy()

    act(() => button?.click())

    expect(onChange).toHaveBeenCalledWith(light?.colors)
  })

  it('does not propagate invalid hex and shows the error', () => {
    const { onChange } = render()
    const hex = container.querySelector<HTMLInputElement>('#theme-hex-bg')
    expect(hex).not.toBeNull()

    act(() => setInputValue(hex!, '#12'))

    expect(onChange).not.toHaveBeenCalled()
    expect(hex!.getAttribute('aria-invalid')).toBe('true')
    expect(hex!.getAttribute('aria-describedby')).toBe('theme-error-bg')
    expect(container.textContent).toContain('Use o formato #rrggbb.')
  })
})
