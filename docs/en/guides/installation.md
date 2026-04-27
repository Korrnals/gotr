# Installation

Language: [Русский](../../ru/guides/installation.md) | English

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

## Quick install (Linux/macOS)

```bash
# Unix
curl -s -L https://github.com/Korrnals/gotr/releases/latest/download/gotr-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o gotr && chmod +x gotr && sudo mv gotr /usr/local/bin/
```

> **Note:** for Windows download the .exe manually from [Releases](https://github.com/Korrnals/gotr/releases).

## Build from source

### Requirements

- Go 1.21+
- (Optional) UPX for compression

### Option 1: Simple build

```bash
git clone https://github.com/Korrnals/gotr.git
cd gotr
go build -ldflags="-s -w" -o gotr
sudo mv gotr /usr/local/bin/
```

### Option 2: Via Makefile (recommended)

```bash
git clone https://github.com/Korrnals/gotr.git
cd gotr

# Build and install
make install

# Other targets:
make build          # build only
make test           # run tests
make compress       # UPX compression
make build-compressed  # build + compress
make clean          # cleanup
make release        # build for all platforms
```

### Build with a version

```bash
# Without a tag — version "dev"
make build
# gotr version → dev

# With a tag
git tag v2.0.0
make build
# gotr version → v2.0.0
```

## Verify installation

```bash
gotr --help
gotr version
```

---

← [Guides](index.md) · [Documentation](../index.md)
