# Операторская карточка: живой прогон миграции

> ⚠️ Критично: выполнять шаги строго по порядку.
>
> ⚠️ Ничего не выполнять без явного подтверждения ответственного.
>
> ⚠️ Только `[GOTR-TEST]` проекты. Никаких продуктивных проектов.

## Переменные (заполнить один раз перед стартом)

Скопировать блок ниже, подставить значения и выполнить в терминале.

```bash
# IDs проектов
export SRC_PROJECT_ID=""
export DST_PROJECT_ID=""

# IDs suite/section в SRC
export SRC_SUITE_CORE_ID=""
export SRC_SUITE_EDGE_ID=""
export SRC_SECTION_1_ID=""
export SRC_SECTION_2_ID=""
export SRC_SECTION_3_ID=""

# ID suite в DST (для миграции)
export DST_SUITE_CORE_ID=""

# Снапшот до миграции
export SNAP_ID_BEFORE=""

# Префикс тестовых сущностей
export TEST_PREFIX="[GOTR-TEST]"
```

Проверка, что переменные заполнены:

```bash
echo "SRC_PROJECT_ID=$SRC_PROJECT_ID"
echo "DST_PROJECT_ID=$DST_PROJECT_ID"
echo "SRC_SUITE_CORE_ID=$SRC_SUITE_CORE_ID"
echo "SRC_SECTION_1_ID=$SRC_SECTION_1_ID"
echo "DST_SUITE_CORE_ID=$DST_SUITE_CORE_ID"
echo "SNAP_ID_BEFORE=$SNAP_ID_BEFORE"
```

## 0) Подготовка

- [ ] Собран свежий бинарник
- [ ] Проверен доступ: `./gotr self-test`
- [ ] Подтверждены права на запись (create/update/delete)
- [ ] Зафиксировано время старта

```bash
make build
./gotr version
./gotr self-test
```

## 1) Создать 2 тестовых проекта (интерактивно)

- [ ] Создан SRC проект `[GOTR-TEST] SRC Migration Source`
- [ ] Создан DST проект `[GOTR-TEST] DST Migration Target`
- [ ] Сохранены `SRC_PROJECT_ID` и `DST_PROJECT_ID`

```bash
./gotr add project -i
./gotr add project -i
```

## 2) Наполнить SRC данными (интерактивно)

- [ ] 3 shared steps
- [ ] 2 suites
- [ ] 3 sections
- [ ] 7 cases
- [ ] Все сущности с префиксом `[GOTR-TEST]`
- [ ] Описания содержат пометку:
- [ ] `⚠️ Создано автоматически инструментом gotr для тестирования миграции. Можно удалить.`

```bash
./gotr add shared-step $SRC_PROJECT_ID -i
./gotr add shared-step $SRC_PROJECT_ID -i
./gotr add shared-step $SRC_PROJECT_ID -i

./gotr add suite $SRC_PROJECT_ID -i
./gotr add suite $SRC_PROJECT_ID -i

./gotr add section $SRC_PROJECT_ID -i
./gotr add section $SRC_PROJECT_ID -i
./gotr add section $SRC_PROJECT_ID -i

./gotr add case $SRC_SECTION_1_ID -i
./gotr add case $SRC_SECTION_1_ID -i
./gotr add case $SRC_SECTION_1_ID -i
./gotr add case $SRC_SECTION_2_ID -i
./gotr add case $SRC_SECTION_2_ID -i
./gotr add case $SRC_SECTION_3_ID -i
./gotr add case $SRC_SECTION_3_ID -i
```

## 3) Минимальные данные в DST (интерактивно)

- [ ] Создан suite с тем же именем, что в SRC
- [ ] Создан 1 shared step с тем же именем, что в SRC

```bash
./gotr add suite $DST_PROJECT_ID -i
./gotr add shared-step $DST_PROJECT_ID -i
```

## 4) Snapshot до миграции

- [ ] Snapshot создан
- [ ] Сохранён `SNAP_ID_BEFORE`

```bash
./gotr snap create -i
```

## 5) Полная миграция (интерактивно)

- [ ] Выполнен `sync full -i`
- [ ] Подтверждены все промпты критичных действий
- [ ] В логе есть `Filter result`
- [ ] В логе есть `Migration summary`
- [ ] В логе есть `Migration report saved: ...`

```bash
./gotr sync full -i
```

## 6) Проверка отчёта

- [ ] Список отчётов получен
- [ ] Открыт `latest`
- [ ] Проверены source/destination, stats, snapshot reference
- [ ] Сохранён путь к отчёту

```bash
./gotr report list
./gotr report view latest
```

## 7) Откат (интерактивно)

- [ ] Запущен rollback
- [ ] Подтверждён откат к `SNAP_ID_BEFORE`
- [ ] Проверено, что DST вернулся в исходное состояние

```bash
./gotr snap rollback -i
```

## 8) Очистка (только после явного OK)

- [ ] Получено письменное подтверждение на удаление
- [ ] Удалён SRC
- [ ] Удалён DST
- [ ] Проверено отсутствие `[GOTR-TEST]` проектов

```bash
./gotr delete project $SRC_PROJECT_ID -i
./gotr delete project $DST_PROJECT_ID -i
./gotr get projects --non-interactive | grep "GOTR-TEST"
```

## Стоп-правила

- [ ] Любая 4xx/5xx ошибка API -> немедленный стоп
- [ ] Нет snapshot перед миграцией -> миграция запрещена
- [ ] Нет подтверждения ответственного -> деструктивные шаги запрещены
- [ ] Любое сомнение в корректности проекта -> стоп и ручная валидация в UI
