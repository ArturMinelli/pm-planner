# UI/UX Validation (Playwright, fake PontoMais only)

Stack: `cmd/pm-fake-pontomais` (:3850), `pm-dev` with `PM_FAKE_PONTOMAIS_URL=http://127.0.0.1:3850/api` (:3847), vite (:5173). Playwright blocked `*.pontomais.com.br`; `browser_network_requests` filtered on "pontomais" returned nothing (SC-005 ✔).

| Scenario | Result |
|---|---|
| TEST MODE banner visible | ✔ |
| Success flow, punch in list, totals reloaded | ✔ (08:51 shown, originalTimes updated) |
| Double-click | ✔ exactly 1 register call |
| Second punch < 60 s | ✔ blocked, confirm button, 2nd call only after confirm |
| 401 | ✔ "Sessão expirada", retry shown, 1 call |
| 500 (not stored) | ✔ "O ponto NÃO foi registrado" |
| Dropped connection, punch stored | ✔ recovered as registered, 1 punch |
| Dropped connection, not stored | ✔ "Sem conexão. O ponto NÃO foi registrado." |
| Geolocation denied (headless default) | ✔ default office location, labelled |
| Geolocation granted | ✔ detected coordinates sent (-27.59,-48.55), labelled |
| 375px width, keyboard (focus+Enter) | ✔ no horizontal scroll, punch registered |

Screenshots: `.playwright-mcp/us1-success.png`, `.playwright-mcp/us3-mobile.png` (pre-polish).
Notes: empty-day planner shows the existing "usando padrões" fallback (pre-existing behavior, not caused by this feature). Real WebKitGTK geolocation and real production punch are NOT verified.
