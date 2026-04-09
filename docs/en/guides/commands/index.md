# Commands

Language: [Русский](../../../ru/guides/commands/index.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](index.md)
      - [General](#general)
        - [global-flags](global-flags.md)
        - [config](config.md)
        - [completion](completion.md)
        - [self-test](self-test.md)
      - [CRUD Operations](#crud-operations)
        - [add](add.md)
        - [delete](delete.md)
        - [update](update.md)
        - [list](list.md)
        - [export](export.md)
      - [Core Resources](#core-resources)
        - [get](get.md)
        - [sync](sync.md)
        - [compare](compare.md)
        - [cases](cases.md)
        - [run](run.md)
        - [result](result.md)
        - [test](test.md)
        - [tests](tests.md)
        - [attachments](attachments.md)
        - [plans](plans.md)
        - [reports](reports.md)
      - [Special Resources](#special-resources)
        - [bdds](bdds.md)
        - [configurations](configurations.md)
        - [datasets](datasets.md)
        - [groups](groups.md)
        - [labels](labels.md)
        - [milestones](milestones.md)
        - [roles](roles.md)
        - [templates](templates.md)
        - [users](users.md)
        - [variables](variables.md)
    - [Instructions](../instructions/index.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

Reference catalog for all top-level CLI commands.

## Command Structure

Compact command structure below explains what each group includes and when to use it.

### General

Service commands and core CLI configuration.

- [global-flags](global-flags.md) — global flags for connection, output, and runtime behavior.
- [config](config.md) — local client configuration management.
- [completion](completion.md) — shell completion generation for bash/zsh/fish/powershell.
- [self-test](self-test.md) — quick environment and API availability checks.

### CRUD Operations

Universal create/update/delete/list/export operations.

- [add](add.md) — create entities via API.
- [delete](delete.md) — remove resources.
- [update](update.md) — modify existing entities.
- [list](list.md) — retrieve lists and base selections.
- [export](export.md) — export data to files and report formats.

### Core Resources

Domain-oriented namespace commands for daily TestRail workflows.

- [get](get.md) — read resources and reference data.
- [sync](sync.md) — synchronize entities between projects/structures.
- [compare](compare.md) — compare entities and detect diffs.
- [cases](cases.md) — operations with test cases.
- [run](run.md) — operations with test runs.
- [result](result.md) — add and inspect test results.
- [test](test.md) — operations with individual run tests.
- [tests](tests.md) — batch operations with sets of tests.
- [attachments](attachments.md) — upload and fetch attachments.
- [plans](plans.md) — work with test plans.
- [reports](reports.md) — access TestRail reporting endpoints.

### Special Resources

Extended resources and specialized endpoint groups.

- [bdds](bdds.md) — BDD data for cases.
- [configurations](configurations.md) — configurations and parameter sets.
- [datasets](datasets.md) — datasets and related entities.
- [groups](groups.md) — user groups and access scope.
- [labels](labels.md) — labels and categorization.
- [milestones](milestones.md) — milestone lifecycle operations.
- [roles](roles.md) — roles and permissions.
- [templates](templates.md) — case templates.
- [users](users.md) — users and user attributes.
- [variables](variables.md) — variables and values.

---

← [Guides](../index.md) · [Documentation](../../index.md)
