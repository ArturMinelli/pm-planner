## Why

The desktop app no longer overflows at medium widths after `planner-ui-polish`, but columns still feel uneven: the planner summary panel leaves dead space while the journey list could use more room, settings cards stack in a single narrow column on wide screens, and page headers do not share a consistent column rhythm. Users want a cohesive column system that fills the window effectively without changing suggestion logic or backend behavior.

## What Changes

- **Planner 60/40 column ratio**: Journey list becomes the primary column (~60%) and summary the secondary (~40%) at desktop widths, eliminating large empty gaps.
- **Settings two-column layout**: Account, planner defaults, and reminder settings cards arrange in a responsive two-column grid on wide screens; stack to single column below the medium breakpoint.
- **Shared column tokens**: Introduce consistent grid proportions and max-width rules reused across planner and settings pages.
- **Page header alignment**: Page headers and action areas align to the same column grid so date picker and titles do not float in awkward dead zones.
- **No behavior changes**: Suggestion calculation, data flow, and form save logic remain unchanged — layout and structure only.

## Capabilities

### New Capabilities

_None — column proportion requirements extend the existing layout capability._

### Modified Capabilities

- `desktop-ui-layout`: Add requirements for journey-primary column ratios, settings two-column card grid, and app-wide column fill (no large dead space).

## Impact

- **Desktop frontend CSS**: `index.css` (grid-2 proportions, new settings grid, page max-width tokens).
- **Pages**: `PlannerPage.tsx`, `SettingsPage.tsx` (markup wrappers for column grids).
- **Shared UI**: `ui.tsx` (`Page`, `PageHeader` grid alignment if needed).
- **Tests**: Layout structure tests for column classes at wide vs medium breakpoints.
- **No backend or API changes.**
