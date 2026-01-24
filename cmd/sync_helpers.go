package cmd

import "gotr/internal/migration"

// newMigration — seam для тестов; по умолчанию указывает на migration.NewMigration
var newMigration = migration.NewMigration
