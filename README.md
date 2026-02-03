# gotr — CLI Client for TestRail API

[English](README.md) | [Русский](README_ru.md)

`gotr` is a powerful and convenient command-line utility for working with TestRail API v2.  
It allows you to perform GET requests, export data to files, synchronize entities between projects, filter responses through the built-in `jq`, and much more — without the need to install external dependencies.

## 🙏 Acknowledgements

This project uses the following amazing open-source libraries:

- **[spf13/cobra](https://github.com/spf13/cobra)** — CLI application framework
- **[spf13/viper](https://github.com/spf13/viper)** — configuration and environment variables
- **[cheggaaa/pb/v3](https://github.com/cheggaaa/pb)** — progress bars
- **[go.uber.org/zap](https://github.com/uber-go/zap)** — high-performance logging
- **[stretchr/testify](https://github.com/stretchr/testify)** — testing toolkit
- **[embedded jq](https://github.com/itchyny/gojq)** — built-in jq utility for JSON filtering

## 📁 Project Structure

```bash
gotr/
├── cmd/                    # CLI commands
│   ├── get/               # GET commands (cases, suites, projects, etc.)
│   ├── sync/              # SYNC commands (data migration)
│   ├── commands.go        # Centralized command registration
│   ├── root.go            # Root command and configuration
│   ├── config.go          # Config management commands
│   ├── list.go            # List command
│   └── ...                # Other commands
├── docs/                   # Documentation
│   ├── installation.md
│   ├── configuration.md
│   ├── get-commands.md
│   ├── sync-commands.md
│   └── ...
├── embedded/               # Embedded utilities (jq)
├── internal/               # Internal packages
│   ├── client/            # HTTP client for TestRail API
│   ├── migration/         # Migration logic (sync)
│   ├── models/            # Data structures
│   └── utils/             # Utilities
├── pkg/                    # Public packages
├── main.go                 # Entry point
├── go.mod                  # Go modules
└── Makefile               # Build automation
```

## 🚀 Quick Start

```bash
# Installation (Linux/macOS)
curl -s -L https://github.com/Korrnals/gotr/releases/latest/download/gotr-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o gotr && chmod +x gotr && sudo mv gotr /usr/local/bin/

# Verify
gotr --help
```

## ✨ Key Features

- 📡 **Full TestRail API Support** — GET requests to all endpoints
- 🔄 **Synchronization** — migrate cases, shared steps, suites, sections between projects
- 🎯 **Interactive Mode** — no need to remember project and suite IDs
- 📦 **Built-in jq** — filtering without installing external utilities
- 💾 **Export** — save data to JSON with automatic naming
- 🔧 **Flexible Configuration** — flags, env variables, config file
- 🖥️ **Auto-completion** — bash/zsh/fish completion

## 📚 Documentation

Detailed documentation is available in the [`docs/`](docs/) directory:

- [Installation](docs/installation.md)
- [Configuration](docs/configuration.md)
- [GET Commands](docs/get-commands.md)
- [SYNC Commands](docs/sync-commands.md)
- [Interactive Mode](docs/interactive-mode.md)
- [Other Commands](docs/other-commands.md)

## 🎮 Usage Examples

### Interactive Mode

```bash
# Get cases — interactive selection of project and suite
gotr get cases

# Sync cases — interactive selection of source and destination
gotr sync cases

# Full migration
gotr sync full
```

### Getting Data

```bash
# All projects
gotr get projects

# Project cases (with interactive suite selection)
gotr get cases 30

# Or with explicit suite ID
gotr get cases 30 --suite-id 20069

# All cases from all suites in project
gotr get cases 30 --all-suites

# Shared steps
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

# Cases only (with mapping file)
gotr sync cases \
  --src-project 30 --src-suite 20069 \
  --dst-project 31 --dst-suite 19859 \
  --mapping-file mapping.json --approve
```

### Comparing Projects

```bash
# Compare cases between two projects
gotr compare cases --pid1 30 --pid2 31 --field title
```

### Filtering with jq

```bash
# Only id and name of projects
gotr get projects --jq --jq-filter '.[] | {id: .id, name: .name}'

# Pretty output with jq
gotr get case 12345 --jq
```

## ⚙️ Configuration

Configuration priority (from highest to lowest):

1. **Flags** (`--url`, `--username`, `--api-key`)
2. **Env variables** (`TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`)
3. **Config file** (`~/.gotr/config.yaml`)

```bash
# Create config
gotr config init

# View config
gotr config view
```

## 🆕 What's New

### 2026-02-03 — Interactive Mode

- **Interactive selection** for all `get` and `sync` commands — no need to remember IDs
- **Auto-selection** when project has only one suite
- **`--all-suites` flag** for getting cases from all suites
- **Restructuring** of `cmd/` package — improved code organization

### 2026-01-24 — Sync Commands

- New commands `sync suites` and `sync sections`
- Unified flags for all `sync/*` commands
- Unit tests for synchronization

### 2026-01-15 — Get Commands v2.0

- Redesigned `get` command with subcommands
- Positional arguments for IDs
- Improved typing (int64)

## 📦 Installation

See [docs/installation.md](docs/installation.md)

## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📄 License

MIT License — see [LICENSE](LICENSE)
