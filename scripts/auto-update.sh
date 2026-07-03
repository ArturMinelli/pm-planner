#!/usr/bin/env bash
# PM Planner — wrapper de auto-atualização silenciosa (Linux e macOS).
#
# Chamado pelo scheduler do OS (entrada XDG autostart no Linux, LaunchAgent no macOS)
# a cada login. Roda update.sh no máximo uma vez por dia. Toda a saída é gravada em
# log; o script sempre termina com código 0 para nunca interromper o login.
#
# Arquivo de estado : ${XDG_DATA_HOME:-~/.local/share}/pm/autoupdate-last-run
# Arquivo de log    : ${XDG_CACHE_HOME:-~/.cache}/pm/autoupdate.log

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UPDATE_SCRIPT="$SCRIPT_DIR/update.sh"

STATE_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/pm"
STATE_FILE="$STATE_DIR/autoupdate-last-run"
LOG_DIR="${XDG_CACHE_HOME:-$HOME/.cache}/pm"
LOG_FILE="$LOG_DIR/autoupdate.log"
TODAY="$(date +%Y-%m-%d)"

mkdir -p "$STATE_DIR" "$LOG_DIR" 2>/dev/null || true

# Once-per-day guard: sai silenciosamente se já rodou hoje.
if [[ -f "$STATE_FILE" ]] && [[ "$(cat "$STATE_FILE" 2>/dev/null)" == "$TODAY" ]]; then
    exit 0
fi

# Roda update.sh capturando toda a saída no log.
{
    printf '\n=== PM Planner auto-update: %s %s ===\n' "$TODAY" "$(date +%T)"

    if [[ ! -f "$UPDATE_SCRIPT" ]]; then
        printf 'ERRO: update.sh não encontrado em %s\n' "$UPDATE_SCRIPT"
    elif [[ ! -x "$UPDATE_SCRIPT" ]]; then
        chmod +x "$UPDATE_SCRIPT" 2>/dev/null || true
        if [[ ! -x "$UPDATE_SCRIPT" ]]; then
            printf 'ERRO: update.sh não é executável em %s\n' "$UPDATE_SCRIPT"
        fi
    fi

    if [[ -x "$UPDATE_SCRIPT" ]]; then
        if "$UPDATE_SCRIPT"; then
            printf '%s' "$TODAY" > "$STATE_FILE"
            printf '=== Atualização concluída com sucesso ===\n'
        else
            printf '=== Atualização falhou — instalação existente mantida ===\n'
        fi
    fi
} >> "$LOG_FILE" 2>&1 || true

exit 0
