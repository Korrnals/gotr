# gotr — Makefile для сборки и установки утилиты

# Название бинарника
BINARY_NAME=gotr

# Приоритет версии:
# 1. Явно указанная при вызове make (make VERSION=v1.0.0)
# 2. Git tag (если есть)
# 3. "dev"
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "")
ifeq ($(VERSION),)
    VERSION = dev
endif

# Если VERSION всё ещё пустой или содержит постфикс (например, v1.0.0-3-gabc123), оставляем только тег
# Но если явно указана — оставляем как есть
ifneq ($(filter v%,$(VERSION)),)
    # Если это чистый тег (v1.0.0) — ок
    # Если с постфиксом — берём только тег
    VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
endif

# Коммит и дата для дополнительной информации
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Флаги для встраивания версии
LDFLAGS = -ldflags="-s -w -X gotr/cmd.Version=$(VERSION) -X gotr/cmd.Commit=$(COMMIT) -X gotr/cmd.Date=$(DATE)"

# Цель по умолчанию
all: build

# Сборка
build:
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

tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Укажите версию: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Создание тега $(VERSION)"
	git tag -a $(VERSION) -m "Релиз $(VERSION)"
	git push origin $(VERSION)
	@echo "Тег $(VERSION) создан и отправлен"

# Пример использования:
# make tag VERSION=v1.0.0

# Помощь
help:
	@echo "Доступные цели:"
	@echo "  build       — собрать бинарник для текущей платформы"
	@echo "  compress	 — сжать бинарник UPX (если установлен)"
	@echo "  install     — установить в /usr/local/bin (требует sudo)"
	@echo "  test        — запустить тесты"
	@echo "  clean       — удалить бинарник"
	@echo "  release     — собрать для Linux, macOS и Windows"
	@echo "  help        — показать эту справку"

.PHONY: all build test-build test install clean build-linux build-darwin build-windows release help