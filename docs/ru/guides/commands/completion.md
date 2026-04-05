# Команды completion

Language: Русский | [English](../../../en/guides/commands/completion.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](index.md)
      - [Общие](global-flags.md)
        - [global-flags](global-flags.md)
        - [config](config.md)
        - [completion](completion.md)
        - [self-test](self-test.md)
      - [CRUD операции](add.md)
      - [Основные ресурсы](get.md)
      - [Специальные ресурсы](bdds.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

Команда `gotr completion` генерирует shell completion scripts для интерактивного ввода.

## Что делает

- Генерирует автодополнение команд и флагов.
- Снижает количество ошибок при ручном вводе длинных команд.
- Ускоряет работу в shell за счёт интерактивного выбора аргументов.

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Подкоманды

- `gotr completion bash`
- `gotr completion zsh`
- `gotr completion fish`
- `gotr completion powershell`

## Примеры

```bash
# Bash
source <(gotr completion bash)

# Zsh
gotr completion zsh > "${fpath[1]}/_gotr"

# Fish
gotr completion fish > ~/.config/fish/completions/gotr.fish
```

## Полезные флаги

- `--json` для вывода в машинно-обрабатываемом формате.
- `--output` / `--save` для сохранения результата в файл.
- `--verbose` для детальной диагностики выполнения.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
