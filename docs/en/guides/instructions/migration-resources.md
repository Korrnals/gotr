# Instruction: Resources Migration (suites, sections)

Language: [Русский](../../../ru/guides/instructions/migration-resources.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](../commands/index.md)
    - [Instructions](index.md)
      - [Full Migration](migration-full.md)
      - [Partial Migration](migration-partial.md)
      - [Shared Steps Migration](migration-shared-steps.md)
      - [Resources Migration](migration-resources.md)
      - [Getting Data](crud-get.md)
      - [Exporting Data](crud-export.md)
      - [Creating Objects](crud-add.md)
      - [Updating Objects](crud-update.md)
      - [Deleting Objects](crud-delete.md)
      - [Comparing Projects](compare.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

## Overview 🎯

Migration of structural resources: **suites** and **sections** between projects.
Used to prepare the target project before transferring cases.

> [!TIP]
> Resource migration order: suites → sections → shared steps → cases.
> For a complete transfer, use [Full Migration](migration-full.md).

## Prerequisites ✅

- [ ] gotr configured and connected to TestRail (`gotr self-test`)
- [ ] Source and target project IDs are known

---

## Scenario 1: Suite Migration 🚀

Transferring test suites from one project to another.

### Recon

```bash
# View source project suites
gotr get suites 30

# View target project suites
gotr get suites 34
```

### Dry-run

```bash
gotr sync suites \
  --src-project 30 \
  --dst-project 34 \
  --dry-run
```

### Execute

```bash
gotr sync suites \
  --src-project 30 \
  --dst-project 34 \
  --save-mapping --approve
```

### Verify

```bash
gotr get suites 34
gotr compare suites --pid1 30 --pid2 34
```

---

## Scenario 2: Section Migration 🚀

Transferring sections between suites of two projects.

### Recon

```bash
# View source suite sections
gotr export sections -p 30 -s 20069 --save --format json

# View target suite sections
gotr export sections -p 34 -s 19859 --save --format json
```

### Dry-run

```bash
gotr sync sections \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --dry-run
```

### Execute

```bash
gotr sync sections \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --save-mapping --approve
```

### Verify

```bash
gotr compare sections --pid1 30 --pid2 34
```

---

## Syntax 🧩

### sync suites

```bash
gotr sync suites \
  --src-project <ID> \
  --dst-project <ID> \
  [--compare-field <field>] \
  [--dry-run] \
  [--save-mapping] \
  [--approve] \
  [--quiet]
```

### sync sections

```bash
gotr sync sections \
  --src-project <ID> \
  --src-suite <ID> \
  --dst-project <ID> \
  --dst-suite <ID> \
  [--compare-field <field>] \
  [--dry-run] \
  [--save-mapping] \
  [--approve] \
  [--quiet]
```

## Flags ⚙️

| Flag | Description | Default |
| --- | --- | --- |
| `--src-project` | Source project ID | required |
| `--src-suite` | Source suite ID (for sections) | required for sections |
| `--dst-project` | Target project ID | required |
| `--dst-suite` | Target suite ID (for sections) | required for sections |
| `--compare-field` | Field for duplicate detection | `title` |
| `--dry-run` | Show plan without changes | `false` |
| `--save-mapping` | Save mapping to file | `false` |
| `--approve` | Skip confirmation prompt | `false` |
| `--quiet` | Suppress service output | `false` |

## Full Structure Migration Pipeline 🧩

To transfer an entire project structure:

```bash
# 1. Transfer suites
gotr sync suites \
  --src-project 30 --dst-project 34 \
  --save-mapping --approve

# 2. Transfer sections for each suite
gotr sync sections \
  --src-project 30 --src-suite 20069 \
  --dst-project 34 --dst-suite 19859 \
  --save-mapping --approve

# 3. Then — shared steps and cases
# See Full Migration or Shared Steps Migration
```

## FAQ ❓

- ❓ **Question:** What if a suite with the same name already exists?
  > ↪️ **Answer:** gotr detects duplicates by `title` and does not create duplicate suites.
  >
  > ---

- ❓ **Question:** Is nested section hierarchy preserved?
  > ↪️ **Answer:** yes, `sync sections` preserves parent-child relationships between sections.
  >
  > ---

- ❓ **Question:** Do I need to transfer sections separately if I use `sync full`?
  > ↪️ **Answer:** `sync full` handles shared steps and cases, but not sections. If you need the section structure — transfer it separately before `sync full`.

---

← [Instructions](index.md)
