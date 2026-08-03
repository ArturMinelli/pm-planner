## 1. Go core — types and stamp assignment

- [x] 1.1 Add `ClockSlot`, `Journey`, `Day`, and `Summary` types in `pkg/plan/day.go`
- [x] 1.2 Implement `ResolveAnchors(base []string, slotCount int, breakDuration, target time.Duration) []string` for derived anchors beyond index 4
- [x] 1.3 Replace brute-force `assignStampsToPlannerSlots` with order-preserving DP in `pkg/plan/assign.go`; generalize to arbitrary slot count
- [x] 1.4 Port existing assign tests (lone midday punch → Saída 1, 4+ stamps) and add tests for 6-slot / 3-journey assignment
- [x] 1.5 Implement `SuggestDay(date, stamps, periods, target, anchors []string) (Day, error)` replacing `Suggest`; journey count = `max(2, ceil(len(stamps)/2))`
- [x] 1.6 Implement `SolveExit(day Day, target time.Duration, solvedIndex int, date time.Time) Day` with clamp and warning collection
- [x] 1.7 Implement `Summarize(day Day, target time.Duration, date time.Time) Summary` with per-journey spans, total, overtime, and ordering warnings
- [x] 1.8 Implement `DefaultSolvedExit(day Day) int` (last non-registered exit, or -1)
- [x] 1.9 Add unit tests for SolveExit including the screenshot scenario (4 registered punches → no recalc, Saída 2 = 14:41)
- [x] 1.10 Add unit tests for Summarize with 2 and 3 journeys

## 2. Config and app layer

- [x] 2.1 Change `ResolvePlannerAnchors` in `pkg/config/planner.go` to return `[]string`; keep YAML keys unchanged
- [x] 2.2 Update `PlannerPayload` and `PlannerSummary` in `pkg/app/planner.go` to new shapes (`journeys`, `solvedExit`, `journeySpanSecs`)
- [x] 2.3 Refactor `FetchPlannerPayload` to call `SuggestDay` + `DefaultSolvedExit` + `SolveExit` + `Summarize`
- [x] 2.4 Refactor `RecalculatePlanner` to accept `[]Journey` + `solvedExit`; include balance alternative exit time
- [x] 2.5 Update `pkg/app` tests for new payload shape

## 3. Desktop Wails binding

- [x] 3.1 Add `RecalculateRequest` struct and `RecalculateDay` method on `desktop/app.go`
- [x] 3.2 Update `LoadPlanner` return type to new `PlannerPayload`
- [x] 3.3 Update `desktop/frontend/src/types.ts` with `ClockSlot`, `Journey`, new `PlannerPayload` and `PlannerSummary`
- [x] 3.4 Add `recalculateDay` to `desktop/frontend/src/services/backend.ts`

## 4. Desktop UI — journey groups

- [x] 4.1 Create `PlannerJourneyGroup.tsx` with Entrada input, Saída exit field, registrada badges, remove button
- [x] 4.2 Create `PlannerExitField.tsx` merging time input + calculator toggle + balance badge + alternative time (from `PlannerClockoutField`)
- [x] 4.3 Create `PlannerJourneyList.tsx` with journey groups, "+ Adicionar jornada", and originals line
- [x] 4.4 Refactor `PlannerPage.tsx` to hold `journeys[]` + `solvedExit` state; debounced `recalculateDay` on edit
- [x] 4.5 Implement add-journey (seed entry = prev exit + break, set new exit as solved) and remove-journey (disabled when registered)
- [x] 4.6 Refactor `PlannerSummaryPanel.tsx` to render dynamic journey span rows
- [x] 4.7 Add `.journey-group` styles to `index.css`; remove or repurpose `.time-grid`
- [x] 4.8 Delete math from `plannerSummary.ts`; keep only type re-exports if needed
- [x] 4.9 Update `plannerTimes.ts` for dynamic journey order enforcement (`enforceJourneyOrder`)
- [x] 4.10 Update component tests (`PlannerSummaryPanel`, `PlannerClockoutField` → `PlannerExitField`, new journey list tests)
- [x] 4.11 Manual smoke: load 4-punch day, confirm Saída 2 shows 14:41 not 15:41

## 5. CLI TUI

- [x] 5.1 Refactor `pkg/ui/checkins.go` `PlanModel` to hold `[]Journey` + `solvedExit`
- [x] 5.2 Render dynamic journey rows in TUI viewport with solved-exit indicator
- [x] 5.3 Add keyboard shortcuts to add/remove journey (match desktop semantics)
- [x] 5.4 Update `pkg/cmd/plan.go` to use new `FetchPlannerPayload` and `RecalculatePlanner` shapes
- [x] 5.5 Update CLI plan tests if any exist

## 6. Reminders

- [x] 6.1 Replace fixed slot constants in `pkg/reminder/types.go` with `SlotKey` / `SlotLabel` generators
- [x] 6.2 Refactor `BuildDayPlan` in `pkg/reminder/planner.go` to produce dynamic slots from `SuggestDay`
- [x] 6.3 Update `pkg/reminder/schedule.go` for new slot key format
- [x] 6.4 Update `pkg/reminder/planner_test.go` for 2- and 3-journey day plans

## 7. Cleanup and validation

- [x] 7.1 Remove deprecated `Suggestion`, `ComputeOut2`, `SegmentDurations` four-slot functions (or thin wrappers if still needed internally)
- [x] 7.2 Run `go test ./...` and fix any remaining compile errors in `pkg/plan`, `pkg/app`, `pkg/ui`, `pkg/reminder`, `pkg/cmd`
- [x] 7.3 Run `npm test` in `desktop/frontend` and fix failures
- [x] 7.4 Run `openspec validate --change dynamic-journeys --strict` and confirm all artifacts pass
