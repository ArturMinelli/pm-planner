# Tasks: In-App Clock-In

**Input**: `specs/001-app-clock-in/` (plan, spec, research, data-model, contracts, quickstart)
**Execution model**: The orchestrator (Sonnet) dispatches each task, or each `[P]` group, to a Haiku subagent. It reviews the diff, runs the listed check, and only then ticks the box. Subagents MUST NOT call `*.pontomais.com.br`; only the fake server (T004) may stand in for it. Never use real credentials or the HAR's tokens or PII in fixtures.

## Phase 1: Setup

- [X] T001 Read `pkg/api/client.go`, `pkg/api/workday.go`, `cmd/pm-dev/main.go`, `desktop/app.go`, `pkg/server/httpapi/` and `desktop/frontend/src/services/{backend,httpClient}.ts`; write a 10-line findings note to `specs/001-app-clock-in/notes.md` on how `Client`, `AuthHeaders`, `BaseURL` and the frontend backend adapter are wired.
- [X] T002 Geocode "Av. Presidente Kennedy, 861, Campinas, São José, SC, 88102-401" once (public geocoder, address only) and record lat/lng in `specs/001-app-clock-in/notes.md`, marked "NEEDS USER CONFIRMATION".

## Phase 2: Foundational (blocks all stories)

- [X] T003 [P] Create `pkg/api/punch.go` with types `Location{Latitude,Longitude,Accuracy,Address}` and `Punch{ID,Date,Time,Address,Latitude,Longitude,Disabled}`, plus `FetchEmployee(ctx)` (`GET /api/session` → `session.employee` as `json.RawMessage`) and `ReverseGeocode(ctx, lat, lng, employeeID)` (`GET /api/geo_locations` → `address.formatted_address`). Add `pkg/api/punch_test.go` using `httptest`.
- [X] T004 [P] Create `cmd/pm-fake-pontomais/main.go`, a localhost-only fake serving `/api/session`, `/api/geo_locations`, `/api/time_cards/register` (202, stores the punch in memory and counts calls) and `/api/time_cards/current/last_cached`. Scenario is switchable at `POST /__scenario` with {ok, 401, 500, slow, offline-drop}. Fixtures are synthetic with a fake employee id. Also serve `GET /__calls` returning the register call count.
- [X] T005 Add a `--fake` flag to `cmd/pm-dev/main.go`: set `Client.BaseURL` to the fake server and inject static `AuthHeaders` so sign-in never runs. Expose `testMode:true` in the config/status response. Refuse to start `--fake` if the base URL is not localhost.
- [X] T006 [P] Add `desktop/frontend/src/util/location.ts` with `DEFAULT_LOCATION` (coordinates from T002, address text above), `detectLocation(timeoutMs=3000)` (navigator.geolocation, never rejects, returns `{...location, source:'detected'|'default'}`) and `location.test.ts` covering granted, denied, timeout and unsupported.
- [X] T007 [P] Add pt-BR strings (clock in, locating, sending, registered, not registered, retry, session expired, too close confirm, location source detected/default, test-mode banner) to `desktop/frontend/src/locales/pt-BR.json`.

## Phase 3: User Story 1 - Register a clock-in (P1) 🎯 MVP

**Goal**: One click registers a punch and shows it. **Independent test**: against the fake, click "Clock in" and see a new punch plus a success message.

- [X] T008 [US1] Implement `RegisterPunch(ctx, loc Location)` in `pkg/api/punch.go`. It builds the body from research R1 (employee from `FetchEmployee`, `location_edited:false`, `_path:"/registrar-ponto"`, `_device` with a stable generated uuid), sets headers `api-version: 2` and `uuid`, expects 202, and returns the `Punch`. Add tests with `httptest` for 202, 401, 500.
- [X] T009 [US1] Create `pkg/app/clockin.go` with `ClockIn(ctx, req)`. It reverse-geocodes the address (falls back to text on error), calls `RegisterPunch`, and returns `{status, punch, locationSource}`. Hold a mutex so only one call runs at a time. Test it in `pkg/app/clockin_test.go`.
- [X] T010 [US1] Add the `POST /api/clockin` route in `pkg/server/httpapi/` and the `ClockIn` binding in `desktop/app.go`, per `contracts/clockin.md`.
- [X] T011 [US1] Add `clockIn(req)` to `desktop/frontend/src/services/backend.ts` and `httpClient.ts`, plus types in `src/types.ts`.
- [X] T012 [US1] Create `src/hooks/useClockIn.ts` with state idle→locating→sending→success|failed, using `detectLocation`. Ignore calls while busy (FR-002). Add `useClockIn.test.ts`.
- [X] T013 [US1] Create `src/features/planner/ClockInButton.tsx` (+ test). It has a progress state, is disabled while busy, is keyboard accessible with `aria-live` status, and shows the location source. Mount it in `src/pages/PlannerPage.tsx`; on success, reload the planner payload so the punch and totals update (FR-005/006). Show the TEST MODE banner when `testMode` is true (FR-012).

**Checkpoint**: run `go test ./...`, `npm test`, `npm run build` in `desktop/frontend`.

## Phase 4: User Story 2 - Failure feedback and safe retry (P2)

**Goal**: Every failure states clearly whether a punch was registered; retry never duplicates. **Independent test**: fake scenarios 401/500/offline/slow.

- [X] T014 [US2] In `pkg/app/clockin.go`, map errors to the contract codes (`session_expired`, `service_rejected`, `offline`, `uncertain`), each with `registered`. On timeout or network error after sending, call `last_cached` and compare against the pre-send punch list to decide registered vs not. Add table tests.
- [X] T015 [US2] Add a 60 s near-duplicate guard in `pkg/app/clockin.go` returning `409 too_close` unless `confirmNear` is true (FR-009). Test the boundary.
- [X] T016 [US2] Extend `useClockIn.ts` and `ClockInButton.tsx` for failed/uncertain states: show the message from the locale, a Retry button, a confirm dialog for `too_close`, and a sign-in prompt on `session_expired`. Add tests.

## Phase 5: User Story 3 - Automatic location with default (P3)

**Goal**: No location prompts, ever. **Independent test**: geolocation granted vs denied.

- [X] T017 [US3] Verify `detectLocation` is wired into `useClockIn`, that accuracy is sent as `accuracy`, and that the UI shows "detected" vs "default" (FR-004). Add component tests for both.
- [X] T018 [US3] Confirm with the user the coordinates from T002 (or user-supplied ones). Then run them through the fake's `/api/geo_locations` shape and bake the service-format address into `DEFAULT_LOCATION`.

## Phase 6: UI/UX validation with Playwright (orchestrator-owned, no production)

- [X] T019 Start the stack per `quickstart.md` (`pm-fake-pontomais`, `pm-dev --fake`, `vite`). Via Playwright MCP, block `**/*.pontomais.com.br/**`, then verify: TEST MODE banner, success flow, double-click leaves the fake's `/__calls` at 1, 401, 500, slow→uncertain→verify, offline, <60 s confirm, geolocation denied→default.
- [X] T020 After each scenario, check `browser_network_requests`: only localhost hosts. Capture screenshots at desktop and narrow widths and review the layout, focus order and keyboard operation. Record the results in `specs/001-app-clock-in/validation.md`.

## Phase 7: Polish

- [X] T021 [P] Run `go vet ./...`, `go test ./...`, `npm run lint`, `npm test`, `npm run build`, and fix findings.
- [X] T022 Orchestrator review: diff audit that no code path reaches production without an explicit user action, no secrets or PII in fixtures or logs, and that all FR/SC items in `spec.md` are covered.

## Dependencies & Execution

- Phase 1 → Phase 2 → US1 → US2 → US3 (US2 and US3 only need US1's files; US3's T017 can run in parallel with US2 if different people take them) → Phase 6 → Phase 7.
- Parallel: T003/T004/T006/T007 (different files). Within US1, T011 can run alongside T009/T010 once the contract is fixed.
- **MVP**: Phases 1–3 plus a smoke run of T019's success scenario.
- Real-production punching is only ever done by the user manually, after validation.
