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
