## 1. Go — defaults payload builder

- [x] 1.1 Add `BuildDefaultsPlannerPayload(dateStr string) (*PlannerPayload, error)` in `pkg/app/planner.go` using `config.ResolvePlannerAnchors`, 8h30 target, two unregistered journeys, `solvedExit` on last exit, `originalsLine` `(nenhum)`, no balance
- [x] 1.2 Add unit tests in `pkg/app/planner_test.go` for defaults payload shape, custom anchors from config, and solved exit index
- [x] 1.3 Wrap `FetchPlannerPayload` failure in `pkg/cmd/plan.go` with defaults fallback and stderr warning; add test in `pkg/cmd/plan_test.go`

## 2. Desktop backend binding

- [x] 2.1 Update `LoadPlanner` in `desktop/app.go` to try `FetchPlannerPayload` first, fall back to `BuildDefaultsPlannerPayload`, and return a load warning string (new field on payload or wrapper type)
- [x] 2.2 Expose warning to frontend via `PlannerPayload` JSON (`loadWarning` or similar)

## 3. Frontend — auto-fetch lifecycle

- [x] 3.1 Refactor `PlannerPage.tsx`: `useEffect` on `date` triggers fetch on mount and date change; track `fetching` state; guard against stale responses
- [x] 3.2 Move date picker into `PageHeader`; remove `PlannerLoadCard` usage
- [x] 3.3 Show non-blocking warning banner when `loadWarning` is present
- [x] 3.4 Replace `timesDisabled` logic: disabled while `fetching`, enabled after success or fallback

## 4. Frontend — skeleton loading

- [x] 4.1 Add skeleton placeholder styles in `index.css`
- [x] 4.2 Add `loading` prop to `PlannerJourneyList` and `PlannerSummaryPanel`; render skeletons when true
- [x] 4.3 Update component tests for skeleton and post-load editable state

## 5. Cleanup and integration tests

- [x] 5.1 Delete `PlannerLoadCard.tsx` and its test file; fix any remaining imports
- [x] 5.2 Update `PlannerPage` integration tests for auto-load on mount, date-change re-fetch, and fallback warning
- [x] 5.3 Run full Go and frontend test suites; fix regressions
