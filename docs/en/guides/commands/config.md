# Команды config

Команда `gotr config` — управление конфигурацией

## What it does

- Основная операция для config
- Интеграция с другими командами  
- Поддержка интерактивного режима

## When to use

- Для операций config в тестировании
- Когда нужна автоматизация процесса
- В CI/CD конвейерах

## Examples

```bash
gotr config --help      # справка
gotr config --project 30 # базовый запуск
```

## Main flags

| Флаг | Описание |
| --- | --- |
| `--help` | Справка по команде |
| `--verbose` | Детальный вывод |
| `--dry-run` | Предпросмотр |

## FAQ

**Q: Как получить справку?**  
A: `gotr config --help`.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
