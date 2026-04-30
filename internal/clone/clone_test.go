package clone_test

import (
	"context"
	"testing"

	"github.com/yourusername/vaultswap/internal/clone"
	mockprovider "github.com/yourusername/vaultswap/internal/provider/mock"
)

func makeProviders() map[string]interface{ ListSecrets(context.Context) ([]string, error); GetSecret(context.Context, string) (string, error); PutSecret(context.Context, string, string) error; DeleteSecret(context.Context, string) error } {
	return nil // unused helper stub
}

func buildCloner(srcSecrets, dstSecrets map[string]string) (*clone.Cloner, *mockprovider.Mock, *mockprovider.Mock) {
	src := mockprovider.New()
	dst := mockprovider.New()
	ctx := context.Background()
	for k, v := range srcSecrets {
		_ = src.PutSecret(ctx, k, v)
	}
	for k, v := range dstSecrets {
		_ = dst.PutSecret(ctx, k, v)
	}
	providers := map[string]interface {
		ListSecrets(context.Context) ([]string, error)
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
	}{
		"src": src,
		"dst": dst,
	}
	_ = providers
	import_providers := map[string]interface{ ListSecrets(context.Context) ([]string, error); GetSecret(context.Context, string) (string, error); PutSecret(context.Context, string, string) error; DeleteSecret(context.Context, string) error }{"src": src, "dst": dst}
	_ = import_providers
	return clone.New(map[string]interface{ ListSecrets(context.Context) ([]string, error); GetSecret(context.Context, string) (string, error); PutSecret(context.Context, string, string) error; DeleteSecret(context.Context, string) error }{"src": src, "dst": dst}), src, dst
}

func TestClone_Success(t *testing.T) {
	ctx := context.Background()
	src := mockprovider.New()
	dst := mockprovider.New()
	_ = src.PutSecret(ctx, "KEY1", "val1")
	_ = src.PutSecret(ctx, "KEY2", "val2")

	c := clone.New(map[string]interface {
		ListSecrets(context.Context) ([]string, error)
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
	}{"src": src, "dst": dst})

	results, err := c.Clone(ctx, "src", "dst", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Overwrote {
			t.Errorf("key %q should not be marked as overwritten", r.Key)
		}
	}
}

func TestClone_SkipsExistingWithoutOverwrite(t *testing.T) {
	ctx := context.Background()
	src := mockprovider.New()
	dst := mockprovider.New()
	_ = src.PutSecret(ctx, "KEY1", "new")
	_ = dst.PutSecret(ctx, "KEY1", "old")

	c := clone.New(map[string]interface {
		ListSecrets(context.Context) ([]string, error)
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
	}{"src": src, "dst": dst})

	results, err := c.Clone(ctx, "src", "dst", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results (skipped), got %d", len(results))
	}
	val, _ := dst.GetSecret(ctx, "KEY1")
	if val != "old" {
		t.Errorf("expected old value to be preserved, got %q", val)
	}
}

func TestClone_OverwriteExisting(t *testing.T) {
	ctx := context.Background()
	src := mockprovider.New()
	dst := mockprovider.New()
	_ = src.PutSecret(ctx, "KEY1", "new")
	_ = dst.PutSecret(ctx, "KEY1", "old")

	c := clone.New(map[string]interface {
		ListSecrets(context.Context) ([]string, error)
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
	}{"src": src, "dst": dst})

	results, err := c.Clone(ctx, "src", "dst", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Overwrote {
		t.Fatalf("expected 1 overwritten result")
	}
	val, _ := dst.GetSecret(ctx, "KEY1")
	if val != "new" {
		t.Errorf("expected new value, got %q", val)
	}
}

func TestClone_UnknownSourceAlias(t *testing.T) {
	ctx := context.Background()
	c := clone.New(map[string]interface {
		ListSecrets(context.Context) ([]string, error)
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
	}{})
	_, err := c.Clone(ctx, "missing", "dst", false)
	if err == nil {
		t.Fatal("expected error for unknown source alias")
	}
}

func TestClone_UnknownDestAlias(t *testing.T) {
	ctx := context.Background()
	src := mockprovider.New()
	c := clone.New(map[string]interface {
		ListSecrets(context.Context) ([]string, error)
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
	}{"src": src})
	_, err := c.Clone(ctx, "src", "missing", false)
	if err == nil {
		t.Fatal("expected error for unknown destination alias")
	}
}
