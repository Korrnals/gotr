# Другие команды

## Экспорт данных

Экспорт в JSON-файл:

```bash
# Экспорт с автоматическим именем файла
gotr export projects get_projects

# Экспорт с указанием имени файла
gotr export cases get_cases 30 --suite-id 20069 -o my_cases.json

# Файлы сохраняются в директорию .testrail/
```

## Сравнение проектов

Сравнение кейсов между двумя проектами:

```bash
# Сравнение по названию (по умолчанию)
gotr compare cases --pid1 30 --pid2 31

# Сравнение по другому полю
gotr compare cases --pid1 30 --pid2 31 --field priority_id
```

Вывод:

- Кейсы только в проекте 1
- Кейсы только в проекте 2
- Кейсы с отличающимися полями

## Список эндпоинтов

```bash
# Все эндпоинты
gotr list all

# Эндпоинты для конкретного ресурса
gotr list cases
gotr list projects
gotr list suites

# Форматы вывода
gotr list cases --json      # JSON формат
gotr list cases --short     # Краткий вывод (Method URI)
```

## Конфигурация

```bash
# Создать конфиг
gotr config init

# Показать путь к конфигу
gotr config path

# Просмотреть конфиг
gotr config view

# Редактировать конфиг
gotr config edit
```

## Автодополнение

```bash
# Bash
source <(gotr completion bash)

# Zsh
gotr completion zsh > "${fpath[1]}/_gotr"

# Fish
gotr completion fish > ~/.config/fish/completions/gotr.fish
```

## Команды в разработке

Следующие команды находятся в разработке:

- `add` — создание ресурсов
- `delete` — удаление ресурсов
- `update` — обновление ресурсов
- `copy` — копирование между проектами
- `import` — импорт из файлов
