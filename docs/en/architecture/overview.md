# gotr architecture

Language: [Русский](../../ru/architecture/overview.md) | English

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

## What gotr is

`gotr` is a CLI client for the TestRail API v2, built as a multi-layered architecture with a clear separation of responsibilities between layers.

## Overall diagram

```text
┌─────────────────────────────────────────────────────────────┐
│  CLI Layer (cmd/*)                                          │
│  • Argument and flag parsing (flags.*)                      │
│  • Interactive selection (internal/interactive)             │
│  • Output (ui.*, output.*)                                  │
│  • Calls services and the client                            │
└──────────────┬──────────────────────┬───────────────────────┘
               │                      │
    ┌──────────▼──────────┐  ┌────────▼──────────┐
    │  UI Layer           │  │  Flags Layer       │
    │  (internal/ui/)     │  │  (internal/flags/) │
    │  • Table, JSON      │  │  • ValidateID      │
    │  • Info, Success,   │  │  • GetFlag[T]      │
    │    Warning, Error   │  │  • ParseID         │
    │  • Live display     │  └───────────────────┘
    └─────────────────────┘
               │
┌──────────────▼──────────────────────────────────────────────┐
│  Service Layer (internal/service/*)                         │
│  • Business logic                                           │
│  • Data validation                                          │
│  • Data migration (migration)                               │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Concurrency Layer (internal/concurrency/*)                 │
│  • ParallelController — pipeline pagination over suites    │
│  • FetchParallel[T] — light strategy (across projects)     │
│  • FetchParallelBySuite[T] — medium strategy (per suite)   │
│  • ResultAggregator — collects results from goroutines     │
│  • PriorityQueue — prioritises large suites                │
│  • SuiteFetcher — interface for fetch implementations      │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Concurrent Layer (internal/concurrent/*)                   │
│  • WorkerPool — parallel request processing                │
│  • RateLimiter — caps at 180 requests/minute               │
│  • Retry — retries with exponential backoff                │
│  • CircuitBreaker — protection against cascading errors    │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Client Layer (internal/client/*)                           │
│  • HTTPClient — real client                                 │
│  • ClientInterface — abstraction for tests (106 endpoints)  │
│  • MockClient — testing implementation                      │
│  • fetchAllPages[T] — generic paginator (Stage 6.9)        │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Output Layer (internal/output/*)                           │
│  • OutputResult — output + save to file                    │
│  • DryRunPrinter — output for dry-run mode                 │
│  • --save / --save-to handling                              │
└─────────────────────────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  TestRail API v2                                            │
└─────────────────────────────────────────────────────────────┘
```

### Dependency rules

| Layer | May depend on | MUST NOT depend on |
| ---- | ----------------- | -------------------- |
| `cmd/*` | `internal/service`, `internal/client`, `internal/ui`, `internal/flags`, `internal/interactive`, `internal/output`, `pkg/*` | — |
| `internal/service` | `internal/client`, `internal/concurrency`, `internal/concurrent`, `internal/models` | `cmd/*`, `internal/ui` |
| `internal/client` | `internal/concurrent`, `internal/models/data` | `cmd/*`, `internal/service` |
| `internal/ui` | stdlib, `go-pretty/v6` | `internal/client`, `internal/service` |
| `internal/concurrency` | stdlib | `internal/client`, `cmd/*` |
| `pkg/*` | stdlib, `go-pretty/v6` | `internal/*`, `cmd/*` |

### Forbidden

- `service/` → `cmd/` (calling upwards)
- `client/` → `ui/` (the client knows nothing about UI)
- `pkg/` → `internal/` (public must not import private)

## Layers in detail

### 1. CLI Layer (`cmd/`)

**Responsibility:** Accepts user commands, parses arguments, calls services.

### Structure

```text
cmd/
├── root.go              # Root command, Execute(ctx)
├── commands.go          # Registers all subcommands (init())
├── add.go               # gotr add <resource>
├── update.go            # gotr update <resource>
├── delete.go            # gotr delete <resource>
├── list.go              # gotr list <resource>
├── export.go            # gotr export
├── config.go            # gotr config {init|path|view|edit}
├── resources.go         # gotr resources (API endpoints)
├── selftest.go          # gotr selftest
├── completion.go        # gotr completion {bash|zsh|fish}
├── <resource>/          # Subcommands for a resource
│   ├── <resource>.go   #   Register() + getClient()
│   ├── add.go          #   newAddCmd(clientFn)
│   ├── get.go          #   newGetCmd(clientFn)
│   ├── list.go         #   newListCmd(clientFn)
│   ├── update.go       #   newUpdateCmd(clientFn)
│   ├── delete.go       #   newDeleteCmd(clientFn)
│   └── *_test.go       #   Table-driven tests
├── compare/             # gotr compare <resource> --pid1 X --pid2 Y
├── sync/                # gotr sync {cases|sections|suites|...}
└── internal/            # Shared test helpers (testhelper)
```

### Command pattern (Stage 7.0+)

```go
func newXxxCmd(clientFn func(*cobra.Command) client.ClientInterface) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "xxx",
        Short: "Brief description",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            cli := clientFn(cmd)
            ctx := cmd.Context()

            id, err := flags.ValidateRequiredID(args, 0, "resource_id")
            if err != nil { return err }

            result, err := cli.GetXxx(ctx, id)
            if err != nil { return err }

            return output.OutputResult(cmd, result, "xxx")
        },
    }
    output.AddFlag(cmd)
    return cmd
}
```

### Example

```bash
gotr run get 12345 --jq
# cmd/run/get.go → client.GetRun(ctx, 12345) → output via output.OutputResult
```

### 2. Service Layer (`internal/service/`)

**Responsibility:** Business logic, validation, operation orchestration.

### Components

- `RunService` — handling test runs
- `ResultService` — handling test results
- `migration/` — migrating data between projects
  - `types.go` — migration context
  - `fetch.go` — data retrieval
  - `filter.go` — duplicate filtering
  - `import.go` — entity import
  - `export.go` — data export
  - `mapping.go` — ID mapping management

### Validation

```go
// Checks before creating a run:
// - projectID > 0
// - name not empty
// - suite_id > 0 (if specified)
```

### 3. Client Layer (`internal/client/`)

**Responsibility:** HTTP requests to the TestRail API.

### Structure

```text
internal/client/
├── client.go           # HTTPClient — constructor, DoRequest (http.NewRequestWithContext)
├── interfaces.go       # ClientInterface + 14 API groups (106 endpoints)
├── mock.go             # MockClient for testing
├── paginator.go        # Generic fetchAllPages[T] — auto-pagination for list methods (Stage 6.9)
├── request.go          # sendRequest(), debug output
├── accessor.go         # ClientAccessor — lazy init
├── concurrent.go       # Thread-safe wrappers
├── projects.go         # ProjectsAPI (5 endpoints)
├── cases.go            # CasesAPI (14 endpoints)
├── suites.go           # SuitesAPI (5 endpoints)
├── sections.go         # SectionsAPI (5 endpoints)
├── sharedsteps.go      # SharedStepsAPI (6 endpoints)
├── runs.go             # RunsAPI (6 endpoints)
├── results.go          # ResultsAPI (7 endpoints)
├── tests.go            # TestsAPI (3 endpoints)
├── milestones.go       # MilestonesAPI (5 endpoints)
├── plans.go            # PlansAPI (9 endpoints)
├── attachments.go      # AttachmentsAPI (5 endpoints)
├── configs.go          # ConfigurationsAPI (7 endpoints)
├── users.go            # UsersAPI (4 endpoints)
├── reports.go          # ReportsAPI (3 endpoints)
└── extended.go         # ExtendedAPI (21 endpoint)
```

### Key facts (Stage 7.0+)

- Every method takes `ctx context.Context` as the first argument
- List methods use `fetchAllPages[T]` for auto-pagination
- Cancellation through context (Ctrl+C → signal.NotifyContext)
- 14 ISP-aligned interfaces, 106 endpoints, 100% coverage of TestRail API v2

### 4. Concurrent Layer (`internal/concurrent/`)

**Responsibility:** Parallel processing of API requests with load control.

**Why it's needed:** The TestRail API has a limit of 180 requests per minute. Sequential processing of many suites/cases takes minutes. The Concurrent Layer allows you to:

- Run requests in parallel (up to 5 at a time)
- Automatically throttle the request rate (150 req/min)
- Automatically retry on errors
- Protect the API from overload

### Speed-up example

```text
Loading cases from 10 suites:
- Sequential:  10 requests × 1 sec = 10 seconds
- Parallel:    10 requests / 5 workers = ~2 seconds (5x speed-up)
```

### Components

- **WorkerPool** — manages a goroutine pool (3-5 workers)
- **RateLimiter** — token bucket (150 requests/minute)
- **Retry** — retry with exponential backoff (1s, 2s, 4s...)
- **CircuitBreaker** — blocks requests when errors pile up

More details: [docs/architecture/concurrency.md](./concurrency.md)

### 5. Concurrency Layer (`internal/concurrency/`)

**Responsibility:** Unified parallelization strategies for all `compare` subcommands.

### Components

- **FetchParallel[T]** — light strategy: load a resource from N projects in parallel (generic, Go 1.24+)
- **FetchParallelBySuite[T]** — medium strategy: load per-suite resources in parallel (sections etc.)
- **ParallelController** — heavy strategy: pipeline pagination over suites (cases)
- **ResultAggregator** — thread-safe aggregation of results from goroutines
- **PriorityQueue** — heap-based queue, large suites are processed first
- **SuiteFetcher** — interface (`FetchPageCtx`) so real and mock implementations can be plugged in
- **ProgressReporter** — universal progress interface (`OnItemComplete`, `OnBatchReceived`, `OnError`)
- **PaginatedProgressReporter** — extension for strategies with pagination (`OnPageFetched`)

**FetchOption pattern:** `WithReporter()`, `WithContinueOnError()`, `WithMaxConcurrency(n)`

**Configuration:** `--parallel-suites`, `--parallel-pages`, `--page-retries`, `--rate-limit`, `--timeout`

More details: [docs/architecture/recursive-parallelization-plan.md](./recursive-parallelization-plan.md)

### 6. UI Layer (`internal/ui/` + `pkg/reporter/`)

**Responsibility:** Unified output for all commands (Stage 8.0).

### Components

- **internal/ui/display.go** — ANSI live display with dynamic tasks (used by `compare cases`)
  - `New()`, `SetHeader()`, `AddTask()`, `Finish()` — lifecycle
  - Implements `ProgressReporter`, `PaginatedProgressReporter` from `concurrency/`

- **internal/ui/table.go** — static data output
  - `NewTable(cmd)` — go-pretty table honouring `--format` (table/json/csv/md/html)
  - `Table(cmd, t)` — table rendering
  - `JSON(cmd, data)` — JSON output honouring `--quiet`
  - `IsJSON(cmd)`, `IsQuiet(cmd)` — format checks

- **internal/ui/helpers.go** — styled messages (Stage 8.0)
  - `Info(w, msg)` — ℹ️ information
  - `Success(w, msg)` — ✅ success
  - `Warning(w, msg)` — ⚠️ warning
  - `Error(w, msg)` — ❌ error
  - `Phase(w, msg)` — 🔄 phase
  - `Stat(w, icon, label, val)` — statistic
  - `Section(w, msg)` — section heading
  - `Preview(w, title, fields)` — preview pane

- **pkg/reporter/** — builder pattern for structured reports (ANSI + go-pretty)

### Rules

- All user-facing output goes through `ui.*` (except interactive prompts and debug)
- Emoji prefixes only inside `ui.*` — no hard-coded emoji in `cmd/`
- The first argument is always an `io.Writer` (typically `os.Stdout`)

### 7. Flags Layer (`internal/flags/`)

**Responsibility:** Type-safe validation of CLI arguments and flags.

### Functions

```go
flags.ValidateRequiredID(args, index, name)   // Parses an ID from args
flags.GetFlagInt64(cmd, name)                 // int64 flag
flags.GetFlagString(cmd, name)                // string flag
flags.GetFlagBool(cmd, name)                  // bool flag
flags.ParseID(s)                              // string → int64
```

### 8. Output Layer (`internal/output/`)

**Responsibility:** Saving results to files, dry-run, formatting.

### Functions

```go
output.AddFlag(cmd)                       // Registers --save, --save-to
output.OutputResult(cmd, data, resource)  // Output + save
output.Output(cmd, data, dir, format)     // Save to ~/.gotr/exports/
output.NewDryRunPrinter(cmd)              // Output for dry-run
```

### Usage

- `gotr compare cases` — live display with real-time progress
- All 13 compare subcommands — reporter for the final summary
- `gotr compare all` — go-pretty table + reporter for the summary table
- `ui.Infof(os.Stderr, ...)` — styled messages

### 7. Progress Runtime (`internal/ui/runtime.go`)

**Responsibility:** A single runtime for progress and statuses across CLI commands.

### Usage

- `ui.RunWithStatus(...)` — simple status-style operations
- `ui.NewOperation(...)` + `AddTask(...)` — multi-phase and streaming operations

### Highlights

- Honours `--quiet`
- Supports phases and task-level progress
- Used by compare/sync/get and other commands

### 8. Interactive Layer (`internal/interactive/`)

**Responsibility:** Interactive selection of projects, suites and runs.

### Usage

- `gotr run list` — pick a project → list of runs
- `gotr result list` — pick a project → pick a run → results
- `gotr get cases` — pick a project → pick a suite

### 9. Models (`internal/models/data/`)

**Responsibility:** DTOs (Data Transfer Objects) for the API.

### Main structures

- `Project`, `Suite`, `Section`, `Case`
- `Run`, `Test`, `Result`
- `SharedStep`, `Milestone`, `Plan`
- `Attachment`, `Config`, `User`
- `Report`, `Group`, `Role`, `Dataset`
- `Status`, `Priority` — constants

### 10. Utilities (`internal/utils/`)

**Responsibility:** Helper functions.

### Components

- `helpers.go` — ID parsing, result output, saving to a file
- `log.go` — log directories

## Data flow

### Example 1: Creating a test run

```text
User
    ↓
gotr run create 30 --suite-id 100 --name "Smoke"
    ↓
CLI Layer (cmd/run/create.go)
    ↓
RunService.Create(projectID=30, req={suite_id:100, name:"Smoke"})
    ↓
Validation: projectID>0? suite_id>0? name not empty?
    ↓
HTTPClient.AddRun(30, req)
    ↓
POST /index.php?/api/v2/add_run/30
    ↓
TestRail API
```

### Example 2: Data migration (sync full)

```text
User
    ↓
gotr sync full --src-project 30 --dst-project 31
    ↓
CLI Layer (cmd/sync/sync_full.go)
    ↓
migration.NewMigration(client, 30, 0, 31, 0, "title", logDir)
    ↓
Migration.FetchSharedStepsData()  // Fetch data
    ↓
Migration.FilterSharedSteps()     // Filter duplicates
    ↓
Migration.ImportSharedSteps()     // Import
    ↓
Same flow for cases
    ↓
TestRail API (src) → Migration → TestRail API (dst)
```

## Full project structure

```text
gotr/
├── cmd/                          # CLI commands (Cobra)
│   ├── root.go                  #   Execute(ctx), initConfig()
│   ├── commands.go              #   init() — registers all subcommands
│   ├── add.go                   #   gotr add <endpoint>
│   ├── update.go                #   gotr update <endpoint>
│   ├── delete.go                #   gotr delete <endpoint>
│   ├── list.go                  #   gotr list <resource>
│   ├── export.go                #   gotr export <resource>
│   ├── config.go                #   gotr config {init|path|view|edit}
│   ├── resources.go             #   gotr resources
│   ├── selftest.go              #   gotr selftest
│   ├── completion.go            #   gotr completion {bash|zsh|fish}
│   ├── attachments/             #   gotr attachments {add|get|list|delete}
│   ├── bdds/                    #   gotr bdds {add|get}
│   ├── cases/                   #   gotr cases {add|get|list|update|delete|bulk}
│   ├── compare/                 #   gotr compare {cases|suites|sections|...}
│   ├── configurations/          #   gotr configurations {add|list|update|delete}
│   ├── datasets/                #   gotr datasets
│   ├── get/                     #   gotr get {project|suite|case|...}
│   ├── groups/                  #   gotr groups
│   ├── labels/                  #   gotr labels
│   ├── milestones/              #   gotr milestones
│   ├── plans/                   #   gotr plans
│   ├── reports/                 #   gotr reports
│   ├── result/                  #   gotr result {get|list|add}
│   ├── roles/                   #   gotr roles
│   ├── run/                     #   gotr run {get|list|create|update|close|delete}
│   ├── sync/                    #   gotr sync {full|cases|shared-steps|suites|sections}
│   ├── templates/               #   gotr templates
│   ├── test/                    #   gotr test
│   ├── tests/                   #   gotr tests
│   ├── users/                   #   gotr users
│   ├── variables/               #   gotr variables
│   └── internal/                #   Test helpers (testhelper)
├── docs/                         # Documentation (Russian)
│   ├── index.md                 #   Documentation hub
│   ├── guides/                  #   Installation, configuration, commands
│   │   ├── commands/            #     GET/SYNC/other
│   │   ├── interactive-mode.md  #     Interactive mode
│   │   └── progress.md          #     Progress system
│   ├── architecture/            #   Architecture and standards
│   │   ├── overview.md          #     This file
│   │   ├── concurrency.md       #     Parallel processing
│   │   ├── standards.md         #     Coding standards
│   │   └── recursive-parallelization-plan.md
│   ├── operations/              #   Release process
│   └── reports/                 #   Audits and stage reports
├── embedded/                     # Embedded utilities
│   └── jq_embed.go             #   Embedded jq
├── internal/                     # Internal code
│   ├── client/                  #   HTTP client for the TestRail API
│   │   ├── client.go           #     HTTPClient (DoRequest + http.NewRequestWithContext)
│   │   ├── interfaces.go       #     ClientInterface (14 interfaces, 106 endpoints)
│   │   ├── mock.go             #     MockClient for tests
│   │   ├── paginator.go        #     fetchAllPages[T] — generic paginator
│   │   ├── accessor.go         #     ClientAccessor (lazy init)
│   │   └── <domain>.go         #     Endpoints grouped by domain
│   ├── concurrent/             #   Parallel processing (low-level)
│   │   ├── pool.go            #     WorkerPool, ParallelMap
│   │   ├── limiter.go         #     RateLimiter (150 req/min)
│   │   ├── retry.go           #     Retry with backoff
│   │   └── circuit.go         #     CircuitBreaker
│   ├── concurrency/            #   Parallelization strategies (high-level)
│   │   ├── types.go           #     ProgressReporter, FetchOption
│   │   ├── fetch_parallel.go  #     FetchParallel[T]
│   │   ├── fetch_by_suite.go  #     FetchParallelBySuite[T]
│   │   ├── controller.go      #     ParallelController (pipeline)
│   │   ├── priority_queue.go  #     PriorityQueue (heap)
│   │   └── aggregator.go      #     ResultAggregator
│   ├── ui/                     #   Unified output (Stage 8.0)
│   │   ├── display.go         #     ANSI live display + Task reporter
│   │   ├── table.go           #     Table(), JSON(), NewTable(), --format
│   │   └── helpers.go         #     Info, Success, Warning, Error, Phase, Preview
│   ├── flags/                  #   Flag and argument validation (Stage 8.0)
│   │   └── helpers.go         #     ValidateRequiredID, GetFlag*, ParseID
│   ├── output/                 #   Saving results
│   │   ├── save.go            #     OutputResult, AddFlag, SaveToFile
│   │   ├── dryrun.go          #     DryRunPrinter
│   │   ├── filename.go        #     GenerateTimestamp, BuildFilename
│   │   └── paths.go           #     GetExportsDir, EnsureDir
│   ├── interactive/            #   Interactive selection
│   │   ├── interactive.go     #     SelectProject, SelectSuite, SelectRun
│   │   └── wizard.go          #     InteractiveWizard
│   ├── paths/                  #   Path management
│   │   └── paths.go           #     BaseDir, ConfigFile, EnsureAllDirs
│   ├── log/                    #   Structured logging (zap)
│   │   └── logger.go          #     Init, L(), WithField, WithFields
│   ├── selftest/               #   Self-diagnostics
│   │   ├── types.go           #     CheckResult, Report
│   │   └── checks.go          #     ConfigChecker, BaseDirChecker, ...
│   ├── service/                #   Business logic
│   │   ├── run.go              #     RunService
│   │   ├── result.go           #     ResultService
│   │   └── migration/          #     Data migration
│   │       ├── types.go       #       Migration struct
│   │       ├── fetch.go       #       Data retrieval
│   │       ├── filter.go      #       Filtering
│   │       ├── import.go      #       Import
│   │       ├── export.go      #       Export
│   │       ├── mapping.go     #       ID mapping
│   │       └── log.go         #       Logging
│   ├── models/                 #   Data models
│   │   └── data/              #     DTOs for the TestRail API
│   └── utils/                  #   Utilities (legacy, shrinking)
│       ├── helpers.go         #     ParseID, SaveToFile, LoadMapping
│       └── log.go             #     LogDir
├── pkg/                          # Public packages
│   ├── reporter/               #   Builder-pattern reporter (Section/Stat/Print)
│   │   └── reporter.go
│   └── testrailapi/            #   API endpoint descriptions
│       └── api_paths.go
├── main.go                       # Entry point (signal.NotifyContext + ExecuteContext)
├── go.mod                        # Go modules
├── Makefile                     # Build
├── CHANGELOG.md                 # Change history
└── README.md                    # Main documentation
```

## Available commands

### Fetching data (`gotr get`)

| Command | Description |
|---------|----------|
| `gotr get projects` | All projects |
| `gotr get project <id>` | A specific project |
| `gotr get suites [project-id]` | Project suites |
| `gotr get suite <id>` | A specific suite |
| `gotr get cases [project-id]` | Cases (interactive suite picker) |
| `gotr get case <id>` | A specific case |
| `gotr get sharedsteps <project-id>` | Shared steps |
| `gotr get sections <project-id>` | Sections |

### Managing test runs (`gotr run`)

| Command | Description |
|---------|----------|
| `gotr run get <id>` | Get information about a run |
| `gotr run list [project-id]` | List runs (interactive selection) |
| `gotr run create <project-id>` | Create a run |
| `gotr run update <id>` | Update a run |
| `gotr run close <id>` | Close a run |
| `gotr run delete <id>` | Delete a run |

### Managing results (`gotr result`)

| Command | Description |
|---------|----------|
| `gotr result get <test-id>` | Get test results |
| `gotr result get-case <run-id> <case-id>` | Get case results |
| `gotr result list [--run-id <id>]` | List results (interactive) |
| `gotr result add <test-id>` | Add a result |
| `gotr result add-case <run-id>` | Add a result for a case |
| `gotr result add-bulk <run-id>` | Bulk add from a file |

### Comparing projects (`gotr compare`)

| Command | Description |
|---------|----------|
| `gotr compare cases --pid1 X --pid2 Y` | Compare cases (in parallel) |
| `gotr compare suites --pid1 X --pid2 Y` | Compare suites |
| `gotr compare sections --pid1 X --pid2 Y` | Compare sections |
| `gotr compare runs --pid1 X --pid2 Y` | Compare runs |
| `gotr compare all --pid1 X --pid2 Y` | Full comparison (13 resources) |

### Data migration (`gotr sync`)

| Command | Description |
|---------|----------|
| `gotr sync full` | Full migration (shared-steps + cases) |
| `gotr sync cases` | Migrate cases |
| `gotr sync shared-steps` | Migrate shared steps |
| `gotr sync suites` | Migrate suites |
| `gotr sync sections` | Migrate sections |

### Other commands

| Command | Description |
|---------|----------|
| `gotr add <endpoint>` | POST requests to the API |
| `gotr update <endpoint>` | POST/PATCH requests |
| `gotr delete <endpoint>` | DELETE requests |
| `gotr list <resource>` | List API endpoints |
| `gotr export <resource>` | Export data (JSON) |
| `gotr config init` | Initialise the configuration |
| `gotr selftest` | Self-diagnostics |
| `gotr completion {bash\|zsh\|fish}` | Shell completion |

### Global flags

| Flag | Short | Type | Description |
|------|----------|-----|----------|
| `--url` | — | string | TestRail instance URL |
| `--username` | `-u` | string | User email |
| `--api-key` | `-k` | string | API key |
| `--config` | `-c` | bool | Create a default config |
| `--format` | `-f` | string | Format: `table`, `json`, `csv`, `md`, `html` |
| `--quiet` | `-q` | bool | Quiet mode (CI/CD) |
| `--debug` | `-d` | bool | Debug output (hidden) |
| `--insecure` | — | bool | Skip TLS verification |
| `--save` | — | bool | Save the result to a file |
| `--save-to` | — | string | Save to a specific path |

## Why this architecture

### Benefits

1. **Clear separation** — each layer only knows about its own level
2. **Testability** — services can be tested with MockClient without real HTTP requests
3. **Extensibility** — adding a new command does not require changes in the client
4. **Reuse** — a single service is used across different commands
5. **Interactivity** — a single selection mechanism in `internal/interactive/`

### Adding retry

If TestRail returns "rate limit", retry is added in the Service Layer only:

```go
func (s *RunService) Get(id int64) (*data.Run, error) {
    var run *data.Run
    err := retry.Do(3, func() error {
        var err error
        run, err = s.client.GetRun(id)
        return err
    })
    return run, err
}
```

CLI commands need no changes!

## For developers

Where to make changes:

| Task | Location |
|--------|---------|
| New command | `cmd/<group>/*.go` |
| New validation | `internal/service/*.go` |
| New API method | `internal/client/*.go` + `interfaces.go` |
| New data structure | `internal/models/data/*.go` |
| Interactive selection | `internal/interactive/wizard.go` |
| Parallel processing (generic) | `internal/concurrent/*.go` |
| Concurrency strategies | `internal/concurrency/*.go` |
| Generic compare factory | `cmd/compare/simple.go` |
| Unified output (compare) | `pkg/reporter/*.go` |
| Progress runtime | `internal/ui/runtime.go` |

Detailed technical documentation: `.github/copilot/instructions/`

### Parallel processing in commands

To speed up operations over many suites/cases, use `internal/concurrent`:

```go
// Example: parallel loading of cases
import "github.com/Korrnals/gotr/internal/concurrent"

func fetchAllCases(client ClientInterface, projectID int64) ([]Case, error) {
    suites, _ := client.GetSuites(projectID)
    
    // Load cases from all suites in parallel
    results, err := concurrent.ParallelMap(suites, 5,
        func(suite Suite, index int) ([]Case, error) {
            return client.GetCases(projectID, suite.ID, 0)
        })
    
    // Collect results
    var allCases []Case
    for _, r := range results {
        if r.Error == nil {
            allCases = append(allCases, r.Data...)
        }
    }
    return allCases, err
}
```

Rate limiting (150 req/min) and retry are applied automatically.

---

← [Architecture](index.md) · [Documentation](../index.md)
