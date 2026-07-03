#!/usr/bin/env bash
# PM Planner — script de setup para macOS e Linux.
# Instala dependências, obtém o código-fonte e compila/instala a CLI e o app desktop.

set -euo pipefail

readonly GO_VERSION="1.23.3"
readonly WAILS_PKG="github.com/wailsapp/wails/v2/cmd/wails@latest"
readonly REPO_URL="https://github.com/ArturMinelli/pm-planner.git"
readonly REPO_TARBALL_URL="https://github.com/ArturMinelli/pm-planner/archive/refs/heads/main.tar.gz"
case "$(uname -s)" in
	Darwin) DEFAULT_INSTALL_DIR="$HOME/Library/Application Support/pm-planner" ;;
	*)      DEFAULT_INSTALL_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/pm-planner" ;;
esac
readonly DEFAULT_INSTALL_DIR
readonly PATH_MARKER="# pm-planner setup"

CHECK_ONLY=0
NO_AUTOUPDATE=0

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
	cat <<'EOF'
Uso: setup.sh [opções]

Instala dependências, compila e instala a CLI (pm) e o app desktop do PM Planner.

Opções:
  --check-only    Apenas verifica dependências, sem instalar nem compilar
  --no-autoupdate Remove o registro de auto-atualização do OS (padrão: registrar)
  --help          Exibe esta ajuda

Variáveis de ambiente (úteis com curl | bash):
  PM_NO_AUTOUPDATE=1  Equivalente a --no-autoupdate

Exemplos:
  curl -fsSL https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.sh | bash
  git clone https://github.com/ArturMinelli/pm-planner.git && cd pm-planner
  ./scripts/setup.sh
  ./scripts/setup.sh --no-autoupdate
EOF
}

apply_env_flags() {
	case "${PM_NO_AUTOUPDATE:-}" in
		1|true|yes|TRUE|YES|True) NO_AUTOUPDATE=1 ;;
	esac
}

parse_args() {
	while [[ $# -gt 0 ]]; do
		case "$1" in
			--build)         shift ;; # obsoleto: compilação é o comportamento padrão
			--check-only)    CHECK_ONLY=1; shift ;;
			--autoupdate)    shift ;; # obsoleto: auto-atualização já é o padrão
			--no-autoupdate) NO_AUTOUPDATE=1; shift ;;
			--help|-h)       usage; exit 0 ;;
			*)               die "Opção desconhecida: $1 (use --help)" ;;
		esac
	done
}

detect_os() {
	case "$(uname -s)" in
		Darwin) echo "darwin" ;;
		Linux)  echo "linux" ;;
		*)      die "Sistema operacional não suportado por setup.sh: $(uname -s). Use scripts/setup.ps1 no Windows." ;;
	esac
}

detect_linux_pm() {
	if command -v apt-get >/dev/null 2>&1; then
		echo "apt"
	elif command -v dnf >/dev/null 2>&1; then
		echo "dnf"
	elif command -v pacman >/dev/null 2>&1; then
		echo "pacman"
	else
		echo "unknown"
	fi
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

	# Tenta a partir do diretório do script.
	# Ignora quando executado via pipe (BASH_SOURCE[0] é vazio ou /dev/stdin).
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
	ok "Código-fonte extraído em $target"
}

clone_project_repo() {
	local target="$1"
	rm -rf "$target"
	info "Clonando repositório..."
	git clone "$REPO_URL" "$target"
	ok "Repositório clonado em $target"
}

ensure_project_root() {
	local root=""

	if root="$(find_project_root 2>/dev/null)"; then
		echo "$root"
		return 0
	fi

	if is_project_root "$DEFAULT_INSTALL_DIR"; then
		ok "Repositório encontrado em $DEFAULT_INSTALL_DIR"
		echo "$DEFAULT_INSTALL_DIR"
		return 0
	fi

	if [[ "$CHECK_ONLY" -eq 1 ]]; then
		return 1
	fi

	if [[ -d "$DEFAULT_INSTALL_DIR" ]]; then
		warn "Removendo instalação inválida em $DEFAULT_INSTALL_DIR"
		rm -rf "$DEFAULT_INSTALL_DIR"
	fi

	info "Obtendo código-fonte em $DEFAULT_INSTALL_DIR..."
	if command -v git >/dev/null 2>&1; then
		clone_project_repo "$DEFAULT_INSTALL_DIR"
		echo "$DEFAULT_INSTALL_DIR"
		return 0
	fi

	fetch_project_tarball "$DEFAULT_INSTALL_DIR"
	echo "$DEFAULT_INSTALL_DIR"
}

ensure_path_block() {
	local go_bin=""
	if command -v go >/dev/null 2>&1; then
		go_bin="$(go env GOPATH 2>/dev/null)/bin"
	fi
	# Default GOPATH when go is not yet on PATH (fresh install).
	local default_go_bin="$HOME/go/bin"

	local paths=()
	[[ -d "$HOME/.local/go/bin" ]] && paths+=("$HOME/.local/go/bin")
	[[ -n "$go_bin" && "$go_bin" != "/bin" ]] && paths+=("$go_bin")
	# Always include the default Go bin so freshly-installed binaries are found.
	[[ "$default_go_bin" != "$go_bin" ]] && paths+=("$default_go_bin")

	for p in "${paths[@]}"; do
		case ":$PATH:" in
			*":$p:"*) ;;
			*) export PATH="$PATH:$p" ;;
		esac
	done

	if [[ "$CHECK_ONLY" -eq 1 ]]; then
		return 0
	fi

	local shell_rc=""
	if [[ -f "$HOME/.zshrc" ]]; then
		shell_rc="$HOME/.zshrc"
	elif [[ -f "$HOME/.bashrc" ]]; then
		shell_rc="$HOME/.bashrc"
	fi

	if [[ -z "$shell_rc" ]]; then
		return 0
	fi

	if grep -q "$PATH_MARKER" "$shell_rc" 2>/dev/null; then
		return 0
	fi

	{
		echo ""
		echo "$PATH_MARKER"
		echo 'export PATH="$PATH:$HOME/.local/go/bin"'
		echo 'export PATH="$PATH:$HOME/go/bin"'
		echo 'if command -v go >/dev/null 2>&1; then export PATH="$PATH:$(go env GOPATH)/bin"; fi'
	} >>"$shell_rc"
	warn "PATH atualizado em $shell_rc — reinicie o terminal ou execute: source $shell_rc"
}

go_version_ok() {
	if ! command -v go >/dev/null 2>&1; then
		return 1
	fi
	local ver
	ver="$(go version | awk '{print $3}' | sed 's/^go//')"
	local major minor
	major="${ver%%.*}"
	local rest="${ver#*.}"
	minor="${rest%%.*}"
	[[ "$major" -gt 1 ]] || { [[ "$major" -eq 1 ]] && [[ "$minor" -ge 23 ]]; }
}

install_go_tarball() {
	local arch os tarball url
	arch="$(uname -m)"
	case "$arch" in
		x86_64)  arch="amd64" ;;
		aarch64) arch="arm64" ;;
		arm64)   arch="arm64" ;;
		*)       die "Arquitetura não suportada para instalação automática do Go: $arch" ;;
	esac
	os="$1"

	tarball="go${GO_VERSION}.${os}-${arch}.tar.gz"
	url="https://go.dev/dl/${tarball}"

	info "Baixando Go ${GO_VERSION}..."
	local tmp
	tmp="$(mktemp -d)"

	curl -fsSL "$url" -o "$tmp/$tarball"
	rm -rf "$HOME/.local/go"
	mkdir -p "$HOME/.local"
	tar -C "$HOME/.local" -xzf "$tmp/$tarball"
	rm -rf "$tmp"
	export PATH="$HOME/.local/go/bin:$PATH"
	ok "Go ${GO_VERSION} instalado em ~/.local/go"
}

install_go_darwin() {
	if command -v brew >/dev/null 2>&1; then
		info "Instalando Go via Homebrew..."
		brew install go
		ok "Go instalado via Homebrew"
		return 0
	fi
	install_go_tarball "darwin"
}

install_go() {
	if go_version_ok; then
		ok "Go $(go version | awk '{print $3}') já instalado"
		return 0
	fi
	if [[ "$CHECK_ONLY" -eq 1 ]]; then
		warn "Go 1.23+ não encontrado"
		return 1
	fi
	info "Instalando Go 1.23+..."
	case "$(detect_os)" in
		darwin) install_go_darwin ;;
		linux)  install_go_tarball "linux" ;;
	esac
}

have_node() {
	command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1
}

install_node() {
	if have_node; then
		ok "Node.js $(node --version) e npm $(npm --version) já instalados"
		return 0
	fi
	if [[ "$CHECK_ONLY" -eq 1 ]]; then
		warn "Node.js/npm não encontrados"
		return 1
	fi

	case "$(detect_os)" in
		darwin)
			if command -v brew >/dev/null 2>&1; then
				info "Instalando Node.js via Homebrew..."
				brew install node
				ok "Node.js instalado"
				return 0
			fi
			;;
		linux)
			local pm
			pm="$(detect_linux_pm)"
			case "$pm" in
				apt)
					info "Instalando Node.js 20.x via NodeSource..."
					curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
					sudo apt-get install -y nodejs
					ok "Node.js instalado"
					return 0
					;;
				dnf)
					info "Instalando Node.js via dnf..."
					sudo dnf install -y nodejs npm
					ok "Node.js instalado"
					return 0
					;;
				pacman)
					info "Instalando Node.js via pacman..."
					sudo pacman -S --noconfirm nodejs npm
					ok "Node.js instalado"
					return 0
					;;
			esac
			;;
	esac

	warn "Instale Node.js manualmente: https://nodejs.org/"
	return 1
}

pkg_config_exists() {
	command -v pkg-config >/dev/null 2>&1 && pkg-config --exists "$1" 2>/dev/null
}

install_linux_platform_deps() {
	local pm missing=0
	pm="$(detect_linux_pm)"

	if pkg_config_exists "gtk+-3.0" && { pkg_config_exists "webkit2gtk-4.1" || pkg_config_exists "webkit2gtk-4.0"; }; then
		ok "GTK e WebKitGTK já instalados"
		return 0
	fi

	if [[ "$CHECK_ONLY" -eq 1 ]]; then
		warn "Dependências Linux (GTK/WebKitGTK) ausentes"
		return 1
	fi

	info "Instalando dependências nativas do Linux (${pm})..."
	case "$pm" in
		apt)
			sudo apt-get update -qq
			if sudo apt-get install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev 2>/dev/null; then
				ok "Pacotes apt instalados (webkit 4.1)"
			elif sudo apt-get install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.0-dev; then
				ok "Pacotes apt instalados (webkit 4.0)"
			else
				missing=1
			fi
			;;
		dnf)
			if sudo dnf install -y gcc gtk3-devel webkit2gtk4.1-devel 2>/dev/null; then
				ok "Pacotes dnf instalados (webkit 4.1)"
			elif sudo dnf install -y gcc gtk3-devel webkit2gtk4.0-devel; then
				ok "Pacotes dnf instalados (webkit 4.0)"
			else
				missing=1
			fi
			;;
		pacman)
			if sudo pacman -S --noconfirm base-devel gtk3 webkit2gtk-4.1 2>/dev/null; then
				ok "Pacotes pacman instalados (webkit 4.1)"
			elif sudo pacman -S --noconfirm base-devel gtk3 webkit2gtk-4.0; then
				ok "Pacotes pacman instalados (webkit 4.0)"
			else
				missing=1
			fi
			;;
		*)
			warn "Gerenciador de pacotes não reconhecido. Instale manualmente (veja README.md):"
			echo "  Debian/Ubuntu: sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev"
			echo "  Fedora:        sudo dnf install gcc gtk3-devel webkit2gtk4.1-devel"
			echo "  Arch:          sudo pacman -S base-devel gtk3 webkit2gtk-4.1"
			return 1
			;;
	esac

	[[ "$missing" -eq 0 ]] || return 1
}

install_darwin_platform_deps() {
	if xcode-select -p >/dev/null 2>&1; then
		ok "Xcode Command Line Tools instaladas"
		return 0
	fi
	if [[ "$CHECK_ONLY" -eq 1 ]]; then
		warn "Xcode Command Line Tools ausentes"
		return 1
	fi
	warn "Xcode Command Line Tools são necessárias. Execute:"
	echo "  xcode-select --install"
	echo "  (uma janela do sistema será aberta — conclua a instalação e execute este script novamente)"
	return 1
}

install_platform_deps() {
	case "$(detect_os)" in
		darwin) install_darwin_platform_deps ;;
		linux)  install_linux_platform_deps ;;
	esac
}

install_wails() {
	if command -v wails >/dev/null 2>&1; then
		ok "Wails CLI já instalada ($(wails version 2>/dev/null | head -1 | tr -d '\n' || echo disponível))"
		return 0
	fi
	if [[ "$CHECK_ONLY" -eq 1 ]]; then
		warn "Wails CLI não encontrada"
		return 1
	fi
	if ! command -v go >/dev/null 2>&1; then
		warn "Go não disponível — não é possível instalar Wails"
		return 1
	fi
	info "Instalando Wails CLI..."
	go install "$WAILS_PKG"
	ensure_path_block
	if command -v wails >/dev/null 2>&1; then
		ok "Wails CLI instalada"
	else
		warn "Wails instalada em \$(go env GOPATH)/bin — adicione ao PATH e reinicie o terminal"
		return 1
	fi
}

run_doctor() {
	local root=""
	if root="$(find_project_root 2>/dev/null)"; then
		:
	elif is_project_root "$DEFAULT_INSTALL_DIR"; then
		root="$DEFAULT_INSTALL_DIR"
	fi

	if [[ -z "$root" ]]; then
		return 0
	fi
	if ! command -v go >/dev/null 2>&1; then
		return 0
	fi
	info "Executando verificação do projeto (project doctor)..."
	(cd "$root" && go run ./cmd/pm project doctor)
}

build_and_install_apps() {
	local root
	if ! root="$(ensure_project_root)"; then
		die "Não foi possível obter o código-fonte. Verifique conexão de rede ou clone manualmente:
  git clone $REPO_URL $DEFAULT_INSTALL_DIR"
	fi

	if ! command -v go >/dev/null 2>&1; then
		die "Go não disponível — não é possível compilar"
	fi

	info "Instalando CLI (pm)..."
	(cd "$root" && go install ./cmd/pm)
	# Refresh PATH so the freshly-installed pm binary is available in this session.
	ensure_path_block
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

	ok "CLI e app desktop instalados com sucesso!"
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
	local root=""
	if root="$(find_project_root 2>/dev/null)"; then
		:
	elif is_project_root "$DEFAULT_INSTALL_DIR"; then
		root="$DEFAULT_INSTALL_DIR"
	fi

	echo ""
	info "Próximos passos:"

	if [[ "$CHECK_ONLY" -eq 1 ]]; then
		echo "  Execute sem --check-only para instalar dependências, compilar e instalar CLI + desktop"
		return
	fi

	echo "  pm --version          # CLI (reinicie o terminal ou: source ~/.zshrc)"
	echo "  pm list               # listar pontos do dia"
	if [[ -n "$root" ]]; then
		echo "  Código-fonte em: $root"
	fi

	case "$(detect_os)" in
		darwin) echo "  App desktop: ~/Applications/PM Planner.app (Launchpad)" ;;
		linux)  echo "  App desktop: menu de aplicativos (PM Planner)" ;;
	esac
}

main() {
	apply_env_flags
	parse_args "$@"

	echo ""
	info "PM Planner — setup ($(detect_os))"
	echo ""

	local failed=0
	install_go       || failed=1
	ensure_path_block
	install_node     || failed=1
	install_platform_deps || failed=1
	install_wails    || failed=1

	if [[ "$failed" -ne 0 && "$CHECK_ONLY" -eq 1 ]]; then
		die "Dependências ausentes detectadas (use sem --check-only para instalar)"
	fi

	if [[ "$failed" -ne 0 && "$CHECK_ONLY" -eq 0 ]]; then
		warn "Algumas dependências precisam de atenção manual (veja mensagens acima)"
	fi

	if [[ "$CHECK_ONLY" -eq 0 ]]; then
		build_and_install_apps
	fi

	run_doctor || true

	if [[ "$NO_AUTOUPDATE" -eq 1 ]]; then
		uninstall_autoupdate
	elif [[ "$CHECK_ONLY" -eq 0 ]]; then
		local au_root=""
		au_root="$(find_project_root 2>/dev/null)" || true
		if [[ -z "$au_root" ]] && is_project_root "$DEFAULT_INSTALL_DIR"; then
			au_root="$DEFAULT_INSTALL_DIR"
		fi
		if [[ -n "$au_root" ]]; then
			install_autoupdate "$au_root"
		else
			warn "Não foi possível localizar o diretório do projeto para registrar a auto-atualização"
		fi
	fi

	print_next_steps
	echo ""
	ok "Setup concluído."
}

main "$@"
