package sync

import (
	"context"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/spf13/cobra"
)

// GetClientFunc is the function type for obtaining a client.
type GetClientFunc = client.GetClientFunc

// Cmd is the parent command for migration.
var Cmd = &cobra.Command{
	Use:   "sync",
	Short: "Синхронизация данных TestRail между проектами",
	Long: `Родительская команда для миграции данных между проектами TestRail.

Поддерживает интерактивный режим выбора проектов и сьютов.
Если параметры не указаны — будет предложен выбор из списка.

Подкоманды:
	• shared-steps — миграция общих шагов (генерирует mapping)
	• cases        — миграция кейсов (требует mapping)
	• full         — полная миграция (shared-steps + cases за один проход)
	• suites       — миграция suites между проектами
	• sections     — миграция sections между сюитами

Логи и mapping сохраняются в директории: .testrail (лог-файлы находятся в .testrail/logs/)

Примеры:
	# Интерактивный режим (выбор всех параметров)
	gotr sync full
	gotr sync cases

	# Через флаги
	gotr sync full --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve --save-mapping
	gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --mapping shared_steps_mapping.json --dry-run
	gotr sync shared-steps --src-project 30 --src-suite 20069 --dst-project 31 --approve --output shared_steps_mapping.json
`,
}

var clientAccessor *client.Accessor

// SetGetClientForTests sets getClient for tests.
func SetGetClientForTests(fn GetClientFunc) {
	if clientAccessor == nil {
		clientAccessor = client.NewAccessor(fn)
	} else {
		clientAccessor.SetClientForTests(fn)
	}
}

// testClientKey is the context key for mock client in tests.
var testClientKey = &struct{}{}

// SetTestClient sets the mock client for tests.
func SetTestClient(cmd *cobra.Command, mockClient client.ClientInterface) {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.TODO()
	}
	cmd.SetContext(context.WithValue(ctx, testClientKey, mockClient))
}

// getClientSafe safely calls getClient with a nil check.
// Fallback: gets client from context (for tests).
func getClientSafe(cmd *cobra.Command) client.ClientInterface {
	if clientAccessor != nil {
		if c := clientAccessor.GetClientSafe(cmd.Context()); c != nil {
			return c
		}
	}
	// Fallback for old tests — get from context by old key
	if ctx := cmd.Context(); ctx != nil {
		if v := ctx.Value(testHTTPClientKey); v != nil {
			if c, ok := v.(client.ClientInterface); ok {
				return c
			}
		}
	}
	return nil
}

// getClientInterface safely returns ClientInterface (for tests with MockClient).
func getClientInterface(cmd *cobra.Command) client.ClientInterface {
	// First check the new key for mock clients
	if v := cmd.Context().Value(testClientKey); v != nil {
		if c, ok := v.(client.ClientInterface); ok {
			return c
		}
	}
	// Fallback: use regular getClientSafe
	return getClientSafe(cmd)
}

// Register registers the sync command and all its subcommands.
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
	clientAccessor = client.NewAccessor(clientFn)
	rootCmd.AddCommand(Cmd)

	Cmd.AddCommand(sharedStepsCmd)
	Cmd.AddCommand(casesCmd)
	Cmd.AddCommand(fullCmd)
	Cmd.AddCommand(suitesCmd)
	Cmd.AddCommand(sectionsCmd)

	// Flags for sync cases
	casesCmd.Flags().Int64("src-project", 0, "ID source проекта (откуда копировать)")
	casesCmd.Flags().Int64("src-suite", 0, "ID source сьюта")
	casesCmd.Flags().Int64("dst-project", 0, "ID destination проекта (куда копировать)")
	casesCmd.Flags().Int64("dst-suite", 0, "ID destination сьюта")
	casesCmd.Flags().String("compare-field", "title", "Поле для поиска дубликатов")
	casesCmd.Flags().String("mapping-file", "", "Файл mapping для замены shared_step_id")
	casesCmd.Flags().Bool("dry-run", false, "Просмотр без импорта")
	casesCmd.Flags().String("output", "", "Дополнительный JSON файл с результатами")

	// Flags for sync shared-steps
	sharedStepsCmd.Flags().Int64("src-project", 0, "ID source проекта")
	sharedStepsCmd.Flags().Int64("src-suite", 0, "ID source сьюта")
	sharedStepsCmd.Flags().Int64("dst-project", 0, "ID destination проекта")
	sharedStepsCmd.Flags().String("compare-field", "title", "Поле для поиска дубликатов")
	sharedStepsCmd.Flags().Bool("approve", false, "Автоматическое подтверждение")
	sharedStepsCmd.Flags().Bool("save-mapping", false, "Сохранить mapping автоматически")
	sharedStepsCmd.Flags().Bool("save-filtered", false, "Сохранить filtered список автоматически")
	sharedStepsCmd.Flags().String("output", "", "Файл для сохранения mapping")
	sharedStepsCmd.Flags().Bool("dry-run", false, "Просмотр без импорта")

	// Flags for sync sections
	sectionsCmd.Flags().Int64("src-project", 0, "ID source проекта")
	sectionsCmd.Flags().Int64("src-suite", 0, "ID source сьюта")
	sectionsCmd.Flags().Int64("dst-project", 0, "ID destination проекта")
	sectionsCmd.Flags().Int64("dst-suite", 0, "ID destination сьюта")
	sectionsCmd.Flags().String("compare-field", "title", "Поле для поиска дубликатов")
	sectionsCmd.Flags().Bool("approve", false, "Автоматическое подтверждение")
	sectionsCmd.Flags().Bool("dry-run", false, "Просмотр без импорта")
	sectionsCmd.Flags().Bool("save-mapping", false, "Сохранить mapping автоматически")

	// Flags for sync full
	fullCmd.Flags().Int64("src-project", 0, "ID source проекта")
	fullCmd.Flags().Int64("src-suite", 0, "ID source сьюта")
	fullCmd.Flags().Int64("dst-project", 0, "ID destination проекта")
	fullCmd.Flags().Int64("dst-suite", 0, "ID destination сьюта")
	fullCmd.Flags().String("compare-field", "title", "Поле для поиска дубликатов")
	fullCmd.Flags().Bool("approve", false, "Автоматическое подтверждение")
	fullCmd.Flags().Bool("save-mapping", false, "Сохранить mapping автоматически")
	fullCmd.Flags().Bool("dry-run", false, "Просмотр без импорта")
}
