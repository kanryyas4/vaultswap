package export_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourusername/vaultswap/internal/export"
	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/provider/mock"
)

func makeProviders(t *testing.T) map[string]provider.Provider {
	t.Helper()
	p := mock.New()
	_ = p.PutSecret("DB_PASS", "hunter2")
	_ = p.PutSecret("API_KEY", "abc123")
	return map[string]provider.Provider{"src": p}
}

func TestExport_JSON(t *testing.T) {
	providers := makeProviders(t)
	ex := export.New(providers)
	dest := filepath.Join(t.TempDir(), "out.json")

	if err := ex.Export("src", dest, export.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(dest)
	var got map[string]string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got["DB_PASS"] != "hunter2" || got["API_KEY"] != "abc123" {
		t.Errorf("unexpected secrets: %v", got)
	}
}

func TestExport_Dotenv(t *testing.T) {
	providers := makeProviders(t)
	ex := export.New(providers)
	dest := filepath.Join(t.TempDir(), "out.env")

	if err := ex.Export("src", dest, export.FormatDotenv); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(dest)
	content := string(data)
	if !strings.Contains(content, "DB_PASS=hunter2") || !strings.Contains(content, "API_KEY=abc123") {
		t.Errorf("unexpected dotenv content: %s", content)
	}
}

func TestExport_UnknownAlias(t *testing.T) {
	ex := export.New(map[string]provider.Provider{})
	err := ex.Export("ghost", "/tmp/out.json", export.FormatJSON)
	if err == nil || !strings.Contains(err.Error(), "unknown provider alias") {
		t.Errorf("expected unknown alias error, got: %v", err)
	}
}

func TestExport_UnsupportedFormat(t *testing.T) {
	providers := makeProviders(t)
	ex := export.New(providers)
	dest := filepath.Join(t.TempDir(), "out.xml")
	err := ex.Export("src", dest, export.Format("xml"))
	if err == nil || !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected unsupported format error, got: %v", err)
	}
}
