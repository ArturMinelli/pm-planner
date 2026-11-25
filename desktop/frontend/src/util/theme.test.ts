import { describe, expect, it } from 'vitest'
import { PRESETS, THEME_KEYS, buildThemeVars, contrastRatio, isHex } from './theme'

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
    expect(ids).toEqual(['auto', 'light', 'dark', 'ocean', 'warm'])
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
