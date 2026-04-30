package rollback_test

import (
	"context"
	"testing"

	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/provider/mock"
	"github.com/yourusername/vaultswap/internal/rollback"
)

func makeProviders(t *testing.T) map[string]provider.Provider {
	t.Helper()
	p, _ := mock.New(map[string]string{})
	return map[string]provider.Provider{"dev": p}
}

func TestCapture_Success(t *testing.T) {
	ctx := context.Background()
	p, _ := mock.New(map[string]string{"DB_PASS": "secret1", "API_KEY": "key123"})
	providers := map[string]provider.Provider{"dev": p}

	rb := rollback.New(providers)
	snap, err := rb.Capture(ctx, "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Alias != "dev" {
		t.Errorf("expected alias dev, got %s", snap.Alias)
	}
	if snap.Secrets["DB_PASS"] != "secret1" {
		t.Errorf("expected DB_PASS=secret1, got %s", snap.Secrets["DB_PASS"])
	}
}

func TestCapture_UnknownAlias(t *testing.T) {
	rb := rollback.New(makeProviders(t))
	_, err := rb.Capture(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown alias")
	}
}

func TestRestore_Success(t *testing.T) {
	ctx := context.Background()
	p, _ := mock.New(map[string]string{"DB_PASS": "old"})
	providers := map[string]provider.Provider{"dev": p}

	rb := rollback.New(providers)
	snap := &rollback.Snapshot{
		Alias:   "dev",
		Secrets: map[string]string{"DB_PASS": "restored", "NEW_KEY": "val"},
	}
	if err := rb.Restore(ctx, snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := p.GetSecret(ctx, "DB_PASS")
	if v != "restored" {
		t.Errorf("expected restored, got %s", v)
	}
}

func TestRestore_UnknownAlias(t *testing.T) {
	rb := rollback.New(makeProviders(t))
	snap := &rollback.Snapshot{Alias: "ghost", Secrets: map[string]string{}}
	if err := rb.Restore(context.Background(), snap); err == nil {
		t.Fatal("expected error for unknown alias")
	}
}

func TestCaptureAndRestore_RoundTrip(t *testing.T) {
	ctx := context.Background()
	p, _ := mock.New(map[string]string{"TOKEN": "original"})
	providers := map[string]provider.Provider{"prod": p}
	rb := rollback.New(providers)

	snap, _ := rb.Capture(ctx, "prod")
	_ = p.PutSecret(ctx, "TOKEN", "mutated")
	_ = rb.Restore(ctx, snap)

	v, _ := p.GetSecret(ctx, "TOKEN")
	if v != "original" {
		t.Errorf("expected original after restore, got %s", v)
	}
}
