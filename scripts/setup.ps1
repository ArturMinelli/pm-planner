# PM Planner — script de setup para Windows.
# Instala dependências, obtém o código-fonte e compila/instala a CLI e o app desktop.

[CmdletBinding()]
param(
    [switch]$Build,        # obsoleto: compilação é o comportamento padrão
    [switch]$CheckOnly,
    [switch]$Autoupdate,
    [switch]$NoAutoupdate,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

$GoMinVersion = [version]"1.23.0"
$WailsPkg = "github.com/wailsapp/wails/v2/cmd/wails@latest"
$RepoUrl = "https://github.com/ArturMinelli/pm-planner.git"
$RepoTarballUrl = "https://github.com/ArturMinelli/pm-planner/archive/refs/heads/main.tar.gz"
$DefaultInstallDir = Join-Path $env:USERPROFILE "pm-planner"
$PathMarker = "# pm-planner setup"

function Write-Info($Message)  { Write-Host "→ $Message" -ForegroundColor Blue }
function Write-Ok($Message)    { Write-Host "✓ $Message" -ForegroundColor Green }
function Write-Warn($Message)  { Write-Host "! $Message" -ForegroundColor Yellow }
function Write-Fail($Message)  { Write-Host "✗ $Message" -ForegroundColor Red }

function Show-Usage {
    @"
Uso: setup.ps1 [opções]

Instala dependências, compila e instala a CLI (pm) e o app desktop do PM Planner.

Opções:
  -CheckOnly     Apenas verifica dependências, sem instalar nem compilar
  -Autoupdate    Registra auto-atualização diária via Agendador de Tarefas após o setup
  -NoAutoupdate  Remove o registro de auto-atualização do Agendador de Tarefas
  -Help          Exibe esta ajuda

Exemplos:
  irm https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.ps1 | iex
  git clone https://github.com/ArturMinelli/pm-planner.git; cd pm-planner
  .\scripts\setup.ps1
  .\scripts\setup.ps1 -Autoupdate

Nota: revise o script antes de executar com irm | iex.
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
        Write-Ok "Código-fonte extraído em $Target"
    } finally {
        if (Test-Path $tempDir) {
            Remove-Item -Recurse -Force $tempDir
        }
    }
}

function Clone-ProjectRepo {
    param([string]$Target)

    if (Test-Path $Target) {
        Remove-Item -Recurse -Force $Target
    }
    Write-Info "Clonando repositório..."
    git clone $RepoUrl $Target
    Write-Ok "Repositório clonado em $Target"
}

function Ensure-ProjectRoot {
    $root = Find-ProjectRoot
    if ($root) { return $root }

    if (Test-ProjectRoot $DefaultInstallDir) {
        Write-Ok "Repositório encontrado em $DefaultInstallDir"
        return $DefaultInstallDir
    }

    if ($CheckOnly) { return $null }

    if (Test-Path $DefaultInstallDir) {
        Write-Warn "Removendo instalação inválida em $DefaultInstallDir"
        Remove-Item -Recurse -Force $DefaultInstallDir
    }

    Write-Info "Obtendo código-fonte em $DefaultInstallDir..."
    if (Get-Command git -ErrorAction SilentlyContinue) {
        Clone-ProjectRepo $DefaultInstallDir
        return $DefaultInstallDir
    }

    Get-ProjectTarball $DefaultInstallDir
    return $DefaultInstallDir
}

function Test-Winget {
    return [bool](Get-Command winget -ErrorAction SilentlyContinue)
}

function Install-WingetPackage([string]$Id, [string]$Label) {
    if (-not (Test-Winget)) {
        Write-Warn "winget não encontrado — instale $Label manualmente"
        return $false
    }
    Write-Info "Instalando $Label via winget ($Id)..."
    winget install --id $Id -e --accept-source-agreements --accept-package-agreements
    Write-Ok "$Label instalado"
    return $true
}

function Refresh-Path {
  $machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $env:Path = "$machinePath;$userPath"
}

function Ensure-GoBinPath {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) { return }

    $goBin = Join-Path (go env GOPATH) "bin"
    if ($env:Path -notlike "*$goBin*") {
        $env:Path = "$env:Path;$goBin"
    }

    if ($CheckOnly) { return }

    $profilePath = $PROFILE
    if (-not $profilePath) { return }

    $profileDir = Split-Path -Parent $profilePath
    if (-not (Test-Path $profileDir)) {
        New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
    }

    if (Test-Path $profilePath) {
        $content = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
        if ($content -and $content.Contains($PathMarker)) { return }
    }

    @"

$PathMarker
`$goBin = Join-Path (go env GOPATH) "bin"
if (Test-Path `$goBin) { `$env:Path = "`$env:Path;`$goBin" }
"@ | Add-Content -Path $profilePath

    Write-Warn "PATH atualizado em $profilePath — reinicie o PowerShell"
}

function Test-GoVersion {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) { return $false }
    $raw = (go version) -replace '^go version go', '' -replace ' .+$', ''
    try {
        $ver = [version]$raw
        return $ver -ge $GoMinVersion
    } catch {
        return $false
    }
}

function Install-Go {
    if (Test-GoVersion) {
        Write-Ok "Go $(go version) já instalado"
        return $true
    }
    if ($CheckOnly) {
        Write-Warn "Go 1.23+ não encontrado"
        return $false
    }
    return (Install-WingetPackage "GoLang.Go" "Go")
}

function Test-Node {
    return [bool](Get-Command node -ErrorAction SilentlyContinue) -and
           [bool](Get-Command npm -ErrorAction SilentlyContinue)
}

function Install-Node {
    if (Test-Node) {
        Write-Ok "Node.js $(node --version) e npm $(npm --version) já instalados"
        return $true
    }
    if ($CheckOnly) {
        Write-Warn "Node.js/npm não encontrados"
        return $false
    }
    return (Install-WingetPackage "OpenJS.NodeJS.LTS" "Node.js LTS")
}

function Test-Gcc {
    return [bool](Get-Command gcc -ErrorAction SilentlyContinue)
}

function Install-Gcc {
    if (Test-Gcc) {
        Write-Ok "GCC já disponível ($(gcc --version | Select-Object -First 1))"
        return $true
    }
    if ($CheckOnly) {
        Write-Warn "GCC (CGO) não encontrado"
        return $false
    }
    if (Install-WingetPackage "BrechtSanders.WinLibs.POSIX.UCRT" "WinLibs GCC") {
        Refresh-Path
        if (Test-Gcc) { return $true }
    }
    Write-Warn "Instale GCC manualmente: https://jmeubank.github.io/tdm-gcc/"
    return $false
}

function Test-WebView2 {
    $paths = @(
        "HKLM:\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7CE4A}",
        "HKLM:\SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7CE4A}"
    )
    foreach ($p in $paths) {
        if (Test-Path $p) { return $true }
    }
    return $false
}

function Install-WebView2 {
    if (Test-WebView2) {
        Write-Ok "WebView2 Runtime já instalado"
        return $true
    }
    if ($CheckOnly) {
        Write-Warn "WebView2 Runtime não detectado"
        return $false
    }
    if (Install-WingetPackage "Microsoft.EdgeWebView2Runtime" "WebView2 Runtime") {
        return $true
    }
    Write-Warn "Baixe o WebView2 Runtime: https://developer.microsoft.com/microsoft-edge/webview2"
    return $false
}

function Install-Wails {
    if (Get-Command wails -ErrorAction SilentlyContinue) {
        $ver = wails version 2>$null | Select-Object -First 1
        Write-Ok "Wails CLI já instalada ($ver)"
        return $true
    }
    if ($CheckOnly) {
        Write-Warn "Wails CLI não encontrada"
        return $false
    }
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Warn "Go não disponível — não é possível instalar Wails"
        return $false
    }
    Write-Info "Instalando Wails CLI..."
    go install $WailsPkg
    Ensure-GoBinPath
    Refresh-Path
    if (Get-Command wails -ErrorAction SilentlyContinue) {
        Write-Ok "Wails CLI instalada"
        return $true
    }
    Write-Warn "Wails instalada em $(go env GOPATH)\bin — adicione ao PATH e reinicie o terminal"
    return $false
}

function Invoke-Doctor {
    $root = Find-ProjectRoot
    if (-not $root -and (Test-ProjectRoot $DefaultInstallDir)) {
        $root = $DefaultInstallDir
    }
    if (-not $root) { return }
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) { return }
    Write-Info "Executando verificação do projeto (project doctor)..."
    Push-Location $root
    try {
        go run ./cmd/pm project doctor
    } finally {
        Pop-Location
    }
}

function Invoke-BuildAndInstallApps {
    $root = Ensure-ProjectRoot
    if (-not $root) {
        throw @"
Não foi possível obter o código-fonte. Verifique conexão de rede ou clone manualmente:
  git clone $RepoUrl $DefaultInstallDir
"@
    }

    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "Go não disponível — não é possível compilar"
    }

    Push-Location $root
    try {
        Write-Info "Instalando CLI (pm)..."
        go install ./cmd/pm
        Ensure-GoBinPath
        Refresh-Path
        Write-Ok "CLI instalada em $(Join-Path (go env GOPATH) 'bin\pm')"

        Write-Info "Compilando app desktop..."
        go run ./cmd/pm project build desktop
        Write-Info "Instalando app desktop..."
        go run ./cmd/pm project install desktop --skip-build
    } finally {
        Pop-Location
    }
    Write-Ok "CLI e app desktop instalados com sucesso!"
}

function Show-NextSteps {
    $root = Find-ProjectRoot
    if (-not $root -and (Test-ProjectRoot $DefaultInstallDir)) {
        $root = $DefaultInstallDir
    }

    Write-Host ""
    Write-Info "Próximos passos:"

    if ($CheckOnly) {
        Write-Host "  Execute sem -CheckOnly para instalar dependências, compilar e instalar CLI + desktop"
        return
    }

    Write-Host "  pm --version          # CLI (reinicie o PowerShell)"
    Write-Host "  pm list               # listar pontos do dia"
    if ($root) {
        Write-Host "  Código-fonte em: $root"
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
Write-Info "PM Planner — setup (windows)"
Write-Host ""

$failed = $false
if (-not (Install-Go))       { $failed = $true }
Refresh-Path
Ensure-GoBinPath
if (-not (Install-Node))     { $failed = $true }
if (-not (Install-Gcc))      { $failed = $true }
if (-not (Install-WebView2))  { $failed = $true }
if (-not (Install-Wails))    { $failed = $true }

if ($failed -and $CheckOnly) {
    Write-Fail "Dependências ausentes detectadas (use sem -CheckOnly para instalar)"
    exit 1
}

if ($failed -and -not $CheckOnly) {
    Write-Warn "Algumas dependências precisam de atenção manual (veja mensagens acima)"
}

if (-not $CheckOnly) {
    Invoke-BuildAndInstallApps
}

try {
    Invoke-Doctor
} catch {
    Write-Warn "Verificação do projeto falhou: $_"
}

if ($Autoupdate) {
    $root = Find-ProjectRoot
    if (-not $root -and (Test-ProjectRoot $DefaultInstallDir)) { $root = $DefaultInstallDir }
    if ($root) {
        Install-Autoupdate $root
    } else {
        Write-Warn "Não foi possível localizar o diretório do projeto para registrar a auto-atualização"
    }
} elseif ($NoAutoupdate) {
    Uninstall-Autoupdate
}

Show-NextSteps
Write-Host ""
Write-Ok "Setup concluído."
