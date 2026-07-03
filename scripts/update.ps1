# PM Planner — atualiza instalação existente (Windows).
# Obtém a versão mais recente do código-fonte e recompila/reinstala CLI e desktop.

[CmdletBinding()]
param(
    [switch]$Autoupdate,
    [switch]$NoAutoupdate,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

$RepoUrl = "https://github.com/ArturMinelli/pm-planner.git"
$RepoTarballUrl = "https://github.com/ArturMinelli/pm-planner/archive/refs/heads/main.tar.gz"
$DefaultInstallDir = Join-Path $env:USERPROFILE "pm-planner"
$SetupUrl = "https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.ps1"

$script:DaemonWasRunning = $false

function Write-Info($Message)  { Write-Host "→ $Message" -ForegroundColor Blue }
function Write-Ok($Message)    { Write-Host "✓ $Message" -ForegroundColor Green }
function Write-Warn($Message)  { Write-Host "! $Message" -ForegroundColor Yellow }
function Write-Fail($Message)  { Write-Host "✗ $Message" -ForegroundColor Red }

function Show-Usage {
    @"
Uso: update.ps1 [opções]

Atualiza uma instalação existente do PM Planner: baixa código novo e
recompila/reinstala a CLI (pm) e o app desktop.

Opções:
  -NoAutoupdate  Remove o registro de auto-atualização do Agendador de Tarefas (padrão: registrar)
  -Help          Exibe esta ajuda

Exemplos:
  irm https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/update.ps1 | iex
  cd $DefaultInstallDir; .\scripts\update.ps1
  cd $DefaultInstallDir; .\scripts\update.ps1 -NoAutoupdate
"@
}

if ($Help) {
    Show-Usage
    exit 0
}

function Test-ProjectRoot([string]$Dir) {
    $goMod = Join-Path $Dir "go.mod"
    $wailsJson = Join-Path $Dir "desktop\wails.json"
    if (-not (Test-Path $goMod) -or -not (Test-Path $wailsJson)) { return $false }
    return (Select-String -Path $goMod -Pattern '^module pm-cli$' -Quiet)
}

function Find-ProjectRoot {
    $dir = Get-Location
    while ($dir) {
        if (Test-ProjectRoot $dir.FullName) { return $dir.FullName }
        $parent = $dir.Parent
        if (-not $parent -or $parent.FullName -eq $dir.FullName) { break }
        $dir = $parent
    }

    $scriptDir = Split-Path -Parent $MyInvocation.PSCommandPath
    $candidate = Split-Path -Parent $scriptDir
    if (Test-ProjectRoot $candidate) { return $candidate }

    return $null
}

function Find-ExistingInstall {
    $root = Find-ProjectRoot
    if ($root) { return $root }
    if (Test-ProjectRoot $DefaultInstallDir) { return $DefaultInstallDir }
    return $null
}

function Get-ProjectTarball {
    param([string]$Target)

    $tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("pm-planner-" + [guid]::NewGuid().ToString())
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        Write-Info "Baixando código-fonte (tarball)..."
        $archivePath = Join-Path $tempDir "pm-planner.tar.gz"
        Invoke-WebRequest -Uri $RepoTarballUrl -OutFile $archivePath -UseBasicParsing
        tar -xzf $archivePath -C $tempDir
        if (Test-Path $Target) {
            Remove-Item -Recurse -Force $Target
        }
        Move-Item (Join-Path $tempDir "pm-planner-main") $Target
        Write-Ok "Código-fonte atualizado em $Target"
    } finally {
        if (Test-Path $tempDir) {
            Remove-Item -Recurse -Force $tempDir
        }
    }
}

function Update-GitRepo {
    param([string]$Root)

    $branch = "main"
    Write-Info "Atualizando repositório git em $Root..."
    Push-Location $Root
    try {
        if (-not (Test-Path ".git")) {
            throw "Diretório $Root não é um repositório git válido"
        }

        $hasOrigin = $false
        try {
            git remote get-url origin | Out-Null
            $hasOrigin = $true
        } catch {
            $hasOrigin = $false
        }

        if ($hasOrigin) {
            git fetch origin --prune
        } else {
            Write-Warn "Remote origin ausente — adicionando $RepoUrl"
            git remote add origin $RepoUrl
            git fetch origin --prune
        }

        git show-ref --verify --quiet "refs/heads/$branch" 2>$null
        $localBranchExists = $LASTEXITCODE -eq 0
        git show-ref --verify --quiet "refs/remotes/origin/$branch" 2>$null
        $remoteBranchExists = $LASTEXITCODE -eq 0

        if ($localBranchExists) {
            git checkout $branch
        } elseif ($remoteBranchExists) {
            git checkout -B $branch "origin/$branch"
        } else {
            throw "Branch $branch não encontrada no repositório"
        }

        git diff --quiet
        if ($LASTEXITCODE -ne 0) {
            throw @"
Há alterações locais em $Root. Faça commit, stash ou descarte antes de atualizar:
  git -C $Root status
  git -C $Root stash -u
"@
        }
        git diff --cached --quiet
        if ($LASTEXITCODE -ne 0) {
            throw @"
Há alterações locais em $Root. Faça commit, stash ou descarte antes de atualizar:
  git -C $Root status
  git -C $Root stash -u
"@
        }

        git pull --ff-only origin $branch
        if ($LASTEXITCODE -ne 0) {
            throw "git pull falhou em $Root. Resolva conflitos manualmente e execute o update novamente."
        }
    } finally {
        Pop-Location
    }
    Write-Ok "Repositório git atualizado (branch $branch)"
}

function Update-Source {
    param([string]$Root)

    if (Test-Path (Join-Path $Root ".git")) {
        Update-GitRepo $Root
        return
    }

    Write-Warn "Instalação sem git detectada — substituindo por tarball mais recente"
    Get-ProjectTarball $Root
}

function Get-ConfigDir {
    return Join-Path $env:APPDATA "pm"
}

function Get-DaemonPidFile {
    return Join-Path (Get-ConfigDir) "reminder-daemon.pid"
}

function Test-DaemonRunning {
    $pidFile = Get-DaemonPidFile
    if (-not (Test-Path $pidFile)) {
        return $false
    }

    $pidText = (Get-Content -Raw $pidFile).Trim()
    if ($pidText -notmatch '^\d+$') {
        return $false
    }

    $processId = [int]$pidText
    return $null -ne (Get-Process -Id $processId -ErrorAction SilentlyContinue)
}

function Get-InstalledDesktopBinary {
    return Join-Path $env:LOCALAPPDATA "Programs\PM Planner\pm-desktop.exe"
}

function Stop-ReminderDaemon {
    $pidFile = Get-DaemonPidFile
    if (-not (Test-Path $pidFile)) {
        return
    }

    $pidText = (Get-Content -Raw $pidFile).Trim()
    if ($pidText -match '^\d+$') {
        $processId = [int]$pidText
        $process = Get-Process -Id $processId -ErrorAction SilentlyContinue
        if ($null -ne $process) {
            Write-Info "Encerrando daemon de lembretes (PID $processId)..."
            Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
            Write-Ok "Daemon encerrado"
        }
    }

    Remove-Item -Force $pidFile -ErrorAction SilentlyContinue
}

function Stop-DesktopApp {
    Write-Info "Encerrando app desktop PM Planner..."
    $processes = Get-Process -Name "pm-desktop" -ErrorAction SilentlyContinue
    if ($null -eq $processes) {
        Write-Ok "App desktop encerrado"
        return
    }

    $processes | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 1

    $remaining = Get-Process -Name "pm-desktop" -ErrorAction SilentlyContinue
    if ($null -ne $remaining) {
        throw "Não foi possível encerrar o app desktop. Feche o PM Planner manualmente e execute o update novamente."
    }

    Write-Ok "App desktop encerrado"
}

function Stop-RunningProcesses {
    if (Test-DaemonRunning) {
        $script:DaemonWasRunning = $true
    }
    Stop-ReminderDaemon
    Stop-DesktopApp
}

function Start-ReminderDaemon {
    $binary = Get-InstalledDesktopBinary
    if (-not (Test-Path $binary)) {
        Write-Warn "Binário desktop não encontrado em $binary — daemon não reiniciado."
        return
    }

    Write-Info "Reiniciando daemon de lembretes..."
    Start-Process -FilePath $binary -ArgumentList "--daemon" -WindowStyle Hidden
    Write-Ok "Daemon de lembretes reiniciado"
}

function Restart-DaemonIfNeeded {
    if (-not $script:DaemonWasRunning) {
        return
    }
    Start-ReminderDaemon
}

function Refresh-Path {
    $machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $env:Path = "$machinePath;$userPath"
}

function Ensure-SessionPath {
    Refresh-Path
    if (Get-Command go -ErrorAction SilentlyContinue) {
        $goBin = Join-Path (go env GOPATH) "bin"
        if ($env:Path -notlike "*$goBin*") {
            $env:Path = "$env:Path;$goBin"
        }
    }
}

function Invoke-RebuildAndInstall {
    param([string]$Root)

    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw @"
Go não encontrado no PATH. Execute o setup novamente:
  irm $SetupUrl | iex
"@
    }

    Push-Location $Root
    try {
        Write-Info "Instalando CLI (pm)..."
        go install ./cmd/pm
        Ensure-SessionPath
        Write-Ok "CLI instalada em $(Join-Path (go env GOPATH) 'bin\pm')"

        Write-Info "Compilando app desktop..."
        go run ./cmd/pm project build desktop
        Write-Info "Instalando app desktop..."
        go run ./cmd/pm project install desktop --skip-build
    } finally {
        Pop-Location
    }
    Write-Ok "CLI e app desktop reinstalados com sucesso!"
}

function Show-NextSteps {
    param([string]$Root)

    Write-Host ""
    Write-Info "Atualização concluída."
    Write-Host "  pm --version          # CLI atualizada"
    if ($Root) {
        Write-Host "  Código-fonte em: $Root"
    }
    Write-Host "  App desktop: Menu Iniciar (PM Planner)"
}

$AutoupdateTaskName = "PM Planner Auto Update"

function Install-Autoupdate {
    param([string]$Root)

    $wrapper = Join-Path $Root "scripts\auto-update.ps1"
    if (-not (Test-Path $wrapper)) {
        Write-Warn "auto-update.ps1 não encontrado em $wrapper"
        return
    }

    $action   = New-ScheduledTaskAction `
                    -Execute "powershell.exe" `
                    -Argument "-WindowStyle Hidden -ExecutionPolicy Bypass -NonInteractive -File `"$wrapper`""
    $trigger  = New-ScheduledTaskTrigger -AtLogOn
    $settings = New-ScheduledTaskSettingsSet `
                    -ExecutionTimeLimit (New-TimeSpan -Hours 1) `
                    -StartWhenAvailable

    Register-ScheduledTask `
        -TaskName $AutoupdateTaskName `
        -Action $action `
        -Trigger $trigger `
        -Settings $settings `
        -RunLevel Limited `
        -Force | Out-Null

    Write-Ok "Auto-atualização registrada no Agendador de Tarefas ('$AutoupdateTaskName')"
    $logFile = Join-Path $env:LOCALAPPDATA "pm\autoupdate.log"
    Write-Info "Log de auto-atualização: $logFile"
}

function Uninstall-Autoupdate {
    $task = Get-ScheduledTask -TaskName $AutoupdateTaskName -ErrorAction SilentlyContinue
    if ($null -ne $task) {
        Unregister-ScheduledTask -TaskName $AutoupdateTaskName -Confirm:$false
        Write-Ok "Auto-atualização removida do Agendador de Tarefas"
    } else {
        Write-Warn "Auto-atualização não estava registrada"
    }
}

# --- main ---
Write-Host ""
Write-Info "PM Planner — update (windows)"
Write-Host ""

$root = Find-ExistingInstall
if (-not $root) {
    Write-Fail @"
Instalação não encontrada. Execute o setup primeiro:
  irm $SetupUrl | iex

Ou, se já clonou manualmente, execute este script de dentro de pm-planner/.
"@
    exit 1
}

Write-Ok "Instalação encontrada em $root"
Ensure-SessionPath

Update-Source $root
Stop-RunningProcesses
Invoke-RebuildAndInstall $root
Restart-DaemonIfNeeded

if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Info "Verificando ambiente (project doctor)..."
    try {
        Push-Location $root
        go run ./cmd/pm project doctor
    } catch {
        Write-Warn "Verificação do projeto reportou problemas: $_"
    } finally {
        Pop-Location
    }
}

if ($NoAutoupdate) {
    Uninstall-Autoupdate
} else {
    Install-Autoupdate $root
}

Show-NextSteps $root
Write-Host ""
Write-Ok "Update concluído."
