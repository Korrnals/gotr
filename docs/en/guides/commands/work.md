# Command: work

Language: [Русский](../../../ru/guides/commands/work.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](index.md)
      - [General](global-flags.md)
        - [global-flags](global-flags.md)
        - [config](config.md)
        - [completion](completion.md)
        - [self-test](self-test.md)
        - [snap](snap.md)
        - [work](work.md)
      - [CRUD Operations](add.md)
      - [Core Resources](get.md)
      - [Special Resources](bdds.md)
    - [Instructions](../instructions/index.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)


## Overview 🎯

Interactive navigation hub for working with TestRail. Combines all key operations
(compare, sync, snap, CRUD) into a single menu with cross-navigation between commands.

> [!TIP]
> `gotr work` is the recommended entry point for interactive work.
> All subcommands are available through the hierarchical menu.

## Syntax 🧩

```bash
gotr work
```

## How It Works

### Main Menu

Running `gotr work` displays a server picker, then the main menu:

```text
? What to do?
  ← Exit
  ─────────────────
  📊 Compare — compare projects
  🔄 Sync — migrate data
  📦 Snap — snapshots and rollback
  📋 Get — read resources
  ➕ Add — create resources
  ✏️  Update — modify resources
  🗑  Delete — remove resources
  📤 Export — export resources
```

### Group Hubs

Selecting a group opens a submenu with specific subcommands:

**Compare:**
```text
? Compare — what to compare?
  ← Back
  ─────────────────
  All resources (full scan)
  Cases
  ...
```

**Snap:**
```text
? Snap — what to do?
  ← Back
  ─────────────────
  📋 List snapshots
  🔍 View snapshot details
  ↻ Rollback a snapshot
  📤 Export snapshot
  🗑  Delete snapshot
```

### Cross-Navigation

After executing an operation, the post-action menu offers navigation to related commands:

- **📊 Compare** — jump to project comparison
- **🔄 Sync** — jump to migration
- **📦 Snap** — jump to snapshot management

This enables a seamless workflow: compare → sync → snap → compare, without leaving the interactive mode.

### Session

On server connection, a banner is displayed:

```text
Server: https://your.testrail.io
```

Session context (project, suite) is inherited between commands within the hub.

## Examples 🚀

### ▶️ Launch the navigation hub

```bash
gotr work
```

### ▶️ Typical workflow

1. `gotr work` → select server
2. **Compare** → compare projects
3. From post-action → **Sync** → migrate data
4. From post-action → **Snap** → check snapshots
5. ← Back → return to main menu

## 🧪 Pre-run Checklist

- [ ] gotr is configured (`gotr self-test`)
- [ ] Interactive mode is enabled (no `--non-interactive`)

## FAQ ❓

- ❓ **Question:** How is `gotr work` different from direct commands?
  > ↪️ **Answer:** `gotr work` is a navigation wrapper. The same commands (`gotr compare all`, `gotr snap list`) work directly. The hub adds menus, cross-navigation, and session context.

- ❓ **Question:** Can I use `gotr work` in CI/scripts?
  > ↪️ **Answer:** No. `gotr work` is strictly interactive. For CI, use direct commands with flags.

---

← [Commands Index](index.md) · [snap](snap.md)
