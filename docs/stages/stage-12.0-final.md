# Stage 12.0 — Финальный рефакторинг

> Полная интеграция всех стандартов и паттернов. Финальная проверка.
> Проект доводится до production-ready состояния.

---

## Цель

Комплексная проверка и доработка проекта после завершения Stages 9–11:
- Убедиться, что ВСЕ стандарты соблюдены
- Устранить технический долг, накопившийся в процессе рефакторинга
- Подготовить проект к стабильному релизу

---

## Зависимости

```
Stage 9.0 (Стандарты)     ──► Stage 12.0
Stage 10.0 (Аудит)        ──► Stage 12.0
Stage 11.0 (Interactive)  ──► Stage 12.0
```

Stage 12.0 начинается ТОЛЬКО после полного завершения Stages 9–11.

---

## Подзадачи

### 12.1 — Кросс-пакетный аудит зависимостей

**Цель:** Убедиться, что слоистая архитектура не нарушена.

```bash
# Построить граф зависимостей
go mod graph
# Проверить отсутствие циклов
# cmd/ → service/ → client/  (OK)
# cmd/ → client/              (ЗАПРЕЩЕНО)
# service/ → cmd/             (ЗАПРЕЩЕНО)
```

**Проверки:**
- [ ] cmd/ НЕ импортирует internal/client/ напрямую
- [ ] service/ НЕ импортирует cmd/
- [ ] internal/ui/ НЕ импортирует internal/client/
- [ ] internal/flags/ НЕ импортирует internal/service/
- [ ] Нет циклических зависимостей

---

### 12.2 — Linter проход

**Цель:** Пройти полный набор линтеров без ошибок.

```yaml
# .golangci.yml — финальная конфигурация
linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - gosec
    - gocyclo
    - dupl
    - misspell
    - goconst
    - gofmt
    - goimports
    - depguard
    - bodyclose
    - noctx
    - exportloopref
    - exhaustive
    - prealloc
    - unparam
    - nakedret
    - gocritic
    - revive

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  goconst:
    min-len: 3
    min-occurrences: 3
  depguard:
    rules:
      main:
        deny:
          - pkg: "os/exec"
            desc: "use internal/editor or dedicated wrapper"
          - pkg: "log"
            desc: "use internal/log (zap)"
```

**Критерий:** 0 ошибок при `golangci-lint run --config .golangci.yml ./...`

---

### 12.3 — Удаление кириллицы из кода

**Цель:** Весь пользовательский текст в коде — на английском.

```bash
# Найти оставшуюся кириллицу
grep -rn '[а-яА-ЯёЁ]' --include='*.go' cmd/ internal/ pkg/ | grep -v _test.go | grep -v '// '
```

**Правило:**
- Комментарии `//` — допускается (для разработчиков)
- Строковые литералы (fmt.Print, errors.New) — ТОЛЬКО английский
- Логи (zap) — ТОЛЬКО английский
- Тесты — допускается кириллица в описаниях тестов

---

### 12.4 — Документация кода

**Цель:** Каждый экспортируемый символ документирован.

| Объект | Требование |
| ------ | ---------- |
| Пакет | `doc.go` с описанием назначения пакета |
| Экспортируемый тип | GoDoc-комментарий |
| Экспортируемая функция | GoDoc-комментарий с описанием параметров |
| Интерфейс | GoDoc + описание каждого метода |
| Константы / Enum | GoDoc для группы и каждого значения |

```bash
# Проверить покрытие GoDoc
# Для каждого пакета:
go doc ./internal/client/
go doc ./internal/service/migration/
go doc ./internal/ui/
```

---

### 12.5 — Финальное тестирование

```bash
# 1. Все тесты проходят
go test ./...

# 2. Нет data races
go test -race ./...

# 3. Покрытие
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1
# Цель: >= 85% overall

# 4. Бенчмарки (critical paths)
go test -bench=. ./internal/client/ ./internal/concurrent/ ./cmd/compare/

# 5. Сборка для всех платформ
GOOS=linux GOARCH=amd64 go build -o /dev/null .
GOOS=darwin GOARCH=amd64 go build -o /dev/null .
GOOS=darwin GOARCH=arm64 go build -o /dev/null .
```

---

### 12.6 — Обновление зависимостей

```bash
# Проверить устаревшие зависимости
go list -m -u all

# Обновить minor/patch
go get -u ./...
go mod tidy

# Проверить уязвимости
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

---

### 12.7 — Binary / Release проверка

```bash
# Проверить размер бинарника
go build -o gotr .
ls -lh gotr
# Убедиться: embedded jq не раздул бинарник чрезмерно

# Проверить selftest
./gotr selftest

# Проверить completions
./gotr completion bash > /dev/null
./gotr completion zsh > /dev/null
./gotr completion fish > /dev/null
```

---

### 12.8 — Финальный README / Документация

| Документ | Обновить |
| -------- | -------- |
| README.md | Актуализировать: фичи, установка, примеры |
| README_ru.md | Синхронизировать с README.md |
| docs/architecture.md | Отразить финальную архитектуру (после Stages 9–11) |
| docs/configuration.md | Добавить новые флаги/опции |
| docs/interactive-mode.md | Обновить с учётом Stage 11.0 |
| CHANGELOG.md | Добавить записи для Stages 9–12 |

---

### 12.9 — CI/CD подготовка

```yaml
# GitHub Actions: .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go vet ./...
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} go build -o gotr .
```

---

### 12.10 — Git cleanup

**Цель:** Подготовить репозиторий к чистой истории.

- [ ] Все feature branches смёрджены в main
- [ ] Теги для каждой стабильной версии
- [ ] `.gitignore` актуален (gotr binary, coverage.out, .env)
- [ ] Нет секретов в истории (API keys, passwords)
- [ ] CHANGELOG.md полный и актуальный

---

## Метрики финального проекта (целевые)

| Метрика | Цель |
| ------- | ---- |
| Файлов .go | — (без цели, только качество) |
| Покрытие тестами | ≥ 85% overall, ≥ 80% per package |
| `golangci-lint` ошибок | 0 |
| `go vet` предупреждений | 0 |
| Data races (`-race`) | 0 |
| Уязвимостей (`govulncheck`) | 0 |
| Кириллица в строках кода | 0 |
| Циклических зависимостей | 0 |
| Средняя цикломатическая сложность | ≤ 10 |
| Максимальная цикломатическая сложность | ≤ 15 |

---

## Definition of Done — Stage 12.0

- [ ] Все метрики из таблицы выше достигнуты
- [ ] Документация полная и актуальная
- [ ] CI pipeline настроен и проходит
- [ ] `selftest` command проходит успешно
- [ ] Бинарник собирается для linux/amd64, darwin/amd64, darwin/arm64
- [ ] Все completions (bash, zsh, fish) генерируются без ошибок
- [ ] README.md и README_ru.md синхронизированы
- [ ] Создан release tag с CHANGELOG
