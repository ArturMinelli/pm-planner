# PM Planner — wrapper de auto-atualização silenciosa (Windows).
#
# Chamado pelo Agendador de Tarefas do Windows a cada logon.
# Roda update.ps1 no máximo uma vez por dia. Toda a saída é gravada em log;
# o script sempre termina com código 0 para nunca bloquear o logon.
#
# Arquivo de estado : %LOCALAPPDATA%\pm\autoupdate-last-run.txt
# Arquivo de log    : %LOCALAPPDATA%\pm\autoupdate.log

$ErrorActionPreference = "SilentlyContinue"

$ScriptDir   = Split-Path -Parent $MyInvocation.MyCommand.Path
$UpdateScript = Join-Path $ScriptDir "update.ps1"
$StateDir    = Join-Path $env:LOCALAPPDATA "pm"
$StateFile   = Join-Path $StateDir "autoupdate-last-run.txt"
$LogDir      = $StateDir
$LogFile     = Join-Path $LogDir "autoupdate.log"
$Today       = Get-Date -Format "yyyy-MM-dd"

if (-not (Test-Path $StateDir)) {
    New-Item -ItemType Directory -Path $StateDir -Force | Out-Null
}

# Once-per-day guard: sai silenciosamente se já rodou hoje.
if ((Test-Path $StateFile)) {
    $lastRun = (Get-Content -Raw $StateFile -ErrorAction SilentlyContinue).Trim()
    if ($lastRun -eq $Today) {
        exit 0
    }
}

function Write-Log([string]$Message) {
    Add-Content -Path $LogFile -Value $Message -ErrorAction SilentlyContinue
}

$timestamp = Get-Date -Format "HH:mm:ss"
Write-Log ""
Write-Log "=== PM Planner auto-update: $Today $timestamp ==="

if (-not (Test-Path $UpdateScript)) {
    Write-Log "ERRO: update.ps1 não encontrado em $UpdateScript"
    exit 0
}

try {
    $output = & powershell.exe `
        -ExecutionPolicy Bypass `
        -NonInteractive `
        -WindowStyle Hidden `
        -File "$UpdateScript" 2>&1

    $exitCode = $LASTEXITCODE

    ($output | Out-String).Trim() | ForEach-Object {
        if ($_ -ne "") { Write-Log $_ }
    }

    if ($exitCode -eq 0) {
        Set-Content -Path $StateFile -Value $Today -NoNewline -ErrorAction SilentlyContinue
        Write-Log "=== Atualização concluída com sucesso ==="
    } else {
        Write-Log "=== Atualização falhou (código $exitCode) — instalação existente mantida ==="
    }
} catch {
    Write-Log "ERRO inesperado: $_"
}

exit 0
