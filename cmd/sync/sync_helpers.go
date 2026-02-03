package sync

import "github.com/Korrnals/gotr/internal/service/migration"

// newMigration — seam для тестов; по умолчанию указывает на migration.NewMigration
var newMigration = migration.NewMigration
