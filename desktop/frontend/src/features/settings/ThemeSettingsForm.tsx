import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button, Card } from '../../components/ui'
import {
  contrastRatio,
  isHex,
  isProtectedPreset,
  PRESETS,
  slugifyThemeId,
  THEME_KEYS,
  type ResolvedPreset,
  type ThemeKey,
} from '../../util/theme'

type ThemeSettingsFormProps = {
  /** Current draft; an empty object means Automático (CSS defaults). */
  colors: Record<string, string>
  presets: ResolvedPreset[]
  /** Preset id matching the draft, 'auto' for an empty draft, '' for custom colors. */
  selectedId: string
  busy: boolean
  onChange: (colors: Record<string, string>) => void
  onSelectPreset: (id: string) => void
  onSaveAs: (name: string, colors: Record<string, string>) => void
  onDeletePreset: (id: string) => void
  onSave: () => void
  onReset: () => void
}

/** Computed CSS default for a token; the swatch needs a valid hex. */
function cssDefault(key: string): string {
  const value = getComputedStyle(document.documentElement).getPropertyValue(`--${key}`).trim()
  return isHex(value) ? value : '#000000'
}

export default function ThemeSettingsForm({
  colors,
  presets,
  selectedId,
  busy,
  onChange,
  onSelectPreset,
  onSaveAs,
  onDeletePreset,
  onSave,
  onReset,
}: ThemeSettingsFormProps) {
  const { t } = useTranslation()
  // Hex text as typed, so invalid input stays visible without reaching onChange.
  const [typed, setTyped] = useState<Partial<Record<ThemeKey, string>>>({})
  const [saveName, setSaveName] = useState('')
  // The color editor stays folded away until the user asks for it.
  const [mode, setMode] = useState<'closed' | 'edit' | 'new'>('closed')

  const effective = (key: ThemeKey) => colors[key] ?? cssDefault(key)
  const complete = () =>
    Object.fromEntries(THEME_KEYS.map((key) => [key, effective(key)])) as Record<string, string>

  const presetLabel = (preset: ResolvedPreset) => preset.label ?? t(preset.labelKey!)
  const selectedPreset = presets.find((preset) => preset.id === selectedId)
  // Saving under a custom theme's own name replaces it, so editing prefills that name.
  const nameFor = (preset?: ResolvedPreset) => (preset?.custom ? presetLabel(preset) : '')

  const choosePreset = (id: string) => {
    setTyped({})
    onSelectPreset(id)
    if (mode === 'edit') setSaveName(nameFor(presets.find((preset) => preset.id === id)))
  }
  const openEdit = () => {
    setSaveName(nameFor(selectedPreset))
    setMode('edit')
  }
  const openNew = () => {
    setSaveName('')
    setMode('new')
  }
  const closeEditor = () => setMode('closed')
  const setColor = (key: ThemeKey, value: string) => {
    setTyped((current) => ({ ...current, [key]: undefined }))
    onChange({ ...complete(), [key]: value })
  }
  const setHex = (key: ThemeKey, value: string) => {
    // only invalid text needs local state; a valid hex already lives in `colors`
    setTyped((current) => ({ ...current, [key]: isHex(value) ? undefined : value }))
    if (isHex(value)) onChange({ ...complete(), [key]: value })
  }

  const hasInvalidHex = THEME_KEYS.some((key) => {
    const draftHex = typed[key]
    return draftHex !== undefined && !isHex(draftHex)
  })
  const slug = slugifyThemeId(saveName)
  const overwrites = presets.some((preset) => preset.id === slug)
  const canSaveAs = slug !== '' && !hasInvalidHex

  const submitSaveAs = () => {
    if (!canSaveAs) return
    onSaveAs(saveName.trim(), complete())
    setSaveName('')
    setMode('closed')
  }

  const ratio = contrastRatio(effective('text'), effective('bg'))
  // Claro/Escuro may be removed from the picker; the Automático swatch still needs their colors.
  const builtinColors = (id: string) =>
    presets.find((preset) => preset.id === id)?.colors ?? PRESETS.find((preset) => preset.id === id)?.colors
  const light = builtinColors('light')
  const dark = builtinColors('dark')

  return (
    <Card collapsible title={t('settings.theme.title')} intro={t('settings.theme.intro')}>
      <form
        className="form-stack"
        onSubmit={(event) => {
          event.preventDefault()
          onSave()
        }}
      >
        <div className="theme-presets" role="group" aria-label={t('settings.theme.presetsLabel')}>
          {presets.map((preset) => {
            const selected = selectedId === preset.id
            const label = presetLabel(preset)
            const dots = preset.colors ?? null
            return (
              <span key={preset.id} className="theme-preset-wrap" data-custom={preset.custom || undefined}>
                <button
                  type="button"
                  className="theme-preset"
                  aria-pressed={selected}
                  title={preset.custom ? t('settings.theme.customBadge') : undefined}
                  onClick={() => choosePreset(preset.id)}
                >
                  <span className="theme-dots" aria-hidden="true">
                    {dots ? (
                      (['bg', 'panel', 'accent'] as const).map((key) => (
                        <span key={key} className="theme-dot" style={{ background: dots[key] }} />
                      ))
                    ) : (
                      <span
                        className="theme-dot wide"
                        style={{ background: `linear-gradient(90deg, ${light?.bg} 50%, ${dark?.bg} 50%)` }}
                      />
                    )}
                  </span>
                  {label}
                </button>
                {isProtectedPreset(preset) ? null : (
                  <button
                    type="button"
                    className="theme-preset-delete"
                    aria-label={t('settings.theme.deletePreset', { name: label })}
                    disabled={busy}
                    onClick={() => { setTyped({}); onDeletePreset(preset.id) }}
                  >
                    ×
                  </button>
                )}
              </span>
            )
          })}
          <button
            type="button"
            className="theme-preset theme-preset-new"
            aria-pressed={mode === 'new'}
            onClick={openNew}
          >
            <span aria-hidden="true">+</span> {t('settings.theme.newTheme')}
          </button>
        </div>

        {mode === 'closed' ? null : (
          <div className="theme-editor">
            <div className="theme-editor-head">
              <strong>
                {mode === 'new'
                  ? t('settings.theme.editorNew')
                  : t('settings.theme.editorEditing', {
                      name: selectedPreset ? presetLabel(selectedPreset) : t('settings.theme.customColors'),
                    })}
              </strong>
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

            <div className="theme-saveas">
              <label className="theme-saveas-field">
                <span className="theme-saveas-label">
                  {mode === 'new' ? t('settings.theme.saveAsLabel') : t('settings.theme.nameLabel')}
                </span>
                <input
                  type="text"
                  maxLength={32}
                  autoComplete="off"
                  placeholder={t('settings.theme.saveAsPlaceholder')}
                  value={saveName}
                  onChange={(event) => setSaveName(event.target.value)}
                  onKeyDown={(event) => {
                    // Enter saves the theme here instead of submitting the main form.
                    if (event.key !== 'Enter') return
                    event.preventDefault()
                    submitSaveAs()
                  }}
                />
              </label>
              <Button type="button" variant="secondary" onClick={submitSaveAs} disabled={!canSaveAs}>
                {t('settings.theme.saveAs')}
              </Button>
              {overwrites ? (
                <small className="theme-saveas-hint">
                  {t('settings.theme.saveAsOverwrite', { name: saveName.trim() })}
                </small>
              ) : null}
            </div>
          </div>
        )}

        {ratio < 4.5 ? (
          <p className="theme-warning" role="status">
            {t('settings.theme.lowContrast', { ratio: ratio.toFixed(2) })}
          </p>
        ) : null}

        <div className="btn-row theme-actions">
          <Button type="submit" variant="primary" disabled={busy} aria-busy={busy}>
            {t('settings.theme.save')}
          </Button>
          <div className="theme-actions-side">
            <Button
              type="button"
              variant="secondary"
              aria-expanded={mode !== 'closed'}
              onClick={mode === 'closed' ? openEdit : closeEditor}
            >
              <svg width="15" height="15" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
                <path
                  d="M11.2 2.3a1.6 1.6 0 0 1 2.25 2.25L5.4 12.6 2 13.5l.9-3.4z"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="1.4"
                  strokeLinejoin="round"
                />
              </svg>
              {mode === 'closed' ? t('settings.theme.edit') : t('settings.theme.closeEditor')}
            </Button>
            <Button variant="secondary" onClick={() => { setTyped({}); onReset() }} disabled={busy}>
              {t('settings.theme.reset')}
            </Button>
          </div>
        </div>
      </form>
    </Card>
  )
}
