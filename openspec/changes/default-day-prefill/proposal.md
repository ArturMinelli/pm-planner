## Why

The planner desktop page requires the user to pick a date and click **Carregar** before any fields become editable. That extra step blocks quick planning and duplicates what should happen automatically. When the API is unavailable, the app is unusable instead of falling back to the planner defaults already configured on the Settings page.

## What Changes

- **Remove the manual load step**: drop `PlannerLoadCard` and the **Carregar** button; the planner auto-fetches the selected day on page open and whenever the date changes.
- **Move the date picker** into the page header (next to the title).
- **Show skeleton placeholders** while fetching; journey and summary fields stay disabled until the fetch completes or fails.
- **On successful fetch**: populate journeys, target, balance, and summary from the API (same full pipeline as today's `FetchPlannerPayload` / `pm plan`).
- **On fetch failure**: fall back to planner defaults from Settings (anchor times, max daily extra, balance credit multiplier) with a fixed 8h30 base target; show a non-blocking warning; enable editing with two default journeys, last exit as solved exit, and originals line `(nenhum)`.
- **Date change** always triggers a fresh full fetch and **full replace** of planner state (user edits on the previous date are not preserved).
- **CLI parity**: `pm plan` keeps auto-fetching; on API failure it falls back to the same settings defaults instead of exiting with an error.

## Capabilities

### New Capabilities

- `planner-auto-load`: Automatic day loading on mount and date change, settings-based fallback when the API fails, and removal of the manual load UI.

### Modified Capabilities

<!-- No existing specs in openspec/specs/ yet -->

## Impact

- **Desktop frontend**: `PlannerPage.tsx` (auto-fetch lifecycle, header date picker, skeleton loading, remove load card); `PlannerLoadCard` deleted or unused; `PlannerJourneyList` / `PlannerSummaryPanel` skeleton states.
- **Go app layer**: new or extended binding to build a defaults-only `PlannerPayload` from `config.ResolvePlannerAnchors` and built-in target when fetch fails; shared between desktop and CLI.
- **CLI**: `pkg/cmd/plan.go` — catch fetch errors and render defaults-based plan instead of failing.
- **Tests**: frontend tests for auto-load, skeleton, and fallback UX; Go tests for defaults payload builder and CLI fallback path.

## Non-goals

- A `--defaults` flag to intentionally skip API fetch (fallback-on-failure only).
- Persisting user edits across date changes.
- New Settings fields for default base target (8h30 remains the fallback constant).
- Changing reminder scheduling behavior.
