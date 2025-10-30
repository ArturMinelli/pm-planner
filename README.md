# pm-cli

A small Cobra-based CLI to call PontoMais API.

## Install

```
go mod tidy
go build -o pm
```

## Configuration

Place a YAML config file at `$HOME/.pm/config.yaml`:

```yaml
email: "you@example.com"
password: "your-password"
# optional: cache TTL in hours if API does not provide expiry (default: 8)
# cache_ttl_hours: 8
```

The CLI will log in automatically using your email/password and cache the
authentication headers in `$HOME/.pm/session.json` for reuse. If the API returns
an `Expiry` header it will be respected; otherwise the TTL is used.

## Usage

```
# List points for today
./pm list

# List points for a specific date
./pm list --date 2024-06-11

# Credentials are read from config; no auth flags are required
./pm list --date 2024-06-11
```

The `list` command calls:

```
GET https://api.pontomais.com.br/api/time_card_control/current/work_days/{YYYY-MM-DD}
```
