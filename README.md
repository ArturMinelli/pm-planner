# PM Planner

O PM Planner é um auxiliar do PontoMais que permite planejar os horários de ponto do dia de trabalho antes de bater o ponto de verdade. Ele vem em duas formas que compartilham a mesma lógica Go e arquivo de configuração:

- **`pm`** — CLI para o terminal
- **`pm-desktop`** — App desktop Wails/React com interface gráfica

---

## Índice

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
  - [Planejar um Dia de Trabalho](#planejar-um-dia-de-trabalho)
  - [Versão](#versão)
- [Usando o App Desktop (GUI)](#usando-o-app-desktop-gui)
  - [Página do Planner](#página-do-planner)
  - [Página de Configurações](#página-de-configurações)
- [Desenvolvimento](#desenvolvimento)
- [Solução de Problemas](#solução-de-problemas)

---

## Requisitos

### macOS

- Go 1.23+
- Node.js/npm (apenas para o app desktop)
- Xcode Command Line Tools (necessário para CGO):

```bash
xcode-select --install
```

- Wails CLI (obrigatório para compilar e instalar o app desktop):

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Linux

- Go 1.23+
- Node.js/npm (apenas para o app desktop)
- CGO, GTK e headers de desenvolvimento do WebKitGTK:

```bash
sudo apt install build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

> Distribuições mais antigas podem precisar de `libwebkit2gtk-4.0-dev` com as build tags ajustadas.

- Wails CLI (apenas para desenvolvimento com hot reload):

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Windows

- Go 1.23+
- Node.js/npm (apenas para o app desktop)
- [GCC via TDM-GCC ou MinGW-w64](https://jmeubank.github.io/tdm-gcc/) (necessário para CGO — o compilador `gcc` deve estar no `PATH`)
- WebView2 Runtime (já incluído no Windows 10/11; para versões anteriores, baixe o instalador em [developer.microsoft.com/microsoft-edge/webview2](https://developer.microsoft.com/microsoft-edge/webview2))
- Wails CLI (apenas para desenvolvimento com hot reload):

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

---

Execute o comando de diagnóstico para verificar o ambiente:

```bash
go run ./cmd/pm project doctor
```

---

## Instalação

### Compilar a CLI

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
email: "voce@empresa.com"
password: "sua-senha"
cache_ttl_hours: 8
planner:
  in1: "08:00"
  out1: "12:00"
  in2: "13:30"
  out2: "18:00"
```

| Campo | Descrição |
| --- | --- |
| `email` / `password` | Credenciais de login do PontoMais |
| `cache_ttl_hours` | Horas em que o cache de sessão local permanece válido quando a API não fornece uma data de expiração. Padrão: 8. |
| `planner.in1/out1/in2/out2` | Horários âncora padrão para os quatro campos de ponto. Usados tanto pelo `pm plan` quanto pelo planner do app desktop como sugestão inicial para o dia. |

Os headers de autenticação são armazenados em cache como `session.json` no mesmo diretório e reutilizados até expirarem.

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

### Planejar um Dia de Trabalho

Carrega um dia de trabalho e edita interativamente os quatro campos de horário de ponto:

```bash
pm plan                        # hoje, UI interativa Bubble Tea
pm plan --date 2026-06-03      # data específica
pm plan --live=false           # usa o formulário huh mais simples em vez do TUI
```

O planner:
1. Busca o dia de trabalho na API.
2. Mapeia os registros de ponto existentes para **Entrada 1**, **Saída 1**, **Entrada 2** e **Saída 2** usando os horários âncora configurados.
3. Permite editar os três primeiros campos; **Saída 2 é calculada** a partir das horas-meta do dia e das três entradas fornecidas.
4. Exibe um resumo: 1ª jornada, 2ª jornada, total trabalhado e horas extras.

Os horários âncora padrão (`08:00`, `12:00`, `13:30`, `18:00`) são os valores iniciais quando não há ponto correspondente. Altere-os no `config.yaml` ou pela página de Configurações do app desktop.

### Versão

```bash
pm version
```

Imprime a versão do binário, o hash do commit e a data de build.

---

## Usando o App Desktop (GUI)

Inicie o app desktop:

```bash
bin/pm-desktop
```

Ou, se o lançador desktop foi instalado no Linux, abra **PM — Planner** pelo menu de aplicativos.

O app tem duas páginas acessíveis pela barra lateral.

### Página do Planner

Esta é a tela principal, equivalente ao `pm plan` no terminal.

1. **Selecione uma data** usando o seletor de datas no topo. Por padrão, é o dia atual.
2. Clique em **Carregar Dia** para buscar os dados do dia de trabalho na API do PontoMais.
3. O card **Marcações Sugeridas** exibe os quatro campos de horário de ponto:
   - **Entrada 1**, **Saída 1** e **Entrada 2** são editáveis. Use os botões de passo (▲/▼) para ajustar em incrementos de 15 minutos, ou digite o horário diretamente.
   - **Saída 2 Calculada** é somente leitura e atualiza automaticamente conforme os outros três campos são alterados.
4. O card **Resumo** atualiza em tempo real enquanto você edita, exibindo:
   - **Meta do Dia** — horas necessárias para o dia
   - **1ª Jornada** — duração do primeiro turno (Entrada 1 → Saída 1)
   - **2ª Jornada** — duração do segundo turno (Entrada 2 → Saída 2)
   - **Total** — tempo total trabalhado
   - **Hora Extra** — horas extras (negativo se houver déficit)

O planner exibe a linha de ponto bruta da API acima dos campos de entrada para que você possa comparar os horários sugeridos com os registros reais.

### Página de Configurações

Clique no ícone de configurações na barra lateral para configurar o app.

**Conta e Sessão:**

- Insira seu **e-mail** e **senha** do PontoMais.
- Opcionalmente, defina o **Cache TTL em Horas** — quantas horas o cache de sessão local permanece válido quando a API não fornece data de expiração.
- Clique em **Salvar Conta** para gravar no `config.yaml` compartilhado.
- Clique em **Testar Login** para verificar as credenciais contra a API antes de salvar.

**Horários Padrão do Planner:**

- Ajuste os quatro horários âncora (**Entrada 1**, **Saída 1**, **Entrada 2**, **Saída 2**) usados como padrão quando o planner não tem ponto correspondente para iniciar.
- Horários consecutivos devem ter pelo menos 15 minutos de diferença.
- Clique em **Restaurar Padrões** para voltar a `08:00 / 12:00 / 13:30 / 18:00`.
- Clique em **Salvar Horários** para persistir no `config.yaml`. Esses valores também são utilizados pelo `pm plan` na CLI.

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
| `wails` não encontrado durante o desenvolvimento | Instale com `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| Build no Linux não encontra o WebKitGTK | Instale `libwebkit2gtk-4.1-dev`; distribuições mais antigas podem precisar de `libwebkit2gtk-4.0-dev` com build tags ajustadas |
| `linux/386` tenta linkar contra GTK/WebKit 64 bits | Use `pm project build desktop`; o comando força `GOARCH=amd64` no Linux x86\_64, a menos que `--force-go-host` seja passado |
| Assets do frontend ausentes em tempo de execução | Execute `pm project build desktop` sem `--skip-frontend` |
| App desktop exibe banner "Modo navegador" | O frontend foi aberto em um navegador comum em vez de pelo `pm-desktop`; chamadas reais à API só funcionam dentro do shell Wails |
| Erro de autenticação no `pm plan` ou `pm list` | Verifique o `config.yaml` com e-mail e senha corretos; use Configurações → **Testar Login** no app desktop para verificar |
| `gcc` não encontrado no Windows | Instale o [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) e certifique-se de que está no `PATH` |
| WebView2 não encontrado no Windows | Baixe e instale o [WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2) da Microsoft |
