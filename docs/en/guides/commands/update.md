# Команды update

Команда `gotr update` — обновление существующих элементов

## What it does

- Основная операция для update
- Интеграция с другими командами  
- Поддержка интерактивного режима

## When to use

- Для операций update в тестировании
- Когда нужна автоматизация процесса
- В CI/CD конвейерах

## Examples

```bash
gotr update --help      # справка
gotr update --project 30 # базовый запуск
```

## Main flags

| Флаг | Описание |
| --- | --- |
| `--help` | Справка по команде |
| `--verbose` | Детальный вывод |
| `--dry-run` | Предпросмотр |

## FAQ

**Q: Как получить справку?**  
A: `gotr update --help`.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
