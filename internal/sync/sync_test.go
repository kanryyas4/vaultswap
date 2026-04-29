package sync_test

import (
	"context"
	"errors"
	"testing"

	mockprovider "github.com/yourusername/vaultswap/internal/provider/mock"
	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/sync"
)

func makeProviders() map[string]provider.Provider {
	return map[string]provider.Provider{
		"src": mockprovider.New(),
		"dst": mockprovider.New(),
	}
}

func TestSync_Success(t *testing.T) {
	providers := makeProviders()
	ctx := context.Background()

	_ = providers["src"].PutSecret(ctx, "db/password", "s3cr3t")

	s := sync.New(providers)
	results := s.Sync(ctx, []sync.SecretPair{
		{SourceAlias: "src", SourceKey: "db/password", DestAlias: "dst", DestKey: "db/password"},
	})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Success {
		t.Fatalf("expected success, got err: %v", results[0].Err)
	}

	val, err := providers["dst"].GetSecret(ctx, "db/password")
	if err != nil {
		t.Fatalf("unexpected error reading dst: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected %q, got %q", "s3cr3t", val)
	}
}

func TestSync_UnknownSourceAlias(t *testing.T) {
	s := sync.New(makeProviders())
	results := s.Sync(context.Background(), []sync.SecretPair{
		{SourceAlias: "missing", SourceKey: "k", DestAlias: "dst", DestKey: "k"},
	})
	if results[0].Success {
		t.Fatal("expected failure for unknown source alias")
	}
	if results[0].Err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestSync_UnknownDestAlias(t *testing.T) {
	s := sync.New(makeProviders())
	results := s.Sync(context.Background(), []sync.SecretPair{
		{SourceAlias: "src", SourceKey: "k", DestAlias: "nowhere", DestKey: "k"},
	})
	if results[0].Success {
		t.Fatal("expected failure for unknown dest alias")
	}
}

func TestSync_MissingSourceSecret(t *testing.T) {
	s := sync.New(makeProviders())
	results := s.Sync(context.Background(), []sync.SecretPair{
		{SourceAlias: "src", SourceKey: "nonexistent", DestAlias: "dst", DestKey: "nonexistent"},
	})
	if results[0].Success {
		t.Fatal("expected failure when source secret missing")
	}
	if !errors.Is(results[0].Err, results[0].Err) {
		t.Fatal("expected wrapped error")
	}
}
