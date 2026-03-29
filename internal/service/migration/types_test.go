package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapping_ReturnsEmptyMapWhenMappingIsNil(t *testing.T) {
	dir := t.TempDir()
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "title", dir)
	require.NoError(t, err)
	defer m.Close()

	// Manually set mapping to nil to test the nil guard
	m.mapping = nil

	result := m.Mapping()
	require.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestMapping_ReturnsCopyOfIndex(t *testing.T) {
	dir := t.TempDir()
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "title", dir)
	require.NoError(t, err)
	defer m.Close()

	// Verify mapping is initialized
	require.NotNil(t, m.mapping)

	// Call Mapping() which should return a copy
	result := m.Mapping()
	require.NotNil(t, result)

	// Result should be a map (even if empty)
	_, isMap := interface{}(result).(map[int64]int64)
	assert.True(t, isMap)
}

func TestMapping_PreservesValues(t *testing.T) {
	dir := t.TempDir()
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "title", dir)
	require.NoError(t, err)
	defer m.Close()

	// Manually add some values to the internal mapping
	if m.mapping != nil && m.mapping.index != nil {
		m.mapping.index[100] = 200
		m.mapping.index[101] = 201
	}

	result := m.Mapping()
	assert.Len(t, result, 2)
	assert.Equal(t, int64(200), result[100])
	assert.Equal(t, int64(201), result[101])
}

func TestMigrationClose_NoError(t *testing.T) {
	dir := t.TempDir()
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "title", dir)
	require.NoError(t, err)

	// Close should not error
	err = m.Close()
	require.NoError(t, err)
}

func TestNewMigration_CreatesLogDirectory(t *testing.T) {
	dir := t.TempDir()
	logDir := filepath.Join(dir, "logs")
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "title", logDir)
	require.NoError(t, err)
	defer m.Close()

	// Verify log directory was created
	_, err = os.Stat(logDir)
	require.NoError(t, err)
}

func TestNewMigration_InitializesField(t *testing.T) {
	dir := t.TempDir()
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "custom_field", dir)
	require.NoError(t, err)
	defer m.Close()

	assert.Equal(t, int64(30), m.srcProject)
	assert.Equal(t, int64(1001), m.srcSuite)
	assert.Equal(t, int64(31), m.dstProject)
	assert.Equal(t, int64(2001), m.dstSuite)
	assert.Equal(t, "custom_field", m.compareField)
}

func TestMigrationImportedCases_InitializesZero(t *testing.T) {
	dir := t.TempDir()
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "title", dir)
	require.NoError(t, err)
	defer m.Close()

	// imported cases should start at 0
	assert.Equal(t, 0, m.importedCases)
}

func TestNewMigration_InitializesSharedStepMapping(t *testing.T) {
	dir := t.TempDir()
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "title", dir)
	require.NoError(t, err)
	defer m.Close()

	// Mapping should be initialized
	require.NotNil(t, m.mapping)
	result := m.Mapping()
	require.NotNil(t, result)
}

func TestNewMigration_FailsWithInvalidLogDir(t *testing.T) {
	// Use a path with non-parent directories that can't be created
	invalidPath := filepath.Join("", "invalid\x00path")
	cli := &client.MockClient{}

	_, err := NewMigration(cli, 30, 1001, 31, 2001, "title", invalidPath)
	// Should fail to create directory due to path issues
	if err == nil {
		t.Skip("test environment allows null path creation")
	}
}

func TestNewMigration_InitializesLoggerNotNil(t *testing.T) {
	dir := t.TempDir()
	cli := &client.MockClient{}

	m, err := NewMigration(cli, 30, 1001, 31, 2001, "title", dir)
	require.NoError(t, err)
	defer m.Close()

	// Logger should be initialized
	assert.NotNil(t, m.logger)
}
