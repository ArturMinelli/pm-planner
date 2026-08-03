## Why

The settings page currently arranges account, planner defaults, and reminder cards in a two-column grid on wide screens (`page-column-layout`). That layout splits related configuration across columns and makes long forms (especially reminders) harder to scan. Users want every settings section stacked vertically at all viewport widths, using the full main content width like the planner page.

## What Changes

- **Settings single-column stack**: Account, planner defaults, and reminders cards always display in one column, in order, at every viewport width.
- **Full-width settings cards**: Settings cards stretch to the main content width (same `Page` container as planner), not a narrow centered column.
- **Remove two-column settings grid CSS**: Drop `grid-template-columns: repeat(2, …)` and the reminders full-width span rule at ≥801px.
- **Planner unchanged**: Journey/summary 60/40 column ratio stays as implemented.
- **No behavior changes**: Form fields, save logic, and API calls remain unchanged — layout only.

## Capabilities

### New Capabilities

_None — layout requirement updates extend the existing desktop UI layout capability._

### Modified Capabilities

- `desktop-ui-layout`: Replace the two-column settings card grid requirement with always-single-column stacking and full-width cards.

## Impact

- **Desktop frontend CSS**: `index.css` (`.settings-grid` rules at ≥801px).
- **Pages**: `SettingsPage.tsx` (wrapper may stay; CSS drives layout).
- **Tests**: `SettingsPage.test.tsx` — update assertions if they expect two-column structure.
- **No backend or API changes.**
