# Implementation Plan: In-App Clock-In

**Branch**: `001-app-clock-in` | **Date**: 2026-11-17 | **Spec**: [spec.md](spec.md)

## Summary

Add a "Clock in" action to the planner. The Go backend (`pkg/api`) gains `RegisterPunch`, which replays the PontoMais `POST /api/time_cards/register` call seen in the HAR. It is exposed through the Wails app binding and the dev HTTP server, and consumed by a new frontend hook and button. Location comes from the webview geolocation (3 s timeout) and falls back to a built-in default. The address string is resolved through the service's own read-only `GET /api/geo_locations`. A fake PontoMais server plus injected auth makes UI validation in Playwright impossible to reach production.

## Technical Context

**Language/Version**: Go 1.23 (backend, `pm-cli` module + `desktop` Wails v2.12 module); TypeScript/React + Vite (frontend)
**Primary Dependencies**: existing only (net/http, Wails, React, vitest). No new dependencies.
**Storage**: none new (session cache already in `pkg/auth`)
**Testing**: `go test` (httptest fake server), `vitest`, Playwright MCP against `vite` + `pm-dev --fake`
**Target Platform**: Wails desktop (linux/mac/windows) + dev browser mode
**Project Type**: desktop-app (Go backend + React frontend)
**Performance Goals**: definitive result ≤ 5 s (SC-002); location detection must not add more than 3 s
**Constraints**: Never call production in tests (SC-005); exactly-once submission; no PII/tokens in logs or fixtures
**Scale/Scope**: 1 button, 1 hook, 1 backend method, 1 fake server

## Constitution Check

`.specify/memory/constitution.md` is still the unfilled template, so there are no enforceable gates. Self-imposed gates: no new dependencies, reuse the existing `Client`/auth/session code, tests for each new non-trivial branch. **Pass.**

## Project Structure

### Documentation (this feature)

```text
specs/001-app-clock-in/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/clockin.md
└── tasks.md   # /speckit-tasks
```

### Source Code

```text
pkg/api/punch.go            # RegisterPunch, ReverseGeocode, FetchEmployee (+ punch_test.go)
pkg/app/clockin.go          # orchestration: location resolve, dedupe guard, post-verify
pkg/server/httpapi/         # POST /api/clockin route
desktop/app.go              # ClockIn binding
cmd/pm-fake-pontomais/      # fake service for tests/UI validation (localhost only)
cmd/pm-dev/                 # --fake flag: BaseURL -> fake, static AuthHeaders
desktop/frontend/src/
├── services/backend.ts, httpClient.ts   # clockIn()
├── hooks/useClockIn.ts                  # state: idle|locating|sending|success|failed|uncertain
├── features/planner/ClockInButton.tsx   # + tests
├── util/location.ts                     # geolocation w/ timeout + DEFAULT_LOCATION
└── locales/pt-BR.json (only locale)
```

**Structure Decision**: extend the existing Go-core + React layout; no new modules except the fake server command.

## Complexity Tracking

None.
