package purge_test

import (
	"context"
	"testing"

	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/provider/mock"
	"github.com/yourusername/vaultswap/internal/purge"
)

func makeProviders(t *testing.T) map[string]provider.Provider {
	t.Helper()
	p := mock.New()
	return map[string]provider.Provider{"dev": p}
}

func seed(t *testing.T, prov provider.Provider, kvs map[string]string) {
	t.Helper()
	for k, v := range kvs {
		if err := prov.Put(context.Background(), k, v); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func TestPurge_ByPrefix(t *testing.T) {
	providers := makeProviders(t)
	seed(t, providers["dev"], map[string]string{
		"app/db": "pass",
		"app/api": "key",
		"other/x": "val",
	})

	pur := purge.New(providers)
	res, err := pur.Run(context.Background(), "dev", purge.Options{Prefix: "app/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Deleted) != 2 {
		t.Fatalf("expected 2 deleted, got %d", len(res.Deleted))
	}
	if len(res.Skipped) != 0 {
		t.Fatalf("expected 0 skipped, got %d", len(res.Skipped))
	}
}

func TestPurge_ExplicitKeys(t *testing.T) {
	providers := makeProviders(t)
	seed(t, providers["dev"], map[string]string{"a": "1", "b": "2", "c": "3"})

	pur := purge.New(providers)
	res, err := pur.Run(context.Background(), "dev", purge.Options{Keys: []string{"a", "c"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Deleted) != 2 {
		t.Fatalf("expected 2 deleted, got %d", len(res.Deleted))
	}
}

func TestPurge_DryRun(t *testing.T) {
	providers := makeProviders(t)
	seed(t, providers["dev"], map[string]string{"x": "1", "y": "2"})

	pur := purge.New(providers)
	res, err := pur.Run(context.Background(), "dev", purge.Options{DryRun: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Deleted) != 0 {
		t.Fatalf("expected 0 deleted in dry-run, got %d", len(res.Deleted))
	}
	if len(res.Skipped) != 2 {
		t.Fatalf("expected 2 skipped, got %d", len(res.Skipped))
	}
}

func TestPurge_UnknownAlias(t *testing.T) {
	pur := purge.New(makeProviders(t))
	_, err := pur.Run(context.Background(), "nope", purge.Options{})
	if err == nil {
		t.Fatal("expected error for unknown alias")
	}
}
