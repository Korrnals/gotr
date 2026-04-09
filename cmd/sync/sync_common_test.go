package sync

import (
	"context"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/Korrnals/gotr/internal/service/migration"
)

// init sets getClient for tests
func init() {
	SetGetClientForTests(func(ctx context.Context) client.ClientInterface {
		// In tests, mockClient is used via newMigration
		return nil
	})
}

// newMigrationFactoryFromMock creates a migration factory with a mock client
func newMigrationFactoryFromMock(t *testing.T, mock client.ClientInterface) func(client.ClientInterface, int64, int64, int64, int64, string, string) (*migration.Migration, error) {
	return func(cli client.ClientInterface, srcProject, srcSuite, dstProject, dstSuite int64, compareField, logDir string) (*migration.Migration, error) {
		tmp := t.TempDir()
		return migration.NewMigration(mock, srcProject, srcSuite, dstProject, dstSuite, compareField, tmp)
	}
}
