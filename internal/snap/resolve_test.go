package snap

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// RegisterFlags / ResolveName / ResolveLabel
// ---------------------------------------------------------------------------

func TestRegisterFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	assert.NotNil(t, cmd.Flags().Lookup(FlagSnapshot))
	assert.NotNil(t, cmd.Flags().Lookup(FlagSnapName))
	assert.NotNil(t, cmd.Flags().Lookup(FlagSnapLabel))
}

func TestResolveName(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	assert.Equal(t, "", ResolveName(cmd))

	_ = cmd.Flags().Set(FlagSnapName, "my-snap")
	assert.Equal(t, "my-snap", ResolveName(cmd))
}

func TestResolveLabel(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	assert.Equal(t, "", ResolveLabel(cmd))

	_ = cmd.Flags().Set(FlagSnapLabel, "deploy-v2")
	assert.Equal(t, "deploy-v2", ResolveLabel(cmd))
}

func TestResolveDecision_ExplicitFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	_ = cmd.Flags().Set(FlagSnapshot, "false")
	assert.False(t, ResolveDecision(cmd))

	_ = cmd.Flags().Set(FlagSnapshot, "true")
	assert.True(t, ResolveDecision(cmd))
}

func TestResolveDecision_Config(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)

	viper.Set("snap.enabled", false)
	assert.False(t, ResolveDecision(cmd))

	viper.Set("snap.enabled", true)
	assert.True(t, ResolveDecision(cmd))
}

func TestResolveDecision_Default(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{Use: "test"}
	RegisterFlags(cmd)
	assert.True(t, ResolveDecision(cmd))
}

// ---------------------------------------------------------------------------
// ReadConfig
// ---------------------------------------------------------------------------

func TestReadConfig_Defaults(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cfg := ReadConfig()
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 30, cfg.RetentionDays)
	assert.Equal(t, 100, cfg.MaxSnapshots)
}

func TestReadConfig_CustomValues(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	viper.Set("snap.enabled", false)
	viper.Set("snap.retention.default_ttl_days", 7)
	viper.Set("snap.max_snapshots", 50)
	viper.Set("snap.attachments.save_binary", "always")
	viper.Set("snap.attachments.max_file_mb", 20)
	viper.Set("snap.attachments.compress", true)
	viper.Set("snap.attachments.prompt_above_threshold", true)

	cfg := ReadConfig()
	assert.False(t, cfg.Enabled)
	assert.Equal(t, 7, cfg.RetentionDays)
	assert.Equal(t, 50, cfg.MaxSnapshots)
	assert.Equal(t, "always", cfg.AttachSaveBinary)
	assert.Equal(t, 20, cfg.AttachMaxFileMB)
	assert.True(t, cfg.AttachCompress)
	assert.True(t, cfg.AttachPromptAboveThresh)
}

// ---------------------------------------------------------------------------
// InfoBanner
// ---------------------------------------------------------------------------

func TestInfoBanner_Enabled(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = old }()

	InfoBanner(true)
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	assert.Contains(t, buf.String(), "enabled")
}

func TestInfoBanner_Disabled(t *testing.T) {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = old }()

	InfoBanner(false)
	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	assert.Contains(t, buf.String(), "disabled")
}

// ---------------------------------------------------------------------------
// Manifest.Latest edge cases
// ---------------------------------------------------------------------------

func TestManifest_Latest_Empty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store, err := NewStore()
	assert.NoError(t, err)
	manifest, err := LoadManifest(store)
	assert.NoError(t, err)

	// No entries — should return nil.
	assert.Nil(t, manifest.Latest())
}

func TestManifest_Latest_WithEntries(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	store, err := NewStore()
	assert.NoError(t, err)
	manifest, err := LoadManifest(store)
	assert.NoError(t, err)

	meta := BuildMeta(OpUpdate, "case", []int64{42}, Tier1, 1, 1, "test", nil, "https://example.com")
	_, err = TakeSnapshot(t.Context(), store, manifest, meta, nil)
	assert.NoError(t, err)

	entry := manifest.Latest()
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.ID)
}
