package config

import (
	"os"
	"path/filepath"
	"testing"

	"jira-worklog-publisher/internal/models"
)

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")
	defer SetConfigPathForTest(func() string { return path })()

	cfg := Load()
	if cfg.BaseURL != "" || cfg.Username != "" || cfg.APIToken != "" {
		t.Errorf("Load() missing file should return empty config, got %+v", cfg)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatal(err)
	}
	defer SetConfigPathForTest(func() string { return path })()

	cfg := Load()
	if cfg.BaseURL != "" || cfg.Username != "" {
		t.Errorf("Load() invalid JSON should return empty config, got %+v", cfg)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	want := models.JiraConfig{
		BaseURL:  "https://test.atlassian.net",
		Username: "u@test.com",
		APIToken: "token",
	}
	data := `{"baseURL":"https://test.atlassian.net","username":"u@test.com","apiToken":"token"}`
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	defer SetConfigPathForTest(func() string { return path })()

	cfg := Load()
	if cfg.BaseURL != want.BaseURL || cfg.Username != want.Username || cfg.APIToken != want.APIToken {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
}

func TestSave_ThenLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	defer SetConfigPathForTest(func() string { return path })()

	cfg := models.JiraConfig{
		BaseURL:  "https://save-test.atlassian.net",
		Username: "save@test.com",
		APIToken: "saved-token",
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := Load()
	if loaded.BaseURL != cfg.BaseURL || loaded.Username != cfg.Username || loaded.APIToken != cfg.APIToken {
		t.Errorf("Load() after Save = %+v, want %+v", loaded, cfg)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file stat: %v", err)
	}
	if info.Mode().Perm()&0077 != 0 {
		t.Errorf("config file should be owner-only (0600), got %o", info.Mode().Perm())
	}
}
