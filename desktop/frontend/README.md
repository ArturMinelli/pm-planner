# PM Planner Frontend

React + Vite UI for the Wails desktop shell and local browser development.

## Desktop build

Use the root CLI for normal workflows:

```bash
go run ../../cmd/pm project build desktop
go run ../../cmd/pm project dev desktop
```

## Browser development

`npm run dev` starts the local Go API (`cmd/pm-dev`) and Vite together. Open the
Vite URL in your browser — the UI uses the same config file and domain logic as
the desktop app via `/api` (proxied to `127.0.0.1:3847`).

```bash
npm ci
npm run dev
```

## Frontend-only checks

```bash
npm ci
npm run lint
npm run build
npm test
```

Production users run `pm-desktop`; browser mode is for local development only.
