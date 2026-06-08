# PM Planner — script de setup para Windows.
# Detecta dependências ausentes, instala via winget quando possível e
# opcionalmente compila e instala o app desktop (-Build).

[CmdletBinding()]
param(
    [switch]$Build,
    [switch]$CheckOnly,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

$GoMinVersion = [version]"1.23.0"
$WailsPkg = "github.com/wailsapp/wails/v2/cmd/wails@latest"
$RepoUrl = "https://github.com/ArturMinelli/pm-planner.git"
$PathMarker = "# pm-planner setup"

function Write-Info($Message)  { Write-Host "→ $Message" -ForegroundColor Blue }
function Write-Ok($Message)    { Write-Host "✓ $Message" -ForegroundColor Green }
function Write-Warn($Message)  { Write-Host "! $Message" -ForegroundColor Yellow }
function Write-Fail($Message)  { Write-Host "✗ $Message" -ForegroundColor Red }

function Show-Usage {
    @"
Uso: setup.ps1 [opções]

Instala dependências do PM Planner (Go, Node.js, GCC, WebView2, Wails).

Opções:
  -Build       Compila e instala o app desktop (requer estar no repositório)
  -CheckOnly   Apenas verifica dependências, sem instalar
  -Help        Exibe esta ajuda

Exemplos:
  irm https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.ps1 | iex
  git clone https://github.com/ArturMinelli/pm-planner.git; cd pm-planner
  .\scripts\setup.ps1 -Build

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

function Invoke-BuildAndInstall {
    $root = Find-ProjectRoot
    if (-not $root) {
        throw @"
Opção -Build requer o repositório pm-planner. Clone primeiro:
  git clone $RepoUrl
  cd pm-planner
  .\scripts\setup.ps1 -Build
"@
    }

    Write-Info "Compilando app desktop..."
    Push-Location $root
    try {
        go run ./cmd/pm project build desktop
        Write-Info "Instalando app desktop..."
        go run ./cmd/pm project install desktop --skip-build
    } finally {
        Pop-Location
    }
    Write-Ok "App desktop compilado e instalado com sucesso!"
}

function Show-NextSteps {
    $root = Find-ProjectRoot
    if (-not $root) {
        Write-Host ""
        Write-Info "Próximos passos:"
        Write-Host "  git clone $RepoUrl"
        Write-Host "  cd pm-planner"
        Write-Host "  .\scripts\setup.ps1 -Build"
        Write-Host ""
        Write-Host "  Ou verifique o ambiente após clonar:"
        Write-Host "  go run ./cmd/pm project doctor"
        return
    }
    if (-not $Build) {
        Write-Host ""
        Write-Info "Próximos passos:"
        Write-Host "  .\scripts\setup.ps1 -Build    # compilar e instalar o app desktop"
        Write-Host "  go run ./cmd/pm project doctor"
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

try {
    Invoke-Doctor
} catch {
    Write-Warn "Verificação do projeto falhou: $_"
}

if ($Build) {
    Invoke-BuildAndInstall
}

Show-NextSteps
Write-Host ""
Write-Ok "Setup concluído."
