import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button, Card } from '../../components/ui'
import {
  contrastRatio,
  isHex,
  PRESETS,
  THEME_KEYS,
  type ThemeKey,
} from '../../util/theme'

type ThemeSettingsFormProps = {
  /** Current draft; an empty object means Automático (CSS defaults). */
  colors: Record<string, string>
  busy: boolean
  onChange: (colors: Record<string, string>) => void
  onSave: () => void
  onReset: () => void
}

const LIGHT = PRESETS.find((preset) => preset.id === 'light')?.colors
const DARK = PRESETS.find((preset) => preset.id === 'dark')?.colors

/** Computed CSS default for a token; the swatch needs a valid hex. */
function cssDefault(key: string): string {
  const value = getComputedStyle(document.documentElement).getPropertyValue(`--${key}`).trim()
  return isHex(value) ? value : '#000000'
}

function isSelected(colors: Record<string, string>, presetColors: Record<ThemeKey, string> | null) {
  if (!presetColors) return Object.keys(colors).length === 0
  return THEME_KEYS.every((key) => colors[key]?.toLowerCase() === presetColors[key])
}

export default function ThemeSettingsForm({
  colors,
  busy,
  onChange,
  onSave,
  onReset,
}: ThemeSettingsFormProps) {
  const { t } = useTranslation()
  // Hex text as typed, so invalid input stays visible without reaching onChange.
  const [typed, setTyped] = useState<Partial<Record<ThemeKey, string>>>({})

  const effective = (key: ThemeKey) => colors[key] ?? cssDefault(key)
  const complete = () =>
    Object.fromEntries(THEME_KEYS.map((key) => [key, effective(key)])) as Record<string, string>

  const applyPreset = (presetColors: Record<ThemeKey, string> | null) => {
    setTyped({})
    onChange(presetColors ?? {})
  }
  const setColor = (key: ThemeKey, value: string) => {
    setTyped((current) => ({ ...current, [key]: undefined }))
    onChange({ ...complete(), [key]: value })
  }
  const setHex = (key: ThemeKey, value: string) => {
    setTyped((current) => ({ ...current, [key]: value }))
    if (isHex(value)) onChange({ ...complete(), [key]: value })
  }

  const ratio = contrastRatio(effective('text'), effective('bg'))

  return (
    <Card title={t('settings.theme.title')} intro={t('settings.theme.intro')}>
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          onSave()
        }}
      >
        <div className="theme-presets" role="group" aria-label={t('settings.theme.presetsLabel')}>
          {PRESETS.map((preset) => {
            const selected = isSelected(colors, preset.colors)
            const dots = preset.colors ?? null
            return (
              <button
                key={preset.id}
                type="button"
                className="theme-preset"
                aria-pressed={selected}
                onClick={() => applyPreset(preset.colors)}
              >
                <span className="theme-dots" aria-hidden="true">
                  {dots ? (
                    (['bg', 'panel', 'accent'] as const).map((key) => (
                      <span key={key} className="theme-dot" style={{ background: dots[key] }} />
                    ))
                  ) : (
                    <span
                      className="theme-dot wide"
                      style={{ background: `linear-gradient(90deg, ${LIGHT?.bg} 50%, ${DARK?.bg} 50%)` }}
                    />
                  )}
                </span>
                {t(preset.labelKey)}
              </button>
            )
          })}
        </div>

        <div className="theme-color-grid" role="group" aria-label={t('settings.theme.colorsLabel')}>
          {THEME_KEYS.map((key) => {
            const labelId = `theme-name-${key}`
            const hexId = `theme-hex-${key}`
            const errorId = `theme-error-${key}`
            const draftHex = typed[key]
            const hexValue = draftHex ?? effective(key)
            const invalid = draftHex !== undefined && !isHex(draftHex)
            return (
              <div className="theme-color-row" key={key}>
                <span id={labelId} className="theme-name">
                  {t(`settings.theme.keys.${key}`)}
                </span>
                <input
                  className="theme-swatch"
                  type="color"
                  aria-labelledby={labelId}
                  value={effective(key)}
                  onChange={(event) => setColor(key, event.target.value)}
                />
                <input
                  id={hexId}
                  className="theme-hex"
                  type="text"
                  aria-labelledby={labelId}
                  aria-invalid={invalid}
                  aria-describedby={invalid ? errorId : undefined}
                  maxLength={7}
                  spellCheck={false}
                  autoComplete="off"
                  value={hexValue}
                  onChange={(event) => setHex(key, event.target.value)}
                />
                {invalid ? (
                  <small id={errorId} className="theme-error">
                    {t('settings.theme.invalidHex')}
                  </small>
                ) : null}
              </div>
            )
          })}
        </div>

        {ratio < 4.5 ? (
          <p className="theme-warning" role="status">
            {t('settings.theme.lowContrast', { ratio: ratio.toFixed(2) })}
          </p>
        ) : null}

        <div className="btn-row">
          <Button variant="secondary" onClick={onReset} disabled={busy}>
            {t('settings.theme.reset')}
          </Button>
          <Button type="submit" variant="primary" disabled={busy} aria-busy={busy}>
            {t('settings.theme.save')}
          </Button>
        </div>
      </form>
    </Card>
  )
}
