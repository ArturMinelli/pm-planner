## Why

The desktop planner still feels sluggish and confusing: suggestion times update only after a debounced backend round-trip, entry slots do not always show immediate anchor-based suggestions when journeys are added or edited, and the current journey-card layout (calculator toggle, balance badges, alternative exit, registered badges) overwhelms users who only want to see “what time should I punch next.” The dynamic-journeys model is in place, but the experience does not feel instant, adaptive, or intuitive.

## What Changes

- **Instant local suggestions**: Compute all entry suggestions from anchors (and derived anchors for journey 3+) synchronously on every edit, add/remove journey, or solved-exit toggle — without waiting on a debounced `RecalculateDay` call. Backend recalculation remains for summary totals, balance, and alternative exit only.
- **Adaptive algorithm for N journeys**: Generalize anchor resolution and default slot seeding so any journey count (2, 3, 4, …) gets coherent entry/exit suggestions; adding a journey immediately seeds entry from previous exit + break and exit from solve or anchor share.
- **Clear solved-slot UX**: One exit is always calculated when any exit is unsolved; the active solved exit is visually distinct (filled suggestion chip, not a disabled text field). Calculator control becomes a simple “calcular esta saída” affordance with plain-language state.
- **Clean suggestion presentation**: Replace cluttered per-field chrome with a unified timeline-style list — each slot shows label, suggested time prominently, optional `registrada` pill, and inline ordering warnings only when relevant. Balance/alternative exit moves to summary panel.
- **Planner page UX redesign**: Simpler two-column layout — left: “Horários do dia” timeline; right: “Resumo” with target, spans, overtime, and bank adjustment. Date picker stays in header; remove redundant copy and skeleton noise.
- **Remove 400 ms debounce** for time-display updates; keep async backend call for summary metrics only.

## Capabilities

### New Capabilities

- `instant-planner-suggestions`: Immediate, journey-count-adaptive suggestion times for every clock slot, solved-exit selection, and clean desktop presentation.

### Modified Capabilities

- `journey-planning`: Entry slots SHALL receive immediate anchor-derived suggestions; solved exit display and summary/balance separation updated to match the redesigned UX.

## Impact

- **Desktop frontend**: `PlannerPage.tsx` (split local suggest vs backend summary), new `plannerSuggestions.ts` (anchor/solve mirror or thin client), `PlannerJourneyList`, `PlannerJourneyGroup`, `PlannerExitField`, `PlannerSummaryPanel`, `index.css`.
- **Go core** (minor): Export or reuse `ResolveAnchors`, `SolveExit`, `DefaultSolvedExit` via existing `RecalculateDay` binding; optional lightweight `SuggestSlots` endpoint if client mirror is rejected.
- **Tests**: TS tests for instant suggestion on add/edit/toggle; component tests for timeline layout and solved-exit highlight.
- **No breaking API changes** to `PlannerPayload` JSON shapes.
