## Context

Dynamic journeys and Go-backed `RecalculateDay` already exist. The desktop still debounces recalculation (400 ms), updates visible times only from backend responses for solved exits, and stacks balance badges, calculator buttons, and alternative times on each Saída row. Users perceive lag and clutter. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**

- Synchronous slot suggestion pipeline on the frontend mirroring Go `ResolveAnchors`, `SolveExit`, and `DefaultSolvedExit`
- Timeline-style journey UI with calculada highlight and summary-owned balance
- Remove debounce for display; async backend only for summary metrics

**Non-Goals:**

- Changing Go JSON payload shapes or reminder slot keys
- Solving entry slots from daily target (entries stay anchor-suggested only)
- Redesigning Settings or CLI TUI in this change

## Decisions

### 1. Local suggestion mirror in TypeScript

**Choice:** Add `plannerSuggestions.ts` implementing the same anchor resolution, default solved exit, and exit solve logic as `pkg/plan/day.go`, using `baseTargetSecs` and config anchors from the loaded payload.

**Alternatives:** New Wails `SuggestSlots` RPC on every keystroke — rejected due to latency and IPC overhead. Keep debounced backend only — rejected; fails instant UX goal.

**Rationale:** Math is deterministic and small; duplicating the slice of Go logic keeps UI instant while `RecalculateDay` remains authoritative for spans and balance.

### 2. State split: `journeys` (user input) vs `displayJourneys` (after local solve)

**Choice:** On each edit, apply local solve to produce `displayJourneys` for rendering; keep `journeys` as raw user-edited values for registered slots and manual edits. Merge solved exit time from local solve into display layer only.

**Alternative:** Single state updated only from backend — rejected (lag).

### 3. Timeline component structure

**Choice:** Replace card-per-journey grid with `PlannerTimeline` — one row per slot (Entrada N, Saída N) grouped visually by journey with a light divider. Calculator button becomes text link "Calcular" / badge "Calculada".

**Balance/alternative exit:** Move to `PlannerSummaryPanel` only.

### 4. Backend recalculate without debounce for summary

**Choice:** Fire `recalculateDay` immediately on edit (no 400 ms debounce) but do not block slot display on it; use `requestId` to ignore stale responses.

## Risks / Trade-offs

- **[Risk] TS/Go drift** → Mitigation: shared test fixtures (same inputs → same HH:MM) in Go tests and TS tests; optional thin export of anchor list from backend on load.
- **[Risk] Double network calls on rapid typing** → Mitigation: abort prior summary request via ref counter; summary can trail slightly behind slots.
- **[Risk] Large CSS churn** → Mitigation: scope new classes under `.planner-timeline`; remove obsolete `.clockout-*` where unused.

## Migration Plan

1. Ship frontend + CSS changes; no config migration.
2. Rebuild desktop via `scripts/setup.sh` from repo root.
3. Rollback: reinstall previous build from git tag if needed.

## Open Questions

None — grilling session will confirm timeline vs card layout preference before coding.
