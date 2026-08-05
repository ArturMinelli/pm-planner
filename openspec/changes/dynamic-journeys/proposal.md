## Why

The planner hardcodes exactly two journeys (four clock-in slots: Entrada 1, Saída 1, Entrada 2, Saída 2). Days with more than four punches, or users who split work into three or more segments, cannot be planned accurately. A related bug makes this worse: when all four slots already have real punches, the desktop still recalculates the last exit from the daily target instead of showing the registered time — e.g. suggesting 15:41 when 14:41 was already clocked out.

## What Changes

- Replace the fixed four-slot model with **dynamic journeys**: each journey is an entry + exit pair; default is 2 journeys, user can add or remove per day.
- Introduce a **solved exit** concept: exactly one exit slot is calculated from the daily target at a time; the user picks which exit via an inline calculator toggle on each Saída field.
- **Registered punches** from PontoMais are pre-filled, badged `registrada`, and editable — but never chosen as the solved slot; if the default solved exit is already punched, auto-shift to the last exit without a punch.
- When every slot has a registered punch, nothing is solved; summary shows the real total and overtime.
- **BREAKING**: `PlannerPayload` JSON drops `in1`/`out1`/`in2`/`out2` in favor of `journeys[]` + `solvedExit`; `PlannerSummary` drops `firstSpanSecs`/`secondSpanSecs` in favor of `journeySpans[]`.
- **BREAKING**: Reminder slot keys change from fixed `in1`/`out1`/`in2`/`out2` to generated `in{n}`/`out{n}`.
- Move all plan math to Go as the single source of truth; add a Wails `RecalculateDay` binding; remove duplicated TypeScript math in `plannerSummary.ts`.
- Redesign the desktop "Marcações Sugeridas" card as journey groups with add/remove controls; update Resumo to list N journey spans.
- Adapt CLI TUI (`pm plan`) and reminder scheduler to the new dynamic model.
- Keep Settings anchor times unchanged (four defaults for journeys 1–2); derive anchors for journey 3+ from previous exit + break.

## Capabilities

### New Capabilities

- `journey-planning`: Dynamic multi-journey day planning — journey count, stamp assignment, solved-exit selection, registered-punch handling, summary calculation, and validation across desktop, CLI, and reminders.

### Modified Capabilities

<!-- No existing specs in openspec/specs/ yet -->

## Impact

- **Go core**: `pkg/plan/` — new `ClockSlot`, `Journey`, `Day` types; generalized `AssignStampsToSlots` (DP); `SolveExit`, `Summarize`; remove `Suggestion{In1,Out1,In2,Out2}`.
- **App layer**: `pkg/app/planner.go` — new payload/summary shapes; `RecalculatePlanner` generalized.
- **Config**: `pkg/config/planner.go` — `ResolvePlannerAnchors` returns a slice; YAML keys unchanged.
- **Desktop**: `desktop/app.go` (new binding), `desktop/frontend/src/pages/PlannerPage.tsx`, planner feature components, `types.ts`; delete `plannerSummary.ts` math.
- **CLI**: `pkg/cmd/plan.go`, `pkg/ui/checkins.go` — dynamic journey UI.
- **Reminders**: `pkg/reminder/types.go`, `pkg/reminder/planner.go`, `pkg/reminder/schedule.go` — dynamic slot keys and labels.
- **Tests**: Go unit tests for assign/solve/summarize; TS component tests for journey groups and solved-exit toggle.

## Non-goals

- Changing how punches are submitted to PontoMais (read-only planning tool).
- Per-journey anchor configuration in Settings (only the existing four anchors + derived).
- Solving for entry slots (only exits can be calculated).
- Persisting per-day journey count across sessions (resets on reload based on stamp count).
