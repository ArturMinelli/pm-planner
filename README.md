# pm-cli

A small Cobra-based CLI to call PontoMais API and a **desktop app** (`pm-desktop`)
for the same workflows with a React UI.

## Install — CLI

```
go mod tidy
go build -o pm
```

Or `make cli` to emit `bin/pm`.

## Desktop app (Wails + React)

The desktop binary lives under [`desktop/`](desktop/) and shares Go packages
(`pkg/api`, `pkg/auth`, `pkg/plan`, `pkg/config`) with the CLI.

**Prerequisites:** Node.js/npm for the frontend build. On Linux the WebView links with
CGO, so install a C toolchain and GTK/WebKitGTK **development** packages, then verify
with `wails doctor` ([Wails prerequisites](https://wails.io/docs/gettingstarted/installation)).

Example on Debian/Ubuntu-based systems:

```bash
sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
# If 4.1 is unavailable, try: libwebkit2gtk-4.0-dev (see output of `wails doctor`)
```

On **64‑bit** Ubuntu/Zorin you should use **64‑bit Go** (`linux-amd64`; `go env GOARCH`
should print `amd64`). A **linux-386** Go install on the same machine often breaks
**CGO**/`wails dev`: the compiler targets 32‑bit object code but only **amd64**
`libc6-dev` headers are present, which produces `bits/libc-header-start.h` missing/fatal
errors.

This repo’s **`Makefile` `desktop` target** detects **386 Go on an x86_64 kernel** and
passes **`GOARCH=amd64`** for the final `go build` so it links against your system’s
64‑bit WebKit/GTK (set **`DESKTOP_FORCE_GOHOST=1`** if you truly need a 386 binary).

Build (frontend + Go):

```
make desktop
```

Without `make` (Node + Go only), mirror the Makefile (tags + arch matter):

```
(cd desktop/frontend && npm ci && npm run build)
mkdir -p bin
# If `go env GOARCH` is 386 on an x86_64 PC, add: GOARCH=amd64 GOOS=linux
(cd desktop && CGO_ENABLED=1 go build -tags production,webkit2_41 -o ../bin/pm-desktop .)
```

Production builds **must** use Go build tag **`production`** (otherwise the binary exits
with: `Wails applications will not build without the correct build tags`). **`wails dev`**
passes **`dev`** for you.

Produces `bin/pm-desktop`. The desktop shell is a separate Go module
([`desktop/go.mod`](desktop/go.mod)); it uses `replace` to import shared `pm-cli`
packages.

Install it on your `PATH`, for example:

```
cp bin/pm-desktop ~/bin/
```

### Linux application menu

A starter desktop entry template is provided at
[`packaging/pm-desktop.desktop`](packaging/pm-desktop.desktop). Copy or symlink it:

```
mkdir -p ~/.local/share/applications
cp packaging/pm-desktop.desktop ~/.local/share/applications/
```

Then assign an icon named `pm-desktop` (same `Icon=` value in the `.desktop`
file)—for example add `pm-desktop.svg` under `~/.local/share/icons/hicolor/scalable/apps/`,
or change `Icon=` in the desktop file to an existing theme icon name.

Ensure `Exec=` resolves (`pm-desktop` on `PATH`). Log out or run
`update-desktop-database ~/.local/share/applications` if needed.

### One-step build + Zorin / GNOME menu (Super search)

From the repo root:

```bash
chmod +x packaging/install-user-menu.sh
./packaging/install-user-menu.sh
```

This installs `~/.local/bin/pm-desktop`, the clock-style icon, and
`~/.local/share/applications/pm-desktop.desktop` with **Keywords** tuned for search
(try **PM**, **Planner**, **PontoMais**, **Relogio**, **pmcli**, etc.).  
If the launcher does not show up immediately, press **Super** twice, or log out/in.

### Optional: Wails CLI

If you install the [Wails CLI](https://wails.io/), you can run live reload development
from the `desktop/` directory:

```bash
cd desktop
wails dev
```

**Important:** `wails dev` sets the build **architecture to `runtime.GOARCH` of the
`wails` executable** (not `GOARCH` in your shell). If you previously installed the Wails
CLI with **386 Go**, `wails dev` keeps compiling **386** and you get linker errors against
64‑bit `/usr/lib/x86_64-linux-gnu` libraries. Fix: install **[linux-amd64 Go](https://go.dev/dl/)**,
put it first on **`PATH`**, verify `go env GOARCH` → **`amd64`**, then run
`go install github.com/wailsapp/wails/v2/cmd/wails@latest` again so `$(which wails)` is
an **amd64** binary.

Packaging via `wails build` targets **darwin** and **`linux/amd64`**, **`linux/arm64`**, etc.;
**`linux/386` is intentionally skipped** by the upstream CLI (`platform … not supported`).
On a 32-bit Go install you can still try `CGO_ENABLED=1 go build -tags production`
if your distro ships matching WebKit/GTK libs; otherwise use **`GOARCH=amd64`** Go (and
matching dependencies) so `wails build`/`wails doctor` match a supported Linux arch.

### Desktop troubleshooting

| Symptom | Fix |
|--------|-----|
| `Wails applications will not build without the correct build tags` at **runtime** | Rebuild with `go build -tags production ./desktop` or use `make desktop` / `wails build`. |
| `gcc: not found` / `runtime/cgo` errors when building | Install `build-essential` (or Xcode CLI tools on macOS); ensure `CGO_ENABLED=1`. |
| `wails doctor` reports missing **libwebkit** | Install WebKitGTK dev packages (see above); run `sudo apt update` after adding repos if needed. |
| `Package webkit2gtk-4.0 was not found` (pkg-config during build) | Ubuntu 24+ only ships **4.1**; this project sets `"build:tags": "webkit2_41"` in [`desktop/wails.json`](desktop/wails.json). Manual `go build` must use **`-tags production,webkit2_41`**. |
| `ld: … incompatível` / `-m32` link against `x86_64-linux-gnu` `.so` | Linking **32‑bit** output with **64‑bit** system GTK/WebKit. Prefer **amd64 Go**; **`make desktop`** forces **`GOARCH=amd64`** when it sees **386 Go** on an **x86_64** kernel. |
| **`wails dev`** still uses **`linux_386/link`** / `-m32` | The Wails CLI was built as **386**. Reinstall it using **amd64 Go** (see **Wails CLI** section above). Setting `GOARCH=amd64` in the shell is **not** enough for `wails dev`. |
| `bits/libc-header-start.h: No such file` (CGO / **`wails dev`**) | Usually **32‑bit Go** (`go env GOARCH` → `386`) on a **64‑bit** OS with only **`libc6-dev:amd64`**. **Recommended:** install **[Go linux-amd64](https://go.dev/dl/)**, put it on `PATH`, run `go env GOARCH` → `amd64`, reinstall **wails**, then **`wails dev`**. **Alternative:** install 32‑bit C headers: `sudo apt install gcc-multilib libc6-dev-i386` (you still need **i386** WebKitGTK `-dev` to match—often not worth it). |
| Blank window / crash at startup | Confirm `frontend/dist` exists (`npm run build` in `desktop/frontend`). |

## Configuration

Place a YAML config file at `$HOME/.config/pm/config.yaml`:

```yaml
email: "you@example.com"
password: "your-password"
# optional: cache TTL in hours if API does not provide expiry (default: 8)
# cache_ttl_hours: 8
```

The CLI and desktop app load and save **the same file**. Desktop changes take effect
through Viper reload so CLI commands see updated credentials immediately.

The CLI will log in automatically using your email/password and cache the
authentication headers in `$HOME/.config/pm/session.json` for reuse. If the API returns
an `Expiry` header it will be respected; otherwise the TTL is used.

## Usage

```
# List points for today
./pm list

# List points for a specific date
./pm list --date 2024-06-11

# Credentials are read from config; no auth flags are required
./pm list --date 2024-06-11
```

The `list` command calls:

```
GET https://api.pontomais.com.br/api/time_card_control/current/work_days/{YYYY-MM-DD}
```

## Planner

Interactive terminal planner (`pm plan`, Bubble Tea live mode) mirrors the desktop
planner screen; shared logic lives in [`pkg/app/planner.go`](pkg/app/planner.go).

When loading a day, [`plan.Suggest`](pkg/plan/compute.go) assigns each punch to the planner
slots that minimize total deviation from configurable anchors (**order preserving**, no forced
ordering by punch index). Factory defaults: `08:00`, `12:00`, `13:30`, `18:00` (Entrada 1 →
Saída 1 → Entrada 2 → Saída 2). A lone lunch punch (for example 12:28) maps nearest **12:00**
(Saída 1); **Entrada 1 stays the morning anchor** when no stamp matches that slot. Shorthand
`0008` refers to **08:00**, not midnight. You can still edit the suggested times in the UI.

Configure anchors in `~/.config/pm/config.yaml` (desktop **Configurações → Horários padrão do
planner**, or edit YAML directly):

```yaml
planner:
  in1: "08:00"
  out1: "12:00"
  in2: "13:30"
  out2: "18:00"
```

Times must be strictly increasing with at least 15 minutes between consecutive slots. Partial
or invalid values fall back per slot to the built-in defaults when loading a day; saving from
the settings UI validates the full set.
