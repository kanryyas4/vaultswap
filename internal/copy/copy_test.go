package copy_test

import (
	"context"
	"testing"

	"github.com/vaultswap/vaultswap/internal/audit"
	"github.com/vaultswap/vaultswap/internal/copy"
	"github.com/vaultswap/vaultswap/internal/provider"
	"github.com/vaultswap/vaultswap/internal/provider/mock"
)

func makeProviders(t *testing.T) map[string]provider.Provider {
	t.Helper()
	a := mock.New("alpha")
	b := mock.New("beta")
	return map[string]provider.Provider{"alpha": a, "beta": b}
}

func buildCopier(t *testing.T, providers map[string]provider.Provider) *copy.Copier {
	t.Helper()
	a := audit.New()
	return copy.New(providers, a)
}

func TestCopy_Success(t *testing.T) {
	ctx := context.Background()
	providers := makeProviders(t)
	_ = providers["alpha"].(interface {
		PutSecret(context.Context, string, string) error
	}).PutSecret(ctx, "db/password", "s3cr3t")

	c := buildCopier(t, providers)
	err := c.Copy(ctx, copy.Options{
		SourceAlias: "alpha",
		SourceKey:   "db/password",
		DestAlias:   "beta",
		DestKey:     "db/password",
		Overwrite:   false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := providers["beta"].GetSecret(ctx, "db/password")
	if err != nil {
		t.Fatalf("secret not found in dest: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected %q, got %q", "s3cr3t", val)
	}
}

func TestCopy_UnknownSourceAlias(t *testing.T) {
	c := buildCopier(t, makeProviders(t))
	err := c.Copy(context.Background(), copy.Options{
		SourceAlias: "missing",
		DestAlias:   "beta",
		SourceKey:   "k",
	})
	if err == nil {
		t.Fatal("expected error for unknown source alias")
	}
}

func TestCopy_UnknownDestAlias(t *testing.T) {
	c := buildCopier(t, makeProviders(t))
	err := c.Copy(context.Background(), copy.Options{
		SourceAlias: "alpha",
		DestAlias:   "missing",
		SourceKey:   "k",
	})
	if err == nil {
		t.Fatal("expected error for unknown dest alias")
	}
}

func TestCopy_NoOverwrite(t *testing.T) {
	ctx := context.Background()
	providers := makeProviders(t)
	_ = providers["alpha"].(interface {
		PutSecret(context.Context, string, string) error
	}).PutSecret(ctx, "key", "v1")
	_ = providers["beta"].(interface {
		PutSecret(context.Context, string, string) error
	}).PutSecret(ctx, "key", "existing")

	c := buildCopier(t, providers)
	err := c.Copy(ctx, copy.Options{
		SourceAlias: "alpha",
		SourceKey:   "key",
		DestAlias:   "beta",
		DestKey:     "key",
		Overwrite:   false,
	})
	if err == nil {
		t.Fatal("expected error when dest key exists and overwrite is false")
	}
}
