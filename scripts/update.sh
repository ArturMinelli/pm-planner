#!/usr/bin/env bash
# PM Planner — atualiza instalação existente (macOS e Linux).
# Obtém a versão mais recente do código-fonte e recompila/reinstala CLI e desktop.

set -euo pipefail

readonly REPO_URL="https://github.com/ArturMinelli/pm-planner.git"
readonly REPO_TARBALL_URL="https://github.com/ArturMinelli/pm-planner/archive/refs/heads/main.tar.gz"
case "$(uname -s)" in
	Darwin) DEFAULT_INSTALL_DIR="$HOME/Library/Application Support/pm-planner" ;;
	*)      DEFAULT_INSTALL_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/pm-planner" ;;
esac
readonly DEFAULT_INSTALL_DIR
readonly LEGACY_INSTALL_DIR="$HOME/pm-planner"
readonly SETUP_URL="https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.sh"

daemon_was_running=false
autoupdate_flag=0
no_autoupdate_flag=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()  { printf '%b\n' "${BLUE}→${NC} $*" >&2; }
ok()    { printf '%b\n' "${GREEN}✓${NC} $*" >&2; }
warn()  { printf '%b\n' "${YELLOW}!${NC} $*" >&2; }
fail()  { printf '%b\n' "${RED}✗${NC} $*" >&2; }
die()   { fail "$*"; exit 1; }

usage() {
	cat <<EOF
Uso: update.sh [opções]

Atualiza uma instalação existente do PM Planner: baixa código novo e
recompila/reinstala a CLI (pm) e o app desktop.

Opções:
  --autoupdate    Registra auto-atualização diária via scheduler do OS
  --no-autoupdate Remove o registro de auto-atualização do OS
  --help          Exibe esta ajuda

Exemplos:
  curl -fsSL https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/update.sh | bash
  cd ~/pm-planner && ./scripts/update.sh
  cd ~/pm-planner && ./scripts/update.sh --autoupdate
EOF
}

parse_args() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
			--autoupdate)    autoupdate_flag=1; shift ;;
			--no-autoupdate) no_autoupdate_flag=1; shift ;;
			--help|-h)       usage; exit 0 ;;
			*)               die "Opção desconhecida: $1 (use --help)" ;;
		esac
	done
}

detect_os() {
	case "$(uname -s)" in
		Darwin) echo "darwin" ;;
		Linux)  echo "linux" ;;
		*)      die "Sistema operacional não suportado por update.sh: $(uname -s). Use scripts/update.ps1 no Windows." ;;
	esac
}

is_project_root() {
	[[ -f "$1/go.mod" ]] && grep -q '^module pm-cli$' "$1/go.mod" && [[ -f "$1/desktop/wails.json" ]]
}

find_project_root() {
	local dir="${1:-$(pwd)}"
	dir="$(cd "$dir" && pwd)"
	while [[ "$dir" != "/" ]]; do
		if is_project_root "$dir"; then
			echo "$dir"
			return 0
		fi
		dir="$(dirname "$dir")"
	done

	local bash_source="${BASH_SOURCE[0]:-}"
	if [[ -n "$bash_source" && "$bash_source" != "/dev/stdin" && -f "$bash_source" ]]; then
		local script_dir candidate
		script_dir="$(cd "$(dirname "$bash_source")" && pwd)"
		candidate="$(dirname "$script_dir")"
		if is_project_root "$candidate"; then
			echo "$candidate"
			return 0
		fi
	fi
	return 1
}

find_existing_install() {
	local root=""
	if root="$(find_project_root 2>/dev/null)"; then
		echo "$root"
		return 0
	fi
	if is_project_root "$DEFAULT_INSTALL_DIR"; then
		echo "$DEFAULT_INSTALL_DIR"
		return 0
	fi
	if is_project_root "$LEGACY_INSTALL_DIR"; then
		echo "$LEGACY_INSTALL_DIR"
		return 0
	fi
	return 1
}

migrate_install_dir() {
	if ! is_project_root "$LEGACY_INSTALL_DIR"; then
		return 0
	fi

	local new_dir="$DEFAULT_INSTALL_DIR"

	if [[ "$(cd "$LEGACY_INSTALL_DIR" && pwd)" == "$(cd "$new_dir" 2>/dev/null && pwd)" ]]; then
		return 0
	fi

	if is_project_root "$new_dir"; then
		warn "Destino $new_dir já contém uma instalação válida. Removendo diretório legado $LEGACY_INSTALL_DIR..."
		rm -rf "$LEGACY_INSTALL_DIR"
		ok "Diretório legado removido"
		return 0
	fi

	info "Migrando instalação de $LEGACY_INSTALL_DIR → $new_dir ..."
	mkdir -p "$(dirname "$new_dir")"
	mv "$LEGACY_INSTALL_DIR" "$new_dir"
	ok "Instalação migrada para $new_dir"

	local new_wrapper="$new_dir/scripts/auto-update.sh"
	case "$(detect_os)" in
		linux)
			local desktop_file
			desktop_file="$(_autoupdate_desktop_file)"
			if [[ -f "$desktop_file" ]]; then
				sed -i "s|Exec=bash \"[^\"]*auto-update\.sh\"|Exec=bash \"$new_wrapper\"|" "$desktop_file"
				ok "Autostart atualizado: $desktop_file"
			fi
			;;
		darwin)
			local plist_file
			plist_file="$(_autoupdate_plist_file)"
			if [[ -f "$plist_file" ]]; then
				launchctl unload "$plist_file" 2>/dev/null || true
				sed -i '' "s|<string>[^<]*auto-update\.sh</string>|<string>$new_wrapper</string>|" "$plist_file"
				launchctl load "$plist_file" 2>/dev/null || true
				ok "LaunchAgent atualizado: $plist_file"
			fi
			;;
	esac
}

fetch_project_tarball() {
	local target="$1"
	local tmp
	tmp="$(mktemp -d)"

	info "Baixando código-fonte (tarball)..."
	curl -fsSL "$REPO_TARBALL_URL" -o "$tmp/pm-planner.tar.gz"
	tar -xzf "$tmp/pm-planner.tar.gz" -C "$tmp"
	rm -rf "$target"
	mv "$tmp/pm-planner-main" "$target"
	rm -rf "$tmp"
	ok "Código-fonte atualizado em $target"
}

update_git_repo() {
	local root="$1"
	local branch="main"

	info "Atualizando repositório git em $root..."
	(
		cd "$root"
		if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
			die "Diretório $root não é um repositório git válido"
		fi

		if git remote get-url origin >/dev/null 2>&1; then
			git fetch origin --prune
		else
			warn "Remote origin ausente — adicionando $REPO_URL"
			git remote add origin "$REPO_URL"
			git fetch origin --prune
		fi

		if git show-ref --verify --quiet "refs/heads/$branch"; then
			git checkout "$branch"
		elif git show-ref --verify --quiet "refs/remotes/origin/$branch"; then
			git checkout -B "$branch" "origin/$branch"
		else
			die "Branch $branch não encontrada no repositório"
		fi

		if ! git diff --quiet || ! git diff --cached --quiet; then
			die "Há alterações locais em $root. Faça commit, stash ou descarte antes de atualizar:
  git -C $root status
  git -C $root stash -u   # guardar alterações temporariamente"
		fi

		if ! git pull --ff-only origin "$branch"; then
			die "git pull falhou em $root. Resolva conflitos manualmente e execute o update novamente."
		fi
	)
	ok "Repositório git atualizado (branch $branch)"
}

update_source() {
	local root="$1"

	if [[ -d "$root/.git" ]]; then
		update_git_repo "$root"
		return 0
	fi

	warn "Instalação sem git detectada — substituindo por tarball mais recente"
	fetch_project_tarball "$root"
}

config_dir() {
	if [[ -n "${XDG_CONFIG_HOME:-}" ]]; then
		echo "${XDG_CONFIG_HOME}/pm"
	else
		echo "${HOME}/.config/pm"
	fi
}

daemon_pid_file() {
	echo "$(config_dir)/reminder-daemon.pid"
}

daemon_is_running() {
	local pid_file pid
	pid_file="$(daemon_pid_file)"
	[[ -f "$pid_file" ]] || return 1
	pid="$(tr -d '[:space:]' < "$pid_file")"
	[[ "$pid" =~ ^[0-9]+$ ]] || return 1
	kill -0 "$pid" 2>/dev/null
}

installed_desktop_binary() {
	case "$(detect_os)" in
		darwin)
			echo "$HOME/Applications/PM Planner.app/Contents/MacOS/pm-desktop"
			;;
		linux)
			local bin_home="${XDG_BIN_HOME:-$HOME/.local/bin}"
			echo "$bin_home/pm-desktop"
			;;
	esac
}

wait_for_process_exit() {
	local pid="$1"
	local max_waits="${2:-10}"
	local waited=0
	while kill -0 "$pid" 2>/dev/null && [[ $waited -lt $max_waits ]]; do
		sleep 0.5
		waited=$((waited + 1))
	done
}

pm_desktop_processes_running() {
	pgrep -x pm-desktop >/dev/null 2>&1
}

stop_pm_desktop_processes() {
	if ! pm_desktop_processes_running; then
		return 0
	fi
	pkill -TERM -x pm-desktop 2>/dev/null || true
	local waited=0
	while pm_desktop_processes_running && [[ $waited -lt 10 ]]; do
		sleep 0.5
		waited=$((waited + 1))
	done
	if ! pm_desktop_processes_running; then
		return 0
	fi
	warn "App desktop ainda em execução; forçando encerramento..."
	pkill -KILL -x pm-desktop 2>/dev/null || true
	sleep 0.5
}

stop_reminder_daemon() {
	local pid_file pid
	pid_file="$(daemon_pid_file)"
	[[ -f "$pid_file" ]] || return 0
	pid="$(tr -d '[:space:]' < "$pid_file")"
	if [[ "$pid" =~ ^[0-9]+$ ]] && kill -0 "$pid" 2>/dev/null; then
		info "Encerrando daemon de lembretes (PID $pid)..."
		kill -TERM "$pid" 2>/dev/null || true
		wait_for_process_exit "$pid" 10
		if kill -0 "$pid" 2>/dev/null; then
			warn "Daemon não respondeu ao SIGTERM; forçando encerramento..."
			kill -KILL "$pid" 2>/dev/null || true
		fi
		ok "Daemon encerrado"
	fi
	rm -f "$pid_file"
}

stop_desktop_app() {
	info "Encerrando app desktop PM Planner..."
	case "$(detect_os)" in
		darwin)
			osascript -e 'tell application "PM Planner" to quit' 2>/dev/null || true
			sleep 1
			;;
	esac
	stop_pm_desktop_processes
	if pm_desktop_processes_running; then
		die "Não foi possível encerrar o app desktop. Feche o PM Planner manualmente e execute o update novamente."
	fi
	ok "App desktop encerrado"
}

stop_running_processes() {
	if daemon_is_running; then
		daemon_was_running=true
	fi
	stop_reminder_daemon
	stop_desktop_app
}

start_reminder_daemon() {
	local binary
	binary="$(installed_desktop_binary)"
	if [[ ! -x "$binary" ]]; then
		warn "Binário desktop não encontrado em $binary — daemon não reiniciado."
		return 0
	fi
	info "Reiniciando daemon de lembretes..."
	nohup "$binary" --daemon >/dev/null 2>&1 &
	disown 2>/dev/null || true
	ok "Daemon de lembretes reiniciado"
}

maybe_restart_daemon() {
	if [[ "$daemon_was_running" != true ]]; then
		return 0
	fi
	start_reminder_daemon
}

ensure_session_path() {
	local paths=()
	[[ -d "$HOME/.local/go/bin" ]] && paths+=("$HOME/.local/go/bin")
	[[ -d "$HOME/go/bin" ]] && paths+=("$HOME/go/bin")
	if command -v go >/dev/null 2>&1; then
		paths+=("$(go env GOPATH 2>/dev/null)/bin")
	fi

	local path_entry
	for path_entry in "${paths[@]}"; do
		case ":$PATH:" in
			*":$path_entry:"*) ;;
			*) export PATH="$PATH:$path_entry" ;;
		esac
	done
}

rebuild_and_install() {
	local root="$1"

	if ! command -v go >/dev/null 2>&1; then
		die "Go não encontrado no PATH. Execute o setup novamente:
  curl -fsSL $SETUP_URL | bash"
	fi

	info "Instalando CLI (pm)..."
	(cd "$root" && go install ./cmd/pm)
	ensure_session_path
	local gopath_bin
	gopath_bin="$(go env GOPATH)/bin"
	case ":$PATH:" in
		*":$gopath_bin:"*) ;;
		*) export PATH="$PATH:$gopath_bin" ;;
	esac
	ok "CLI instalada em $gopath_bin/pm"

	info "Compilando app desktop..."
	(cd "$root" && go run ./cmd/pm project build desktop)

	info "Instalando app desktop..."
	(cd "$root" && go run ./cmd/pm project install desktop --skip-build)

	ok "CLI e app desktop reinstalados com sucesso!"
}

_autoupdate_desktop_file() {
	echo "${XDG_CONFIG_HOME:-$HOME/.config}/autostart/pm-planner-autoupdate.desktop"
}

_autoupdate_plist_file() {
	echo "$HOME/Library/LaunchAgents/com.pm-planner.autoupdate.plist"
}

install_autoupdate() {
	local root="$1"
	local wrapper="$root/scripts/auto-update.sh"

	chmod +x "$wrapper" 2>/dev/null || true

	case "$(detect_os)" in
		linux)
			local autostart_dir="${XDG_CONFIG_HOME:-$HOME/.config}/autostart"
			local desktop_file="$autostart_dir/pm-planner-autoupdate.desktop"
			mkdir -p "$autostart_dir"
			cat > "$desktop_file" <<EOF
[Desktop Entry]
Type=Application
Name=PM Planner Auto Update
Comment=Atualiza o PM Planner automaticamente uma vez por dia ao fazer login
Exec=bash "$wrapper"
Hidden=false
NoDisplay=true
X-GNOME-Autostart-enabled=true
EOF
			ok "Auto-atualização registrada: $desktop_file"
			;;
		darwin)
			local agents_dir="$HOME/Library/LaunchAgents"
			local plist_file="$agents_dir/com.pm-planner.autoupdate.plist"
			mkdir -p "$agents_dir"
			cat > "$plist_file" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.pm-planner.autoupdate</string>
	<key>ProgramArguments</key>
	<array>
		<string>/bin/bash</string>
		<string>$wrapper</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>StandardOutPath</key>
	<string>/dev/null</string>
	<key>StandardErrorPath</key>
	<string>/dev/null</string>
</dict>
</plist>
EOF
			launchctl load "$plist_file" 2>/dev/null || true
			ok "Auto-atualização registrada: $plist_file"
			;;
	esac
	info "Log de auto-atualização: ${XDG_CACHE_HOME:-$HOME/.cache}/pm/autoupdate.log"
}

uninstall_autoupdate() {
	case "$(detect_os)" in
		linux)
			local desktop_file
			desktop_file="$(_autoupdate_desktop_file)"
			if [[ -f "$desktop_file" ]]; then
				rm -f "$desktop_file"
				ok "Auto-atualização removida ($desktop_file)"
			else
				warn "Auto-atualização não estava registrada"
			fi
			;;
		darwin)
			local plist_file
			plist_file="$(_autoupdate_plist_file)"
			if [[ -f "$plist_file" ]]; then
				launchctl unload "$plist_file" 2>/dev/null || true
				rm -f "$plist_file"
				ok "Auto-atualização removida ($plist_file)"
			else
				warn "Auto-atualização não estava registrada"
			fi
			;;
	esac
}

print_next_steps() {
	local root="$1"
	echo ""
	info "Atualização concluída."
	echo "  pm --version          # CLI atualizada"
	if [[ -n "$root" ]]; then
		echo "  Código-fonte em: $root"
	fi
	case "$(detect_os)" in
		darwin) echo "  App desktop: ~/Applications/PM Planner.app (Launchpad)" ;;
		linux)  echo "  App desktop: menu de aplicativos (PM Planner)" ;;
	esac
}

main() {
	parse_args "$@"

	echo ""
	info "PM Planner — update ($(detect_os))"
	echo ""

	migrate_install_dir

	local root
	if ! root="$(find_existing_install)"; then
		die "Instalação não encontrada. Execute o setup primeiro:
  curl -fsSL $SETUP_URL | bash

Ou, se já clonou manualmente, execute este script de dentro de pm-planner/."
	fi

	ok "Instalação encontrada em $root"
	ensure_session_path

	update_source "$root"
	stop_running_processes
	rebuild_and_install "$root"
	maybe_restart_daemon

	if command -v go >/dev/null 2>&1; then
		info "Verificando ambiente (project doctor)..."
		(cd "$root" && go run ./cmd/pm project doctor) || warn "Verificação do projeto reportou problemas (veja acima)"
	fi

	if [[ "$autoupdate_flag" -eq 1 ]]; then
		install_autoupdate "$root"
	elif [[ "$no_autoupdate_flag" -eq 1 ]]; then
		uninstall_autoupdate
	fi

	print_next_steps "$root"
	echo ""
	ok "Update concluído."
}

main "$@"
