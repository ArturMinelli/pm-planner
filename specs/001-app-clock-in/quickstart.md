# Quickstart / Validation

1. Start fake service and dev API: `go run ./cmd/pm-fake-pontomais` and `go run ./cmd/pm-dev --fake` (from `desktop/frontend`: `npm run dev` for the UI).
2. Confirm the "TEST MODE" banner shows.
3. Playwright scenarios (fake scenario via query/flag): success; double-click (exactly 1 register call on the fake); 401; 500; timeout → uncertain → verify; offline; punch < 60 s apart → confirm dialog; geolocation denied → default location shown.
4. After each scenario assert `browser_network_requests` contains no host other than localhost.
5. Unit: `go test ./...` and `cd desktop/frontend && npm test`.
Expected: all scenarios pass, fake server logs show one register per intended punch, zero external requests.
