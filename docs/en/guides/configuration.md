# Configuration

Language: [Русский](../../ru/guides/configuration.md) | English

## Navigation

- [Documentation](../index.md)
  - [Guides](index.md)
    - [Installation](installation.md)
    - [Configuration](configuration.md)
    - [Interactive Mode](interactive-mode.md)
    - [Progress](progress.md)
    - [Commands Index](commands/index.md)
    - [Instructions](instructions/index.md)
  - [Architecture](../architecture/index.md)
  - [Operations](../operations/index.md)
  - [Reports](../reports/index.md)
- [Home](../../../README.md)

## Configuration sources

Priority (highest to lowest):

1. **Command-line flags** (`--url`, `--username`, `--api-key`)
2. **Environment variables** (`TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`)
3. **Config file** (`~/.gotr/config/default.yaml`)

## Environment variables

```bash
export TESTRAIL_BASE_URL="https://testrail.example.com"
export TESTRAIL_USERNAME="user@example.com"
export TESTRAIL_API_KEY="your_api_key"
```

## Config file

### Creating the config

```bash
# Create the default config
gotr config init

# Show the config path
gotr config path

# View contents
gotr config view

# Edit
gotr config edit
```

Security notes:
- `gotr config init` creates the file with mode 0600 (read/write for the owner only).
- `gotr config view` masks sensitive keys `api_key`, `password`, `token`, `authorization` as `"***"`.

### `config.yaml` structure

```yaml
base_url: "https://testrail.example.com"
username: "user@example.com"
api_key: "your_api_key"

# Optional parameters
insecure: false      # skip TLS verification (legacy; see tls.insecure)
jq_format: false     # enable jq formatting by default
debug: false         # debug output

# TLS (v3.3.0+)
tls:
  insecure: false                     # equivalent of top-level insecure; OR-merge
  ca_bundle: "/etc/ssl/corp-ca.pem"   # corporate CA; preferred over insecure=true

# UI / warnings (v3.3.0+)
ui:
  # Per-key suppression list: tls_insecure | deprecation | flat_layout
  suppress_warnings: []

# Retention / auto-cleanup (v3.3.0+)
# Disabled by default; gotr cleanup applies this policy manually.
retention:
  reports:
    enabled: false
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]   # whitelist: these categories are never pruned
    dry_run: true
  snaps:
    enabled: false
    max_age_days: 180
    max_count: 100
    dry_run: false
  exports:
    enabled: false
    max_age_days: 30

# Compare command settings (performance tuning)
compare:
  # Auto-detect environment: auto | cloud | server
  deployment: "auto"

  # For cloud: professional | enterprise
  cloud_tier: "professional"

  # Request rate limit per minute: -1 = auto (per profile), 0 = unlimited, >0 = fixed
  rate_limit: -1

  cases:
    parallel_suites: 10    # Suites fetched in parallel
    parallel_pages: 6      # Pages fetched in parallel within a suite
    page_retries: 5        # Retries per page in the main stage
    timeout: "30m"         # Timeout for the whole operation
    auto_retry_failed_pages: true  # Auto-retry problematic pages

    retry:
      attempts: 5          # Attempts per page in the targeted retry
      workers: 12          # Parallel workers for the retry stage
      delay: "200ms"       # Delay between attempts
```

### File location

Config lookup, in priority order:

1. `~/.gotr/config/default.yaml`
2. `./config.yaml` (current directory)

## Global flags

| Flag | Description | Environment variable |
|------|-------------|---------------------|
| `--url` | TestRail URL | `TESTRAIL_BASE_URL` |
| `-u, --username` | User email | `TESTRAIL_USERNAME` |
| `-k, --api-key` | API key | `TESTRAIL_API_KEY` |
| `-i, --insecure` | Skip TLS verification (legacy, see `tls.insecure`) | - |
| `--show-warnings` | Show all warnings, ignoring `ui.suppress_warnings` | - |
| `-d, --debug` | Debug output | `TESTRAIL_DEBUG` |

## TLS and corporate CA (v3.3.0+)

The preferred way to work with private CAs is to load a PEM bundle:

```yaml
tls:
  ca_bundle: "/etc/ssl/corp-ca.pem"
```

The bundle is read once at start, parsed with `x509.NewCertPool` and
plugged into `tls.Config.RootCAs` of the internal HTTP client (see the
architecture doc on warnings + TLS). This is safer than
`insecure=true` and avoids MITM problems in CI.

The `tls.insecure` key is equivalent to top-level `insecure` and the
legacy `--insecure` flag; the sources OR-merge, so existing configs
keep working unchanged. When `insecure` is active, a warning banner is
printed to stderr (key `tls_insecure`).

## Warnings: suppress_warnings (v3.3.0+)

Some non-critical warnings can be suppressed per key:

```yaml
ui:
  suppress_warnings:
    - tls_insecure   # banner about insecure=true
    - flat_layout    # one-shot hint about hierarchy migration
```

Allowed keys: `tls_insecure`, `deprecation`, `flat_layout`. On the
first emission of any warning a hint is appended to stderr:
"add '<key>' to ui.suppress_warnings to silence this warning".

The `--show-warnings` flag is a runtime override that prints all
warnings regardless of `suppress_warnings` (handy for CI validation).

## Retention/cleanup (v3.3.0+)

Retention policies are described in the config and applied manually
via `gotr cleanup` (nothing is deleted automatically by default):

```yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]
    dry_run: false
```

- `max_age_days` — delete files older than N days (0 = unlimited).
- `max_count` — keep no more than N newest files per category
  (0 = unlimited).
- `keep_categories` — whitelist of categories ignored by retention
  (typically `coverage`).
- `dry_run: true` — print the plan only, do not delete.

Run:

```bash
gotr cleanup reports --dry-run
gotr cleanup all
```

See [gotr cleanup](commands/cleanup.md).

## Compare flags

These flags are available for all `compare` subcommands (`cases`,
`all`, `retry-failed-pages`, etc.):

| Flag | Description | Config key |
|------|-------------|------------|
| `--rate-limit` | Requests per minute (-1=auto, 0=unlimited) | `compare.rate_limit` |
| `--parallel-suites` | Parallel suites (default 10) | `compare.cases.parallel_suites` |
| `--parallel-pages` | Parallel pages (default 6) | `compare.cases.parallel_pages` |
| `--page-retries` | Retries per page (default 5) | `compare.cases.page_retries` |
| `--timeout` | Operation timeout (default 30m) | `compare.cases.timeout` |
| `--retry-attempts` | Auto-retry attempts (default 3) | `compare.cases.retry.attempts` |
| `--retry-workers` | Auto-retry workers (default 12) | `compare.cases.retry.workers` |
| `--retry-delay` | Auto-retry delay (default 200ms) | `compare.cases.retry.delay` |

**Priority:** CLI flag > YAML config > default.

## Usage examples

```bash
# Via flags
gotr get projects --url https://testrail.example.com --username user@example.com --api-key xxx

# Via env
gotr get projects

# Via config
gotr get projects
```

---

← [Guides](index.md) · [Documentation](../index.md)
