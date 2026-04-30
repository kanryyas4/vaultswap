package importpkg_test

import (
	"encoding/json"
	"os"
	"testing"

	importpkg "github.com/yourusername/vaultswap/internal/import"
	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/provider/mock"
)

func makeProviders(t *testing.T) map[string]provider.Provider {
	t.Helper()
	p := mock.New()
	return map[string]provider.Provider{"vault": p}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "import-*")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestImport_JSON(t *testing.T) {
	providers := makeProviders(t)
	im := importpkg.New(providers)

	secrets := map[string]string{"DB_PASS": "s3cr3t", "API_KEY": "abc123"}
	raw, _ := json.Marshal(secrets)
	path := writeTempFile(t, string(raw))

	if err := im.Import("vault", path, "json"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for k, want := range secrets {
		got, err := providers["vault"].Get(k)
		if err != nil || got != want {
			t.Errorf("key %q: got %q, want %q (err: %v)", k, got, want, err)
		}
	}
}

func TestImport_Dotenv(t *testing.T) {
	providers := makeProviders(t)
	im := importpkg.New(providers)

	content := "# comment\nDB_HOST=localhost\nDB_PORT=\"5432\"\n"
	path := writeTempFile(t, content)

	if err := im.Import("vault", path, "dotenv"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v, _ := providers["vault"].Get("DB_HOST"); v != "localhost" {
		t.Errorf("DB_HOST: got %q, want %q", v, "localhost")
	}
	if v, _ := providers["vault"].Get("DB_PORT"); v != "5432" {
		t.Errorf("DB_PORT: got %q, want %q", v, "5432")
	}
}

func TestImport_UnknownAlias(t *testing.T) {
	im := importpkg.New(makeProviders(t))
	err := im.Import("nonexistent", "/dev/null", "json")
	if err == nil {
		t.Fatal("expected error for unknown alias")
	}
}

func TestImport_UnsupportedFormat(t *testing.T) {
	providers := makeProviders(t)
	im := importpkg.New(providers)
	path := writeTempFile(t, "")
	err := im.Import("vault", path, "yaml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestImport_InvalidJSON(t *testing.T) {
	providers := makeProviders(t)
	im := importpkg.New(providers)
	path := writeTempFile(t, "{not valid json")
	err := im.Import("vault", path, "json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
