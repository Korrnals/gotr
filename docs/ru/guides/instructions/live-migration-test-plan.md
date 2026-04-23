# Живой тест-план миграции на боевом сервере

> ⚠️ **КРИТИЧЕСКИ ВАЖНО — ПРОЧИТАТЬ ПЕРЕД НАЧАЛОМ**
>
> Этот документ описывает обязательный план проверки инструмента `gotr` на реальном TestRail-сервере
> **перед любой боевой миграцией данных**.
>
> - **Каждый шаг выполняется строго по порядку**
> - **Ничего не делается без явного подтверждения ответственного лица**
> - **Любое отклонение от плана — немедленная остановка**
> - **При малейшем сомнении — стоп, консультация, только потом продолжение**
>
> Несоблюдение этих правил может привести к необратимой потере данных в продуктивных проектах.

---

## Цель документа

Стандартизированный пошаговый тест-план для проверки полного цикла миграции gotr:
создание тестовых данных → полная миграция → snapshot → откат → очистка.

Выполняется **один раз перед боевой миграцией** на изолированных тестовых проектах.
После успешного прохождения — можно переходить к реальной миграции.

Краткая версия для оператора во время живого прогона:
- [Операторская карточка живого прогона](live-migration-operator-card.md)

---

## Актуальный профиль текущей кампании

Для текущей серии проверок используются фиксированные тестовые проекты:

- SRC: `Test1` (ID: `48`)
- DST: `Test2` (ID: `49`)

Правила для этого профиля:

- Проекты `48/49` **не удаляются** в конце шага
- Выполняется только очистка тестовых сущностей внутри проектов
- После устранения найденных багов выполняется **повторный полный прогон**:
  - интерактивный режим
  - неинтерактивный режим

---

## Режим автономного интерактивного прогона

Во время интерактивных шагов:

- поля `name/title/description/refs` и другие содержательные поля заполняются автоматически исполнителем
- оператор не тратит время на согласование текстовых полей
- оператор согласует только **этапы** и изменения стратегии

Базовая политика ответов на промпты:

- `Create snapshot before migration?` → `Yes`
- `Continue?` → `Yes` (если не запланирован останов)
- `Save mapping?` → `Yes`
- `Save result to file?` → `Yes` (или `No` по сценарию, но единообразно в рамках шага)

Цель: тестировать логику и надежность цепочек, а не ручной ввод текстов.

---

## Предварительные условия

### Технические

- [ ] Собран свежий бинарник: `make build` → проверить версию `./gotr version`
- [ ] API-ключ с правами **администратора** (создание/удаление проектов, suite, section, case, shared steps)
- [ ] Доступность сервера: `./gotr self-test` → все проверки зелёные
- [ ] Свободное место в `~/.gotr/` для снапшотов (минимум 100 МБ)

### Организационные

- [ ] Ответственное лицо за тест определено и присутствует
- [ ] Никаких параллельных миграций в это время на сервере
- [ ] Зафиксировано время начала теста

---

## Полностью интерактивный маршрут (без `--non-interactive`)

Если цель проверки — именно интерактивный UX gotr, используйте этот сквозной маршрут.

Ключевое правило:
- Во всех командах ниже используется только `-i` (интерактивный режим)
- Флаг `--non-interactive` в этом сценарии не используется

```bash
# Шаг 1: создать 2 тестовых проекта
./gotr add project -i
./gotr add project -i

# Шаг 2: наполнить SRC данными
./gotr add shared-step <SRC_PROJECT_ID> -i
./gotr add shared-step <SRC_PROJECT_ID> -i
./gotr add shared-step <SRC_PROJECT_ID> -i

./gotr add suite <SRC_PROJECT_ID> -i
./gotr add suite <SRC_PROJECT_ID> -i

./gotr add section <SRC_PROJECT_ID> -i
./gotr add section <SRC_PROJECT_ID> -i
./gotr add section <SRC_PROJECT_ID> -i

./gotr add case <SRC_SECTION_ID_1> -i
./gotr add case <SRC_SECTION_ID_1> -i
./gotr add case <SRC_SECTION_ID_1> -i
./gotr add case <SRC_SECTION_ID_2> -i
./gotr add case <SRC_SECTION_ID_2> -i
./gotr add case <SRC_SECTION_ID_3> -i
./gotr add case <SRC_SECTION_ID_3> -i

# Минимальные данные в DST
./gotr add suite <DST_PROJECT_ID> -i
./gotr add shared-step <DST_PROJECT_ID> -i

# Шаг 3: snapshot до миграции
./gotr snap create -i

# Шаг 4: полная миграция (интерактивно)
./gotr sync full -i

# Шаг 5: отчёт
./gotr report list
./gotr report view latest

# Шаг 6: rollback (интерактивно)
./gotr snap rollback -i

# Шаг 7: удаление тестовых проектов (только после явного OK)
./gotr delete project <SRC_PROJECT_ID> -i
./gotr delete project <DST_PROJECT_ID> -i
```

Что обязательно фиксировать во время интерактивного маршрута:
- ID созданных проектов, suite, section, case, shared step
- ID снапшота до миграции
- Путь к migration report (`~/.gotr/reports/migration-*.md`)
- Все подтверждения `Yes/No` на критичных шагах

---

## Шаг 0 — Baseline: фиксация состояния сервера

**Цель:** знать точное состояние до теста, чтобы после всё вернуть как было.

```bash
# Получить полный список проектов (сохранить вывод!)
./gotr get projects --non-interactive --format json > /tmp/gotr-test-baseline-projects.json

# Убедиться что тестовых проектов ещё нет
cat /tmp/gotr-test-baseline-projects.json | grep -i "GOTR-TEST"
# Ожидаемый результат: пустой вывод
```

**⛔ Стоп:** если `GOTR-TEST` уже есть в списке — разобраться откуда они, прежде чем продолжать.

---

## Шаг 1 — Создание тестовых проектов

**Создаются два изолированных проекта исключительно для теста.**

Все названия содержат префикс `[GOTR-TEST]` — чтобы их можно было однозначно идентифицировать и удалить.

```bash
# Создать SRC (источник миграции)
./gotr add project \
  --name "[GOTR-TEST] SRC Migration Source" \
  --announcement "⚠️ Тестовый проект-ИСТОЧНИК для проверки gotr миграции. Создан автоматически. Удалить после тестирования." \
  --non-interactive

# Сохранить ID проекта из вывода! Например: "id": 99
export GOTR_TEST_SRC_ID=<id из вывода>

# Создать DST (приёмник миграции)
./gotr add project \
  --name "[GOTR-TEST] DST Migration Target" \
  --announcement "⚠️ Тестовый проект-ПРИЁМНИК для проверки gotr миграции. Создан автоматически. Удалить после тестирования." \
  --non-interactive

export GOTR_TEST_DST_ID=<id из вывода>

echo "SRC=$GOTR_TEST_SRC_ID  DST=$GOTR_TEST_DST_ID"
```

**Проверка:**
```bash
./gotr get project $GOTR_TEST_SRC_ID --non-interactive
./gotr get project $GOTR_TEST_DST_ID --non-interactive
```

**⛔ Стоп:** оба проекта должны быть видны и содержать правильные имена.

---

## Шаг 2 — Наполнение SRC тестовыми данными

**Принцип наполнения:**
- Часть данных **совпадает по имени** в SRC и DST — чтобы проверить логику "уже существует / пропустить"
- Часть данных **уникальна для SRC** — чтобы проверить импорт новых
- Все названия с префиксом `[GOTR-TEST]`

### 2.1 — Shared steps (сначала, т.к. нужны при создании кейсов)

```bash
# Shared step 1 — будет совпадать в DST
./gotr add shared-step $GOTR_TEST_SRC_ID \
  --title "[GOTR-TEST] Open browser and navigate" \
  --non-interactive

# Shared step 2 — уникальный для SRC
./gotr add shared-step $GOTR_TEST_SRC_ID \
  --title "[GOTR-TEST] Fill login form with credentials" \
  --non-interactive

# Shared step 3 — уникальный для SRC
./gotr add shared-step $GOTR_TEST_SRC_ID \
  --title "[GOTR-TEST] Verify success notification" \
  --non-interactive
```

Сохранить ID всех shared steps из вывода.

### 2.2 — Suite в SRC

```bash
# Suite 1 — основной (совпадёт с тем что создадим в DST)
./gotr add suite $GOTR_TEST_SRC_ID \
  --name "[GOTR-TEST] Core Functionality" \
  --description "⚠️ Тестовый suite gotr. Удалить после теста." \
  --non-interactive

export GOTR_TEST_SRC_SUITE1_ID=<id из вывода>

# Suite 2 — уникальный для SRC
./gotr add suite $GOTR_TEST_SRC_ID \
  --name "[GOTR-TEST] Edge Cases" \
  --description "⚠️ Тестовый suite gotr. Удалить после теста." \
  --non-interactive

export GOTR_TEST_SRC_SUITE2_ID=<id из вывода>
```

### 2.3 — Sections

```bash
# В Suite 1
./gotr add section $GOTR_TEST_SRC_ID \
  --suite-id $GOTR_TEST_SRC_SUITE1_ID \
  --name "[GOTR-TEST] Authentication" \
  --non-interactive
export GOTR_TEST_SRC_SEC_AUTH=<id>

./gotr add section $GOTR_TEST_SRC_ID \
  --suite-id $GOTR_TEST_SRC_SUITE1_ID \
  --name "[GOTR-TEST] Payments" \
  --non-interactive
export GOTR_TEST_SRC_SEC_PAY=<id>

# В Suite 2
./gotr add section $GOTR_TEST_SRC_ID \
  --suite-id $GOTR_TEST_SRC_SUITE2_ID \
  --name "[GOTR-TEST] Boundary Values" \
  --non-interactive
export GOTR_TEST_SRC_SEC_BV=<id>
```

### 2.4 — Test cases

```bash
# Auth section — 3 кейса
./gotr add case $GOTR_TEST_SRC_SEC_AUTH \
  --title "[GOTR-TEST] Login with valid credentials" \
  --non-interactive

./gotr add case $GOTR_TEST_SRC_SEC_AUTH \
  --title "[GOTR-TEST] Login with invalid password" \
  --non-interactive

./gotr add case $GOTR_TEST_SRC_SEC_AUTH \
  --title "[GOTR-TEST] Logout" \
  --non-interactive

# Payments section — 2 кейса
./gotr add case $GOTR_TEST_SRC_SEC_PAY \
  --title "[GOTR-TEST] Successful payment" \
  --non-interactive

./gotr add case $GOTR_TEST_SRC_SEC_PAY \
  --title "[GOTR-TEST] Payment with insufficient funds" \
  --non-interactive

# Edge Cases section — 2 кейса
./gotr add case $GOTR_TEST_SRC_SEC_BV \
  --title "[GOTR-TEST] Max length input" \
  --non-interactive

./gotr add case $GOTR_TEST_SRC_SEC_BV \
  --title "[GOTR-TEST] Empty input validation" \
  --non-interactive
```

### 2.5 — Минимальные данные в DST (для проверки "уже существует")

```bash
# Suite с тем же именем что в SRC — чтобы sync мог его использовать как DST
./gotr add suite $GOTR_TEST_DST_ID \
  --name "[GOTR-TEST] Core Functionality" \
  --non-interactive
export GOTR_TEST_DST_SUITE_ID=<id>

# Один shared step с совпадающим именем
./gotr add shared-step $GOTR_TEST_DST_ID \
  --title "[GOTR-TEST] Open browser and navigate" \
  --non-interactive
```

**Проверка наполнения:**
```bash
./gotr get suites $GOTR_TEST_SRC_ID --non-interactive
./gotr get cases $GOTR_TEST_SRC_ID $GOTR_TEST_SRC_SUITE1_ID --non-interactive | grep -c '"id"'
./gotr get sharedsteps $GOTR_TEST_SRC_ID --non-interactive | grep -c '"id"'
# Ожидается: 2 suite, 7 кейсов, 3 shared steps в SRC
```

**⛔ Стоп:** перед следующим шагом убедиться что все данные созданы корректно.

---

## Шаг 3 — Создание snapshot состояния DST перед миграцией

```bash
./gotr snap create \
  --snap-label "before-gotr-test-migration" \
  --non-interactive
```

Сохранить ID снапшота из вывода — он понадобится для отката.

---

## Шаг 4A — Полная миграция SRC → DST (флаговый/полуавтоматический режим)

```bash
./gotr sync full \
  --src-project $GOTR_TEST_SRC_ID \
  --src-suite $GOTR_TEST_SRC_SUITE1_ID \
  --dst-project $GOTR_TEST_DST_ID \
  --dst-suite $GOTR_TEST_DST_SUITE_ID \
  --approve \
  --save-mapping \
  --snapshot
```

Этот вариант удобен для CI и повторяемых прогонов.

---

## Шаг 4B — Полная миграция SRC → DST (полностью интерактивный режим)

> Использовать, когда нужно протестировать именно поведение диалогов/подсказок CLI.
>
> Допустимо запускать как основной сценарий вместо шага 4A.

```bash
./gotr sync full -i
```

### Ожидаемый интерактивный сценарий

- [ ] Выбран SRC project: `[GOTR-TEST] SRC Migration Source`
- [ ] Выбран SRC suite: `[GOTR-TEST] Core Functionality`
- [ ] Выбран DST project: `[GOTR-TEST] DST Migration Target`
- [ ] Выбран DST suite: `[GOTR-TEST] Core Functionality`
- [ ] На вопрос подтверждения импорта дан ответ `Yes`
- [ ] На вопрос сохранения mapping дан ответ `Yes`
- [ ] На вопрос создания snapshot дан ответ `Yes`
- [ ] Сформирован итоговый migration report

### Что фиксировать в логах во время интерактивного прогона

- [ ] Блок `Filter result` (source/target/matched/new)
- [ ] Блок `Migration summary`
- [ ] Строку вида `Snapshot saved: ...`
- [ ] Строку вида `Migration report saved: ~/.gotr/reports/migration-*.md`

### Минимальный пример ответов на промпты

Ниже шаблон, фактические формулировки промптов могут немного отличаться по версии CLI:

```text
? Select source project: [GOTR-TEST] SRC Migration Source
? Select source suite: [GOTR-TEST] Core Functionality
? Select destination project: [GOTR-TEST] DST Migration Target
? Select destination suite: [GOTR-TEST] Core Functionality
? Continue with migration? Yes
? Save mapping file? Yes
? Save snapshot before/after migration? Yes
```

После выполнения шага 4B дальнейшие проверки и шаги (отчёт, откат, очистка) выполняются так же, как описано ниже.

**Проверить в процессе:**
- [ ] Отобразился filter result (source/target/matched/new)
- [ ] Показан migration summary перед импортом
- [ ] Снапшот создан автоматически
- [ ] В конце: `Migration report saved: ~/.gotr/reports/migration-*.md`

**Проверка после:**
```bash
./gotr get cases $GOTR_TEST_DST_ID $GOTR_TEST_DST_SUITE_ID --non-interactive | grep -c '"id"'
# Ожидается: кейсы из SRC перенесены

./gotr get sharedsteps $GOTR_TEST_DST_ID --non-interactive | grep -c '"id"'
# Shared steps 2 и 3 должны появиться; shared step 1 — уже был
```

**⛔ Стоп:** сверить кол-во кейсов в DST с тем что было в SRC.

---

## Шаг 5 — Просмотр отчёта миграции

```bash
# Список всех отчётов
./gotr report list

# Просмотр последнего
./gotr report view latest
```

**Что проверить в отчёте:**
- [ ] Тип миграции корректный (`full`)
- [ ] Source и Destination проекты верные
- [ ] Статистика: created / matched / failed
- [ ] Ссылка на снапшот присутствует
- [ ] Время выполнения отображается

Скопировать путь к файлу отчёта и открыть браузером / передать для проверки.

---

## Шаг 6 — Откат миграции через snapshot

```bash
# Посмотреть список снапшотов
./gotr snap list

# Откатить к снапшоту "до миграции"
./gotr snap rollback --snap-id <snap-id из шага 3>
```

**Проверка после отката:**
```bash
./gotr get cases $GOTR_TEST_DST_ID $GOTR_TEST_DST_SUITE_ID --non-interactive | grep -c '"id"'
# Ожидается: DST вернулся к состоянию ДО миграции (0 или только те кейсы что были до)

./gotr get sharedsteps $GOTR_TEST_DST_ID --non-interactive | grep -c '"id"'
# Ожидается: только 1 shared step (который был до)
```

**⛔ Стоп:** обязательно убедиться что откат прошёл корректно до перехода к удалению.

---

## Шаг 7 — Очистка тестовых сущностей (без удаления проектов)

> ⛔ **ВЫПОЛНЯЕТСЯ ТОЛЬКО С ЯВНОГО ПИСЬМЕННОГО ОДОБРЕНИЯ ОТВЕТСТВЕННОГО ЛИЦА**
>
> Для текущего профиля (`Test1=48`, `Test2=49`) сами проекты не удаляются.

Удаляются только тестовые сущности с префиксом `[GOTR-TEST]`:

```bash
# Примерно в таком порядке:
# 1) runs/plans (если создавались)
# 2) cases
# 3) sections
# 4) suites
# 5) shared steps (если есть отдельные)

# Проверить что в проектах не осталось тестовых suites/shared steps
./gotr get suites 48 --non-interactive --format json
./gotr get suites 49 --non-interactive --format json
./gotr get sharedsteps 48 --non-interactive --format json
./gotr get sharedsteps 49 --non-interactive --format json
```

**Ожидаемый результат:** в ответах нет тестовых сущностей текущей кампании.

---

## Шаг 8 — Повторный полный прогон после фикса багов

После исправления найденных дефектов выполнить повторно:

1. Полный интерактивный маршрут (`-i`) с автозаполнением полей
2. Полный неинтерактивный маршрут (`--non-interactive`) с явными флагами

Минимальный набор команд для rerun:

```bash
# Interactive
./gotr sync suites
./gotr sync sections
./gotr sync shared-steps
./gotr sync cases
./gotr sync full

# Non-interactive (пример)
./gotr sync suites --src-project 48 --dst-project 49 --approve --save-mapping --snapshot
./gotr sync sections --src-project 48 --src-suite <SRC_SUITE> --dst-project 49 --dst-suite <DST_SUITE> --approve --save-mapping --snapshot
./gotr sync shared-steps --src-project 48 --dst-project 49 --approve --save-mapping --snapshot
./gotr sync cases --src-project 48 --src-suite <SRC_SUITE> --dst-project 49 --dst-suite <DST_SUITE> --snapshot
./gotr sync full --src-project 48 --src-suite <SRC_SUITE> --dst-project 49 --dst-suite <DST_SUITE> --approve --save-mapping --snapshot
```

---

## Чеклист: критерии успешного прохождения теста

| # | Критерий | Результат |
|---|---|---|
| 1 | Бинарник собран без ошибок | ✓/✗ |
| 2 | Self-test зелёный | ✓/✗ |
| 3 | Оба тестовых проекта созданы | ✓/✗ |
| 4 | SRC наполнен: 2 suite, 7 кейсов, 3 shared steps | ✓/✗ |
| 5 | Снапшот DST создан до миграции | ✓/✗ |
| 6 | Миграция выполнена (шаг 4A или 4B), кейсы перенесены | ✓/✗ |
| 7 | Интерактивный сценарий (шаг 4B) пройден без ошибок | ✓/✗ |
| 8 | Отчёт миграции сохранён и читаем | ✓/✗ |
| 9 | Откат через снапшот выполнен корректно | ✓/✗ |
| 10 | DST вернулся к состоянию до миграции | ✓/✗ |
| 11 | Тестовые сущности очищены, проекты 48/49 сохранены | ✓/✗ |
| 12 | Повторный прогон после фиксов: interactive + non-interactive | ✓/✗ |

**Только при всех ✓ — можно переходить к реальной миграции.**

---

## Строгие правила выполнения

> Эти правила — не рекомендации. Это обязательные условия.

- **Никаких действий без явного "ОК" от ответственного лица** перед каждым деструктивным шагом (удаление, откат, перезапись)
- **Никакого параллельного выполнения** — шаги строго последовательные
- **Любая ошибка API — немедленная остановка**, анализ причины, только потом продолжение
- **Не использовать production проекты** как SRC или DST — только специально созданные `[GOTR-TEST]` проекты
- **Снапшот до миграции — обязателен** — без него нет отката
- **Сохранять вывод каждой команды** — для анализа если что-то пойдёт не так
- **Не торопиться** — лучше потратить час на тест, чем потерять данные в бою

---

## Восстановление при сбое

Если что-то пошло не так в любом шаге:

1. **Не паниковать** — зафиксировать точное сообщение об ошибке
2. Проверить снапшоты: `./gotr snap list`
3. Если снапшот есть — откатить: `./gotr snap rollback --snap-id <id>`
4. Если снапшота нет — обратиться к администратору TestRail
5. Проверить состояние DST вручную в веб-интерфейсе

---

*Документ создан в рамках проекта gotr. Актуальная версия — в репозитории: `docs/ru/guides/instructions/live-migration-test-plan.md`*
