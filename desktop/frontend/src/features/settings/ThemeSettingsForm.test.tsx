/**
 * @vitest-environment jsdom
 */
import { act } from 'react'
import { createRoot, type Root } from 'react-dom/client'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import ThemeSettingsForm from './ThemeSettingsForm'
import { resolvePresets, THEME_KEYS } from '../../util/theme'

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
    const handlers = {
      onChange: vi.fn(),
      onSelectPreset: vi.fn(),
      onSaveAs: vi.fn(),
      onDeletePreset: vi.fn(),
      onSave: vi.fn(),
      onReset: vi.fn(),
    }
    act(() => {
      root.render(
        <ThemeSettingsForm
          colors={{}}
          presets={resolvePresets()}
          selectedId="auto"
          busy={false}
          {...handlers}
          {...props}
        />,
      )
    })
    return handlers
  }

  function clickButton(text: string) {
    const button = [...container.querySelectorAll('button')].find((node) => node.textContent?.includes(text))
    expect(button, `button "${text}"`).toBeTruthy()
    act(() => button!.click())
  }

  it('keeps the color editor folded until Editar cores is clicked', () => {
    render()

    expect(container.querySelectorAll('input[type="color"]')).toHaveLength(0)
    expect(container.querySelector('input[placeholder="Ex.: Meia-noite"]')).toBeNull()

    clickButton('Editar cores')
    expect(container.querySelectorAll('input[type="color"]')).toHaveLength(THEME_KEYS.length)

    clickButton('Fechar')
    expect(container.querySelectorAll('input[type="color"]')).toHaveLength(0)
  })

  it('opens an empty editor from Novo tema and closes it after saving', () => {
    const { onSaveAs } = render()

    clickButton('Novo tema')
    expect(container.textContent).toContain('Novo tema: ajuste as cores e dê um nome')
    const input = container.querySelector<HTMLInputElement>('input[placeholder="Ex.: Meia-noite"]')
    expect(input?.value).toBe('')

    act(() => setInputValue(input!, 'Meu tema'))
    clickButton('Salvar como tema')

    expect(onSaveAs).toHaveBeenCalledWith('Meu tema', expect.any(Object))
    expect(container.querySelectorAll('input[type="color"]')).toHaveLength(0)
  })

  it('prefills a custom theme name when editing it', () => {
    const presets = resolvePresets({ midnight: { label: 'Meia-noite', colors: { bg: '#000000' } } })
    render({ presets, selectedId: 'midnight' })

    clickButton('Editar cores')

    expect(container.textContent).toContain('Editando Meia-noite')
    expect(container.querySelector<HTMLInputElement>('input[placeholder="Ex.: Meia-noite"]')?.value).toBe(
      'Meia-noite',
    )
  })

  it('renders a labelled row for each theme color', () => {
    render()
    clickButton('Editar cores')

    for (const key of THEME_KEYS) {
      const picker = container.querySelector<HTMLInputElement>(`input[type="color"][aria-labelledby="theme-name-${key}"]`)
      const hex = container.querySelector<HTMLInputElement>(`#theme-hex-${key}`)
      expect(picker, `color picker for ${key}`).not.toBeNull()
      expect(hex, `hex input for ${key}`).not.toBeNull()
    }
    expect(container.querySelectorAll('input[type="color"]')).toHaveLength(THEME_KEYS.length)
  })

  it('calls onSelectPreset with the id when a preset is clicked', () => {
    const { onSelectPreset } = render()
    const button = [...container.querySelectorAll('button')].find(
      (node) => node.textContent === 'Claro',
    )
    expect(button).toBeTruthy()

    act(() => button?.click())

    expect(onSelectPreset).toHaveBeenCalledWith('light')
  })

  it('offers a delete button on built-in presets, except Claro and Escuro', () => {
    const { onDeletePreset } = render()

    const removes = [...container.querySelectorAll<HTMLButtonElement>('.theme-preset-delete')]
    expect(removes.map((node) => node.getAttribute('aria-label'))).toEqual([
      'Excluir tema Automático',
      'Excluir tema Areia',
    ])
    act(() => removes[1].click())

    expect(onDeletePreset).toHaveBeenCalledWith('warm')
  })

  it('keeps Claro and Escuro in the picker without a delete button', () => {
    render()

    const chips = [...container.querySelectorAll('.theme-preset-wrap')]
    for (const name of ['Claro', 'Escuro']) {
      const chip = chips.find((node) => node.textContent?.includes(name))
      expect(chip, name).toBeTruthy()
      expect(chip!.querySelector('.theme-preset-delete')).toBeNull()
    }
  })

  it('marks the selected preset and shows user presets with a delete button', () => {
    const presets = resolvePresets({
      midnight: { label: 'Meia-noite', colors: { bg: '#000000' } },
    })
    const { onDeletePreset } = render({ presets, selectedId: 'midnight' })

    const chip = [...container.querySelectorAll('button.theme-preset')].find(
      (node) => node.textContent === 'Meia-noite',
    )
    expect(chip?.getAttribute('aria-pressed')).toBe('true')
    expect(container.querySelector('.theme-preset-wrap[data-custom]')).not.toBeNull()

    const remove = container.querySelector<HTMLButtonElement>(
      '.theme-preset-wrap[data-custom] .theme-preset-delete',
    )
    expect(remove?.getAttribute('aria-label')).toBe('Excluir tema Meia-noite')
    act(() => remove?.click())

    expect(onDeletePreset).toHaveBeenCalledWith('midnight')
  })

  it('saves the full set of colors under a name and clears the input', () => {
    const { onSaveAs } = render()
    clickButton('Editar cores')
    const input = container.querySelector<HTMLInputElement>('input[placeholder="Ex.: Meia-noite"]')
    expect(input).not.toBeNull()

    act(() => setInputValue(input!, 'Meia Noite'))
    const save = [...container.querySelectorAll('button')].find(
      (node) => node.textContent === 'Salvar como tema',
    )
    act(() => save?.click())

    expect(onSaveAs).toHaveBeenCalledTimes(1)
    const [name, colors] = onSaveAs.mock.calls[0]
    expect(name).toBe('Meia Noite')
    expect(Object.keys(colors).sort()).toEqual([...THEME_KEYS].sort())
    // the editor folds away after saving, taking the name input with it
    expect(container.querySelector('input[placeholder="Ex.: Meia-noite"]')).toBeNull()
  })

  it('saves on Enter in the name input without submitting the main form', () => {
    const { onSaveAs, onSave } = render()
    clickButton('Editar cores')
    const input = container.querySelector<HTMLInputElement>('input[placeholder="Ex.: Meia-noite"]')

    act(() => setInputValue(input!, 'Praia'))
    act(() => {
      input!.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }))
    })

    expect(onSaveAs).toHaveBeenCalledWith('Praia', expect.any(Object))
    expect(onSave).not.toHaveBeenCalled()
  })

  it('disables saving when the name has no usable characters', () => {
    const { onSaveAs } = render()
    clickButton('Editar cores')
    const input = container.querySelector<HTMLInputElement>('input[placeholder="Ex.: Meia-noite"]')

    act(() => setInputValue(input!, '!!!'))
    const save = [...container.querySelectorAll('button')].find(
      (node) => node.textContent === 'Salvar como tema',
    )

    expect(save?.disabled).toBe(true)
    expect(onSaveAs).not.toHaveBeenCalled()
  })

  it('does not propagate invalid hex and shows the error', () => {
    const { onChange } = render()
    clickButton('Editar cores')
    const hex = container.querySelector<HTMLInputElement>('#theme-hex-bg')
    expect(hex).not.toBeNull()

    act(() => setInputValue(hex!, '#12'))

    expect(onChange).not.toHaveBeenCalled()
    expect(hex!.getAttribute('aria-invalid')).toBe('true')
    expect(hex!.getAttribute('aria-describedby')).toBe('theme-error-bg')
    expect(container.textContent).toContain('Use o formato #rrggbb.')
  })
})
