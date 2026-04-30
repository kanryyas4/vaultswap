package promote_test

import (
	"context"
	"testing"

	"github.com/yourusername/vaultswap/internal/provider"
	mockprovider "github.com/yourusername/vaultswap/internal/provider/mock"
	"github.com/yourusername/vaultswap/internal/promote"
)

func makeProviders(t *testing.T) map[string]provider.Provider {
	t.Helper()
	return map[string]provider.Provider{
		"staging": mockprovider.New(),
		"prod":    mockprovider.New(),
	}
}

func TestPromote_Success(t *testing.T) {
	ctx := context.Background()
	providers := makeProviders(t)
	_ = providers["staging"].PutSecret(ctx, "staging/db_pass", "s3cr3t")
	_ = providers["staging"].PutSecret(ctx, "staging/api_key", "abc123")

	p := promote.New(providers)
	res, err := p.Promote(ctx, promote.Options{
		SourceAlias:  "staging",
		DestAlias:    "prod",
		SourcePrefix: "staging/",
		DestPrefix:   "prod/",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Promoted) != 2 {
		t.Fatalf("expected 2 promoted, got %d", len(res.Promoted))
	}
	val, err := providers["prod"].GetSecret(ctx, "prod/db_pass")
	if err != nil || val != "s3cr3t" {
		t.Fatalf("expected prod/db_pass=s3cr3t, got %q %v", val, err)
	}
}

func TestPromote_SkipsExistingWithoutOverwrite(t *testing.T) {
	ctx := context.Background()
	providers := makeProviders(t)
	_ = providers["staging"].PutSecret(ctx, "key", "new")
	_ = providers["prod"].PutSecret(ctx, "key", "old")

	p := promote.New(providers)
	res, err := p.Promote(ctx, promote.Options{
		SourceAlias: "staging",
		DestAlias:   "prod",
		Overwrite:   false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Skipped) != 1 {
		t.Fatalf("expected 1 skipped, got %d", len(res.Skipped))
	}
	val, _ := providers["prod"].GetSecret(ctx, "key")
	if val != "old" {
		t.Fatalf("expected old value preserved, got %q", val)
	}
}

func TestPromote_DryRun(t *testing.T) {
	ctx := context.Background()
	providers := makeProviders(t)
	_ = providers["staging"].PutSecret(ctx, "secret", "value")

	p := promote.New(providers)
	res, err := p.Promote(ctx, promote.Options{
		SourceAlias: "staging",
		DestAlias:   "prod",
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Promoted) != 1 || !res.DryRun {
		t.Fatalf("expected 1 dry-run promoted entry")
	}
	if _, err := providers["prod"].GetSecret(ctx, "secret"); err == nil {
		t.Fatal("dry-run should not write to dest")
	}
}

func TestPromote_UnknownSourceAlias(t *testing.T) {
	p := promote.New(makeProviders(t))
	_, err := p.Promote(context.Background(), promote.Options{
		SourceAlias: "nope",
		DestAlias:   "prod",
	})
	if err == nil {
		t.Fatal("expected error for unknown source alias")
	}
}

func TestPromote_UnknownDestAlias(t *testing.T) {
	p := promote.New(makeProviders(t))
	_, err := p.Promote(context.Background(), promote.Options{
		SourceAlias: "staging",
		DestAlias:   "nope",
	})
	if err == nil {
		t.Fatal("expected error for unknown dest alias")
	}
}
