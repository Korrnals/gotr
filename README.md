```
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

# gotr — CLI Client for TestRail API

[English](README.md) | [Русский](README_ru.md)

[![Version](https://img.shields.io/badge/version-2.7.0-blue.svg)](CHANGELOG.md)
[![Go Version](https://img.shields.io/badge/go-1.25.6-blue.svg)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A professional command-line interface for TestRail API v2. Designed for QA engineers and test automation specialists who need efficient data management, migration capabilities, and seamless integration with CI/CD pipelines.

> **Latest Release: v2.7.0** — Stage 4 Complete: 106/106 API endpoints implemented. See [CHANGELOG](CHANGELOG.md) for details

## Overview

`gotr` provides a comprehensive toolkit for TestRail operations:

- **Data Operations** — Retrieve and manage test cases, suites, sections, shared steps, runs, results, milestones, plans, and more
- **Complete API Coverage** — All 106 TestRail API v2 endpoints implemented (Stage 4 complete)
- **Project Synchronization** — Migrate entities between projects with intelligent duplicate detection
- **Interactive Workflow** — Guided selection of projects and suites eliminates the need to memorize IDs
- **Real-time Progress** — Visual progress bars with channel-based updates for all long-running operations
- **Built-in Processing** — JSON filtering with embedded `jq`, progress tracking, and structured logging
- **Flexible Configuration** — Support for flags, environment variables, and configuration files

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
| **Full API Coverage** | 106/106 TestRail API v2 endpoints implemented |
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

## Documentation

- [Installation Guide](docs/installation.md)
- [Configuration](docs/configuration.md)
- [GET Commands](docs/get-commands.md)
- [SYNC Commands](docs/sync-commands.md)
- [Interactive Mode](docs/interactive-mode.md)
- [Progress Monitoring](docs/progress.md) — Universal progress system for long-running operations

## Project Structure

```
gotr/
├── cmd/                          # CLI commands
│   ├── common/                   #   Shared components
│   │   ├── client.go            #     Unified client access
│   │   └── flags.go             #     Common flag parsing
│   ├── get/                     #   GET commands (cases, suites, projects)
│   ├── run/                     #   Test run management
│   ├── result/                  #   Test results management
│   └── sync/                    #   Data migration commands
├── docs/                         # Documentation
│   ├── architecture.md          #   Detailed architecture
│   ├── get-commands.md          #   GET command reference
│   ├── sync-commands.md         #   SYNC command reference
│   └── ...
├── internal/
│   ├── client/                  #   TestRail API client
│   │   ├── interfaces.go       #     ClientInterface (106 endpoints, 14 APIs)
│   │   ├── mock.go             #     MockClient for testing
│   │   └── *.go                #     API implementations
│   ├── interactive/            #   Interactive selection
│   ├── service/                #   Business logic
│   │   ├── run.go              #     RunService
│   │   ├── result.go           #     ResultService
│   │   └── migration/          #     Data migration engine
│   ├── models/                 #   Data models
│   │   └── data/              #     API DTOs
│   └── utils/                  #   Utilities
├── pkg/                          # Public packages
│   └── testrailapi/            #   API endpoint definitions
└── main.go                       # Entry point
```

See [docs/architecture.md](docs/architecture.md) for complete structure.

## What's New in v2.7.0 (Stage 4 Complete)

### Complete API Coverage

All 106 TestRail API v2 endpoints are now implemented:

- **Tests API** (3 endpoints) — GetTest, GetTests, UpdateTest
- **Milestones API** (5 endpoints) — Full CRUD for milestones
- **Plans API** (9 endpoints) — Full CRUD + plan entries management
- **Attachments API** (5 endpoints) — File uploads to cases/plans/results/runs
- **Configurations API** (7 endpoints) — Test configurations management
- **Users & Reference** (7 endpoints) — Users, Priorities, Statuses, Templates
- **Reports API** (3 endpoints) — Report generation and retrieval
- **Extended APIs** (21 endpoints) — Groups, Roles, Datasets, Variables, BDDs, Labels

## What's New in v2.5.0

### Architecture Improvements
- **Unified Client Interface** — Single `ClientInterface` across all packages eliminates code duplication
- **Enhanced Test Coverage** — All sync tests now use interface-based mocking (10 new tests, 0 skipped)
- **Refactored Common Package** — Eliminated `getClientSafe` duplication across command packages

### Interactive Features
- **Interactive Selection** — Visual pickers for projects and suites in `run list` and `result list`
- **Streamlined Workflow** — Reduced friction for common operations

See [CHANGELOG](CHANGELOG.md) for complete history.

## Installation

Detailed installation instructions: [docs/installation.md](docs/installation.md)

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.

## Acknowledgements

This project is built with the following open-source libraries:

| Library | Purpose |
|---------|---------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [spf13/viper](https://github.com/spf13/viper) | Configuration management |
| [cheggaaa/pb/v3](https://github.com/cheggaaa/pb) | Progress bars |
| [go.uber.org/zap](https://github.com/uber-go/zap) | Structured logging |
| [stretchr/testify](https://github.com/stretchr/testify) | Testing toolkit |
| [itchyny/gojq](https://github.com/itchyny/gojq) | Embedded JSON processor |

## License

MIT License — see [LICENSE](LICENSE)
