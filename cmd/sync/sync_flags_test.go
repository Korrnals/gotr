package sync

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// ==================== Тесты для addSyncFlags ====================

func TestAddSyncFlags(t *testing.T) {
	cmd := &cobra.Command{}
	addSyncFlags(cmd)

	// Проверяем, что все флаги добавлены
	flags := []string{
		"src-project",
		"src-suite",
		"dst-project",
		"dst-suite",
		"compare-field",
		"dry-run",
		"approve",
		"save-mapping",
		"mapping-file",
		"output",
	}

	for _, flag := range flags {
		f := cmd.Flags().Lookup(flag)
		assert.NotNil(t, f, "Flag %s should be registered", flag)
	}
}

func TestAddSyncFlags_DefaultValues(t *testing.T) {
	cmd := &cobra.Command{}
	addSyncFlags(cmd)

	// Проверяем значения по умолчанию
	compareField, err := cmd.Flags().GetString("compare-field")
	assert.NoError(t, err)
	assert.Equal(t, "title", compareField)

	dryRun, err := cmd.Flags().GetBool("dry-run")
	assert.NoError(t, err)
	assert.False(t, dryRun)

	approve, err := cmd.Flags().GetBool("approve")
	assert.NoError(t, err)
	assert.False(t, approve)

	saveMapping, err := cmd.Flags().GetBool("save-mapping")
	assert.NoError(t, err)
	assert.False(t, saveMapping)

	srcProject, err := cmd.Flags().GetInt64("src-project")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), srcProject)

	srcSuite, err := cmd.Flags().GetInt64("src-suite")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), srcSuite)

	dstProject, err := cmd.Flags().GetInt64("dst-project")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), dstProject)

	dstSuite, err := cmd.Flags().GetInt64("dst-suite")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), dstSuite)
}

func TestAddSyncFlags_SetValues(t *testing.T) {
	cmd := &cobra.Command{}
	addSyncFlags(cmd)

	// Устанавливаем значения флагов
	cmd.Flags().Set("src-project", "10")
	cmd.Flags().Set("src-suite", "20")
	cmd.Flags().Set("dst-project", "30")
	cmd.Flags().Set("dst-suite", "40")
	cmd.Flags().Set("compare-field", "name")
	cmd.Flags().Set("dry-run", "true")
	cmd.Flags().Set("approve", "true")
	cmd.Flags().Set("save-mapping", "true")
	cmd.Flags().Set("mapping-file", "mapping.json")
	cmd.Flags().Set("output", "output.json")

	// Проверяем установленные значения
	srcProject, _ := cmd.Flags().GetInt64("src-project")
	assert.Equal(t, int64(10), srcProject)

	srcSuite, _ := cmd.Flags().GetInt64("src-suite")
	assert.Equal(t, int64(20), srcSuite)

	dstProject, _ := cmd.Flags().GetInt64("dst-project")
	assert.Equal(t, int64(30), dstProject)

	dstSuite, _ := cmd.Flags().GetInt64("dst-suite")
	assert.Equal(t, int64(40), dstSuite)

	compareField, _ := cmd.Flags().GetString("compare-field")
	assert.Equal(t, "name", compareField)

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	assert.True(t, dryRun)

	approve, _ := cmd.Flags().GetBool("approve")
	assert.True(t, approve)

	saveMapping, _ := cmd.Flags().GetBool("save-mapping")
	assert.True(t, saveMapping)

	mappingFile, _ := cmd.Flags().GetString("mapping-file")
	assert.Equal(t, "mapping.json", mappingFile)

	output, _ := cmd.Flags().GetString("save")
	assert.Equal(t, "output.json", output)
}
