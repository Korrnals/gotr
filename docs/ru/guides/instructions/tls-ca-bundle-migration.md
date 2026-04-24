# Инструкция: миграция TLS с `insecure` на `ca_bundle`

Language: Русский | [English](../../../en/guides/instructions/tls-ca-bundle-migration.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Инструкции](index.md)
      - [TLS: insecure → ca_bundle](tls-ca-bundle-migration.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

Как корректно уйти от `insecure: true` к корпоративному `ca_bundle` без
простоев и без правок CI.

## Когда это нужно

- Корпоративный TestRail за MITM-прокси с самоподписанным корневым CA.
- SRE-требование: избавиться от `--insecure` / `insecure: true` в prod.
- Нужно избавиться от предупреждения `WARNING: TLS verification disabled`
  в логах без потери безопасности.

## Кейс A: Один корпоративный CA

**Исходник:**

```yaml
# ~/.gotr/config/default.yaml (v3.2-style)
insecure: true
```

**Шаги:**

```bash
# 1. Получить PEM CA (у SRE/PKI-команды)
sudo cp corporate-ca.pem /etc/ssl/corp-ca.pem
sudo chmod 0644 /etc/ssl/corp-ca.pem

# 2. Сделать бэкап конфига
cp ~/.gotr/config/default.yaml ~/.gotr/config/default.yaml.bak
```

**Новый конфиг:**

```yaml
# ~/.gotr/config/default.yaml (v3.3-style)
tls:
  insecure: false
  ca_bundle: "/etc/ssl/corp-ca.pem"
```

**Валидация:**

```bash
gotr self-test            # базовая проверка TLS-рукопожатия
gotr get projects         # полноценный запрос
```

В stderr баннер `tls_insecure` **не должен появляться**.

## Кейс B: Оставить insecure как fallback для одной машины

```yaml
tls:
  ca_bundle: "/etc/ssl/corp-ca.pem"
```

```bash
# На машине, где CA ещё не раскатан — разово
gotr get projects --insecure
```

Top-level `insecure` и `--insecure` OR-мерджатся с `tls.insecure` —
существующие скрипты продолжают работать.

## Кейс C: Подавить баннер, если insecure необходим в CI

```yaml
tls:
  insecure: true   # временное решение — задокументировано
ui:
  suppress_warnings: [tls_insecure]
```

```bash
# Показать баннер локально, игнорируя suppress
gotr get projects --show-warnings
```

## Кейс D: Несколько CA

PEM-бандл — это конкатенация PEM-сертификатов. Просто сложите их в
один файл:

```bash
cat corp-root.pem corp-intermediate.pem > /etc/ssl/corp-ca.pem
```

`x509.CertPool.AppendCertsFromPEM` прочитает все блоки.

## Troubleshooting

| Симптом | Причина | Решение |
| --- | --- | --- |
| `x509: certificate signed by unknown authority` | ca_bundle не включает корневой CA | Добавьте root-cert в PEM |
| `open /etc/ssl/corp-ca.pem: permission denied` | chmod/ownership | `chmod 0644`, проверить группу |
| `failed to parse CA PEM` | файл пустой или некорректный | `openssl x509 -in <file> -text -noout` |
| Баннер про TLS всё ещё появляется | `insecure=true` где-то активен | `gotr config view`, проверить env `TESTRAIL_INSECURE` / legacy `insecure:` |

## Критерии успеха

- `gotr self-test` и `gotr get projects` — exit 0 без TLS-warning.
- `go tool dist` / `openssl s_client` к TestRail host успешны через
  тот же PEM (sanity check).
- В конфиге нет top-level `insecure: true`.

## См. также

- [Конфигурация → TLS и ca_bundle](../configuration.md).
- [Migration guide v3.3 → TLS](../migration-guide-v3.3.md).
- [Архитектура: UX polish v3.3.0 → ca_bundle](../../architecture/ux-polish-v3.3.0.md).

---

← [Инструкции](index.md) · [Документация](../../index.md)
