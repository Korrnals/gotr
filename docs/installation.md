# Установка

## Быстрая установка (Linux/macOS)

```bash
# Unix
curl -s -L https://github.com/Korrnals/gotr/releases/latest/download/gotr-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o gotr && chmod +x gotr && sudo mv gotr /usr/local/bin/
```

> **Примечание:** Для Windows скачайте .exe вручную из [Releases](https://github.com/Korrnals/gotr/releases).

## Сборка из исходников

### Требования

- Go 1.21+
- (Опционально) UPX для сжатия

### Вариант 1: Простая сборка

```bash
git clone https://github.com/Korrnals/gotr.git
cd gotr
go build -ldflags="-s -w" -o gotr
sudo mv gotr /usr/local/bin/
```

### Вариант 2: Через Makefile (рекомендуется)

```bash
git clone https://github.com/Korrnals/gotr.git
cd gotr

# Сборка и установка
make install

# Другие команды:
make build          # только сборка
make test           # запуск тестов
make compress       # сжатие UPX
make build-compressed  # сборка + сжатие
make clean          # очистка
make release        # сборка для всех платформ
```

### Сборка с версией

```bash
# Без тега - версия "dev"
make build
# gotr version - dev

# С тегом
git tag v2.0.0
make build
# gotr version - v2.0.0
```

## Проверка установки

```bash
gotr --help
gotr version
```
