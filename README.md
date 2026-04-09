# gotr — CLI Client for TestRail API

```text
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║     ██████╗  ██████╗ ████████╗██████╗                    ║
║    ██╔════╝ ██╔═══██╗╚══██╔══╝██╔══██╗                   ║
║    ██║  ███╗██║   ██║   ██║   ██████╔╝                   ║
║    ██║   ██║██║   ██║   ██║   ██╔══██╗                   ║
║    ╚██████╔╝╚██████╔╝   ██║   ██║  ██║                   ║
║     ╚═════╝  ╚═════╝    ╚═╝   ╚═╝  ╚═╝                   ║
║                                                          ║
║           CLI Client for TestRail API v2                 ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
```

[English](README.md) | [Русский](README_ru.md)

[![Version](https://img.shields.io/badge/version-3.0.0-blue.svg)](CHANGELOG.md)
[![Go Version](https://img.shields.io/badge/go-1.25.0-blue.svg)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A professional command-line interface for TestRail API v2. Designed for QA engineers and test automation specialists who need efficient data management, migration capabilities, and seamless integration with CI/CD pipelines.

> **Latest Release: v3.0.0** — Stage 13.5 Complete: Quality hardening, 7 audit rounds, 0 lint findings, full test coverage. See [CHANGELOG](CHANGELOG.md) for details

## Overview

`gotr` provides a comprehensive toolkit for TestRail operations:

- **Data Operations** — Retrieve and manage test cases, suites, sections, shared steps, runs, results, milestones, plans, and more
- **Complete API Coverage** — All 121 TestRail API v2 endpoints implemented (Stage 4 complete)
- **Project Synchronization** — Migrate entities between projects with intelligent duplicate detection
- **Interactive Workflow** — Guided selection of projects and suites eliminates the need to memorize IDs
- **Real-time Progress** — Visual progress bars with channel-based updates for all long-running operations
- **Built-in Processing** — JSON filtering with embedded `jq`, progress tracking, and structured logging
- **Flexible Configuration** — Support for flags, environment variables, and configuration files

## Navigation

- [Documentation](docs/index.md)
  - [Guides](docs/en/guides/index.md)
    - [Installation](docs/en/guides/installation.md)
    - [Configuration](docs/en/guides/configuration.md)
    - [Interactive Mode](docs/en/guides/interactive-mode.md)
    - [Progress](docs/en/guides/progress.md)
    - [Commands Index](docs/en/guides/commands/index.md)
      - [Command groups](docs/en/guides/commands/index.md#command-groups-and-subgroups)
    - [Instructions](docs/en/guides/instructions/index.md)
  - [Architecture](docs/en/architecture/index.md)
  - [Operations](docs/en/operations/index.md)
  - [Reports](docs/en/reports/index.md)
- [Home](README.md)

## Quick Start

```bash
# Install (Linux/macOS)
curl -sL https://github.com/Korrnals/gotr/releases/latest/download/gotr-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o gotr
chmod +x gotr && sudo mv gotr /usr/local/bin/

# Initialize configuration
gotr config init

# Verify installation
gotr self-test
```

## Key Features

| Feature | Description |
|---------|-------------|
| **Full API Coverage** | 121/121 TestRail API v2 endpoints implemented |
| **Interactive Mode** | Visual selection for projects, suites, and migration targets |
| **Data Synchronization** | Migrate cases, shared steps, suites, and sections between projects |
| **Test Run Management** | Create runs, add results, and track test execution |
| **Built-in jq** | Filter and transform JSON without external dependencies |
| **Real-time Progress** | Channel-based progress bars with live updates for parallel operations |
| **Shell Completion** | Auto-completion for bash, zsh, and fish |
| **Comprehensive Logging** | Structured JSON logs for audit and debugging |

## Usage Examples

### Interactive Mode

```bash
# Get cases with interactive project/suite selection
gotr get cases

# Sync with guided workflow
gotr sync full
```

### Data Retrieval

```bash
# List all projects
gotr get projects

# Get cases from specific project and suite
gotr get cases 30 --suite-id 20069

# Get cases from all suites in project
gotr get cases 30 --all-suites

# Get shared steps
gotr get sharedsteps 30
```

### Synchronization

```bash
# Full migration (shared steps + cases)
gotr sync full \
  --src-project 30 --src-suite 20069 \
  --dst-project 31 --dst-suite 19859 \
  --approve --save-mapping

# Shared steps only
gotr sync shared-steps \
  --src-project 30 --dst-project 31 \
  --approve --save-mapping

# Cases with existing mapping
gotr sync cases \
  --src-project 30 --src-suite 20069 \
  --dst-project 31 --dst-suite 19859 \
  --mapping-file mapping.json --approve
```

### Project Comparison

Compare resources between two projects to identify differences and similarities:

```bash
# Compare all resources between projects
gotr compare all --pid1 30 --pid2 34

# Compare specific resource types
gotr compare cases --pid1 30 --pid2 34
gotr compare suites --pid1 30 --pid2 34
gotr compare sharedsteps --pid1 30 --pid2 34

# Save comparison results
gotr compare all --pid1 30 --pid2 34 --save
gotr compare cases --pid1 30 --pid2 34 --save-to results.json --format json

# Auto-detect format from file extension
gotr compare all --pid1 30 --pid2 34 --save-to comparison.yaml
```

**Supported resources:** `cases`, `suites`, `sections`, `sharedsteps`, `runs`, `plans`, `milestones`, `datasets`, `groups`, `labels`, `templates`, `configurations`, `all`

#### Performance Tuning

```bash
# Server (без rate-limit, максимальная скорость)
gotr compare cases --pid1 30 --pid2 34 --rate-limit 0

# Cloud Enterprise (повышенный лимит)
gotr compare cases --pid1 30 --pid2 34 --rate-limit 300

# Больше параллелизма
gotr compare cases --pid1 30 --pid2 34 --parallel-suites 10 --parallel-pages 6
```

Automatic deployment detection: gotr определяет `cloud/server` по URL и подбирает rate-limit автоматически. Настраивается в конфиге (`compare.deployment`, `compare.cloud_tier`).

#### Точечный дозабор failed pages

```bash
# Если часть страниц не загрузилась — дозабрать только их
gotr compare retry-failed-pages --from ~/.gotr/exports/compare/failed_pages_2026-03-03_10-15-00.json
```

По умолчанию compare cases автоматически пытается дозабрать проблемные страницы.

### Test Runs and Results

```bash
# Create test run
gotr run add 30 --name "Regression Suite" --case-ids "1,2,3,4,5"

# Add test result
gotr result add 12345 --status-id 1 --comment "Test passed"

# List test results
gotr result list --run-id 100
```

### JSON Filtering

```bash
# Extract specific fields
gotr get projects --jq --jq-filter '.[] | {id: .id, name: .name}'

# Pretty print with jq
gotr get case 12345 --jq
```

## Debugging

For troubleshooting and detailed execution information, use the `--debug` (or `-d`) flag:

```bash
# Show debug output for any command
gotr compare cases --pid1 30 --pid2 34 --debug
gotr sync cases --src-project 30 --dst-project 31 --debug
gotr get cases --project-id 30 --debug

# Debug output includes:
# - API request details
# - Progress tracking information
# - Timing for each operation phase
# - Suite/case processing details
```

> **Note:** The `--debug` flag is hidden from autocompletion but available in all commands.

## Configuration

Configuration priority (highest to lowest):

1. **Command-line flags** (`--url`, `--username`, `--api-key`)
2. **Environment variables** (`TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`)
3. **Configuration file** (`~/.gotr/config/default.yaml`)

```bash
# Initialize configuration
gotr config init

# View current configuration
gotr config view
```

## Project Structure

```text
gotr/
├── cmd/                          # CLI commands (29 subcommands)
│   ├── internal/testhelper/     #   Shared test utilities
│   ├── get/                     #   GET commands (cases, suites, projects)
│   ├── run/                     #   Test run management
│   ├── result/                  #   Test results management
│   ├── compare/                 #   Cross-project comparison
│   ├── sync/                    #   Data migration commands
│   └── ...                      #   Other resource subcommands
├── docs/                         # Documentation (EN + RU)
│   ├── en/                      #   English docs
│   └── ru/                      #   Russian docs
├── embedded/                     # Embedded binaries (jq)
├── internal/
│   ├── client/                  #   TestRail API client
│   │   ├── interfaces.go       #     ClientInterface (130+ methods, 16 APIs)
│   │   ├── mock.go             #     MockClient for testing
│   │   └── *.go                #     API implementations
│   ├── concurrency/            #   Domain-level parallel orchestration
│   │   ├── controller.go       #     ParallelController — suite/page streaming
│   │   └── simple.go           #     FetchParallel[T], FetchParallelBySuite[T]
│   ├── concurrent/             #   Low-level concurrency primitives
│   │   ├── pool.go             #     WorkerPool
│   │   ├── limiter.go          #     AdaptiveRateLimiter (180 req/min)
│   │   └── retry.go            #     Exponential backoff retry
│   ├── interactive/            #   Interactive prompts (survey)
│   ├── service/                #   Business logic
│   │   ├── run.go              #     RunService
│   │   ├── result.go           #     ResultService
│   │   └── migration/          #     Data migration engine
│   ├── models/                 #   Data models
│   │   ├── data/              #     API DTOs
│   │   └── config/            #     Configuration model
│   ├── output/                 #   Output formatting (JSON/YAML/table)
│   ├── ui/                     #   Terminal UI (progress, preview)
│   ├── flags/                  #   Common flag parsing
│   ├── log/                    #   Structured logging (zap)
│   └── paths/                  #   Path utilities
├── pkg/                          # Public packages
│   ├── testrailapi/            #   API endpoint definitions (135 endpoints)
│   └── reporter/               #   Unified statistics reporter
└── main.go                       # Entry point
```

See [docs/en/architecture/overview.md](docs/en/architecture/overview.md) for complete structure.

## What's New in v3.0.0

- **135 TestRail API endpoints** defined, 98% implemented in client
- **29 CLI commands** covering all major TestRail resources
- **Streaming parallel pagination** with adaptive rate limiting (180 req/min)
- **100% test coverage** in 35/42 packages, min 97.4% across all packages
- **Zero golangci-lint issues** with gocyclo ≤15 threshold
- **Full EN/RU documentation** with 125 doc pages

See [CHANGELOG](CHANGELOG.md) for full release history.

## Installation

Detailed installation instructions: [docs/en/guides/installation.md](docs/en/guides/installation.md)

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.

## Acknowledgements

This project is built with the following open-source libraries:

| Library | Purpose |
|---------|---------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [spf13/viper](https://github.com/spf13/viper) | Configuration management |
| [go.uber.org/zap](https://github.com/uber-go/zap) | Structured logging |
| [stretchr/testify](https://github.com/stretchr/testify) | Testing toolkit |
| [AlecAivazis/survey/v2](https://github.com/AlecAivazis/survey) | Interactive prompts |
| [jedib0t/go-pretty/v6](https://github.com/jedib0t/go-pretty) | Table output formatting |
| [fatih/color](https://github.com/fatih/color) | Colored terminal output |
| [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync) | Concurrency utilities |
| [golang.org/x/time](https://pkg.go.dev/golang.org/x/time) | Rate limiting |

### Embedded Tools

| Tool | Purpose |
|------|----------|
| [jq](https://github.com/jqlang/jq) | Lightweight JSON processor, embedded as a static binary for `--jq` / `--jq-filter` support |

## License

MIT License — see [LICENSE](LICENSE)
