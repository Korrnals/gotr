package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewAndPathString(t *testing.T) {
	cfg := New("/tmp/test.yaml")
	if cfg.PathString() != "/tmp/test.yaml" {
		t.Fatalf("PathString = %q", cfg.PathString())
	}
	if cfg.Data == nil {
		t.Fatal("expected Data to be initialized")
	}
}

func TestDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := Default()
	if err != nil {
		t.Fatalf("Default error: %v", err)
	}
	want := filepath.Join(home, ".gotr", "config", "default.yaml")
	if cfg.Path != want {
		t.Fatalf("Default path = %q, want %q", cfg.Path, want)
	}
}

func TestDefault_HomeDirError(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")
	origXdg := os.Getenv("XDG_CONFIG_HOME")

	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	_, err := Default()

	if origHome == "" && origUserProfile == "" && origHomeDrive == "" && origHomePath == "" && origXdg == "" {
		if err == nil {
			t.Skip("os.UserHomeDir resolved home on this platform/user setup")
		}
		return
	}

	if err == nil {
		t.Skip("os.UserHomeDir resolved home from system user database")
	}
}

func TestWithDefaultsAndIsValid(t *testing.T) {
	cfg := New("ignored").WithDefaults()

	if cfg.Data.BaseURL != DefaultBaseURL || cfg.Data.Username != DefaultUsername || cfg.Data.APIKey != DefaultAPIKey {
		t.Fatal("default values not applied")
	}
	if cfg.IsValid() {
		t.Fatal("config with placeholders must be invalid")
	}

	cfg.Data.BaseURL = "https://example.testrail.io"
	cfg.Data.Username = "user@example.com"
	cfg.Data.APIKey = "api-key"
	if !cfg.IsValid() {
		t.Fatal("config with real values must be valid")
	}
}

func TestConfig_IsValid_TableDriven(t *testing.T) {
	valid := &ConfigData{
		BaseURL:  "https://example.testrail.io",
		Username: "user@example.com",
		APIKey:   "api-key",
	}

	tests := []struct {
		name string
		data *ConfigData
		want bool
	}{
		{name: "nil data", data: nil, want: false},
		{name: "empty values", data: &ConfigData{}, want: false},
		{name: "default placeholders", data: &ConfigData{BaseURL: DefaultBaseURL, Username: DefaultUsername, APIKey: DefaultAPIKey}, want: false},
		{name: "empty base url", data: &ConfigData{BaseURL: "", Username: valid.Username, APIKey: valid.APIKey}, want: false},
		{name: "default base url", data: &ConfigData{BaseURL: DefaultBaseURL, Username: valid.Username, APIKey: valid.APIKey}, want: false},
		{name: "empty username", data: &ConfigData{BaseURL: valid.BaseURL, Username: "", APIKey: valid.APIKey}, want: false},
		{name: "default username", data: &ConfigData{BaseURL: valid.BaseURL, Username: DefaultUsername, APIKey: valid.APIKey}, want: false},
		{name: "empty api key", data: &ConfigData{BaseURL: valid.BaseURL, Username: valid.Username, APIKey: ""}, want: false},
		{name: "default api key", data: &ConfigData{BaseURL: valid.BaseURL, Username: valid.Username, APIKey: DefaultAPIKey}, want: false},
		{name: "all real values", data: valid, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Data: tt.data}
			if got := cfg.IsValid(); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateAndRenderTemplate(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nested", "default.yaml")

	cfg := New(path).WithDefaults()
	if err := cfg.Create(); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if got := st.Mode().Perm(); got != 0o600 {
		t.Fatalf("file mode = %o, want %o", got, 0o600)
	}

	content := string(b)

	checks := []string{
		"base_url:",
		"username:",
		"api_key:",
		"compare:",
		"auto_retry_failed_pages: true",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("template missing %q", c)
		}
	}
}

func TestCreate_WithCustomInputValues(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "custom", "config.yaml")

	cfg := New(path)
	cfg.Data = &ConfigData{
		BaseURL:  "https://acme.testrail.io",
		Username: "qa@acme.io",
		APIKey:   "secret",
		Insecure: true,
		JqFormat: true,
		Debug:    true,
	}

	if err := cfg.Create(); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if got := st.Mode().Perm(); got != 0o600 {
		t.Fatalf("file mode = %o, want %o", got, 0o600)
	}

	content := string(b)

	checks := []string{
		`base_url: "https://acme.testrail.io"`,
		`username: "qa@acme.io"`,
		`api_key: "secret"`,
		"insecure: true",
		"jq_format: true",
		"debug: true",
	}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("template missing %q", c)
		}
	}
}

func TestCreate_WithNilDataFallsBackToDefaults(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nil-data", "default.yaml")

	cfg := New(path)
	cfg.Data = nil

	if err := cfg.Create(); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	content := string(b)

	if !strings.Contains(content, DefaultBaseURL) ||
		!strings.Contains(content, DefaultUsername) ||
		!strings.Contains(content, DefaultAPIKey) {
		t.Fatal("expected Create with nil data to render defaults")
	}
}

func TestCreate_MkdirAllError(t *testing.T) {
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}

	cfg := New(filepath.Join(blocker, "subdir", "default.yaml")).WithDefaults()
	err := cfg.Create()
	if err == nil {
		t.Fatal("expected Create to fail when parent path contains a file")
	}
	if !strings.Contains(err.Error(), "failed to create directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_WriteFileError(t *testing.T) {
	tmp := t.TempDir()
	pathAsDir := filepath.Join(tmp, "config-as-dir")
	if err := os.MkdirAll(pathAsDir, 0o755); err != nil {
		t.Fatalf("mkdir pathAsDir: %v", err)
	}

	cfg := New(pathAsDir).WithDefaults()
	err := cfg.Create()
	if err == nil {
		t.Fatal("expected Create to fail when target path is a directory")
	}
	if !strings.Contains(err.Error(), "failed to write file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderTemplateNilData(t *testing.T) {
	cfg := New("ignored")
	cfg.Data = nil
	out := cfg.renderTemplate()
	if !strings.Contains(out, DefaultBaseURL) || !strings.Contains(out, DefaultUsername) || !strings.Contains(out, DefaultAPIKey) {
		t.Fatal("renderTemplate should fallback to defaults on nil data")
	}
}

func TestIsDefaultValue(t *testing.T) {
	if !IsDefaultValue("", "x") {
		t.Fatal("empty value should be treated as default")
	}
	if !IsDefaultValue("same", "same") {
		t.Fatal("same value should be treated as default")
	}
	if IsDefaultValue("real", "default") {
		t.Fatal("different non-empty value should not be default")
	}
}

func TestWithDefaults_SnapshotRetention(t *testing.T) {
	cfg := &Config{Data: &ConfigData{}}
	cfg.WithDefaults()

	// Check snapshot config is initialized
	if cfg.Data.Snap == nil {
		t.Fatal("Snap should not be nil after WithDefaults")
	}

	// Check retention is initialized
	if cfg.Data.Snap.Retention == nil {
		t.Fatal("Snap.Retention should not be nil after WithDefaults")
	}

	// Check defaults
	if !cfg.Data.Snap.Enabled {
		t.Fatal("Snap.Enabled should be true")
	}
	if !cfg.Data.Snap.Retention.Enabled {
		t.Fatal("Snap.Retention.Enabled should be true")
	}
	if cfg.Data.Snap.Retention.DefaultTTLDays != 30 {
		t.Fatalf("DefaultTTLDays = %d, want 30", cfg.Data.Snap.Retention.DefaultTTLDays)
	}
	if cfg.Data.Snap.MaxSnapshots != 100 {
		t.Fatalf("MaxSnapshots = %d, want 100", cfg.Data.Snap.MaxSnapshots)
	}

	// Check protected prefixes
	if len(cfg.Data.Snap.Retention.ProtectedPrefixes) != 2 {
		t.Fatalf("ProtectedPrefixes length = %d, want 2", len(cfg.Data.Snap.Retention.ProtectedPrefixes))
	}
	if cfg.Data.Snap.Retention.ProtectedPrefixes[0] != "pinned_" {
		t.Fatalf("ProtectedPrefixes[0] = %q, want pinned_", cfg.Data.Snap.Retention.ProtectedPrefixes[0])
	}
	if cfg.Data.Snap.Retention.ProtectedPrefixes[1] != "archived_" {
		t.Fatalf("ProtectedPrefixes[1] = %q, want archived_", cfg.Data.Snap.Retention.ProtectedPrefixes[1])
	}

	// Check frozen snapshots is empty but initialized
	if cfg.Data.Snap.Retention.FrozenSnapshots == nil {
		t.Fatal("FrozenSnapshots should not be nil")
	}
	if len(cfg.Data.Snap.Retention.FrozenSnapshots) != 0 {
		t.Fatalf("FrozenSnapshots should be empty, got %v", cfg.Data.Snap.Retention.FrozenSnapshots)
	}
}

func TestRenderTemplate_IncludesRetentionSection(t *testing.T) {
	cfg := &Config{Data: &ConfigData{}}
	cfg.WithDefaults()

	out := cfg.renderTemplate()

	// Check retention section exists
	if !strings.Contains(out, "retention:") {
		t.Fatal("renderTemplate should include retention: section")
	}
	if !strings.Contains(out, "default_ttl_days: 30") {
		t.Fatal("renderTemplate should include default_ttl_days: 30")
	}
	if !strings.Contains(out, "pinned_") {
		t.Fatal("renderTemplate should include pinned_ prefix")
	}
	if !strings.Contains(out, "archived_") {
		t.Fatal("renderTemplate should include archived_ prefix")
	}
	if !strings.Contains(out, "frozen_snapshots:") {
		t.Fatal("renderTemplate should include frozen_snapshots")
	}
}

func TestSnapshotRetention_ValidateTTL(t *testing.T) {
	tests := []struct {
		name    string
		ttl     int
		wantErr bool
	}{
		{"positive ttl", 30, false},
		{"zero ttl", 0, true},
		{"negative ttl", -1, true},
		{"large ttl", 365, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := &SnapshotRetention{DefaultTTLDays: tt.ttl}
			err := sr.ValidateTTL()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateTTL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSnapshotRetention_ValidateProtectedPrefixes(t *testing.T) {
	tests := []struct {
		name      string
		prefixes  []string
		wantErr   bool
		errMsg    string
	}{
		{"empty prefixes", []string{}, false, ""},
		{"single prefix", []string{"pinned_"}, false, ""},
		{"multiple unique", []string{"pinned_", "archived_"}, false, ""},
		{"duplicate prefix", []string{"pinned_", "pinned_"}, true, "duplicate"},
		{"three items with dup", []string{"a_", "b_", "a_"}, true, "duplicate"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := &SnapshotRetention{ProtectedPrefixes: tt.prefixes}
			err := sr.ValidateProtectedPrefixes()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateProtectedPrefixes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Fatalf("ValidateProtectedPrefixes() error %q should contain %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestSnapshotRetention_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sr      *SnapshotRetention
		wantErr bool
	}{
		{"nil retention", nil, false},
		{"disabled retention", &SnapshotRetention{Enabled: false, DefaultTTLDays: 0}, false},
		{"valid retention", &SnapshotRetention{Enabled: true, DefaultTTLDays: 30, ProtectedPrefixes: []string{"pinned_"}}, false},
		{"invalid ttl", &SnapshotRetention{Enabled: true, DefaultTTLDays: 0}, true},
		{"invalid prefixes", &SnapshotRetention{Enabled: true, DefaultTTLDays: 30, ProtectedPrefixes: []string{"a_", "a_"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sr.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSnapshotConfig_ValidateMaxSnapshots(t *testing.T) {
	tests := []struct {
		name      string
		max       int
		wantErr   bool
	}{
		{"positive max", 100, false},
		{"zero max", 0, true},
		{"negative max", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &SnapshotConfig{MaxSnapshots: tt.max}
			err := sc.ValidateMaxSnapshots()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateMaxSnapshots() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSnapshotConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		sc      *SnapshotConfig
		wantErr bool
	}{
		{"nil snapshot config", nil, false},
		{"disabled snapshot", &SnapshotConfig{Enabled: false}, false},
		{"valid snapshot config", &SnapshotConfig{
			Enabled:      true,
			MaxSnapshots: 100,
			Retention: &SnapshotRetention{
				Enabled:           true,
				DefaultTTLDays:    30,
				ProtectedPrefixes: []string{"pinned_"},
			},
		}, false},
		{"invalid max snapshots", &SnapshotConfig{Enabled: true, MaxSnapshots: 0}, true},
		{"invalid retention", &SnapshotConfig{
			Enabled:      true,
			MaxSnapshots: 100,
			Retention: &SnapshotRetention{
				Enabled:        true,
				DefaultTTLDays: 0,
			},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sc.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
