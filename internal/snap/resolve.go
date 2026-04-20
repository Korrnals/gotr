package snap

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// FlagSnapshot is the flag name for explicit snapshot override.
const FlagSnapshot = "snapshot"

// FlagSnapName is the flag name for custom snapshot name.
const FlagSnapName = "snap-name"

// FlagSnapLabel is the flag name for a user-defined snapshot label (searchable tag).
const FlagSnapLabel = "snap-label"

// RegisterFlags adds --snapshot, --snap-name, and --snap-label to the given command.
func RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagSnapshot, false, "Force snapshot on/off (overrides config)")
	cmd.Flags().String(FlagSnapName, "", "Custom snapshot name (default: auto-generated)")
	cmd.Flags().String(FlagSnapLabel, "", "Custom label for the snapshot (searchable tag)")
}

// ResolveDecision determines whether to take a snapshot.
// Priority: explicit --snapshot flag > config snap.enabled > default true.
func ResolveDecision(cmd *cobra.Command) bool {
	// 1. Explicit flag has highest priority.
	if cmd.Flags().Changed(FlagSnapshot) {
		val, _ := cmd.Flags().GetBool(FlagSnapshot)
		return val
	}

	// 2. Config value (viper).
	if viper.IsSet("snap.enabled") {
		return viper.GetBool("snap.enabled")
	}

	// 3. Default: enabled.
	return true
}

// ResolveName returns the custom name from --snap-name or empty string.
func ResolveName(cmd *cobra.Command) string {
	name, _ := cmd.Flags().GetString(FlagSnapName)
	return name
}

// ResolveLabel returns the custom label from --snap-label or empty string.
func ResolveLabel(cmd *cobra.Command) string {
	label, _ := cmd.Flags().GetString(FlagSnapLabel)
	return label
}

// ReadConfig loads snap configuration from Viper with defaults.
func ReadConfig() SnapConfig {
	cfg := DefaultConfig()

	if viper.IsSet("snap.enabled") {
		cfg.Enabled = viper.GetBool("snap.enabled")
	}
	if viper.IsSet("snap.retention_days") {
		cfg.RetentionDays = viper.GetInt("snap.retention_days")
	}
	if viper.IsSet("snap.max_snapshots") {
		cfg.MaxSnapshots = viper.GetInt("snap.max_snapshots")
	}
	if viper.IsSet("snap.attachments.save_binary") {
		cfg.AttachSaveBinary = viper.GetString("snap.attachments.save_binary")
	}
	if viper.IsSet("snap.attachments.max_file_mb") {
		cfg.AttachMaxFileMB = viper.GetInt("snap.attachments.max_file_mb")
	}
	if viper.IsSet("snap.attachments.compress") {
		cfg.AttachCompress = viper.GetBool("snap.attachments.compress")
	}
	if viper.IsSet("snap.attachments.prompt_above_threshold") {
		cfg.AttachPromptAboveThresh = viper.GetBool("snap.attachments.prompt_above_threshold")
	}

	return cfg
}
