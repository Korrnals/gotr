package sync

import (
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/service/migration"
	"github.com/spf13/cobra"
)

// init устанавливает getClient для тестов
func init() {
	SetGetClientForTests(func(cmd *cobra.Command) *client.HTTPClient {
		// В тестах используется mockClient через newMigration
		return nil
	})
}

// newMigrationFactoryFromMock создаёт factory для migration с mock клиентом
func newMigrationFactoryFromMock(t *testing.T, mock client.ClientInterface) func(client.ClientInterface, int64, int64, int64, int64, string, string) (*migration.Migration, error) {
	return func(cli client.ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*migration.Migration, error) {
		tmp := t.TempDir()
		return migration.NewMigration(mock, srcProject, srcSuite, dstProject, dstSuite, compareField, tmp)
	}
}
