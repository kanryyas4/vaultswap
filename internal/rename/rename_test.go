package rename_test

import (
	"context"
	"testing"

	"github.com/yourusername/vaultswap/internal/provider"
	mockprovider "github.com/yourusername/vaultswap/internal/provider/mock"
	"github.com/yourusername/vaultswap/internal/rename"
)

func makeProviders(t *testing.T) map[string]provider.Provider {
	t.Helper()
	p1 := mockprovider.New()
	p2 := mockprovider.New()
	return map[string]provider.Provider{
		"src": p1,
		"dst": p2,
	}
}

func TestRename_Success(t *testing.T) {
	ctx := context.Background()
	providers := makeProviders(t)
	_ = providers["src"].PutSecret(ctx, "old-key", "secret-value")

	r := rename.New(providers)
	err := r.Run(ctx, rename.Options{
		SourceAlias: "src",
		DestAlias:   "dst",
		OldKey:      "old-key",
		NewKey:      "new-key",
		DeleteOld:   false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := providers["dst"].GetSecret(ctx, "new-key")
	if err != nil || val != "secret-value" {
		t.Fatalf("expected new-key=secret-value, got %q err=%v", val, err)
	}

	// old key should still exist
	_, err = providers["src"].GetSecret(ctx, "old-key")
	if err != nil {
		t.Fatalf("old key should still exist: %v", err)
	}
}

func TestRename_DeleteOld(t *testing.T) {
	ctx := context.Background()
	providers := makeProviders(t)
	_ = providers["src"].PutSecret(ctx, "old-key", "secret-value")

	r := rename.New(providers)
	err := r.Run(ctx, rename.Options{
		SourceAlias: "src",
		DestAlias:   "dst",
		OldKey:      "old-key",
		NewKey:      "new-key",
		DeleteOld:   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = providers["src"].GetSecret(ctx, "old-key")
	if err == nil {
		t.Fatal("expected old key to be deleted")
	}
}

func TestRename_UnknownSourceAlias(t *testing.T) {
	ctx := context.Background()
	r := rename.New(makeProviders(t))
	err := r.Run(ctx, rename.Options{SourceAlias: "nope", DestAlias: "dst", OldKey: "k", NewKey: "k2"})
	if err == nil {
		t.Fatal("expected error for unknown source alias")
	}
}

func TestRename_UnknownDestAlias(t *testing.T) {
	ctx := context.Background()
	r := rename.New(makeProviders(t))
	err := r.Run(ctx, rename.Options{SourceAlias: "src", DestAlias: "nope", OldKey: "k", NewKey: "k2"})
	if err == nil {
		t.Fatal("expected error for unknown dest alias")
	}
}

func TestRename_MissingSourceSecret(t *testing.T) {
	ctx := context.Background()
	r := rename.New(makeProviders(t))
	err := r.Run(ctx, rename.Options{SourceAlias: "src", DestAlias: "dst", OldKey: "missing", NewKey: "new"})
	if err == nil {
		t.Fatal("expected error for missing source secret")
	}
}
