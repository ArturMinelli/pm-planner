# 002 — Configurable color theme

## Problem
The app follows only the OS light/dark scheme with a fixed teal palette. Users want to choose every color themselves.

## User stories
1. **Pick colors** — In Settings, a "Aparência" card lists every theme color (Fundo, Superfície, Texto, Texto secundário, Borda, Destaque, Sucesso, Alerta, Erro). Each has a native color picker plus a hex text field. Changes preview live across the whole app.
2. **Presets** — One click applies a starting palette: Automático (OS default, no overrides), Claro, Escuro, plus 2 extra tasteful palettes. Presets only fill the colors; the user can keep editing.
3. **Persist** — Saved to the user's config (`theme` key) and applied on every launch, with no flash of the default theme.
4. **Reset** — "Restaurar padrão" clears all overrides (back to Automático). Unsaved preview is reverted on leaving the page.

## Rules
- Derived tokens (soft panel, strong border, accent hover/soft/text, status backgrounds, focus ring) are computed from the 9 picked colors, so a pick never yields unreadable UI.
- Text color on accent buttons is auto-chosen black/white by contrast.
- Invalid hex is rejected inline; never saved.
- Contrast hint: warn (non-blocking) when Texto vs Fundo is under 4.5:1.
- All copy in pt-BR (i18n). Keyboard accessible, labels on every control, visible focus.

## Out of scope
Per-mode (light AND dark) separate custom palettes, import/export, fonts, density.

## Acceptance
- Changing Destaque recolors buttons, links, focus, active nav instantly.
- Reload keeps the theme; "Automático" restores OS light/dark behavior.
- Works at narrow widths (card collapses to one column) and in both OS schemes.
