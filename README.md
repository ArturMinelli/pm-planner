# PM Planner

O PM Planner é um auxiliar do PontoMais que permite planejar os horários de ponto do dia de trabalho antes de bater o ponto de verdade. Ele vem em duas formas que compartilham a mesma lógica Go e arquivo de configuração:

- **`pm`** — CLI para o terminal
- **`pm-desktop`** — App desktop Wails/React com interface gráfica

---

## Índice

- [Instalação rápida](#instalação-rápida-recomendado)
  - [Script de setup](#script-de-setup)
- [Requisitos](#requisitos)
  - [macOS](#macos)
  - [Linux](#linux)
  - [Windows](#windows)
- [Instalação](#instalação)
  - [Compilar a CLI](#compilar-a-cli)
  - [Compilar o App Desktop](#compilar-o-app-desktop)
  - [Instalar o App Desktop](#instalar-o-app-desktop)
- [Configuração](#configuração)
- [Usando a CLI](#usando-a-cli)
  - [Listar um Dia de Trabalho](#listar-um-dia-de-trabalho)
  - [Versão](#versão)
  - [Atualizar](#atualizar)
- [Usando o App Desktop (GUI)](#usando-o-app-desktop-gui)
  - [Página do Planner](#página-do-planner)
  - [Página de Configurações](#página-de-configurações)
- [Desenvolvimento](#desenvolvimento)
- [Solução de Problemas](#solução-de-problemas)

---

## Instalação rápida (recomendado)

Um único comando instala dependências, obtém o código-fonte, compila a CLI (`pm`) e instala o app desktop nativamente.

| Plataforma | Script | Instalação completa |
| --- | --- | --- |
| **macOS / Linux** | [`scripts/setup.sh`](scripts/setup.sh) | `curl -fsSL https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.sh \| bash` |
| **Windows** | [`scripts/setup.ps1`](scripts/setup.ps1) | `irm https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.ps1 \| iex` |

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.sh | bash
```

**Windows** (revise o script antes de executar):

```powershell
irm https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/setup.ps1 | iex
```

O script faz tudo automaticamente:

1. Instala dependências (Go, Node.js, Wails e bibliotecas nativas)
2. Obtém o código-fonte em `~/pm-planner` (clone via git ou download do tarball, se você não estiver dentro de um clone)
3. Instala a CLI `pm` em `$(go env GOPATH)/bin`
4. Compila e instala o app desktop (Launchpad no macOS, menu de aplicativos no Linux, Menu Iniciar no Windows)

Se você já clonou o repositório, execute o script de dentro de `pm-planner/` — ele usa o clone local em vez de baixar novamente:

```bash
./scripts/setup.sh
```

```powershell
.\scripts\setup.ps1
```

Outras flags úteis:

```bash
./scripts/setup.sh --check-only    # apenas verifica dependências, sem instalar nem compilar
./scripts/setup.sh --help
```

```powershell
.\scripts\setup.ps1 -CheckOnly
.\scripts\setup.ps1 -Help
```

Após o setup, reinicie o terminal e teste:

```bash
pm --version
pm list
```

### Atualizar instalação existente

A atualização é sempre manual — o PM Planner não se atualiza sozinho em segundo plano.

**Pelo app desktop (recomendado):** abra **Configurações → Atualizações**, clique em **Verificar Atualizações** e depois em **Atualizar Agora**. O app fecha, atualiza e reabre sozinho, mostrando o resultado ao voltar.

**Pela CLI:**

```bash
pm update           # verifica e, se houver novidade, atualiza
pm update --check   # apenas verifica
```

**Sem a CLI instalada**, os scripts fazem o mesmo trabalho:

| Plataforma | Comando |
| --- | --- |
| **macOS / Linux** | `curl -fsSL https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/update.sh \| bash` |
| **Windows** | `irm https://raw.githubusercontent.com/ArturMinelli/pm-planner/main/scripts/update.ps1 \| iex` |

Em qualquer um dos caminhos, a atualização:

1. Localiza a instalação existente (`~/.local/share/pm-planner`, `~/Library/Application Support/pm-planner`, `%USERPROFILE%\pm-planner` ou o clone atual)
2. Atualiza o código-fonte (`git pull` se houver repositório git; caso contrário, baixa o tarball mais recente)
3. Encerra o app desktop e o daemon de lembretes (se estiverem em execução) para liberar os binários antes da reinstalação
4. Reinstala a CLI (`pm`) e recompila/reinstala o app desktop
5. Reinicia o daemon de lembretes automaticamente se ele estava ativo antes do update

A saída completa fica em `~/.cache/pm/update.log` (`%LOCALAPPDATA%\pm\update.log` no Windows).

Se a instalação ainda não existir, use o [script de setup](#instalação-rápida-recomendado) primeiro.

> Versões anteriores registravam uma auto-atualização diária no login (entrada XDG autostart, LaunchAgent ou Agendador de Tarefas). Ela foi removida: o app desktop, o `setup` e o `update` apagam esse registro automaticamente na primeira execução.

---

## Requisitos

As seções abaixo descrevem a instalação manual de cada dependência. Para a maioria dos usuários, o [script de setup](#script-de-setup) é o caminho mais simples.

### macOS

#### Go 1.23+

Via Homebrew (recomendado):

```bash
brew install go
go version
```

Ou baixe o instalador oficial em [go.dev/dl](https://go.dev/dl/).

#### Node.js e npm

Necessários apenas para o app desktop:

```bash
brew install node
node --version
npm --version
```

#### Xcode Command Line Tools

Necessárias para CGO:

```bash
xcode-select --install
```

#### Wails CLI

Obrigatória para compilar e instalar o app desktop:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Adicione o binário do Go ao PATH (se ainda não estiver):

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
wails version
```

### Linux

#### Go 1.23+

Instalação via tarball oficial (exemplo para amd64):

```bash
curl -fsSL https://go.dev/dl/go1.23.3.linux-amd64.tar.gz -o /tmp/go.tar.gz
rm -rf ~/.local/go
mkdir -p ~/.local
tar -C ~/.local -xzf /tmp/go.tar.gz
echo 'export PATH="$PATH:$HOME/.local/go/bin"' >> ~/.bashrc
source ~/.bashrc
go version
```

Substitua `linux-amd64` por `linux-arm64` em máquinas ARM.

#### Node.js e npm

Necessários apenas para o app desktop.

Debian/Ubuntu (Node 20 LTS via NodeSource):

```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
```

Fedora:

```bash
sudo dnf install -y nodejs npm
```

Arch:

```bash
sudo pacman -S nodejs npm
```

#### CGO, GTK e WebKitGTK

Debian/Ubuntu:

```bash
sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

Fedora:

```bash
sudo dnf install gcc gtk3-devel webkit2gtk4.1-devel
```

Arch:

```bash
sudo pacman -S base-devel gtk3 webkit2gtk-4.1
```

> Distribuições mais antigas podem precisar de `libwebkit2gtk-4.0-dev` / `webkit2gtk-4.0` com as build tags ajustadas.

#### Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.bashrc
source ~/.bashrc
wails version
```

### Windows

#### Go 1.23+

```powershell
winget install GoLang.Go
go version
```

Ou baixe o instalador em [go.dev/dl](https://go.dev/dl/).

#### Node.js e npm

Necessários apenas para o app desktop:

```powershell
winget install OpenJS.NodeJS.LTS
node --version
npm --version
```

#### GCC (CGO)

O compilador `gcc` deve estar no `PATH`:

```powershell
winget install BrechtSanders.WinLibs.POSIX.UCRT
```

Alternativa: [TDM-GCC ou MinGW-w64](https://jmeubank.github.io/tdm-gcc/).

#### WebView2 Runtime

Já incluído no Windows 10/11 na maioria das instalações. Para versões anteriores, baixe em [developer.microsoft.com/microsoft-edge/webview2](https://developer.microsoft.com/microsoft-edge/webview2):

```powershell
winget install Microsoft.EdgeWebView2Runtime
```

#### Wails CLI

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Adicione `%USERPROFILE%\go\bin` ao PATH do usuário e reinicie o terminal:

```powershell
wails version
```

---

Execute o comando de diagnóstico para verificar o ambiente:

```bash
go run ./cmd/pm project doctor
```

---

## Instalação

### Compilar a CLI

O [script de setup](#script-de-setup) já instala a CLI com `go install ./cmd/pm` (binário `pm` em `$(go env GOPATH)/bin`). Para compilar manualmente:

```bash
go run ./cmd/pm project build cli
```

Isso gera o binário `bin/pm`. Adicione-o ao seu `PATH` ou invoque diretamente como `./bin/pm`.

Para compilar para uma plataforma diferente:

```bash
bin/pm project build cli --goos windows --goarch amd64
bin/pm project build cli --output ./dist/pm
```

### Compilar o App Desktop

```bash
go run ./cmd/pm project build desktop
```

Isso executa `npm ci` e `npm run build` dentro de `desktop/frontend/`, compila o app Wails e gera `bin/pm-desktop` (`bin/pm-desktop.exe` no Windows).

Flags úteis:

```bash
bin/pm project build desktop --skip-frontend   # pula as etapas do npm se o frontend já foi compilado
bin/pm project build desktop --force-go-host   # mantém o GOOS/GOARCH do host em vez da correção automática
```

> No Linux x86\_64 o comando de build define automaticamente `GOARCH=amd64` para evitar a incompatibilidade comum com GTK/WebKit 64 bits. Passe `--force-go-host` para sobrescrever esse comportamento.

---

### Instalar o App Desktop

Um único comando compila (se necessário) e instala o app como aplicativo nativo na plataforma atual:

```bash
go run ./cmd/pm project install desktop
```

| Plataforma | O que é instalado |
| --- | --- |
| **macOS** | Bundle `PM Planner.app` em `~/Applications` (visível no Launchpad e no Finder) |
| **Linux** | Binário, ícone e entrada `.desktop` nos diretórios XDG (`~/.local/bin`, `~/.local/share/applications`) |
| **Windows** | Executável em `%LOCALAPPDATA%\Programs\PM Planner\` e atalho no Menu Iniciar |

Flags úteis:

```bash
bin/pm project install desktop --skip-build          # instala a partir de artefatos já compilados
bin/pm project install desktop --system              # macOS: instala em /Applications (pode exigir sudo)
bin/pm project install desktop --desktop-shortcut    # Windows: também cria atalho na Área de Trabalho
```

**macOS** — após a instalação, abra pelo Launchpad ou Finder. Para instalar no diretório de sistema:

```bash
sudo bin/pm project install desktop --system
```

**Linux** — o app aparece no menu de aplicativos como **PM Planner**. O comando legado `install desktop-menu` continua disponível e faz a mesma instalação.

**Windows** — procure **PM Planner** no Menu Iniciar.

Para apenas compilar sem instalar (útil para desenvolvimento):

```bash
bin/pm project build desktop
bin/pm-desktop          # macOS/Linux
bin/pm-desktop.exe      # Windows
```

---

## Configuração

A CLI e o app desktop compartilham um único arquivo `config.yaml` em um diretório nativo da plataforma:

| Plataforma | Caminho |
| --- | --- |
| Linux | `$XDG_CONFIG_HOME/pm/config.yaml` ou `~/.config/pm/config.yaml` |
| macOS | `~/Library/Application Support/pm/config.yaml` |
| Windows | `%AppData%\pm\config.yaml` |

Exemplo de `config.yaml`:

```yaml
login: "voce@empresa.com"
password: "sua-senha"
locale: "pt-BR"
max_daily_extra_minutes: 180
planner:
  in1: "08:00"
  out1: "12:00"
  in2: "13:30"
  out2: "18:00"
reminders:
  enabled: true
  autostart: true
  lead_minutes: [15, 5]
```

| Campo | Descrição |
| --- | --- |
| `login` / `password` | Credenciais de login do PontoMais (e-mail ou CPF). O campo legado `email` ainda é lido se `login` estiver ausente. |
| `locale` | Idioma fixo do app (`pt-BR`). Gravado automaticamente ao salvar configurações; não é editável pela interface. |
| `max_daily_extra_minutes` | Limite de trabalho extra usado para calcular a saída alternativa de quitação de saldo negativo. Padrão: 180 minutos. |
| `planner.in1/out1/in2/out2` | Horários âncora padrão para os quatro campos de ponto. Usados pelo planner do app desktop como sugestão inicial para o dia. |
| `reminders.*` | Lembretes opt-in do app desktop. Quando ativados, `pm-desktop --daemon` verifica os horários recomendados e dispara notificações nativas. `lead_minutes` aceita vários avisos personalizados entre 1 e 240 minutos. |

Os headers de autenticação são armazenados em cache como `session.json` no mesmo diretório e reutilizados até expirarem. O identificador do colaborador descoberto automaticamente pela API também é guardado nesse arquivo protegido; se ficar inválido, o PM Planner o descobre novamente.

> **Segurança:** não faça commit de `config.yaml` ou `session.json` no controle de versão. Se as credenciais foram expostas, troque-as fora deste repositório.

---

## Usando a CLI

Todos os comandos assumem que `pm` está no `PATH`. Se não estiver, substitua `pm` por `./bin/pm`.

### Listar um Dia de Trabalho

Busca e imprime os dados brutos do cartão de ponto para uma data específica:

```bash
pm list                        # hoje
pm list --date 2026-06-03      # data específica (AAAA-MM-DD)
```

A saída é o registro JSON do dia de trabalho da API do PontoMais, útil para depuração ou scripts.

### Versão

```bash
pm version
```

Imprime a versão do binário, o hash do commit e a data de build.

### Atualizar

```bash
pm update           # verifica e, se houver novidade, atualiza
pm update --check   # apenas verifica, sem alterar nada
```

`pm update --check` mostra o commit instalado, sua data e quantas atualizações existem em `origin/main`. A atualização é bloqueada — com o motivo na tela — quando há alterações locais não commitadas, quando Go ou Node.js não estão no PATH, ou quando o repositório remoto está inacessível.

Em instalações feitas a partir do tarball (sem git) não há como comparar versões, e atualizar sempre reinstala a partir do código mais recente.

---

## Usando o App Desktop (GUI)

Inicie o app desktop:

```bash
bin/pm-desktop
```

Ou, se o lançador desktop foi instalado no Linux, abra **PM — Planner** pelo menu de aplicativos.

O mesmo binário também possui modos internos usados pelos lembretes:

```bash
bin/pm-desktop --daemon
```

Normalmente você não precisa chamá-los manualmente: a página de Configurações registra o daemon no autostart do usuário quando os lembretes são ativados.

O app tem duas páginas acessíveis pela barra lateral.

### Página do Planner

Esta é a tela principal do app.

1. **Selecione uma data** usando o seletor de datas no topo. Por padrão, é o dia atual.
2. Clique em **Carregar Dia** para buscar os dados do dia de trabalho na API do PontoMais.
3. O card **Marcações Sugeridas** exibe os quatro campos de horário de ponto:
   - **Entrada 1**, **Saída 1** e **Entrada 2** são editáveis. Use os botões de passo (▲/▼) para ajustar em incrementos de 15 minutos, ou digite o horário diretamente.
   - **Saída 2** é o horário normal, somente leitura, e atualiza automaticamente conforme os outros três campos são alterados.
   - Quando houver uma opção de banco para hoje, um horário alternativo colorido aparece ao lado. O selo compacto mostra o saldo; passe o cursor, toque ou foque o selo para ver o cálculo, o crédito em `1,5x`, o saldo restante e eventual limite diário configurado.
   - Se o saldo estiver indisponível hoje, um selo discreto informa isso sem alterar a saída normal. Recarregar o dia tenta a consulta novamente.
4. O card **Resumo** atualiza em tempo real enquanto você edita, exibindo:
   - **Meta do Dia** — jornada normal informada pelo PontoMais
   - **1ª Jornada** — duração do primeiro turno (Entrada 1 → Saída 1)
   - **2ª Jornada** — duração do segundo turno normal (Entrada 2 → Saída 2)
   - **Total** — tempo total trabalhado
   - **Hora Extra** — horas acima da meta normal

O planner exibe a linha de ponto bruta da API acima dos campos de entrada para que você possa comparar os horários sugeridos com os registros reais.

Os lembretes nativos continuam usando a Saída 2 normal e não são alterados pela alternativa do banco.

### Página de Configurações

Clique no ícone de configurações na barra lateral para configurar o app.

**Conta e Sessão:**

- Insira seu **login** (e-mail ou CPF) e **senha** do PontoMais.
- Clique em **Salvar Conta** para gravar no `config.yaml` compartilhado.
- Clique em **Testar Login** para verificar as credenciais contra a API antes de salvar.

**Planner:**

- Ajuste os quatro horários âncora (**Entrada 1**, **Saída 1**, **Entrada 2**, **Saída 2**) usados como padrão quando o planner não tem ponto correspondente para iniciar.
- Defina o **Limite Diário de Trabalho Extra** usado pela saída alternativa ao pagar saldo negativo. O padrão é 3 horas.
- Horários consecutivos devem ter pelo menos 15 minutos de diferença.
- Clique em **Restaurar Padrões** para voltar a `08:00 / 12:00 / 13:30 / 18:00` e ao limite de 3 horas.
- Clique em **Salvar Planner** para persistir no `config.yaml`.

**Lembretes de Jornada:**

- Ative os lembretes para iniciar o daemon em segundo plano e avisar antes de Entrada 1, Saída 1, Entrada 2 e Saída 2.
- Adicione e remova quantos avisos personalizados quiser, entre 1 minuto e 4 horas antes.
- Quando **Iniciar com o sistema** estiver ativo, o app cria um LaunchAgent no macOS, uma entrada XDG autostart no Linux, ou um atalho Startup no Windows.

**Atualizações:**

- Clique em **Verificar Atualizações** para comparar o commit instalado com `origin/main`. O card mostra a versão instalada e quantas atualizações existem.
- Clique em **Atualizar Agora** para aplicar. O app fecha, a atualização roda em segundo plano e o app reabre sozinho — o resultado aparece como aviso na volta.
- Quando algo impede a atualização (alterações locais não commitadas, Go ou Node.js ausentes, repositório remoto inacessível), o motivo aparece no card e o botão fica desabilitado, antes de o app fechar.

---

## Desenvolvimento

Inicie o app desktop com hot reload:

```bash
go run ./cmd/pm project dev desktop
```

Isso requer a Wails CLI. O frontend React é servido em `http://localhost:5173` via Vite.

Execute testes e linting:

```bash
go test ./...
npm --prefix desktop/frontend ci
npm --prefix desktop/frontend run lint
npm --prefix desktop/frontend run build
(cd desktop && go test .)
```

Limpe os artefatos de build:

```bash
bin/pm project clean
```

---

## Solução de Problemas

| Sintoma | Solução |
| --- | --- |
| `wails` não encontrado após instalação | Adicione `$(go env GOPATH)/bin` ao PATH (`~/.zshrc`, `~/.bashrc` ou perfil do PowerShell) e reinicie o terminal |
| `pm` não encontrado após instalação | Reinicie o terminal (o script adiciona `$(go env GOPATH)/bin` ao PATH) ou execute `source ~/.zshrc` |
| Falha ao obter código-fonte no setup | Verifique conexão de rede; instale `git` ou clone manualmente: `git clone https://github.com/ArturMinelli/pm-planner.git ~/pm-planner` e execute o script novamente |
| Compilação falha com dependências parciais | Conclua instalações pendentes (Xcode CLT no macOS, GTK/WebKit no Linux, GCC/WebView2 no Windows) e execute o script novamente |
| Distribuição Linux não suportada pelo script | Instale manualmente os pacotes listados em [Linux](#linux) ou consulte `wails doctor` |
| PowerShell bloqueia `setup.ps1` | Execute `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned` ou rode `powershell -ExecutionPolicy Bypass -File .\scripts\setup.ps1` |
| `wails` não encontrado durante o desenvolvimento | Instale com `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Build no Linux não encontra o WebKitGTK | Instale `libwebkit2gtk-4.1-dev`; distribuições mais antigas podem precisar de `libwebkit2gtk-4.0-dev` com build tags ajustadas |
| `linux/386` tenta linkar contra GTK/WebKit 64 bits | Use `pm project build desktop`; o comando força `GOARCH=amd64` no Linux x86\_64, a menos que `--force-go-host` seja passado |
| Assets do frontend ausentes em tempo de execução | Execute `pm project build desktop` sem `--skip-frontend` |
| Desenvolvimento no navegador sem API | Rode `npm run dev` em `desktop/frontend` — inicia o servidor Go local (`cmd/pm-dev`) e o Vite com proxy `/api` |
| Erro de autenticação no `pm list` | Verifique o `config.yaml` com login (e-mail ou CPF) e senha corretos; use Configurações → **Testar Login** no app desktop para verificar |
| `gcc` não encontrado no Windows | Instale o [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) e certifique-se de que está no `PATH` |
| WebView2 não encontrado no Windows | Baixe e instale o [WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2) da Microsoft |
