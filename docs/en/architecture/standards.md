# gotr coding standards

Language: [Русский](../../ru/architecture/standards.md) | English

## Navigation

- [Documentation](../index.md)
  - [Guides](../guides/index.md)
    - [Installation](../guides/installation.md)
    - [Configuration](../guides/configuration.md)
    - [Interactive Mode](../guides/interactive-mode.md)
    - [Progress](../guides/progress.md)
    - [Commands Index](../guides/commands/index.md)
      - [General](../guides/commands/index.md#general)
      - [CRUD Operations](../guides/commands/index.md#crud-operations)
      - [Core Resources](../guides/commands/index.md#core-resources)
      - [Special Resources](../guides/commands/index.md#special-resources)
    - [Instructions](../guides/instructions/index.md)
  - [Architecture](index.md)
    - [Overview](overview.md)
    - [Concurrency](concurrency.md)
    - [Standards](standards.md)
    - [Parallelization Plan](recursive-parallelization-plan.md)
  - [Operations](../operations/index.md)
  - [Reports](../reports/index.md)
- [Home](../../../README.md)

## Contents

1. [General principles](#1-general-principles)
2. [Layered architecture](#2-layered-architecture)
3. [Package layout](#3-package-layout)
4. [Coding rules](#4-coding-rules)
5. [Project patterns](#5-project-patterns)
6. [Configuration and environment](#6-configuration-and-environment)
7. [Concurrency and resilience](#7-concurrency-and-resilience)
8. [Logging](#8-logging)
9. [Testing](#9-testing)
10. [Build and release](#10-build-and-release)
11. [Checklist](#11-checklist)

---

## 1. General principles

| Principle | Description |
| --------- | ----------- |
| **Single responsibility** | Each package/file/function does one thing |
| **Explicit > magic** | No global state, no hidden `init()`, DI through arguments |
| **Errors are values** | No `os.Exit`/`log.Fatal` in library code. Panic only in `GetClient*` |
| **Small interfaces** | Interface Segregation: interface lives in the consumer package |
| **Testability** | Mocks via interfaces. No external calls in unit tests |
| **DRY** | Duplication is a bug. Generic factories, helpers `ui.*`, `flags.*` |
| **YAGNI** | No "just-in-case" abstractions. Refactor on the third consumer |

---

## 2. Layered architecture

```text
cmd/ → service/ → client/ → HTTP (TestRail API)
cmd/ → ui.*      (output)
cmd/ → flags.*   (validation)
cmd/ → output.*  (saving)
cmd/ → log.*     (logging)
```

### 2.1. Dependency rules

| Layer | May depend on | Must NOT depend on |
| ----- | ------------- | ------------------ |
| `cmd/*` | `internal/*`, `pkg/*` | — |
| `internal/service` | `internal/client`, `internal/concurrency`, `internal/concurrent`, `internal/models` | `cmd/*`, `internal/ui` |
| `internal/client` | `internal/concurrent`, `internal/models/data` | `cmd/*`, `internal/service` |
| `internal/ui` | stdlib, `go-pretty/v6` | `internal/client`, `internal/service` |
| `internal/concurrency` | stdlib | `internal/client`, `cmd/*` |
| `internal/concurrent` | `golang.org/x/time/rate`, `golang.org/x/sync/errgroup` | `cmd/*`, `internal/service` |
| `internal/log` | `go.uber.org/zap`, `internal/paths` | `cmd/*`, `internal/client` |
| `internal/paths` | stdlib | everything else |
| `pkg/*` | stdlib, `go-pretty/v6` | `internal/*`, `cmd/*` |

### Forbidden

- `service/` → `cmd/` (reaching up across layers)
- `client/` → `ui/` (the client knows nothing about UI)
- `pkg/` → `internal/` (public API does not import private code)
- Circular dependencies between packages

---

## 3. Package layout

### 3.1. `cmd/` — CLI commands

- Constructor: `newXxxCmd(clientFn func(*cobra.Command) client.ClientInterface)`
- Registration: `Register(rootCmd, clientFn)` — called from `commands.go`
- Use `RunE` (not `Run`) — errors are returned to Cobra
- Context: `ctx := cmd.Context()` — passed into every call

**Command registration:** every sub-package calls `Register(rootCmd, clientFn)` in `cmd/commands.go`:

```go
func init() {
    initConfig()
    initGlobalFlags()
    // ...
    cases.Register(rootCmd, GetClientInterface)
    compare.Register(rootCmd, GetClientInterface)
    sync.Register(rootCmd, GetClient)
}
```

### 3.2. `internal/client/` — HTTP client

- Each method: `func (c *HTTPClient) GetXxx(ctx context.Context, ...)`
- List methods: `fetchAllPages[T]` for auto-pagination
- 14 ISP-aligned interfaces, 106 endpoints, 100% coverage of TestRail API v2
- New endpoint: add to `interfaces.go` + `mock.go` + the matching domain file

### HTTP transport (`authTransport`)

`authTransport` is a custom `http.RoundTripper` wrapping the standard transport.
It automatically appends to every request:

- `Authorization: Basic ...` — via `req.SetBasicAuth(username, apiKey)`
- `Content-Type: application/json` — if the caller did not set it
- `User-Agent: Mozilla/5.0 (compatible; gotr/2.7; ...)` — **mandatory**, some TestRail installations return 403/401 without a browser-like header

### Transport tuning

| Parameter | Value | Reason |
| --------- | ----- | ------ |
| `MaxConnsPerHost` | `0` (unlimited) | Concurrency is governed by `concurrent.WorkerPool`, not the transport |
| `MaxIdleConns` | `200` | Pool for connection reuse |
| `MaxIdleConnsPerHost` | `200` | Matches `MaxIdleConns` (single TestRail host) |
| `IdleConnTimeout` | `90s` | Go default |
| `TLSHandshakeTimeout` | `10s` | From `defaultOptions` |
| `Timeout` (Client) | `30s` | From `defaultOptions`, overridable via `WithTimeout()` |

> **Important:** `MaxConnsPerHost=0` (unlimited) is a deliberate decision. With `MaxConnsPerHost=50` and 160 parallel requests (2 projects × 8 suites × 10 pages), 110 requests get queued inside the Go transport. `http.Client.Timeout` includes the queueing time → cascading timeouts → exponential backoff → 3× slowdown.

### Pagination (`paginator.go`)

```go
const paginationLimit = 250  // Standard TestRail API page size
```

- `fetchAllPages[T](ctx, client, endpoint, baseQuery, itemsField)` — fetches all pages
- `decodeListResponse[T](body, itemsField)` — dual-format detection:
  - `{` → Paginated wrapper (TestRail 6.7+): `{"offset":0, "limit":250, "<itemsField>":[...]}`
  - `[` → Flat array (older TestRail Server versions): `[item1, item2, ...]`
- `itemsField` — JSON key name: `"runs"`, `"plans"`, `"sections"`, `"milestones"`, etc.
- Loop: `offset += paginationLimit` while `pageLen >= paginationLimit`

### Functional Options

```go
client.NewClient(baseURL, username, apiKey, debug,
    client.WithSkipTlsVerify(true),   // --insecure
    client.WithTimeout(60*time.Second),
)
```

### 3.3. `internal/service/` — Business logic

The service layer encapsulates business rules and validation:

- `RunService` — create/update/close test runs (parameter validation)
- `ResultService` — work with test results
- `TestService` — operations on tests
- `migration/` — data migration subsystem between TestRail projects:
  - `types.go` — migration context (`Migration` struct)
  - `fetch.go` — fetch source/target data
  - `filter.go` — filter duplicates
  - `import.go` — import entities
  - `export.go` — export data and mappings
  - `mapping.go` — manage ID mappings (source→target)
  - `mapping_loader.go` — load mappings from files
  - `log.go` — log migration operations
  - `migrate.go` — orchestrates the full migration

### 3.4. `internal/ui/` — Output

- `ui.NewTable(cmd)` — create a go-pretty table honouring `--format`
- `ui.Table(cmd, t)` — render a table (table/json/csv/md/html)
- `ui.JSON(cmd, data)` — JSON output honouring `--quiet`
- `ui.IsJSON(cmd)`, `ui.IsQuiet(cmd)` — format checks
- Styled messages (emoji prefixes live only here!):
  - `ui.Info(w, msg)` — ℹ️ information
  - `ui.Success(w, msg)` — ✅ success
  - `ui.Warning(w, msg)` — ⚠️ warning
  - `ui.Error(w, msg)` — ❌ error
  - `ui.Phase(w, msg)` — 🔄 phase
  - `ui.Stat(w, icon, label, val)` — statistic
  - `ui.Section(w, msg)` — section header
  - `ui.Preview(w, title, fields)` — preview pane
- `display.go` — ANSI live display with dynamic tasks (used by `compare cases`):
  - `ui.New(opts...)` — create a Display + background refresh loop (~5 Hz)
  - `d.SetHeader(text)` — header above the tasks
  - `d.AddTask(name, total) *Task` — task tracking; `*Task` implements `parallel.ProgressReporter`
  - `t.OnCasesReceived(n)` — update the case counter
  - `t.OnPageFetched()` — page fetched
  - `t.OnSuiteComplete()` — suite completed
  - `d.Finish()` — stop the refresh loop, render the final frame
  - Options: `WithWriter(w)`, `WithQuiet(true)` — disables output
  - Rendering: ANSI escape codes for in-place line rewriting

### Rules

- All user-facing output goes through `ui.*` (interactive prompts and debug excluded)
- Emoji prefixes only in `ui.*` — no hard-coded emoji in `cmd/`
- The first argument is `io.Writer` (typically `os.Stdout`)

### 3.5. `internal/flags/` — Validation

```go
flags.ValidateRequiredID(args, index, name)   // Parse an ID from arguments
flags.GetFlagInt64(cmd, name)                 // int64 flag
flags.GetFlagString(cmd, name)                // string flag
flags.GetFlagBool(cmd, name)                  // bool flag
flags.ParseID(s)                              // string → int64
```

### 3.6. `internal/output/` — Saving

```go
output.AddFlag(cmd)                       // Register --save, --save-to
output.OutputResult(cmd, data, resource)  // Print + save
output.Output(cmd, data, dir, format)     // Save into ~/.gotr/exports/
```

**DryRunPrinter** — output for `--dry-run` mode (writes to `os.Stderr`):

```go
printer := output.NewDryRunPrinter("sync cases")

// Full operation with HTTP details:
printer.PrintOperation("Create case", "POST", "/api/v2/add_case/1", requestBody)

// Simple operation without a body:
printer.PrintSimple("Delete case", "Would delete case #123")

// Batch operation (shows up to 10 items):
printer.PrintBatch("Sync shared steps", []string{"Step 1", "Step 2", ...})
```

Output: ASCII frames with metadata (Command, Operation, HTTP Method, Endpoint, Request Body).

### 3.7. `internal/log/` — Structured logging

Centralised logging via `go.uber.org/zap`:

```go
log.InitDefault()     // In main.go — initialise with the default config
defer log.Sync()      // Flush buffers on exit
log.L()               // Global logger (fallback → zap.NewNop())
log.Debug("msg")      // Debug message
```

**Configuration** (`log.Config`):

- `Level` — level (`debug`, `info`, `warn`, `error`)
- `JSONFormat` — JSON format (for machine-parseable logs)
- `LogDir` — log directory (default `~/.gotr/logs/`)
- `Development` — development mode (stack traces, line numbers)

### When to use what

- `log.L()` — internal events, audit, diagnostics (written to file)
- `ui.Info()` — user-facing terminal output
- `utils.DebugPrint()` — debug behind the `--debug` flag (to stderr)

### 3.8. `internal/paths/` — Path management

Centralised paths — all gotr directories in a single place:

```text
~/.gotr/                    # BaseDir()
├── config/                 # ConfigDirPath() — configuration
│   └── default.yaml        # Main config file
├── logs/                   # LogsDirPath() — zap logs
├── selftest/               # SelftestDirPath() — selftest reports
├── cache/                  # CacheDirPath() — API cache
├── exports/                # ExportsDirPath() — data exports (--save)
└── temp/                   # TempDirPath() — temporary files (jq)
```

**Rule:** every path goes through `paths.*` — never construct them by hand.

### 3.9. `internal/ui/runtime.go` — Progress runtime

Unified progress/status runtime:

- `ui.RunWithStatus(ctx, cfg, fn)` — status wrapper
- `ui.NewOperation(cfg)` + `AddTask(...)` — multi-phase progress
- Honours `--quiet` automatically

**Used by:** compare/sync/get and the rest of the long-running commands.

### 3.10. `internal/interactive/` — Interactive selection

Dependency: `github.com/AlecAivazis/survey/v2` — interactive terminal prompts.

**Selector functions** (API fetch + selection prompt):

```go
interactive.SelectProjectInteractively(ctx, client)  // Pick a project
interactive.SelectSuiteInteractively(ctx, client, projectID)  // Pick a suite
interactive.SelectRunInteractively(ctx, client, projectID)    // Pick a run
```

**Wizard functions** (`wizard.go`) — forms that create/update resources:

```go
interactive.AskProject(isUpdate)  // → *ProjectAnswers
interactive.AskSuite(isUpdate)    // → *SuiteAnswers
interactive.AskCase(isUpdate)     // → *CaseAnswers
interactive.AskRun(isUpdate)      // → *RunAnswers
```

survey/v2 prompt types: `survey.Input`, `survey.Multiline`, `survey.Confirm`, `survey.Select`.

**Used by:** any command where the user did not provide an ID — it auto-prompts.

### 3.11. `internal/selftest/` — Self-diagnostics

Package backing `gotr selftest` — environment checks:

### Interface

```go
type Checker interface {
    Name() string
    Category() string
    Check() CheckResult
}
```

**Outcomes:** `PASS` (✓), `FAIL` (✗), `WARN` (⚠), `SKIP` (⊘) — with ANSI colours.

### Built-in checks (6 checkers)

| Checker | Category | What it checks |
| ------- | -------- | -------------- |
| `ConfigChecker` | Configuration | Config `~/.gotr/config/default.yaml` exists and is valid |
| `BaseDirChecker` | Configuration | All 6 sub-directories of `~/.gotr/` (config, logs, selftest, cache, exports, temp); auto-creates them via `os.MkdirAll` |
| `BinaryInfoChecker` | System | Version, commit, build date (always PASS) |
| `GoEnvChecker` | System | Go version, OS/arch, CPU count |
| `AllTestsChecker` | Tests | Runs `go test ./... -v`; counts passed/failed/skipped; saves the report into `~/.gotr/selftest/` with a `latest.log` symlink |
| `CoverageChecker` | Coverage | `go test -coverprofile`; parses the coverage percentage; WARN if < 50% |

- Self-healing: `CanFix: true` + `FixCommand: "gotr config init"` (ConfigChecker)
- Report: `Report` struct with timestamp, version, platform, all CheckResult entries, overall Health

### 3.12. `internal/models/` — Data models

**`models/config/`** — application configuration:

```go
type ConfigData struct {
    BaseURL  string `yaml:"base_url"`
    Username string `yaml:"username"`
    APIKey   string `yaml:"api_key"`
    Insecure bool   `yaml:"insecure"`
    JqFormat bool   `yaml:"jq_format"`
    Debug    bool   `yaml:"debug"`
}
```

Builder: `config.Default()` → `.WithDefaults()` → `.Create()`

**`models/data/`** — DTOs for the TestRail API (20+ files):

- Cases, Projects, Runs, Results, Suites, Sections
- SharedSteps, Milestones, Plans, Attachments
- Configs, Users, Reports, Roles, Templates, Tests
- Groups, Labels, Datasets, BDDs
- Status, Priority — constants

### 3.13. `internal/utils/` — Utilities (legacy)

> **⚠️ God-package.** Break-up scheduled for Stage 10.0.

- `DebugPrint(format, args...)` — output under `--debug`
- `OpenEditor(path)` — open a file in `$EDITOR` (fallback: vi/notepad)
- `GetFieldValue(obj, field)` — reflection-based field extraction (case-insensitive)
- `LoadMapping(path)` — load JSON/YAML mappings `map[int64]int64`

### 3.14. `pkg/reporter/` — Report builder

A builder pattern for structured reports (go-pretty):

```go
reporter.New("cases").
    Section("General statistics").
    Stat("⏱️", "Execution time", elapsed).
    Stat("📦", "Total processed", total).
    Section("Comparison results").
    Stat("✅", "Only in P1", 145).
    StatIf("⚠️", "Errors", errs, errs > 0).  // Conditional output
    Print()

// Or the shorthand:
reporter.CompareStats("suites", pid1, pid2, onlyIn1, onlyIn2, common, elapsed).Print()
```

**Emoji → ANSI:** emoji are passed as a hint and the reporter maps them to ANSI-coloured width-1 glyphs (solving alignment across terminals).

### 3.15. `pkg/testrailapi/` — API reference

A structured representation of every TestRail API v2 endpoint:

```go
api := testrailapi.New()
allPaths := api.Paths()           // All endpoints (used by gotr resources)
casePaths := api.Cases.Paths()    // Endpoints for the Cases resource
```

Types: `APIPath{Method, URI, Description, Params}`, 26 resources.

### 3.16. `embedded/` — Embedded jq

Embedded jq binaries via `//go:embed`:

- `jq-linux-amd64`, `jq-macos-amd64`, `jq-windows-i386.exe`
- `RunEmbeddedJQ(rawBody, filter)` — extract → temp file → run → cleanup
- Auto-detects the platform via `runtime.GOOS`
- Invoked for the `--jq` flag (filtering JSON responses)

### 3.17. `cmd/compare/` — Compare subsystem

Architecture of the 13 compare sub-commands:

### Data types

- `CompareResult` — comparison result (OnlyInFirst, OnlyInSecond, Common)
- `ItemInfo` — ID + Name
- `CommonItemInfo` — Name + ID1 + ID2 + IDsMatch

**Configuration profiles** (`config_profile.go`):

- Auto-detects deployment: Cloud vs Server based on the URL
- Cloud rate limits: professional (180 req/min), enterprise (300 req/min)
- Server: no limits
- `resolveCompareCasesRuntimeConfig()` — computes parameters

**Export** — multi-format: JSON, YAML, CSV, Table (auto-detected from the file extension).

**Generic Factory** — `newSimpleCompareCmd[T]()` for DRY across the 13 sub-commands.

### 3.18. `cmd/sync/` — Sync subsystem

Package for syncing data between projects:

### Layout

- `sync.go` — registration and orchestration
- `sync_full.go` — full migration (SharedSteps → Suites → Sections → Cases)
- `sync_cases.go`, `sync_suites.go`, `sync_sections.go`, `sync_shared_steps.go` — per-resource
- `sync_flags.go` — validation of `--src-project`, `--dst-project`
- `sync_helpers.go` — shared helpers
- `interactive.go` — interactive source/destination selection
- `sync_test_helper.go`, `sync_test_skip.go` — test infrastructure

**Sync full ordering:** SharedSteps → Suites → Sections → Cases (dependencies!).

### 3.19. `cmd/internal/testhelper` — Test utilities

A shared package for testing CLI commands (visible only inside `cmd/`):

```go
testhelper.HTTPClientKey                          // Context key for the mock
testhelper.SetupTestCmd(t, mock)                  // Command + mock in context
testhelper.SetupTestCmdWithBuffer(t, mock)        // + an output buffer
testhelper.GetClientForTests(cmd)                 // Pull the mock out of the context
```

---

## 4. Coding rules

### 4.1. Naming

| Element | Style | Example |
| ------- | ----- | ------- |
| Packages | lowercase, single word | `client`, `flags`, `ui` |
| Exported identifiers | CamelCase | `GetCases`, `ValidateRequiredID` |
| Unexported identifiers | camelCase | `fetchAllPages`, `parseResponse` |
| Files | snake_case | `sync_cases.go`, `add_config.go` |
| Tests | `<file>_test.go` | `add_test.go` |
| Constants | CamelCase | `DefaultBaseURL`, `ResultPass` |
| Context key | Typed type | `type ctxKey string` (never raw `string`) |

### 4.2. Error handling

```go
// ✅ Correct — wrap with context
if err != nil {
    return fmt.Errorf("failed to get cases for project %d: %w", projectID, err)
}

// ❌ Bare return
if err != nil { return err }

// ❌ os.Exit / log.Fatal
if err != nil { os.Exit(1) }
```

### Handling hierarchy

1. Library code — returns an `error` with context (`%w`)
2. Service layer — validates input, wraps errors
3. CLI layer (`RunE`) — returns the error to Cobra (Cobra prints to stderr)
4. `GetClient*` — the only place that uses `panic` (cannot continue without a client)

### 4.3. Context

- All I/O functions take `ctx context.Context` as the first argument
- `cmd.Context()` — the context source in the CLI layer
- Ctrl+C is wired through `signal.NotifyContext` (Stage 7.0)
- `http.NewRequestWithContext(ctx, ...)` — used by the HTTP transport
- Context key: always typed, never raw `string`

### 4.4. Language

| Context | Language |
| ------- | -------- |
| Source code (variables, functions) | English |
| User-facing output (UI) | English |
| Errors (`fmt.Errorf`) | English |
| Code comments | English |
| Documentation (`docs/`, `README*.md`) | Russian |

### 4.5. Global flags

| Flag | Short | Type | Description |
| ---- | ----- | ---- | ----------- |
| `--url` | — | string | TestRail base URL |
| `--username` | `-u` | string | User email |
| `--api-key` | `-k` | string | API key |
| `--insecure` | — | bool | Skip TLS verification |
| `--config` | `-c` | bool | Create the default config |
| `--debug` | `-d` | bool | Debug output (hidden) |
| `--quiet` | `-q` | bool | Quiet mode (CI/CD) |
| `--format` | `-f` | string | Format: `table`, `json`, `csv`, `md`, `html` |

### 4.6. Output formats

| Format | Description | Example |
| ------ | ----------- | ------- |
| `table` | ASCII table (go-pretty), default | `gotr cases list 30` |
| `json` | JSON output | `gotr cases list 30 -f json` |
| `csv` | CSV (for Excel/scripts) | `gotr cases list 30 -f csv` |
| `md` | Markdown table | `gotr cases list 30 -f md` |
| `html` | HTML table | `gotr cases list 30 -f html` |

---

## 5. Project patterns

### 5.1. Constructor Injection

```go
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
    casesCmd := newCasesCmd(clientFn)
    rootCmd.AddCommand(casesCmd)
}
```

### 5.2. Generic Factory (DRY)

```go
func newSimpleCompareCmd[T any](cfg simpleCompareCfg[T]) *cobra.Command { ... }
```

### 5.3. Functional Options

```go
// WorkerPool
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),
    concurrent.WithRateLimit(150),
    concurrent.WithProgressMonitor(monitor),
)

// FetchParallel
result, err := concurrency.FetchParallel(ctx, fetchFn,
    concurrency.WithReporter(reporter),
    concurrency.WithMaxConcurrency(5),
    concurrency.WithContinueOnError(),
)
```

### 5.4. Builder Pattern

```go
// Reporter
reporter.New("cases").
    Section("Projects").
    Stat("📁", "Project 1", p1Name).
    Print()

// Config
config.Default().WithDefaults().Create()
```

### 5.5. PersistentPreRunE override

Some commands (for example `config`) do not need a TestRail client. For these we disable `PersistentPreRunE`:

```go
configCmd := &cobra.Command{
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return nil // Skip client creation
    },
}
```

### 5.6. Entry point pattern

```go
func main() {
    log.InitDefault()          // 1. Initialise the logger
    defer log.Sync()           // 2. Flush buffers on exit
    ctx, stop := signal.NotifyContext(context.Background(),
        os.Interrupt, syscall.SIGTERM)
    defer stop()               // 3. Ctrl+C → cancels the context
    cmd.Execute(ctx)           // 4. Run the CLI
}
```

---

## 6. Configuration and environment

### 6.1. Configuration priority

1. **CLI flags** — highest priority (`--url`, `--username`, `--api-key`)
2. **Environment variables** — `TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`
3. **Config file** — `~/.gotr/config/default.yaml`

### 6.2. Viper integration

- `viper.AutomaticEnv()` — auto-reads environment variables
- `viper.BindPFlag(key, flag)` — binds CLI flags to Viper keys
- `TESTRAIL_*` — the prefix for environment variables
- `TESTRAIL_PASSWORD` takes precedence over `TESTRAIL_API_KEY` (backward compatibility — for users migrating from password auth)

### 6.3. Config commands

| Command | Description |
| ------- | ----------- |
| `gotr config init` | Create `~/.gotr/config/default.yaml` with placeholders |
| `gotr config path` | Print the config path |
| `gotr config view` | Print the current config |
| `gotr config edit` | Open the config in `$EDITOR` |

### 6.4. Layout of `~/.gotr/`

```text
~/.gotr/
├── config/default.yaml     # Configuration (base_url, username, api_key, ...)
├── logs/                   # zap logs (daily rotation)
├── selftest/               # gotr selftest reports
├── cache/                  # Cached API responses
├── exports/                # Data exports (--save, --save-to)
└── temp/                   # Temporary files (jq binary, etc.)
```

---

## 7. Concurrency and resilience

Two abstraction levels plus deployment profiles.

### 7.0. Deployment modes and rate limits

gotr automatically detects the TestRail deployment mode and picks an appropriate rate limit.

**Auto-detection** (`config_profile.go`):

```go
func detectDeploymentByURL(baseURL string) string {
    if strings.Contains(url, ".testrail.io") { return "cloud" }
    return "server"
}
```

### Rate-limit profiles

| Deployment | Cloud Tier | Rate Limit | Source |
| ---------- | ---------- | ---------- | ------ |
| **Cloud** | Professional | 180 req/min | TestRail API limit (hardcoded) |
| **Cloud** | Enterprise | 300 req/min | Viper `compare.cloud_rate_limit` (default 300) |
| **Server** | — | 0 (unlimited) | Viper `compare.server_rate_limit` (default 0) |

### Why the WorkerPool default is 150 req/min

TestRail Cloud Professional allows 180 req/min. The WorkerPool defaults to 150 — a conservative limit with ~17% headroom, in order to:

- Avoid 429 (Too Many Requests) under burst load
- Leave headroom for bursts (15% of the rate = 22 tokens)
- Run reliably without tuning on most installations

**Burst formula:** `burst = requestsPerMinute * 15 / 100` (minimum 10).
At 150 req/min: burst = 22. At 300 req/min: burst = 45.

**RateLimiter fallback:** if `requestsPerMinute <= 0`, `NewRateLimiter` falls back to 180 req/min (the Professional ceiling).

### Viper configuration keys (compare)

| Key | Default | Description |
| --- | ------- | ----------- |
| `compare.deployment` | `auto` | `auto`, `cloud`, `server` |
| `compare.cloud_tier` | `professional` | `professional`, `enterprise` |
| `compare.rate_limit` | `-1` (auto) | Explicit limit (overrides the profile) |
| `compare.cloud_rate_limit` | `300` | Cloud limit (when rate_limit=-1) |
| `compare.server_rate_limit` | `0` | Server limit (0 = no limit) |
| `compare.cases.parallel_suites` | `10` | Parallel suites |
| `compare.cases.parallel_pages` | `6` | Parallel pages per suite |
| `compare.cases.page_retries` | `5` | Retries per page |
| `compare.cases.timeout` | `30m` | Total timeout for compare cases |
| `compare.cases.auto_retry_failed_pages` | `true` | Auto-retry failed pages |

**Rate-limit precedence:** CLI flag `--rate-limit` > Viper `compare.rate_limit` > profile (by deployment + tier).

### 7.1. `internal/concurrent/` — Low-level primitives

**WorkerPool** — goroutine pool built on `errgroup`:

```go
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),       // 5 parallel workers (default)
    concurrent.WithRateLimit(150),      // 150 req/min (default)
    concurrent.WithProgressMonitor(m),  // optional Increment() callback
)
pool.Submit(func() error { return doWork() })
err := pool.Wait()
```

**RateLimiter** — token bucket (`golang.org/x/time/rate`):

- req/min → rate/sec conversion: `rate.Limit(float64(rpm) / 60.0)`
- Burst: 15% of the rate (minimum 10)
- Default: 180 req/min (NewRateLimiter fallback when rpm ≤ 0)
- WorkerPool default: 150 req/min (conservative limit)
- API: `Wait()`, `WaitWithTimeout(timeout)`, `Allow()`, `Tokens()`, `Reserve()`

**Retry** — exponential backoff:

```go
config := &concurrent.RetryConfig{
    MaxRetries:   5,           // Number of attempts
    InitialDelay: 1 * time.Second,
    MaxDelay:     30 * time.Second,
    Multiplier:   2.0,         // 1s → 2s → 4s → 8s → 16s
}
err := concurrent.Retry(config, func() error { return apiCall() })
// Or with a context:
err := concurrent.RetryWithContext(ctx, config, fn)
```

**CircuitBreaker** — protection against cascade failures:

- States: Closed → Open (after N failures) → HalfOpen (after the timeout) → Closed
- `NewCircuitBreaker(maxFailures, timeout)`
- `Execute(fn)` — runs the function under protection

### 7.2. `internal/concurrency/` — High-level strategies

For compare sub-commands — three parallelisation strategies:

| Strategy | Pattern | Example |
| -------- | ------- | ------- |
| `FetchParallel[T]` | Light: N projects in parallel | compare suites/milestones |
| `FetchParallelBySuite[T]` | Medium: per-suite parallel | compare sections |
| `ParallelController` | Heavy: pipeline pagination | compare cases |

**FetchOption pattern:** `WithReporter()`, `WithContinueOnError()`, `WithMaxConcurrency(n)`

### Progress interfaces

- `ProgressReporter` — `OnItemComplete`, `OnBatchReceived`, `OnError`
- `PaginatedProgressReporter` — extension: `OnPageFetched`

### Configuration flags (compare cases)

- `--parallel-suites` — per-suite parallelism
- `--parallel-pages` — per-page parallelism
- `--page-retries` — number of page retries
- `--rate-limit` — request rate limit
- `--timeout` — timeout

### 7.3. Concurrency rules

- Goroutines **always** receive a `context.Context`
- `errgroup.WithContext()` — to manage a goroutine group
- Mutexes: `sync.Mutex` for shared state, `sync.Once` for initialisation
- Channels: prefer channels over mutexes for communication
- Race detection: `go test -race ./...` — mandatory in CI

---

## 8. Logging

### 8.1. Three output levels

| Level | Tool | Where | When |
| ----- | ---- | ----- | ---- |
| **User** | `ui.Info()`, `ui.Success()`, etc. | stdout | Always (except `--quiet`) |
| **Debug** | `utils.DebugPrint()` | stderr | Only with `--debug` |
| **Structured** | `log.L().Info()`, `log.L().Error()` | `~/.gotr/logs/` | Always (to file) |

### 8.2. When to use what

```go
// The user must see this in the terminal:
ui.Info(os.Stdout, "Loading cases...")
ui.Success(os.Stdout, "Done!")

// Debug output (only with --debug):
utils.DebugPrint("{syncCases} Processing suite %d", suiteID)

// Diagnostics written to file:
log.L().Info("API call completed",
    zap.Int64("project_id", pid),
    zap.Duration("elapsed", elapsed),
)
log.L().Error("API request failed",
    zap.Error(err),
    zap.String("endpoint", url),
)
```

---

## 9. Testing

### 9.1. General requirements

| Requirement | Description |
| ----------- | ----------- |
| Coverage | ≥ 85% per package (target 90%+) |
| Pattern | Table-driven tests |
| Mocks | `client.MockClient` |
| DI | `newXxxCmd(clientFn)` for injection |
| Naming | `TestXxx_Success`, `TestXxx_Error`, `TestXxx_EdgeCase` |
| No network | Unit tests do not make external calls |
| Race | `go test -race ./...` — 0 data races |

### 9.2. CMD test pattern

The standard pattern for testing CLI commands:

```go
func TestGetCase_Success(t *testing.T) {
    mock := &client.MockClient{
        GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
            return &data.Case{ID: caseID, Title: "Test"}, nil
        },
    }
    cmd := testhelper.SetupTestCmd(t, mock)

    getCmd := newGetCmd(testhelper.GetClientForTests)
    getCmd.SetArgs([]string{"123"})
    getCmd.SetContext(cmd.Context())

    err := getCmd.Execute()
    assert.NoError(t, err)
}
```

### 9.3. Service test pattern

```go
func TestRunService_Create_ValidatesProjectID(t *testing.T) {
    tests := []struct {
        name      string
        projectID int64
        wantErr   bool
    }{
        {"valid", 30, false},
        {"zero", 0, true},
        {"negative", -1, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := NewRunService(mockClient)
            _, err := svc.Create(ctx, tt.projectID, req)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

## 10. Build and release

### 10.1. Makefile targets

| Target | Description |
| ------ | ----------- |
| `make build` | Build using the version from `cmd/root.go` |
| `make build VERSION=v3.0.0` | Build with an explicit version |
| `make test` | Run the tests (`go test ./... -v`) |
| `make install` | Install into `/usr/local/bin` (sudo) |
| `make release` | Cross-compilation: Linux + macOS + Windows |
| `make release-compressed` | Same plus UPX compression |
| `make tag VERSION=v3.0.0` | Create a git tag and push it |
| `make clean` | Remove the binary |

### 10.2. Versioning

The version is injected via `-ldflags` at build time:

```go
var (
    Version = "2.7.0"   // Default value (for go run)
    Commit  = "unknown"
    Date    = "unknown"
)
```

`cmd/root.go` → `rootCmd.Version = Version` → `gotr --version`

### 10.3. Cross-compilation

Supported platforms:

- `linux/amd64`
- `darwin/amd64`
- `windows/amd64`

UPX compression is optional (if installed). For macOS — `--force-macos`.

### 10.4. Linting

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
golangci-lint run --timeout 5m
```

Configuration: `.golangci.yml` (11 linters). Baseline: 208 findings (Stage 9.0).

### 10.5. Dependencies (go.mod)

| Dependency | Version | Purpose |
| ---------- | ------- | ------- |
| `github.com/spf13/cobra` | v1.10.2 | CLI framework (commands, flags, completion) |
| `github.com/spf13/viper` | v1.21.0 | Configuration (yaml, env, pflags) |
| `go.uber.org/zap` | v1.27.1 | Structured logging |
| `github.com/jedib0t/go-pretty/v6` | v6.7.8 | Tables, rendering (table/json/csv/md/html) |
| `github.com/AlecAivazis/survey/v2` | v2.3.7 | Interactive prompts (wizard) |
| `golang.org/x/time` | v0.14.0 | Token bucket rate limiter |
| `golang.org/x/sync` | v0.19.0 | errgroup (WorkerPool) |
| `github.com/fatih/color` | v1.18.0 | ANSI colours (selftest) |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML parsing |
| `github.com/stretchr/testify` | v1.11.1 | Assertions (require, assert) |

---

## 11. Checklist

### A new command

- [ ] Constructor: `newXxxCmd(clientFn)`
- [ ] `ctx := cmd.Context()` is passed into every I/O call
- [ ] Errors are wrapped: `fmt.Errorf("... : %w", err)`
- [ ] Output goes through `ui.*` (not `fmt.Printf`)
- [ ] Validation goes through `flags.*`
- [ ] Saving goes through `output.*`
- [ ] Registered in `commands.go` via `Register(rootCmd, clientFn)`
- [ ] Tests: ≥ success + error + edge case (using `testhelper`)
- [ ] No duplication (≥ 3 occurrences → extract a helper)

### A new API method

- [ ] Signature: `func (c *HTTPClient) GetXxx(ctx context.Context, ...)`
- [ ] Added to `interfaces.go` (the relevant interface)
- [ ] Added to `mock.go` (`MockClient`)
- [ ] List methods: `fetchAllPages[T]`
- [ ] Tested with `MockClient`

### General review

- [ ] No Cyrillic in code (only in `docs/` and `README`)
- [ ] `go vet ./...` — 0 warnings
- [ ] `go build ./...` — 0 errors
- [ ] `go test ./...` — 0 FAIL
- [ ] `go test -race ./...` — 0 data races
- [ ] Paths via `paths.*` (not hard-coded)
- [ ] Logging: `ui.*` for the terminal, `log.L()` for files

---

← [Architecture](index.md) · [Documentation](../index.md)
