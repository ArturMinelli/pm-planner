## Context

See [proposal.md](./proposal.md) for motivation. Today the planner is hardcoded to four slots (`In1/Out1/In2/Out2`) across `pkg/plan`, `pkg/app`, `pkg/ui`, `pkg/reminder`, and the desktop frontend. Stamp assignment uses brute-force `C(n,4)` in `assign.go`. The desktop duplicates `ComputeOut2` / segment math in TypeScript and discards the backend's `out2` on load, causing the registered-punch bug.

Constraints:
- Settings YAML keeps four anchor keys (`in1`–`out2`); no new settings fields.
- Wails v2 desktop; shared Go module `pm-cli` consumed by CLI and desktop.
- Portuguese UI labels (`Entrada N`, `Saída N`, `registrada`).

## Goals / Non-Goals

**Goals:**
- Replace fixed four-slot model with `[]Journey` + single `solvedExit` index.
- Generalize stamp assignment to O(n × slots) DP for arbitrary slot counts.
- Single Go source of truth for solve + summarize; desktop calls `RecalculateDay` binding.
- Journey-group UI with add/remove, solved-exit toggle, registered badges.
- Adapt CLI TUI and reminder scheduler to dynamic slot keys.

**Non-Goals:**
- Persisting per-day journey count across reloads (re-derived from stamp count each load).
- Configurable anchors per journey beyond the existing four.
- Solving entry slots.
- Submitting punches to PontoMais.

## Decisions

### 1. Core types in `pkg/plan`

Replace `Suggestion{In1,Out1,In2,Out2}` with:

```go
type ClockSlot struct {
    Time       string `json:"time"`       // HH:MM; empty when unknown
    Registered bool   `json:"registered"` // matched a real punch
}

type Journey struct {
    Entry ClockSlot `json:"entry"`
    Exit  ClockSlot `json:"exit"`
}

type Day struct {
    Journeys   []Journey `json:"journeys"`
    SolvedExit int       `json:"solvedExit"` // journey index; -1 when none
}
```

`SuggestDay(date, stamps, periods, target, anchors []string) (Day, error)` builds journeys from stamps + defaults. `SolveExit(day, target, solvedIndex) Day` sets the solved exit time. `Summarize(day, target, date) Summary` returns per-journey spans, total, overtime, warnings.

**Alternative considered:** flat `[]string` times array — rejected because journey grouping and registered flags are first-class concerns.

### 2. Stamp assignment via order-preserving DP

Replace `assignStampsToPlannerSlots` brute force with DP:

- State: `(stampIndex, slotIndex)` → min cost to assign stamps `[0..stampIndex]` to slots `[0..slotIndex]`.
- Transition: skip stamp, or assign stamp `i` to slot `j` (only if `stamp[i] >= stamp[i-1]` assigned time).
- Cost: `|stamp - anchor[slot]|`; unassigned slots cost 0 at fill time.
- Slot count = `2 * len(journeys)`.

Preserve existing behavior: when slot 0 (Entrada 1) has no assigned stamp, default to `anchors[0]` not first chronological punch.

**Alternative considered:** greedy nearest-anchor — rejected; fails on the lone-midday-punch case.

### 3. Anchor resolution

`config.ResolvePlannerAnchors` returns `[]string` (min length 4 from YAML). For slot index `s >= 4`:

- Even index (entry): `anchors[s-2] + break` (previous exit anchor + break).
- Odd index (exit): `anchors[s-1] + remainingTargetShare`.

Break from `minBreakFromPeriods` (API shift periods, 1h fallback) — unchanged.

### 4. Solved-exit selection logic

On load:
1. Default `solvedExit` = last journey index whose exit is not registered.
2. If all exits registered → `solvedExit = -1`.

On user toggle: set `solvedExit` to chosen journey index; recalculate via `SolveExit`.

`SolveExit` algorithm:
```
fixedSpan = sum of spans for all journeys except solvedExit
remaining = max(target - fixedSpan, 0)
solvedExit.time = solvedEntry.time + remaining
```
Clamp: if solved time < entry, set to entry and emit ordering warning if next entry exists before solved time.

Balance adjustment: compute alternative solved time using `adjustedTargetSecs` when `balance.AppliesToday` — same pattern as today.

### 5. Payload and API changes

**`PlannerPayload`** (breaking):
```go
type PlannerPayload struct {
    Date           string                  `json:"date"`
    BaseTargetSecs int64                   `json:"baseTargetSecs"`
    Balance        *plan.BalanceAdjustment `json:"balance,omitempty"`
    BalanceError   string                  `json:"balanceError,omitempty"`
    OriginalTimes  []string                `json:"originalTimes"`
    Journeys       []plan.Journey          `json:"journeys"`
    SolvedExit     int                     `json:"solvedExit"`
    OriginalsLine  string                  `json:"originalsLine"`
}
```

**`PlannerSummary`** (breaking):
```go
type PlannerSummary struct {
    Journeys          []plan.Journey `json:"journeys"`          // with solved exit filled
    SolvedExit        int            `json:"solvedExit"`
    JourneySpanSecs   []int64        `json:"journeySpanSecs"`
    TotalSpanSecs     int64          `json:"totalSpanSecs"`
    OvertimeSecs      int64          `json:"overtimeSecs"`
    AlternativeExit   string         `json:"alternativeExit,omitempty"` // solved slot alt time
    Warnings          []string       `json:"warnings,omitempty"`
}
```

New Wails method on `desktop/app.go`:
```go
func (a *App) RecalculateDay(req RecalculateRequest) (*app.PlannerSummary, error)
```
`RecalculateRequest` carries `date`, `baseTargetSecs`, `balance`, `journeys`, `solvedExit`.

Delete `desktop/frontend/src/features/planner/plannerSummary.ts` math; keep only types re-export if needed.

### 6. Desktop UI structure

Replace flat `TIME_FIELDS` array with:

```
PlannerJourneyList
├── PlannerJourneyGroup (×N)
│   ├── label "Nª Jornada"
│   ├── PlannerTimeInput (Entrada) + registrada badge
│   ├── PlannerExitField (Saída) + calculator toggle + registrada badge
│   └── remove button (disabled if registered)
├── Button "+ Adicionar jornada"
└── originals line (unchanged)
```

`PlannerExitField` merges current `PlannerTimeInput` + `PlannerClockoutField` behavior: when `isSolved`, read-only + balance badge + alternative time.

State in `PlannerPage`:
- `journeys: Journey[]`, `solvedExit: number`
- On edit → debounced `backend.recalculateDay(...)`
- On add journey → append journey seeded per spec, set `solvedExit` to new index, recalculate
- On remove → splice journey, adjust `solvedExit` index, recalculate

CSS: new `.journey-group` replaces `.time-grid` for journey rows; reuse existing tokens from `index.css`.

### 7. CLI TUI adaptation

`pkg/ui/checkins.go` `PlanModel`:
- Hold `[]Journey` + `solvedExit` instead of three `textinput.Model`.
- Tab cycles editable fields (all entries + non-solved exits).
- Display N journey rows dynamically in viewport.
- `+`/`-` keys or dedicated keys to add/remove journey (match desktop semantics).

`pkg/cmd/plan.go` non-live form: dynamic huh fields or simplified journey count prompt.

### 8. Reminder slot keys

Replace fixed constants with generators:
```go
func SlotKey(journeyIndex int, isExit bool) string // "in1", "out2", ...
func SlotLabel(journeyIndex int, isExit bool) string // "Entrada 1", ...
```

`BuildDayPlan` produces `len(journeys)*2` slots. Check persisted reminder IDs in `schedule.go` — old `in1|15` format vs new; on first run after upgrade, reminders re-schedule from fresh day plan (acceptable; reminders are same-day ephemeral).

### 9. Backward compatibility

- Config YAML: unchanged keys; `ResolvePlannerAnchors` reads four values into slice.
- No migration for desktop local state (none persisted).
- Breaking JSON only affects in-memory Wails IPC (no external API consumers).

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| DP assignment slower for many stamps | Slot count capped practically at ~10 (5 journeys); n rarely > 20 |
| Debounced recalculate feels laggy | Optimistic UI: show spinner on summary only; keep inputs responsive |
| Reminder ID format change drops pending reminders | Reminders are intraday; reschedule on next fetch |
| CLI TUI layout breaks with many journeys | Scroll viewport; default stays 2 journeys |
| Breaking Wails types break existing desktop build | Update TS types and all consumers in same PR |

## Migration Plan

1. Implement Go core + tests first (no UI breakage yet).
2. Update `pkg/app` payload; add `RecalculateDay`.
3. Update desktop frontend to new types + binding; remove TS math.
4. Update CLI TUI and reminders.
5. Run full test suite (`go test ./...`, `npm test` in desktop frontend).
6. Manual smoke: load screenshot scenario (4 punches → Saída 2 shows 14:41, not 15:41).

Rollback: revert commit; no data migration to undo.

## Open Questions

None — all design decisions were resolved during the grilling session.
