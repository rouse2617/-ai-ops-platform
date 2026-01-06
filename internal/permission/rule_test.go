package permission

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "permissions.yaml")

	configContent := `default: deny
rules:
  - pattern: "test_*"
    action: allow
  - pattern: "dangerous_*"
    action: deny
    message: "危险操作"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Default != ActionDeny {
		t.Errorf("Default = %v; want %v", cfg.Default, ActionDeny)
	}

	if len(cfg.Rules) != 2 {
		t.Errorf("len(Rules) = %d; want 2", len(cfg.Rules))
	}

	if cfg.Rules[0].Pattern != "test_*" {
		t.Errorf("Rules[0].Pattern = %q; want %q", cfg.Rules[0].Pattern, "test_*")
	}

	if cfg.Rules[1].Message != "危险操作" {
		t.Errorf("Rules[1].Message = %q; want %q", cfg.Rules[1].Message, "危险操作")
	}
}

func TestLoadConfig_DefaultValue(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "permissions.yaml")

	configContent := `rules:
  - pattern: "test_*"
    action: allow
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Default != ActionAsk {
		t.Errorf("Default = %v; want %v (default should be ask)", cfg.Default, ActionAsk)
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("LoadConfig() with nonexistent file should return error")
	}
}
