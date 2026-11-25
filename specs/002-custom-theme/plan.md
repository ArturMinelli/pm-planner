# Plan — 002 custom theme

**Storage**: `theme: {colorKey: "#rrggbb"}` in config (Go `config.File` + `fileRaw`, yaml+json, omitempty). Frontend `PmConfig.theme?: Record<string,string>`. Empty/absent = Automático.

**Apply**: `util/theme.ts` builds CSS custom properties (existing names in `index.css :root`) and sets them inline on `document.documentElement`; derived tokens via CSS `color-mix(in srgb, …)` strings; sets `color-scheme` from Fundo luminance. Absent theme → remove inline vars (CSS defaults + prefers-color-scheme apply). Cached in `localStorage` (try/catch) and applied before React renders to avoid flash.

**Wiring**: `ConfigProvider`-adjacent effect applies `config.theme` whenever config changes.

**UI**: `features/settings/ThemeSettingsForm.tsx` (Card, preset chips, color rows, live preview swatch strip, contrast hint, Save/Reset), mounted in `SettingsPage`; draft previews live, reverts to saved on unmount. Styles appended to `index.css` using existing tokens/`ui` components.

**Validation**: vitest for util + Go test for round trip; browser check (playwright) of Settings at desktop/narrow widths, light/dark, and impeccable polish pass.
