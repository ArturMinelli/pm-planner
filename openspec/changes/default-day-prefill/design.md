## Context

Today `PlannerPage` keeps `loaded` null until the user clicks **Carregar**, which calls `backend.loadPlanner(date)` → `FetchPlannerPayload`. Journey fields are disabled (`timesDisabled = busy || !loaded`). Settings already stores planner anchor times and balance tuning via `PlannerDefaultsForm`, and `config.ResolvePlannerAnchors` merges them with built-in defaults — the same anchors used when assigning stamps after a successful fetch.

See `proposal.md` for motivation. Requirements are in `specs/planner-auto-load/spec.md`.

## Goals / Non-Goals

**Goals:**

- Auto-fetch on mount and date change with no manual load step.
- Skeleton loading UX while fetch is in flight; fields disabled until resolved.
- Shared Go helper to build a defaults-only `PlannerPayload` for desktop fallback and CLI fallback.
- Header-integrated date picker; remove `PlannerLoadCard`.

**Non-Goals:**

- Optimistic prefill before fetch (defaults appear only on failure).
- Preserving edits across date changes.
- `--defaults` CLI flag.
- New Settings field for default base target.

## Decisions

### 1. Single fetch lifecycle in `PlannerPage`

**Choice:** `useEffect` keyed on `date` calls `loadPlanner(date)` on mount and every date change. A `fetching` flag drives skeleton state; `loaded` is set on success or fallback.

**Alternatives considered:**
- *Optimistic defaults then replace* — rejected in grilling; user wants wait-for-API with defaults only on failure.
- *Separate shift-only pre-fetch* — rejected; full `FetchPlannerPayload` on every date change.

**Race handling:** Increment a request counter (or abort via `AbortController` if Wails supports it). Ignore stale responses when `date` changes mid-flight.

### 2. `BuildDefaultsPlannerPayload` in Go (`pkg/app/planner.go`)

**Choice:** New exported function that reads config, resolves anchors, builds two unregistered journeys, sets `solvedExit` to `1` (last exit), `baseTargetSecs` to `8h30`, empty `originalTimes`, `originalsLine` to `"(nenhum)"`, no balance.

```go
func BuildDefaultsPlannerPayload(dateStr string) (*PlannerPayload, error)
```

Reuse `config.ResolvePlannerAnchors` and existing `defaultPlannerTarget` constant. Run `plan.SuggestDay` with empty stamps and no periods so solved-exit math is consistent, or construct journeys directly from anchors — prefer `SuggestDay(date, nil, nil, target, anchors)` for one code path.

**Alternatives considered:**
- *Frontend-only fallback* — rejected; CLI parity requires Go-side builder.

### 3. Desktop fetch wrapper

**Choice:** Extend `loadPlanner` binding (or add `loadPlannerWithFallback`) in `desktop/app.go`:

```go
payload, err := app.FetchPlannerPayload(ctx, client, date)
if err != nil {
    payload, err = app.BuildDefaultsPlannerPayload(date)
    // attach warning string for frontend banner
}
```

Return a small wrapper or add `loadWarning string` field on `PlannerPayload` so the UI can show the non-blocking banner without treating fallback as an error.

**Alternative:** Frontend catches error and calls a second `getDefaultsPayload` binding — more round-trips, rejected.

### 4. UI layout changes

**Choice:**
- Add date `<input type="date">` to `PageHeader` via a new optional `actions` slot or inline prop on `PlannerPage`.
- Delete `PlannerLoadCard.tsx` and its test file.
- Add `loading` prop to `PlannerJourneyList` and `PlannerSummaryPanel`; render CSS skeleton blocks (reuse or extend existing utility classes in `index.css`).
- Remove `timesDisabled` dependency on manual load; use `fetching` instead.

### 5. CLI fallback (`pkg/cmd/plan.go`)

**Choice:** Wrap `FetchPlannerPayload` error path with `BuildDefaultsPlannerPayload` and print a stderr note (e.g. `aviso: não foi possível carregar o dia; usando padrões do planner`). Render the same TUI as a successful load.

### 6. Balance on fallback

**Choice:** Omit balance on defaults payload (`balance` nil, `balanceError` empty). Alternative exit hidden. Matches grilling: balance comes from full fetch only; fallback is offline-first.

## Risks / Trade-offs

- **[Race on fast date switching]** → Request-id guard; only apply response matching current `date`.
- **[User edits lost on date change]** → Accepted per grilling (full replace); no confirmation dialog.
- **[Skeleton flash on slow API]** → Acceptable; fields stay disabled until done.
- **[Fallback plan may diverge from real shift target]** → Warning banner sets expectation; user can fix times manually.

## Migration Plan

1. Ship Go `BuildDefaultsPlannerPayload` + desktop/CLI integration behind existing bindings.
2. Update frontend layout (header date, skeletons, remove load card).
3. Update/delete tests referencing `PlannerLoadCard` and manual load flow.
4. No config migration; Settings schema unchanged.

Rollback: revert frontend to manual load card if auto-fetch causes issues; Go helper is additive.

## Open Questions

None — grilling resolved trigger, fallback, CLI parity, and UI placement.
