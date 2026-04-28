package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/vaultswap/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "vaultswap.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return path
}

func TestLoad_Valid(t *testing.T) {
	yaml := `
version: "1"
providers:
  - type: aws
    alias: prod-aws
    options:
      region: us-east-1
  - type: vault
    alias: prod-vault
    options:
      address: https://vault.example.com
`
	path := writeTemp(t, yaml)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(cfg.Providers))
	}
}

func TestLoad_MissingVersion(t *testing.T) {
	yaml := `
providers:
  - type: aws
    alias: prod-aws
`
	path := writeTemp(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing version, got nil")
	}
}

func TestLoad_DuplicateAlias(t *testing.T) {
	yaml := `
version: "1"
providers:
  - type: aws
    alias: same
  - type: gcp
    alias: same
`
	path := writeTemp(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for duplicate alias, got nil")
	}
}

func TestLoad_UnknownProvider(t *testing.T) {
	yaml := `
version: "1"
providers:
  - type: azure
    alias: prod-azure
`
	path := writeTemp(t, yaml)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for unknown provider type, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/vaultswap.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
