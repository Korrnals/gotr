# Команды list

Команда `gotr list` — вывод списков с фильтрацией

## What it does

- Основная операция для list
- Интеграция с другими командами  
- Поддержка интерактивного режима

## When to use

- Для операций list в тестировании
- Когда нужна автоматизация процесса
- В CI/CD конвейерах

## Examples

```bash
gotr list --help      # справка
gotr list --project 30 # базовый запуск
```

## Main flags

| Флаг | Описание |
| --- | --- |
| `--help` | Справка по команде |
| `--verbose` | Детальный вывод |
| `--dry-run` | Предпросмотр |

## FAQ

**Q: Как получить справку?**  
A: `gotr list --help`.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
