package get

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Korrnals/gotr/cmd/common/flags/save"
	"github.com/Korrnals/gotr/internal/client"
	embed "github.com/Korrnals/gotr/embedded"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GetClientFunc — тип функции для получения клиента
type GetClientFunc func(cmd *cobra.Command) *client.HTTPClient

// Cmd — основная команда для GET-запросов
var Cmd = &cobra.Command{
	Use:   "get",
	Short: "GET-запросы к TestRail API",
	Long: `Выполняет GET-запросы к TestRail API.

Подкоманды:
	case               - получить один кейс по ID кейса
	cases              - получить кейсы проекта (требует ID проекта и ID сюиты)
	case-types         - получить список типов кейсов
	case-fields        - получить список полей кейсов
	case-history       - получить историю изменений кейса по ID кейса

	project            - получить один проект по ID проекта
	projects           - получить все проекты

	sharedstep         - получить один shared step по ID шага
	sharedsteps        - получить shared steps проекта (требует ID проекта)
	sharedstep-history - получить историю изменений shared step по ID шага

	suite              - получить одну тест-сюиту по ID сюиты
	suites             - получить тест-сюиты проекта (требует ID проекта)

Примеры:
	gotr get project 30
	gotr get projects

	gotr get case 12345
	gotr get cases 30 --suite-id 20069

	gotr get suite 20069
	gotr get suites 30
	
	gotr get sharedstep 45678
	gotr get sharedsteps 30
`,
}

var getClient GetClientFunc

// SetGetClientForTests устанавливает getClient для тестов
func SetGetClientForTests(fn GetClientFunc) {
	getClient = fn
}

// handleOutput — общая логика вывода, сохранения и jq
func handleOutput(command *cobra.Command, data any, start time.Time) error {
	quiet, _ := command.Flags().GetBool("quiet")
	outputFormat, _ := command.Flags().GetString("type")
	saveFile, _ := command.Flags().GetString("save")
	jqEnabled, _ := command.Flags().GetBool("jq")
	jqFilter, _ := command.Flags().GetString("jq-filter")
	bodyOnly, _ := command.Flags().GetBool("body-only")

	// Если jq не указан явно — берём из конфига
	if !jqEnabled {
		jqEnabled = viper.GetBool("jq_format")
	}

	if saveFile != "" {
		var toSave []byte
		var err error
		if bodyOnly {
			toSave, err = json.MarshalIndent(data, "", "  ")
		} else {
			full := struct {
				Status     string        `json:"status"`
				StatusCode int           `json:"status_code"`
				Duration   time.Duration `json:"duration"`
				Timestamp  time.Time     `json:"timestamp"`
				Data       any           `json:"data"`
			}{
				Status:     "200 OK",
				StatusCode: 200,
				Duration:   time.Since(start),
				Timestamp:  time.Now(),
				Data:       data,
			}
			toSave, err = json.MarshalIndent(full, "", "  ")
		}
		if err != nil {
			return fmt.Errorf("ошибка маршалинга: %w", err)
		}
		if err := os.WriteFile(saveFile, toSave, 0644); err != nil {
			return err
		}
		if !quiet {
			fmt.Printf("Ответ сохранён в %s\n", saveFile)
		}
	}

	if jqEnabled || jqFilter != "" {
		toSave, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("ошибка маршалинга для jq: %w", err)
		}
		if err := embed.RunEmbeddedJQ(toSave, jqFilter); err != nil {
			return err
		}
		return nil
	}

	if !quiet {
		switch outputFormat {
		case "json":
			pretty, _ := json.MarshalIndent(data, "", "  ")
			fmt.Println(string(pretty))
		case "json-full":
			full := struct {
				Status     string        `json:"status"`
				StatusCode int           `json:"status_code"`
				Duration   time.Duration `json:"duration"`
				Timestamp  time.Time     `json:"timestamp"`
				Data       any           `json:"data"`
			}{
				Status:     "200 OK",
				StatusCode: 200,
				Duration:   time.Since(start),
				Timestamp:  time.Now(),
				Data:       data,
			}
			pretty, _ := json.MarshalIndent(full, "", "  ")
			fmt.Println(string(pretty))
		default:
			fmt.Println("Table output not implemented yet")
		}
	}

	return nil
}

// Register регистрирует команду get и все её подкоманды
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	getClient = clientFn
	rootCmd.AddCommand(Cmd)

	// Добавляем подкоманды
	Cmd.AddCommand(casesCmd)
	Cmd.AddCommand(caseCmd)
	Cmd.AddCommand(caseTypesCmd)
	Cmd.AddCommand(caseFieldsCmd)
	Cmd.AddCommand(caseHistoryCmd)
	Cmd.AddCommand(projectsCmd)
	Cmd.AddCommand(projectCmd)
	Cmd.AddCommand(sharedStepsCmd)
	Cmd.AddCommand(sharedStepCmd)
	Cmd.AddCommand(sharedStepHistoryCmd)
	Cmd.AddCommand(suitesCmd)
	Cmd.AddCommand(suiteCmd)

	// Локальные флаги — только для подкоманд get и их детей
	for _, subCmd := range Cmd.Commands() {
		subCmd.Flags().StringP("type", "t", "json", "Формат вывода: json, json-full, table")
		save.AddFlag(subCmd)
		subCmd.Flags().BoolP("quiet", "q", false, "Тихий режим")
		subCmd.Flags().BoolP("jq", "j", false, "Включить jq-форматирование (переопределяет конфиг jq_format)")
		subCmd.Flags().String("jq-filter", "", "jq-фильтр")
		subCmd.Flags().BoolP("body-only", "b", false, "Сохранить только тело ответа (без метаданных)")
	}

	// Специфичные флаги для cases уже определены в конструкторе newCasesCmd
}
