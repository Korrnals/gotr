# Утилита gotr (TestRail)

Утилита для взаимодействия с API TestRail.

## Основной функционал

1. Экспорт/Импорт Test Cases
2. Экспорт/Импорт Test Results

## Установка

## Надо сделать

0. TODO:
   > - реализовать возможность выбора необходимого URI из структуры;
   > - ДОПОЛНИТЬ вложенность в `gotr list {projects, ..} {уровень - конкретного ендпоинта}`;
   > - РЕАЛИЗОВАТЬ команды, для обработки ВЫБРАННЫХ (выбирать из списка субкоманд) URI

1. Базовое: **Замени код, собери проект и попробуй:**

   ```bash
   ./gotr list roles
   ./gotr list datasets
   ./gotr list sharedsteps #Убедись, что выводится корректный список.
   ```

2. Среднее: **Добавь флаг `--json` к `listCmd`:**

   ```go
   jsonOutput := cmd.Flag("json").Value.String() == "true"

   //если он установлен — выводи:
   json.MarshalIndent(paths, "", "  ").
   ```

3. Среднее: **Сделай автодополнение для `Cobra` (используй `ValidArgsFunction`), чтобы при нажатии Tab предлагались все ресурсы из списка выше.**

4. Среднее: **Добавь в `cmd/root.go` глобальный флаг `--verbose` (булевый). Выведи сообщение "Verbose mode enabled", если он установлен (можно в rootCmd.Run или через пре-ран хук).**

5. Продвинутое: **Используй библиотеку `Viper` (от того же автора, что и `Cobra`) для чтения конфига из файла/переменных окружения. Это следующий логичный шаг после флагов.**

6. Продвинутое: **Сделай флаг `--output` (строка: `"pretty"`, `"json"`, `"short"`) вместо двух булевых — это более гибко.**

### Дополнительные ресурсы

> - **Cobra: ValidArgs для автодополнения** — <https://github.com/spf13/cobra#validargs>
>
> - **Форматированный вывод JSON в CLI**  —  <https://pkg.go.dev/encoding/json#MarshalIndent>
>
> - **Cobra + Viper (официальный гайд)**  —  <https://github.com/spf13/cobra#working-with-viper>
>
> - **PersistentFlags vs Flags**  —  <https://pkg.go.dev/github.com/spf13/cobra#Command.PersistentFlags>
>
> - **Cobra: Flags и Run** — <https://github.com/spf13/cobra#working-with-flags>
>
> - **Cobra Completion** — <https://github.com/spf13/cobra#completions>
