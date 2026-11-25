import type { PmConfig, ThemePresetConfig } from '../types'

export const THEME_KEYS =['bg', 'panel', 'text', 'muted', 'border', 'accent', 'ok', 'warning', 'danger'] as const

export type ThemeKey = (typeof THEME_KEYS)[number]

export interface ThemePreset {
  id: string
  labelKey: string
  /** Null means follow the OS color scheme through the CSS defaults. */
  colors: Record<ThemeKey, string> | null
}

const CACHE_KEY = 'pm.theme'

export function isHex(v: string): boolean {
  return /^#[0-9a-f]{6}$/i.test(v)
}

function luminance(hex: string): number {
  const [r, g, b] = [1, 3, 5].map((i) => {
    const c = parseInt(hex.slice(i, i + 2), 16) / 255
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

export function contrastRatio(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x)
  return (hi + 0.05) / (lo + 0.05)
}

export const PRESETS: ThemePreset[] = [
  { id: 'auto', labelKey: 'settings.theme.presets.auto', colors: null },
  {
    id: 'light',
    labelKey: 'settings.theme.presets.light',
    colors: {
      bg: '#f6f7f4', panel: '#ffffff', text: '#17201d', muted: '#64706b', border: '#d9ded8',
      accent: '#087f8c', ok: '#146c36', warning: '#9a5b00', danger: '#ba1a1a',
    },
  },
  {
    id: 'dark',
    labelKey: 'settings.theme.presets.dark',
    colors: {
      bg: '#111412', panel: '#191f1c', text: '#edf3ef', muted: '#a7b2ac', border: '#333d37',
      accent: '#62d7df', ok: '#8bd99f', warning: '#f0c06a', danger: '#ffb4ab',
    },
  },
  {
    id: 'warm',
    labelKey: 'settings.theme.presets.warm',
    colors: {
      bg: '#fbf6ec', panel: '#fffdf8', text: '#3b2a1e', muted: '#6b5646', border: '#e3d6c3',
      accent: '#b5532c', ok: '#2f6b3a', warning: '#8a5a00', danger: '#a4262c',
    },
  },
]

/** Returns the valid picks only, dropping missing or malformed entries. */
function validPicks(colors: Partial<Record<ThemeKey, string>>): Partial<Record<ThemeKey, string>> {
  const picks: Partial<Record<ThemeKey, string>> = {}
  for (const key of THEME_KEYS) {
    const value = colors[key]
    if (value && isHex(value)) picks[key] = value
  }
  return picks
}

export function buildThemeVars(colors: Partial<Record<ThemeKey, string>>): Record<string, string> {
  const p = validPicks(colors)
  const vars: Record<string, string> = {}
  for (const key of THEME_KEYS) {
    if (p[key]) vars[`--${key}`] = p[key]
  }

  const mix = (a: string, pct: number, b: string) => `color-mix(in srgb, ${a} ${pct}%, ${b})`
  const { panel, text, border, accent, ok, warning, danger } = p

  if (panel && text) {
    vars['--panel-soft'] = mix(panel, 96, text)
    vars['--control'] = panel
  }
  if (border && text) vars['--border-strong'] = mix(border, 85, text)
  if (accent && text) vars['--accent-strong'] = mix(accent, 78, text)
  if (accent && panel) vars['--accent-soft'] = mix(accent, 14, panel)
  if (accent) {
    vars['--accent-text'] = contrastRatio('#ffffff', accent) >= contrastRatio('#000000', accent) ? '#ffffff' : '#000000'
    vars['--focus'] = `0 0 0 3px color-mix(in srgb, ${accent} 28%, transparent)`
  }
  if (danger && panel) vars['--danger-bg'] = mix(danger, 12, panel)
  if (ok && panel) vars['--ok-bg'] = mix(ok, 12, panel)
  if (warning && panel) vars['--warning-bg'] = mix(warning, 12, panel)
  if (text) vars['--shadow'] = `0 1px 2px color-mix(in srgb, ${text} 10%, transparent)`

  return vars
}

let appliedVars: string[] = []

export function applyTheme(colors?: Record<string, string> | null): void {
  const root = document.documentElement
  for (const name of appliedVars) root.style.removeProperty(name)

  const picks = validPicks(colors ?? {})
  const vars = buildThemeVars(picks)
  appliedVars = Object.keys(vars)
  for (const [name, value] of Object.entries(vars)) root.style.setProperty(name, value)

  if (picks.bg) {
    const scheme = luminance(picks.bg) < 0.4 ? 'dark' : 'light'
    root.style.colorScheme = scheme
    // CSS reads this to pick shadows that suit the surface (see --card-shadow).
    root.dataset.scheme = scheme
  } else {
    root.style.removeProperty('color-scheme')
    delete root.dataset.scheme
  }
}

/** Cache the saved theme so the next launch paints it before config loads. */
export function cacheTheme(colors?: Record<string, string> | null): void {
  const picks = validPicks(colors ?? {})
  try {
    if (Object.keys(picks).length) localStorage.setItem(CACHE_KEY, JSON.stringify(picks))
    else localStorage.removeItem(CACHE_KEY)
  } catch {
    // Storage can be blocked; the theme still applies for this session.
  }
}

export function loadCachedTheme(): Record<string, string> | null {
  try {
    const raw = localStorage.getItem(CACHE_KEY)
    if (!raw) return null
    const picks = validPicks(JSON.parse(raw) as Partial<Record<ThemeKey, string>>)
    return Object.keys(picks).length ? picks : null
  } catch {
    return null
  }
}

export type ResolvedPreset = {
  id: string
  label?: string
  labelKey?: string
  colors: Record<ThemeKey, string> | null
  custom: boolean
}

const PRESET_ID = /^[a-z0-9_-]+$/

/** Built-ins first, then user presets; a user preset with a built-in id replaces it in place. */
export function resolvePresets(themes?: Record<string, ThemePresetConfig>): ResolvedPreset[] {
  const light = PRESETS.find((p) => p.id === 'light')!.colors!
  const list: ResolvedPreset[] = PRESETS.map((p) => ({ ...p, custom: false }))
  for (const [id, preset] of Object.entries(themes ?? {})) {
    if (!PRESET_ID.test(id)) continue
    const colors = { ...light, ...validPicks(preset.colors ?? {}) } as Record<ThemeKey, string>
    const resolved: ResolvedPreset = { id, label: preset.label || id, colors, custom: true }
    const index = list.findIndex((p) => p.id === id)
    if (index >= 0) list[index] = resolved
    else list.push(resolved)
  }
  return list
}

/** Colors the app should paint with, or null for auto/OS. */
export function resolveEffectiveColors(
  config: Pick<PmConfig, 'theme' | 'theme_name' | 'default_theme' | 'themes'>,
): Record<string, string> | null {
  const picks = validPicks(config.theme ?? {})
  if (Object.keys(picks).length) return picks
  const id = config.theme_name || config.default_theme
  if (!id || id === 'auto') return null
  const preset = resolvePresets(config.themes).find((p) => p.id === id)
  return preset?.colors ?? null
}

export function slugifyThemeId(name: string): string {
  return name
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

/** Built-in themes people can always come back to: they cannot be removed from the picker. */
export const PROTECTED_PRESET_IDS = ['light', 'dark']

export function isProtectedPreset(preset: Pick<ResolvedPreset, 'id' | 'custom'>): boolean {
  return !preset.custom && PROTECTED_PRESET_IDS.includes(preset.id)
}

/** Presets to show in the picker: everything except removed themes, but never Claro/Escuro. */
export function visiblePresets(presets: ResolvedPreset[], hidden: string[] = []): ResolvedPreset[] {
  return presets.filter((preset) => isProtectedPreset(preset) || !hidden.includes(preset.id))
}

/** Id of the preset whose colors match the draft, 'auto' for an empty draft, else ''. */
export function presetIdForColors(presets: ResolvedPreset[], colors: Record<string, string>): string {
  if (Object.keys(colors).length === 0) return 'auto'
  const draft = (key: ThemeKey) => (colors[key] ?? '').toLowerCase()
  const match = presets.find((p) => p.colors && THEME_KEYS.every((k) => p.colors![k].toLowerCase() === draft(k)))
  return match?.id ?? ''
}
