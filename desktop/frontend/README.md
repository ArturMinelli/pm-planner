# PM Planner Frontend

React + Vite UI for the Wails desktop shell.

Use the root CLI for normal workflows:

```bash
go run ../../cmd/pm project build desktop
go run ../../cmd/pm project dev desktop
```

Frontend-only checks:

```bash
npm ci
npm run lint
npm run build
```

The app can render in a browser preview, but API/config operations require the
Wails runtime exposed by `pm-desktop`.
