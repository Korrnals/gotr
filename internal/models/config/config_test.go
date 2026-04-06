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
	if err := os.WriteFile(blocker, []byte("x"), 0600); err != nil {
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
	if err := os.MkdirAll(pathAsDir, 0755); err != nil {
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
