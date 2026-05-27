#!/usr/bin/env bash
set -euo pipefail

# Builds pm-desktop and registers it for the current user:
#   ~/.local/bin/pm-desktop
#   ~/.local/share/applications/pm-desktop.desktop
#   ~/.local/share/icons/hicolor/scalable/apps/pm-desktop.svg

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOM="${HOME}"
BIN="${HOM}/.local/bin"
APP="${HOM}/.local/share/applications"
ICON_DIR="${HOM}/.local/share/icons/hicolor/scalable/apps"

DESKTOP_TEMPLATE="${ROOT}/packaging/pm-desktop.desktop"
ICON_SRC="${ROOT}/packaging/icons/pm-desktop.svg"
DESKTOP_OUT="${APP}/pm-desktop.desktop"

UNAME_M="$(uname -m)"
GOHOST="$(go env GOARCH)"
DESKTOP_GOARCH="${GOHOST}"
if [[ "$(go env GOHOSTARCH 2>/dev/null || true)" ]]; then :; fi
if [[ "${DESKTOP_FORCE_GOHOST:-}" != "1" && "${GOHOST}" == "386" ]]; then
  case "${UNAME_M}" in
    x86_64 | amd64) DESKTOP_GOARCH="amd64" ;;
  esac
fi

echo "→ Building frontend (npm)…"
(cd "${ROOT}/desktop/frontend" && npm ci && npm run build)

echo "→ Building Go binary (GOARCH=${DESKTOP_GOARCH}, tags production,webkit2_41)…"
mkdir -p "${BIN}"
(cd "${ROOT}/desktop" && CGO_ENABLED=1 GOOS=linux GOARCH="${DESKTOP_GOARCH}" \
  go build -tags production,webkit2_41 -o "${BIN}/pm-desktop" .)
chmod +x "${BIN}/pm-desktop"

echo "→ Installing icon…"
mkdir -p "${ICON_DIR}"
cp -f "${ICON_SRC}" "${ICON_DIR}/pm-desktop.svg"
touch "${HOM}/.local/share/icons/hicolor"

echo "→ Installing .desktop entry…"
mkdir -p "${APP}"
sed -e "s|^Exec=.*|Exec=${BIN}/pm-desktop|" \
    -e "s|^TryExec=.*|TryExec=${BIN}/pm-desktop|" \
    "${DESKTOP_TEMPLATE}" > "${DESKTOP_OUT}"
chmod 644 "${DESKTOP_OUT}"

if command -v update-desktop-database >/dev/null 2>&1; then
  echo "→ update-desktop-database…"
  update-desktop-database "${APP}" 2>/dev/null || true
fi
if command -v gtk-update-icon-cache >/dev/null 2>&1; then
  echo "→ gtk-update-icon-cache…"
  gtk-update-icon-cache -f -t "${HOM}/.local/share/icons/hicolor" 2>/dev/null || true
fi

echo
echo "Installed:"
echo "  ${BIN}/pm-desktop"
echo "  ${DESKTOP_OUT}"
echo "  ${ICON_DIR}/pm-desktop.svg"
echo
echo "Open the Zorin menu (Super) and search: PM, Planner, or PontoMais."
echo "If it does not appear immediately, log out and back in once (or restart gnome-shell)."
