# gotr — Makefile для сборки и установки утилиты

# Название бинарника
BINARY_NAME=gotr

# Приоритет версии:
# 1. Явно указанная при вызове make (make VERSION=v2.6.0) - HIGHEST PRIORITY
# 2. Версия из cmd/root.go (извлекается автоматически)
#
# При сборке без VERSION:
#   - Извлекается версия из cmd/root.go
#   - Для релизных версий (без -dev) создается/проверяется git tag
VERSION ?=

# Извлекаем версию из cmd/root.go если не указана явно
ifeq ($(VERSION),)
    CODE_VERSION := $(shell grep -E '^[[:space:]]*Version[[:space:]]*=[[:space:]]*"' cmd/root.go | sed 's/.*"\([^"]*\)".*/\1/')
    VERSION := $(CODE_VERSION)
endif

# Коммит и дата для дополнительной информации
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Флаги для встраивания версии
PACKAGE_PATH = github.com/Korrnals/gotr/cmd
LDFLAGS = -ldflags="-s -w -X '$(PACKAGE_PATH).Version=$(VERSION)' -X '$(PACKAGE_PATH).Commit=$(COMMIT)' -X '$(PACKAGE_PATH).Date=$(DATE)'"

# Цель по умолчанию
all: build

# Синхронизация git tag для релизных версий (без -dev)
# Нормализация тега: убираем префикс v если есть, потом добавляем
VERSION_TAG := $(VERSION:v%=%)
sync-tag:
ifeq ($(findstring -dev,$(VERSION)),)
	@echo "Проверка git tag для релизной версии $(VERSION)..."
	@git fetch --tags 2>/dev/null || true
	@if ! git tag -l "v$(VERSION_TAG)" | grep -q "v$(VERSION_TAG)"; then \
		echo "Создание git tag v$(VERSION_TAG)..."; \
		git tag -a "v$(VERSION_TAG)" -m "Release $(VERSION_TAG)"; \
		echo "✓ Git tag v$(VERSION_TAG) создан локально"; \
		echo "  Для отправки выполните: git push origin v$(VERSION_TAG)"; \
	else \
		echo "✓ Git tag v$(VERSION_TAG) уже существует"; \
	fi
else
	@echo "Dev-версия ($(VERSION)) - git tag не требуется"
endif

# Сборка
build: sync-tag
	@echo "Сборка $(BINARY_NAME) версии $(VERSION) (commit: $(COMMIT))"
	go build $(LDFLAGS) -o $(BINARY_NAME)

# Сжатие бинарника UPX (опционально, если установлен upx)
compress:
	@echo "Сжатие $(BINARY_NAME) с помощью UPX (если установлен)..."
	@if command -v upx >/dev/null 2>&1; then \
		if [ "$(BINARY_NAME)" = "gotr-darwin-amd64" ] || [ "$(BINARY_NAME)" = "gotr-darwin-arm64" ]; then \
			upx --best --force-macos $(BINARY_NAME); \
			echo "$(BINARY_NAME) сжат UPX (с --force-macos для macOS)"; \
		else \
			upx --best $(BINARY_NAME); \
			echo "$(BINARY_NAME) сжат UPX"; \
		fi \
	else \
		echo "UPX не установлен — пропускаем сжатие"; \
	fi

# Сборка с тестом
test-build: test build

# Тестирование
test:
	@echo "Запуск тестов..."
	go test ./... -v

# Установка в /usr/local/bin (требует sudo)
install: build
	@echo "Установка $(BINARY_NAME) в /usr/local/bin..."
	sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "$(BINARY_NAME) успешно установлен!"

# Очистка
clean:
	@echo "Очистка..."
	rm -f $(BINARY_NAME)

# Кросс-компиляция (примеры)
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME).exe

# Полная сборка для всех платформ
release: build-linux build-darwin build-windows

# Сборка релизных бинарников со сжатием UPX
release-compressed: clean
	@echo "Сборка релизных бинарников v$(VERSION)..."
	@echo "Linux amd64..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64
	@if command -v upx >/dev/null 2>&1; then upx --best $(BINARY_NAME)-linux-amd64; fi
	@echo "macOS amd64..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64
	@if command -v upx >/dev/null 2>&1; then upx --best --force-macos $(BINARY_NAME)-darwin-amd64; fi
	@echo "Windows amd64..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe
	@if command -v upx >/dev/null 2>&1; then upx --best $(BINARY_NAME)-windows-amd64.exe; fi
	@echo "Готово!"
	@ls -lh $(BINARY_NAME)-*

# Сборка + сжатие
build-compressed: build compress

# Создание git tag и push
tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Укажите версию: make tag VERSION=v2.6.0"; \
		exit 1; \
	fi
	@echo "Создание тега v$(VERSION)"
	git tag -a "v$(VERSION)" -m "Release $(VERSION)"
	git push origin "v$(VERSION)"
	@echo "Тег v$(VERSION) создан и отправлен"

# Помощь
help:
	@echo "Доступные цели:"
	@echo "  build       — собрать бинарник (версия из cmd/root.go)"
	@echo "  compress    — сжать бинарник UPX (если установлен)"
	@echo "  install     — установить в /usr/local/bin (требует sudo)"
	@echo "  test        — запустить тесты"
	@echo "  clean       — удалить бинарник"
	@echo "  release     — собрать для Linux, macOS и Windows"
	@echo "  tag         — создать и отправить git tag"
	@echo ""
	@echo "Примеры:"
	@echo "  make build                    # Сборка с версией из кода"
	@echo "  make build VERSION=v2.6.0     # Сборка с явной версией"
	@echo "  make tag VERSION=v2.6.0       # Создание релизного тега"

.PHONY: all build test-build test install clean build-linux build-darwin build-windows release help sync-tag
