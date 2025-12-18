# gotr - CLI utility for TestRail API

[English](README.md) | [Русский](README_ru.md)

`gotr` is a powerful and convenient command line tool for working with TestRail API v2.  
Allows you to perform GET requests, export data to files, filter responses through the built-in `jq` and much more - without the need to install external dependencies.

## Features

- Full support for TestRail API GET endpoints
- Built-in `jq` - filtering without installing an external utility
- Export data to JSON files (automatic naming or via `--output`)
- Auto-completion of resources and endpoints
- Flexible flags: `--quiet`, `--type`, `--jq`, `--project-id` and others
- Support for query parameters (suite_id, section_id, etc.)
- Fully self-contained binary - works anywhere Go runs

## Installation

### Download the finished binary with one command (Linux/macOS)

```bash
# Unix
curl -s -L https://github.com/Korrnals/gotr/releases/latest/download/gotr-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o gotr && chmod +x gotr &&
 sudo mv gotr /usr/local/bin/
```

> [!TIP] Note
>
> Replace ***latest*** with a specific version if necessary (for example, ***v1.0.0***).\
> For Windows - download the .exe manually from Releases.
>
> ***Binaries for Linux, macOS and Windows will be available in [Releases](https://github.com/Korrnals/gotr/releases).***

### Option 1: From source (recommended)

```bash
# Clone the repository
git clone https://github.com/Korrnals/gotr.git
cd gotr

# Build the binary (optimized and compressed)
go build -ldflags="-s -w" -o gotr

# (Optional) Compress even more using UPX
upx --best gotr

# Move to PATH
sudo mv gotr /usr/local/bin/
```

### Option 2: Install via Makefile (recommended for developers)

**Makefile** makes it easy to build, test and install.

```bash
# Clone the repository
git clone https://github.com/Korrnals/gotr.git
cd gotr

# Build and install with one command
make install

# Other useful commands:
make build # just build
make test # run tests
make compress # compress UPX (if installed)
make build-compressed # build + compress
make clean # clean
make release # build for all platforms

make install # will build an optimized binary and install it in /usr/local/bin (requires sudo).

# If UPX is installed, use 
make build-compressed # for minimum size (~3-5 MB).

# Example of UPX compression for Windows package
make compress BINARY_NAME=gotr.exe
```

**Build with version:**

```bash
# No tag - version "dev"
make build
# gotr version - dev

# Make a tag
git tag v1.0.0
make build
# gotr version - v1.0.0

# Explicitly indicate the version - priority is higher than the tag
make build VERSION=test-123
# gotr version - test-123
```

---

### Installation on Windows

For Windows:

- No sudo, manual installation in PATH.
- Binary with .exe extension.
- UPX works on Windows.
- Curl one-liner - a little different (PowerShell or cmd).

#### Option 1: Download a ready-made binary with one command (PowerShell)

```powershell
Invoke-WebRequest -Uri https://github.com/Korrnals/gotr/releases/latest/download/gotr.exe -OutFile gotr.exe
# Make it executable (not necessarily on Windows, but for security)
# Move to a directory from PATH (for example, C:\Windows or user bin)
Move-Item gotr.exe C:\Windows\gotr.exe
```

#### Option 2: From source

```powershell
git clone https://github.com/Korrnals/gotr.git
cd gotr
go build -ldflags="-s -w" -o gotr.exe

# (Optional) UPX compression
upx --best gotr.exe

# Move to PATH
Move-Item gotr.exe C:\Windows\
```

#### Option 3: Via Makefile (requires Make for Windows, e.g. Chocolatey: choco install make)

```powershell
git clone https://github.com/Korrnals/gotr.git
cd gotr
make build # build gotr.exe
make compress # compress UPX (if installed)
# Manual installation:
Copy-Item gotr.exe C:\Windows\
```

> [!TIP] Note
> On **Windows** it is recommended to add the directory to **PATH** via "***Settings → System → About → Advanced system settings → Environment variables***".

## Configuration

`gotr` supports several authentication methods:

### Through flags

```bash
gotr --base-url https://your-company.testrail.io/ \
     --username your@email.com \
     --api-key your_api_key \
     get get_projects
```

### Through environment variables

```bash
export TESTRAIL_BASE_URL="https://your-company.testrail.io/"
export TESTRAIL_USERNAME="your@email.com"
export TESTRAIL_API_KEY="your_api_key"

gotr get get_projects
```

### Via config file (coming soon)

---
---

## Usage

### Basic commands

```bash
gotr get <endpoint> [id] # GET request
gotr export <resource> <endpoint> [id] # Export to file
gotr list <resource> # List of available endpoints
```

### Examples

#### Get list of projects

```bash
gotr get get_projects
gotr get get_projects -t table # in table form
gotr get get_projects -j # with embedded jq (formatting)
gotr get get_projects -j -f '.[].name' # project names only
```

#### Get project by ID

```bash
gotr get get_project 30
gotr get get_project --project-id 30 # via flag
gotr get get_project 30 -o project30.json # save to file
```

#### Get cases with filtering

```bash
gotr get get_cases 30 --suite-id 20069
gotr get get_cases 30 --suite-id 20069 --section-id 10
gotr get get_cases --project-id 30 --suite-id 20069
```

#### Data export

```bash
gotr export cases get_cases 30 --suite-id 20069
# The file will be saved in .testrail/cases_30_*.json

gotr export cases get_cases 30 --suite-id 20069 -o my_cases.json
# Save to specified file
```

#### Autocompletion

`gotr` supports resource and endpoint completion:

```bash
gotr get <Tab><Tab> # will suggest endpoints
gotr export cases <Tab> # will offer endpoints for cases
```

---
---

## Flags

### Global

- `--base-url` — TestRail base URL
- `--username` / `-u` — user email
- `--api-key` / `-k` — API key
- `--config` / `-c` — path to the config file
- `--insecure` / `-i` - skip TLS check
- `--jq` / `-j` - output via built-in jq
- `--jq-filter` / `-f` - jq filter
- `--quiet` / `-q` - suppress screen output
- `--type` / `-t` — output format (json, json-full, table)
- `--output` / `-o` - save to file

### Local (for get/export)

- `--project-id` / `-p` — project ID
- `--suite-id` / `-s` — test suite ID
- `--section-id` — section ID
- `--milestone-id` — ID milestone

## License

MIT License - use, modify, distribute freely.

## Authors

- [Korrnals](https://github.com/Korrnals)

## Acknowledgments

- TestRail API
- jqlang/jq - is an excellent tool for working with JSON
- itchyny/gojq - inspiration for built-in jq
- spf13/cobra - CLI basis
- fatih/color — color output

---

⭐ If the utility is useful, put a star on GitHub!  
If you have ideas or bugs, open an issue or PR.

Thanks for using `gotr`! 🚀
