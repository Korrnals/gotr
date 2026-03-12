# Stage 11.0 — Унифицированная интерактивная система

> Создание единой библиотеки интерактивных компонентов на базе
> лучших практик и специализированных Go-библиотек.

---

## Цель

Заменить разрозненные интерактивные реализации единой системой:
- Единый API для промптов, списков, форм
- Тестируемость (mock I/O)
- Поддержка `--non-interactive` / pipe-режима
- Переиспользуемость между всеми командами

---

## Текущее состояние

### Проблемы

1. **Прямой `fmt.Scanf` / `bufio.Scanner`** — не тестируемый, не расширяемый
2. **Дублирование** — каждая команда реализует свой диалог: add, update, sync, compare
3. **Нет стандарта** — разные стили промптов в разных пакетах
4. **Нет `--non-interactive`** — невозможно использовать в CI/CD pipe

### Затронутые пакеты

```
cmd/add.go                   — ручной ввод при добавлении ресурсов
cmd/update.go                — редактирование полей
cmd/sync/                    — Sync Wizard с меню
cmd/compare/                 — выбор scope, параметров
cmd/configurations/          — create/update диалоги
internal/interactive/        — существующие хелперы
internal/utils/              — OpenEditor, PromptForInput и пр.
```

---

## Анализ библиотек

| Библиотека | Описание | Плюсы | Минусы |
| ---------- | -------- | ----- | ------ |
| [charmbracelet/huh](https://github.com/charmbracelet/huh) | Формы / диалоги | Красивый TUI, composable формы, accessibility | Зависимость от Bubble Tea |
| [AlecAivazis/survey/v2](https://github.com/AlecAivazis/survey) | Промпты | Простой API, тестируемый, мощный | Архивирован (2023), нет активной поддержки |
| [manifoldco/promptui](https://github.com/manifoldco/promptui) | Списки/промпты | Минимальный API, лёгкий | Ограниченные возможности |
| **Свой** | internal/interactive | Полный контроль | Время на разработку |

### Рекомендация

**`charmbracelet/huh`** — лучший выбор:
- Активная разработка
- Тестируемый (`huh.NewForm(...).WithStdin(r).WithStdout(w)`)
- Вписывается в Go-экосистему (Charm)
- Поддерживает accessibility
- Composable: формы из групп, группы из полей

Если минимализм критичен — собственная обёртка поверх `bufio.Scanner` с единым API.

---

## Подзадачи

### 11.1 — Анализ и инвентаризация

**Цель:** Полный список всех интерактивных точек в проекте.

```bash
# Найти все интерактивные вызовы
grep -rn 'Scanf\|ReadString\|bufio.NewReader\|bufio.NewScanner' cmd/ internal/ --include='*.go' | grep -v _test.go
grep -rn 'PromptFor\|AskUser\|Confirm\|Select\|Choose' cmd/ internal/ --include='*.go' | grep -v _test.go
```

**Артефакт:** Таблица `[файл, тип интеракции, что вводит пользователь, тестируемость сейчас]`

---

### 11.2 — Проектирование API

**Цель:** Спроектировать `internal/interactive/` API.

**Минимальный контракт:**

```go
package interactive

// Prompter — интерфейс для всех интерактивных операций.
type Prompter interface {
    // Input — запросить текстовый ввод.
    Input(label string, opts ...InputOption) (string, error)
    
    // Confirm — да/нет вопрос.
    Confirm(label string, defaultVal bool) (bool, error)
    
    // Select — выбор из списка.
    Select(label string, items []string, opts ...SelectOption) (int, error)
    
    // MultiSelect — множественный выбор.
    MultiSelect(label string, items []string, opts ...SelectOption) ([]int, error)
    
    // Editor — открыть внешний редактор для ввода.
    Editor(label string, initial string) (string, error)
}

// InputOption — опции для Input.
type InputOption func(*inputConfig)

func WithDefault(v string) InputOption   { ... }
func WithValidation(fn func(string) error) InputOption { ... }
func WithPlaceholder(s string) InputOption { ... }

// SelectOption — опции для Select.
type SelectOption func(*selectConfig)

func WithFilter() SelectOption { ... }
func WithPageSize(n int) SelectOption { ... }
```

---

### 11.3 — Реализация Prompter

```go
// terminal.go — реализация для терминала
type TerminalPrompter struct {
    in  io.Reader
    out io.Writer
}

func NewTerminalPrompter(in io.Reader, out io.Writer) *TerminalPrompter { ... }

// mock.go — реализация для тестов
type MockPrompter struct {
    answers []any
    idx     int
}

func NewMockPrompter(answers ...any) *MockPrompter { ... }

// noninteractive.go — реализация для CI/CD
type NonInteractivePrompter struct{}
// Всегда возвращает defaults или ошибку "requires --interactive"
```

---

### 11.4 — Интеграция в Runner

```go
// В cmd/root.go или cmd/commands.go:
type CommandContext struct {
    Client   func() client.ClientInterface
    Prompter interactive.Prompter
    Output   *output.Writer
}

// При инициализации:
if nonInteractive {
    ctx.Prompter = interactive.NewNonInteractivePrompter()
} else {
    ctx.Prompter = interactive.NewTerminalPrompter(os.Stdin, os.Stdout)
}
```

---

### 11.5 — Миграция `cmd/add.go` (пилот)

Заменить ручные промпты в `cmd/add.go` на `Prompter`:

**До:**
```go
reader := bufio.NewReader(os.Stdin)
fmt.Print("Enter name: ")
name, _ := reader.ReadString('\n')
```

**После:**
```go
name, err := ctx.Prompter.Input("Enter name", interactive.WithValidation(notEmpty))
if err != nil {
    return err
}
```

---

### 11.6 — Миграция остальных команд

| Команда | Файлы | Тип интеракции |
| ------- | ----- | -------------- |
| `cmd/add.go` | add.go | Input (name, desc, etc.) |
| `cmd/update.go` | update.go | Input + Confirm |
| `cmd/sync/` | sync_wizard.go и пр. | Select + MultiSelect + Input |
| `cmd/compare/` | compare.go | Select + Confirm |
| `cmd/configurations/` | add_config.go | Input fields |
| `cmd/run/` | run.go | Confirm (запуск тестов?) |
| `cmd/delete.go` | delete.go | Confirm (подтверждение удаления) |

---

### 11.7 — Формы (вторая итерация)

Для сложных команд (add, update, configurations) — составные формы:

```go
form := interactive.NewForm().
    AddInput("Name", interactive.WithValidation(notEmpty)).
    AddInput("Description").
    AddSelect("Type", []string{"acceptance", "automated", "..."},
        interactive.WithDefault(0)).
    AddConfirm("Create?", true)

results, err := ctx.Prompter.RunForm(form)
```

---

### 11.8 — Тесты

**Требования:**
- Каждый компонент Prompter протестирован (Input, Confirm, Select, MultiSelect, Editor)
- MockPrompter используется во ВСЕХ тестах мигрированных команд
- NonInteractivePrompter: тест на корректные defaults / ошибки
- Покрытие `internal/interactive/` ≥ 95%
- Integration-тест: полный сценарий с MockPrompter для sync wizard

---

## Порядок выполнения

```
11.1 Инвентаризация
  │
  ▼
11.2 API design
  │
  ▼
11.3 Реализация (Terminal + Mock + NonInteractive)
  │
  ├──► 11.4 Интеграция в Runner
  │
  ▼
11.5 Пилот (cmd/add.go)
  │
  ├──► Validate approach
  │
  ▼
11.6 Миграция остальных команд
  │
  ▼
11.7 Формы (вторая итерация)
  │
  ▼
11.8 Тесты + финализация
```

---

## Критерии готовности (Definition of Done)

- [ ] Единый интерфейс `Prompter` в `internal/interactive/`
- [ ] 3 реализации: Terminal, Mock, NonInteractive
- [ ] 0 прямых вызовов `fmt.Scanf` / `bufio.Scanner` для пользовательского ввода
- [ ] Все тесты используют `MockPrompter`
- [ ] Флаг `--non-interactive` работает для всех команд
- [ ] Покрытие ≥ 95% для `internal/interactive/`
- [ ] Документация API с примерами использования
