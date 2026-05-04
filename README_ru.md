# gotr — CLI-клиент для TestRail API

<p align="center">
  <img src="docs/assets/banner.svg" alt="gotr — CLI-клиент для TestRail API v2: миграция · снапшоты · синхронизация · отчёты · автоматизация" width="100%"/>
</p>

<p align="center">
  <a href="README.md">English</a> · <a href="README_ru.md">Русский</a>
</p>

<p align="center">
  <a href="https://github.com/Korrnals/gotr/releases/latest"><img src="https://img.shields.io/badge/release-v3.4.0-blue.svg" alt="Latest Release"/></a>
  <a href="CHANGELOG.md"><img src="https://img.shields.io/badge/next-v3.5.0--dev-orange.svg" alt="Next"/></a>
  <a href="go.mod"><img src="https://img.shields.io/badge/go-1.25-00ADD8.svg?logo=go" alt="Go"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg" alt="License"/></a>
  <a href="docs/index.md"><img src="https://img.shields.io/badge/docs-EN%20%7C%20RU-purple.svg" alt="Docs"/></a>
</p>

> 🛠️ Профессиональный CLI для TestRail API v2 — для QA-инженеров и автоматизаторов, которым нужны быстрые массовые операции, безопасные миграции с откатом и удобная интеграция с CI/CD.

`gotr` — полнофункциональный терминальный клиент, покрывающий весь TestRail API v2 (121 эндпоинт), с операторскими гарантиями безопасности (снапшоты, dry-run, rollback), автодополнением shell, интерактивным режимом и переносимым форматом бандлов для переноса данных между инсталляциями. Любая длительная операция стримит прогресс, любое деструктивное действие обратимо, для каждого сценария есть документированный runbook.

---

## 📚 Навигация

- 📖 [Каталог документации](docs/index.md) — полный индекс EN/RU
- 🚀 [Установка](docs/ru/guides/installation.md) · ⚙️ [Конфигурация](docs/ru/guides/configuration.md) · 💬 [Интерактивный режим](docs/ru/guides/interactive-mode.md)
- 📋 [Каталог команд](docs/ru/guides/commands/index.md) — справочник по каждой подкоманде
- 📘 [Инструкции и runbook'и](docs/ru/guides/instructions/index.md) — эксплуатационные сценарии
- 🏛️ [Архитектура](docs/ru/architecture/index.md) · 📊 [Отчёты](docs/ru/reports/index.md) · 🛠️ [Эксплуатация](docs/ru/operations/index.md)
- 📰 [CHANGELOG](CHANGELOG.md) — история релизов и unreleased-скоуп

---

## 🚀 Установка и быстрый старт

```bash
# 1. Установка (Linux / macOS)
curl -sL https://github.com/Korrnals/gotr/releases/latest/download/gotr-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o gotr
chmod +x gotr && sudo mv gotr /usr/local/bin/

# 2. Инициализация конфигурации (URL · логин · API-ключ)
gotr config init

# 3. Проверка соединения с TestRail
gotr self-test

# 4. Первая команда
gotr get projects
```

**Подробные инструкции** (Windows, пакетные менеджеры, сборка из исходников, troubleshooting): [Гайд по установке](docs/ru/guides/installation.md).

---

## ✨ Ключевые возможности

Каждая строка ведёт на свой справочник; имя ссылки совпадает с реальной подкомандой.

| Возможность | Подкоманда | Что делает |
|---|---|---|
| 🔍 **Чтение ресурсов** | [`get`](docs/ru/guides/commands/get.md) | Кейсы, сьюты, секции, раны, планы, майлстоуны, пользователи и ещё 100+ ресурсов с фильтрами и пагинацией. |
| 🔄 **Кросс-проектная синхронизация** | [`sync`](docs/ru/guides/commands/sync.md) | Миграция кейсов, shared steps, сьютов и секций между проектами с дедупликацией, маппингом и `--verify-coverage`. |
| 🆚 **Сравнение проектов** | [`compare`](docs/ru/guides/commands/compare.md) | Diff кейсов, сьютов, планов, майлстоунов, датасетов; экспорт в JSON / YAML / table. |
| 📸 **Снапшоты и откат** | [`snap`](docs/ru/guides/commands/snap.md) | Снапшот любой мутации, list, restore, GC по per-category TTL. Каждое деструктивное действие по умолчанию делает снапшот. |
| 📎 **Вложения** *(включая массовую очистку)* | [`attachments`](docs/ru/guides/commands/attachments.md) | Загрузка, скачивание, листинг — и **массовая очистка** старых вложений со снапшотом и откатом по умолчанию (категория `cleanup-attachments`, retention 7 дней). |
| 🧹 **Retention и cleanup** | [`cleanup`](docs/ru/guides/commands/cleanup.md) | Настраиваемый retention для отчётов / снапов / экспортов с `--dry-run` и авто-cleanup-хуками. |
| 📊 **Тест-раны и результаты** | [`run`](docs/ru/guides/commands/run.md) · [`result`](docs/ru/guides/commands/result.md) | Создание ранов, массовая загрузка результатов, отслеживание выполнения. |
| ✅ **Операции на уровне теста** | [`test`](docs/ru/guides/commands/test.md) · [`tests`](docs/ru/guides/commands/tests.md) | Просмотр одиночных run-test'ов и батч-операции. |
| 📦 **Переносимые export / import** | [`export`](docs/ru/guides/commands/export.md) | Самодостаточные бандлы (snap / report / migration-archive) с `manifest.json` + `SHA256SUMS`, детерминированные, с поддержкой redaction. Симметричный `import`. |
| 📝 **Жизненный цикл отчётов** | [`report`](docs/ru/guides/commands/report.md) | Категоризированные отчёты (`migrations` / `coverage` / `rollbacks` / `testrail/p<N>`), `report show --print`, рекурсивный листинг, переиндексация INDEX. |
| 🧩 **CRUD-шорткаты** | [`add`](docs/ru/guides/commands/add.md) · [`update`](docs/ru/guides/commands/update.md) · [`delete`](docs/ru/guides/commands/delete.md) · [`list`](docs/ru/guides/commands/list.md) | Универсальные create / update / delete / list по разным типам ресурсов. |
| 💬 **Интерактивный режим** | [interactive-mode](docs/ru/guides/interactive-mode.md) | TTY-guarded survey-промпты для `get` / `sync` / `compare` / `report` / `attachments cleanup` / `export` / `import` — без запоминания ID. |
| 🔧 **Конфигурация и профили** | [`config`](docs/ru/guides/commands/config.md) | YAML-конфиг с переопределениями через env, несколько профилей, поддержка TLS CA-bundle. |
| 🐚 **Shell completion** | [`completion`](docs/ru/guides/commands/completion.md) | bash / zsh / fish / powershell с динамическим `ValidArgsFunction` для файлов, snap ID, путей отчётов. |
| 🩺 **Self-test и диагностика** | [`self-test`](docs/ru/guides/commands/self-test.md) | Проверка соединения с API, валидность конфига, embedded-инструменты. |
| 🪵 **Встроенный JSON-процессор** | `--jq` / `--jq-filter` | Фильтрация и трансформация любого вывода через встроенный `jq` — без внешних зависимостей. |
| 📈 **Streaming-прогресс** | [`progress`](docs/ru/guides/progress.md) | Прогресс-бары на каналах с адаптивным rate-limit (180 req/min) для параллельных загрузок. |
| 🐛 **Debug-трейс** | `--debug` / `-d` | Детали API-запросов, тайминг по фазам, диагностика обработки сьютов/кейсов. |

Полный справочник, плюс команды `bdds` / `configurations` / `datasets` / `groups` / `labels` / `milestones` / `plans` / `roles` / `templates` / `users` / `variables`, см. в [каталоге команд](docs/ru/guides/commands/index.md).

---

## 💡 Примеры (выдержка)

Ниже — самые частые сценарии. Полный набор рецептов (TLS, CI, mapping-файлы, redaction и пр.) — в [каталоге команд](docs/ru/guides/commands/index.md) и [runbook-инструкциях](docs/ru/guides/instructions/index.md).

```bash
# 🔍 Чтение данных
gotr get projects
gotr get cases 30 --suite-id 20069
gotr get sharedsteps 30 --jq --jq-filter '.[] | {id, title}'

# 🔄 Синхронизация между проектами (со снапшотом и верификацией)
gotr sync full \
  --src-project 30 --src-suite 20069 \
  --dst-project 31 --dst-suite 19859 \
  --approve --save-mapping

# 🆚 Сравнение проектов
gotr compare all --pid1 30 --pid2 34 --save
gotr compare cases --pid1 30 --pid2 34 --save-to results.json --format json

# 📎 Массовая очистка старых вложений (по умолчанию со снапшотом и откатом)
gotr attachments cleanup --all-projects --older-than 6M --dry-run
gotr attachments cleanup --project 30 --older-than 6M

# 📸 Откат любой мутации
gotr snap list
gotr snap rollback <snap-id>

# ✅ Создание рана и публикация результатов
gotr run add 30 --name "Regression Suite" --case-ids "1,2,3,4,5"
gotr result add 12345 --status-id 1 --comment "Passed"

# 📦 Переносимый бандл: export → перенос → import
gotr export snap <snap-id>
gotr import snap ~/.gotr/exports/snaps/snap_<id>_<ts>.tar.gz
```

➡️ **Больше примеров**: [каталог команд](docs/ru/guides/commands/index.md) · [интерактивный режим](docs/ru/guides/interactive-mode.md) · [smoke-тестирование](docs/ru/guides/smoke-testing.md).

---

## ⚙️ Конфигурация

Приоритет (от высшего к низшему):

1. **Флаги CLI** — `--url` · `--username` · `--api-key`
2. **Переменные окружения** — `TESTRAIL_BASE_URL` · `TESTRAIL_USERNAME` · `TESTRAIL_API_KEY`
3. **Файл конфигурации** — `~/.gotr/config/default.yaml` (поддерживает несколько профилей)

```bash
gotr config init   # создать дефолтный профиль
gotr config view   # показать итоговую резолюцию
```

Полный справочник: [Гайд по конфигурации](docs/ru/guides/configuration.md) (TLS CA-bundle, retention, suppress_warnings, тюнинг для cloud/server).

---

## 🗂️ Структура проекта

```text
gotr/
├── cmd/                  # CLI-команды (Cobra) — по подпапке на ресурс
│   ├── attachments/      #   upload · list · cleanup (массовая + откат)
│   ├── bundlecmd/        #   export / import (snap · report · migration-archive)
│   ├── cleanup/          #   исполнитель retention (reports · snaps · exports · all)
│   ├── compare/          #   кросс-проектный diff
│   ├── get/              #   read-only чтение ресурсов
│   ├── snap/             #   snapshot list · rollback · gc
│   ├── sync/             #   движок синхронизации проектов
│   ├── report/           #   жизненный цикл отчётов (organize · show · view · list)
│   ├── run/ result/ test/ tests/                  #   домен выполнения
│   └── …                 #   bdds · cases · configurations · datasets · groups
│                         #   labels · milestones · plans · roles · templates
│                         #   users · variables · work
├── internal/
│   ├── client/           #   клиент TestRail API + paginator + ClientInterface
│   ├── service/          #   бизнес-логика (run · result · migration · …)
│   ├── snap/             #   движок снапшотов (entity · backup · rollback)
│   ├── snapbundle/       #   tar.gz-бандлы · manifest · SHA256SUMS
│   ├── reportbundle/     #   zip-бандлы для экспорта отчётов
│   ├── bundle/           #   общая механика бандлов (защита от zip-slip)
│   ├── cleanup/          #   ядро очистки вложений (walker · filter · executor)
│   ├── retention/        #   политики retention (reports · snaps · exports)
│   ├── report/           #   classify · organize · resolve · INDEX
│   ├── exportsorg/       #   мигратор раскладки exports/
│   ├── concurrent/       #   примитивы — WorkerPool · AdaptiveRateLimiter · retry
│   ├── concurrency/      #   доменная оркестрация — ParallelController · streaming
│   ├── interactive/      #   survey-промпты · TTY-guard · MockPrompter
│   ├── output/           #   рендереры JSON · YAML · table
│   ├── ui/               #   прогресс-бары · статус-сообщения · quiet-mode
│   ├── warnings/         #   реестр подавляемых варнингов + first-time tips
│   ├── state/            #   ~/.gotr/state.json (one-shot флаги)
│   ├── flags/            #   общий парсинг флагов
│   ├── log/              #   структурированное логирование (zap)
│   ├── paths/            #   хелперы раскладки ~/.gotr
│   └── models/           #   API DTO + модель конфига
├── pkg/
│   ├── testrailapi/      #   определения API endpoint'ов (135 endpoints)
│   ├── reporter/         #   унифицированный репортер статистики
│   └── snap_smoke/       #   smoke-харнес для снапшотов
├── embedded/             #   встроенный jq (без внешних зависимостей)
├── docs/                 #   документация EN + RU
└── main.go               #   точка входа
```

Углублённо про архитектуру: [docs/ru/architecture/index.md](docs/ru/architecture/index.md).

---

## 🧪 Quality gates

- ✅ `golangci-lint v2.11.4` — 0 issues, порог `gocyclo ≤ 15`
- ✅ `go test ./... -count=1 -timeout 300s` — зелёный
- ✅ Воспроизводимые детерминированные бандлы (фиксированный `ModTime`, стабильная сортировка, `tar.FormatPAX`)
- ✅ Race detector + `govulncheck` в CI

Pre-PR чеклист: `make verify` (test + vet + lint + build + race + vuln).

---

## 🤝 Участие в проекте

Issues и pull requests приветствуются. Пожалуйста, следуйте [протоколу релизов](docs/ru/operations/index.md) и запускайте `make verify` перед открытием PR.

---

## 🙏 Используемые библиотеки

Проект построен на отличных open-source-проектах:

| Библиотека | Назначение |
|---|---|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI-фреймворк |
| [spf13/viper](https://github.com/spf13/viper) | Управление конфигурацией |
| [go.uber.org/zap](https://github.com/uber-go/zap) | Структурированное логирование |
| [stretchr/testify](https://github.com/stretchr/testify) | Тестирование |
| [AlecAivazis/survey/v2](https://github.com/AlecAivazis/survey) | Интерактивные промпты |
| [jedib0t/go-pretty/v6](https://github.com/jedib0t/go-pretty) | Табличный вывод |
| [fatih/color](https://github.com/fatih/color) | Цветной вывод в терминале |
| [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync) · [time](https://pkg.go.dev/golang.org/x/time) | Параллелизм и rate limiting |
| [jq](https://github.com/jqlang/jq) | Встроенный JSON-процессор (`--jq` / `--jq-filter`) |

---

## 📄 Лицензия

[MIT](LICENSE) © Korrnals
