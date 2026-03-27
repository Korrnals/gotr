# Release Branch Flow — Процесс выпуска релизов

## Общая идея

Проект использует **Release Branch Flow** — классическую модель, где:

- Ветка `main` содержит только стабильный, протестированный код
- Для каждого релиза создаётся **release-ветка** (`release-X.Y.Z`)
- Feature-ветки мержатся в release-ветку через **Pull Request**
- Финальный PR из release-ветки в `main` закрывает релиз

### Схема

```
main ───────────────────────────────────────────── main
  │                                                  ↑
  └── release-X.Y.Z ────────────────────────────── PR (финальный)
          ↑            ↑            ↑
          │            │            │
      PR(feat-A)   PR(feat-B)  PR(hotfix)
          ↑            ↑            ↑
          │            │            │
     feature-A    feature-B    hotfix-branch
```

---

## Как создать release-ветку

```bash
# 1. Синхронизируемся с remote
git fetch origin

# 2. Создаём release-ветку от удалённого main
git checkout -b release-X.Y.Z origin/main

# 3. Публикуем
git push -u origin release-X.Y.Z
```

**Важно:**
- Всегда создавать от `origin/main` (не от локального `main`)
- Именование: `release-` + semver, например `release-3.0.0`
- Одна release-ветка на один релиз

---

## Как мержить feature в release

1. Убедиться, что тесты проходят на feature-ветке:
   ```bash
   go test ./...
   go build ./...
   ```

2. Запушить ветку в remote:
   ```bash
   git push -u origin stage-6.7-recursive-parallelization
   ```

3. Создать PR на GitHub:
   - **Base:** `release-X.Y.Z`
   - **Compare:** `stage-N.M-description`
   - Описание: что сделано, что изменилось

4. **Merge PR через GitHub** (не локально!)

5. Feature-ветку **не удалять** — она сохраняется для истории

---

## Как финализировать релиз

1. Убедиться, что все feature-ветки замержены в release-ветку

2. Проверить release-ветку:
   ```bash
   git checkout release-X.Y.Z
   git pull origin release-X.Y.Z

   # Тесты
   go test ./...

   # Сборка
   go build -o gotr ./...

   # Smoke-тесты (опционально)
   ./gotr compare cases -p1 <P1> -p2 <P2>
   ```

3. Обновить `CHANGELOG.md`

4. Создать PR: `release-X.Y.Z` → `main`

5. Merge PR

6. Создать тег:
   ```bash
   git checkout main
   git pull origin main
   git tag -a vX.Y.Z -m "Release X.Y.Z"
   git push origin vX.Y.Z
   ```

---

## Что НЕЛЬЗЯ делать

| Действие | Почему нельзя |
|----------|---------------|
| `git merge feature-X` в `main` локально | Обходит code review и CI |
| `git push origin main` с локальными коммитами | То же самое |
| Удалять feature-ветки сразу после мержа | Теряется история для отладки |
| Мержить напрямую в main без release-ветки | Нарушает процесс стабилизации |

---

## Именование веток

| Тип | Шаблон | Пример |
|-----|--------|--------|
| Release | `release-X.Y.Z` | `release-3.0.0` |
| Feature/Stage | `stage-N.M-description` | `stage-6.8-concurrency-unification` |
| Hotfix | `fix-description` | `fix-reporter-alignment` |

---

## Пример: Release 2.8.0

```
main
  └── release-3.0.0         ← от origin/main
         ↑
         ├── PR: stage-6.7-recursive-parallelization
         │       (параллелизация + reporter rewrite)
         │
         └── PR: stage-6.8-concurrency-unification
                 (generic compare + concurrency package)

release-3.0.0 → PR → main → tag v3.0.0
```

---

## Быстрый чеклист

- [ ] `git fetch origin`
- [ ] `git checkout -b release-X.Y.Z origin/main`
- [ ] `git push -u origin release-X.Y.Z`
- [ ] PR: feature → release — для каждой feature-ветки
- [ ] `go test ./...` на release-ветке
- [ ] `go build ./...` на release-ветке
- [ ] CHANGELOG.md обновлён
- [ ] PR: release → main
- [ ] Tag `vX.Y.Z`

---

## Минимальные Quality Gates (Stage 13)

Перед merge любого PR в release-ветку обязательны проверки:

- `go test ./...`
- `go vet ./...`
- `go build ./...`
- `go test -race ./...` (в CI окружении с `CGO_ENABLED=1` и установленным gcc/clang)
- `govulncheck ./...` (или эквивалентный vulnerability scan)

### Принцип CI parity

- Локальная `verify`-цель Makefile и CI pipeline должны проверять одинаковый набор gates.
- Tagging/release операции не должны быть частью обычной `build`-цели.
- Для release-артефактов должны публиковаться checksums (например SHA256).
