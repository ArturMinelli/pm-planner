# PM Planner

PM Planner is a Go CLI plus a Wails/React desktop app for PontoMais planning
workflows. The CLI owns both product commands and repository maintenance tasks,
so builds do not depend on `make` or shell scripts.

## Requirements

- Go 1.23+
- Node.js/npm for the desktop frontend
- Wails CLI only for live desktop development: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Linux desktop builds need CGO, GTK, and WebKitGTK development headers

On Debian/Ubuntu-based Linux:

```bash
sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

## Project Commands

Run from anywhere inside the repository:

```bash
go run ./cmd/pm project doctor
go run ./cmd/pm project build cli
go run ./cmd/pm project build desktop
go run ./cmd/pm project dev desktop
go run ./cmd/pm project install desktop-menu
go run ./cmd/pm project clean
```

After building the CLI, the same commands are available through `bin/pm`:

```bash
bin/pm project doctor
bin/pm project build desktop
```

Useful flags:

```bash
bin/pm project build cli --goos windows --goarch amd64
bin/pm project build cli --output ./dist/pm
bin/pm project build desktop --skip-frontend
bin/pm project build desktop --force-go-host
bin/pm project install desktop-menu --skip-build
```

Outputs default to `bin/pm` and `bin/pm-desktop` (`.exe` on Windows). Linux
desktop builds automatically use `production,webkit2_41` tags and correct the
common `linux/386` Go-on-`x86_64` mismatch unless `--force-go-host` is set.

## Product Commands

```bash
pm list
pm list --date 2026-06-03
pm plan
pm plan --date 2026-06-03
pm version
```

The planner command shares logic with the desktop planner. Default anchors are
`08:00`, `12:00`, `13:30`, and `18:00`; custom values can be saved from the
desktop Settings page or edited in the YAML config.

## Configuration

The CLI and desktop app share one platform-native config directory:

- Linux: `$XDG_CONFIG_HOME/pm/config.yaml`, or `~/.config/pm/config.yaml`
- macOS: `~/Library/Application Support/pm/config.yaml`
- Windows: `%AppData%\\pm\\config.yaml`

Example:

```yaml
email: "you@example.com"
password: "your-password"
cache_ttl_hours: 8
planner:
  in1: "08:00"
  out1: "12:00"
  in2: "13:30"
  out2: "18:00"
```

Authentication headers are cached as `session.json` in the same directory. If
the API returns an expiry header it is respected; otherwise `cache_ttl_hours`
defaults to 8 hours.

## Desktop App

The desktop shell lives in `desktop/` and imports shared Go packages from the
root module through `desktop/go.mod`.

Development:

```bash
go run ./cmd/pm project dev desktop
```

Production build:

```bash
go run ./cmd/pm project build desktop
```

Linux user-menu install:

```bash
go run ./cmd/pm project install desktop-menu
```

This installs `pm-desktop`, the SVG icon, and the `.desktop` launcher under the
current user’s XDG locations.

## Validation

```bash
go test ./...
npm --prefix desktop/frontend ci
npm --prefix desktop/frontend run lint
npm --prefix desktop/frontend run build
(cd desktop && go test .)
```

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| `wails` not found during dev | Install the Wails CLI with `go install github.com/wailsapp/wails/v2/cmd/wails@latest`. |
| Linux build cannot find WebKitGTK | Install `libwebkit2gtk-4.1-dev`; older distros may need `libwebkit2gtk-4.0-dev` and adjusted tags. |
| `linux/386` links against 64-bit GTK/WebKit | Use `pm project build desktop`; it forces `GOARCH=amd64` on x86_64 Linux unless `--force-go-host` is set. |
| Frontend assets missing | Run `pm project build desktop` without `--skip-frontend`. |

Sensitive API exports and generated build outputs should not be committed. If a
token or password was present in a previous local export, rotate it outside this
repository.
