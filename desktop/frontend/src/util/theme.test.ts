import { describe, expect, it } from 'vitest'
import {
  PRESETS,
  THEME_KEYS,
  applyTheme,
  buildThemeVars,
  contrastRatio,
  isHex,
  presetIdForColors,
  resolveEffectiveColors,
  resolvePresets,
  slugifyThemeId,
  visiblePresets,
} from './theme'

const midnight = {
  label: 'Meia-noite',
  colors: { bg: '#000000', panel: '#0a0a0a', text: '#f5f5f5', muted: '#a0a0a0', border: '#262626', accent: '#7c5cff', ok: '#4ade80', warning: '#facc15', danger: '#f87171' },
}
const light = PRESETS.find((p) => p.id === 'light')!.colors!

describe('theme', () => {
  it('accepts only #rrggbb hex colors', () => {
    expect(isHex('#00ff00')).toBe(true)
    expect(isHex('#AbCdEf')).toBe(true)
    expect(isHex('#FFF')).toBe(false)
    expect(isHex('#GGGGGG')).toBe(false)
    expect(isHex('red')).toBe(false)
  })

  it('computes WCAG contrast', () => {
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 6)
    expect(contrastRatio('#ffffff', '#000000')).toBeCloseTo(21, 6)
  })

  it('returns no vars without valid colors', () => {
    expect(buildThemeVars({})).toEqual({})
    expect(buildThemeVars({ bg: 'nope', accent: '#12345' })).toEqual({})
  })

  it('picks a readable accent text color', () => {
    expect(buildThemeVars({ accent: '#000000' })['--accent-text']).toBe('#ffffff')
    expect(buildThemeVars({ accent: '#ffffff' })['--accent-text']).toBe('#000000')
  })

  it('derives css vars from the picks', () => {
    const vars = buildThemeVars(PRESETS[1].colors!)
    expect(vars['--bg']).toBe('#f6f7f4')
    expect(vars['--panel-soft']).toBe('color-mix(in srgb, #ffffff 96%, #17201d)')
    expect(vars['--focus']).toContain('transparent')
  })

  it('keeps presets valid and readable', () => {
    const ids = PRESETS.map((p) => p.id)
    expect(ids).toEqual(['auto', 'light', 'dark', 'warm'])
    expect(PRESETS[0].colors).toBeNull()
    for (const preset of PRESETS) {
      expect(preset.labelKey).toBe(`settings.theme.presets.${preset.id}`)
      if (!preset.colors) continue
      for (const key of THEME_KEYS) expect(isHex(preset.colors[key])).toBe(true)
      const { bg, panel, text, muted } = preset.colors
      for (const surface of [bg, panel]) {
        expect(contrastRatio(text, surface)).toBeGreaterThanOrEqual(4.5)
        expect(contrastRatio(muted, surface)).toBeGreaterThanOrEqual(4.5)
      }
    }
  })
})

describe('theme presets', () => {
  it('lists built-ins first, then user presets with labels', () => {
    const list = resolvePresets({ midnight })
    expect(list.map((p) => p.id)).toEqual(['auto', 'light', 'dark', 'warm', 'midnight'])
    expect(list[1]).toMatchObject({ custom: false, labelKey: 'settings.theme.presets.light' })
    expect(list[4]).toMatchObject({ custom: true, label: 'Meia-noite', colors: midnight.colors })
    expect(resolvePresets(undefined)).toHaveLength(PRESETS.length)
  })

  it('replaces a built-in with the same id in place', () => {
    const list = resolvePresets({ dark: { label: 'Escuro', colors: { bg: '#123456' } } })
    expect(list.map((p) => p.id)).toEqual(['auto', 'light', 'dark', 'warm'])
    expect(list[2]).toMatchObject({ id: 'dark', label: 'Escuro', custom: true })
    expect(list[2].labelKey).toBeUndefined()
  })

  it('fills missing keys from light and drops invalid hex', () => {
    const [preset] = resolvePresets({ mine: { colors: { bg: '#123456', text: 'nope', accent: '#12345' } } }).slice(-1)
    expect(preset.label).toBe('mine')
    expect(preset.colors!.bg).toBe('#123456')
    expect(preset.colors!.text).toBe(light.text)
    expect(preset.colors!.accent).toBe(light.accent)
    expect(preset.colors).toEqual({ ...light, bg: '#123456' })
  })

  it('skips ids that are not slugs', () => {
    const ids = resolvePresets({ 'Bad Id': midnight, ok_1: midnight }).map((p) => p.id)
    expect(ids).not.toContain('Bad Id')
    expect(ids).toContain('ok_1')
  })

  it('resolves effective colors in the spec order', () => {
    const themes = { midnight }
    const custom = { accent: '#ff8800' }
    expect(resolveEffectiveColors({ theme: custom, theme_name: 'midnight', default_theme: 'warm', themes })).toEqual(custom)
    expect(resolveEffectiveColors({ theme: { accent: 'bad' }, theme_name: 'midnight', themes })).toEqual(midnight.colors)
    expect(resolveEffectiveColors({ theme_name: 'light', default_theme: 'warm' })).toEqual(light)
    expect(resolveEffectiveColors({ default_theme: 'midnight', themes })).toEqual(midnight.colors)
    expect(resolveEffectiveColors({ theme_name: 'auto', default_theme: 'warm' })).toBeNull()
    expect(resolveEffectiveColors({ theme_name: 'gone', default_theme: 'warm' })).toBeNull()
    expect(resolveEffectiveColors({})).toBeNull()
  })

  it('slugifies names into theme ids', () => {
    expect(slugifyThemeId('Meia-Noite Ç')).toBe('meia-noite-c')
    expect(slugifyThemeId('  --Ocean  Blue!! ')).toBe('ocean-blue')
    expect(slugifyThemeId('!!!')).toBe('')
  })

  it('finds the preset matching a draft', () => {
    const presets = resolvePresets({ midnight })
    expect(presetIdForColors(presets, {})).toBe('auto')
    expect(presetIdForColors(presets, { ...midnight.colors, bg: '#000000' })).toBe('midnight')
    expect(presetIdForColors(presets, { ...light })).toBe('light')
    expect(presetIdForColors(presets, { ...light, accent: '#ABCDEF' })).toBe('')
    expect(presetIdForColors(presets, { accent: '#ff8800' })).toBe('')
  })
})

describe('applyTheme color scheme', () => {
  it('tags the page as dark or light from the background, and clears it for Automático', () => {
    const root = document.documentElement

    applyTheme(midnight.colors)
    expect(root.dataset.scheme).toBe('dark')

    applyTheme(light)
    expect(root.dataset.scheme).toBe('light')

    applyTheme(null)
    expect(root.dataset.scheme).toBeUndefined()
  })
})

describe('visiblePresets', () => {
  it('drops hidden themes but always keeps Claro and Escuro', () => {
    const ids = (hidden: string[]) =>
      visiblePresets(resolvePresets({ midnight }), hidden).map((preset) => preset.id)

    expect(ids([])).toEqual(['auto', 'light', 'dark', 'warm', 'midnight'])
    expect(ids(['auto', 'warm', 'midnight'])).toEqual(['light', 'dark'])
    expect(ids(['light', 'dark'])).toEqual(['auto', 'light', 'dark', 'warm', 'midnight'])
  })

  it('does not protect a user theme that reuses a built-in id', () => {
    const presets = resolvePresets({ dark: { label: 'Meia-noite', colors: { bg: '#000000' } } })

    expect(visiblePresets(presets, ['dark']).map((preset) => preset.id)).not.toContain('dark')
  })
})
