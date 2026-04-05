# Команды self-test

Команда `gotr self-test` — самотестирование инструмента.

## What it does

- Основная операция для self-test
- Интеграция с другими командами
- Поддержка интерактивного режима

## When to use

- Для операций self-test в тестировании
- Когда нужна автоматизация процесса
- В CI/CD конвейерах

## Examples

```bash
gotr self-test --help      # справка
gotr self-test --project 30 # базовый запуск
```

## Main flags

| Флаг | Описание |
| --- | --- |
| `--help` | Справка по команде |
| `--verbose` | Детальный вывод |

## FAQ

**Q: Как получить справку?**  
A: `gotr self-test --help`.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
