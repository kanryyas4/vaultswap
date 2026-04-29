package rotate_test

import (
	"context"
	"testing"

	mockprovider "github.com/vaultswap/internal/provider/mock"
	"github.com/vaultswap/internal/rotate"
)

func makeProviders(t *testing.T) map[string]interface{ GetSecret(context.Context, string) (string, error); PutSecret(context.Context, string, string) error; DeleteSecret(context.Context, string) error; ListSecrets(context.Context) ([]string, error) } {
	t.Helper()
	return nil // unused helper kept for symmetry
}

func buildRotator(seeds map[string]string) (*rotate.Rotator, *mockprovider.Mock) {
	m := mockprovider.New()
	ctx := context.Background()
	for k, v := range seeds {
		_ = m.PutSecret(ctx, k, v)
	}
	providers := map[string]interface {
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
		ListSecrets(context.Context) ([]string, error)
	}{
		"vault": m,
	}
	_ = providers
	return rotate.New(map[string]interface {
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
		ListSecrets(context.Context) ([]string, error)
	}{
		"vault": m,
	}), m
}

func TestRotate_Success(t *testing.T) {
	m := mockprovider.New()
	ctx := context.Background()
	_ = m.PutSecret(ctx, "db/password", "old-secret")

	r := rotate.New(map[string]interface {
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
		ListSecrets(context.Context) ([]string, error)
	}{"vault": m})

	res := r.Rotate(ctx, rotate.Options{
		Alias:     "vault",
		SecretKey: "db/password",
		NewValue:  "new-secret",
	})

	if !res.Success {
		t.Fatalf("expected success, got err: %v", res.Err)
	}
	if res.OldValue != "old-secret" {
		t.Errorf("OldValue = %q, want %q", res.OldValue, "old-secret")
	}
	got, _ := m.GetSecret(ctx, "db/password")
	if got != "new-secret" {
		t.Errorf("stored value = %q, want %q", got, "new-secret")
	}
}

func TestRotate_WithBackup(t *testing.T) {
	m := mockprovider.New()
	ctx := context.Background()
	_ = m.PutSecret(ctx, "api/key", "original")

	r := rotate.New(map[string]interface {
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
		ListSecrets(context.Context) ([]string, error)
	}{"vault": m})

	res := r.Rotate(ctx, rotate.Options{
		Alias:     "vault",
		SecretKey: "api/key",
		NewValue:  "rotated",
		BackupKey: "api/key.bak",
	})

	if !res.Success {
		t.Fatalf("expected success, got err: %v", res.Err)
	}
	bak, _ := m.GetSecret(ctx, "api/key.bak")
	if bak != "original" {
		t.Errorf("backup = %q, want %q", bak, "original")
	}
}

func TestRotate_UnknownAlias(t *testing.T) {
	r := rotate.New(map[string]interface {
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
		ListSecrets(context.Context) ([]string, error)
	}{})

	res := r.Rotate(context.Background(), rotate.Options{
		Alias:     "missing",
		SecretKey: "k",
		NewValue:  "v",
	})
	if res.Success {
		t.Fatal("expected failure for unknown alias")
	}
	if res.Err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestRotate_MissingSecret(t *testing.T) {
	m := mockprovider.New()
	r := rotate.New(map[string]interface {
		GetSecret(context.Context, string) (string, error)
		PutSecret(context.Context, string, string) error
		DeleteSecret(context.Context, string) error
		ListSecrets(context.Context) ([]string, error)
	}{"vault": m})

	res := r.Rotate(context.Background(), rotate.Options{
		Alias:     "vault",
		SecretKey: "nonexistent",
		NewValue:  "v",
	})
	if res.Success {
		t.Fatal("expected failure when source secret does not exist")
	}
}
