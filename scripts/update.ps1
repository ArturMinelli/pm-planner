# PM Planner — atualiza instalação existente (Windows).
# Obtém a versão mais recente do código-fonte e recompila/reinstala CLI e desktop.

[CmdletBinding()]
param(
    [switch]$Help
)

$ErrorActionPreference = "Stop"

$RepoUrl = "https://github.com/ArturMinelli/pm-planner.git"
$RepoTarballUrl = "https://github.com/ArturMinelli/pm-planner/archive/refs/heads/main.tar.gz"
$DefaultInstallDir = Join-Path $env:USERPROFILE "pm-planner"
$SetupUrl = "https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.ps1"

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
  -Help  Exibe esta ajuda

Exemplos:
  irm https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/update.ps1 | iex
  cd $DefaultInstallDir; .\scripts\update.ps1
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
Invoke-RebuildAndInstall $root

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

Show-NextSteps $root
Write-Host ""
Write-Ok "Update concluído."
