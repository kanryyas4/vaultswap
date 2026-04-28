package provider_test

import (
	"context"
	"testing"

	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/provider/mock"
)

func TestRegisterAndNew(t *testing.T) {
	provider.Register("mock", func(opts map[string]string) (provider.Provider, error) {
		return mock.New(nil), nil
	})

	p, err := provider.New("mock", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Name() != "mock" {
		t.Errorf("expected name 'mock', got %q", p.Name())
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := provider.New("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

func TestMock_PutAndGet(t *testing.T) {
	m := mock.New(nil)
	ctx := context.Background()

	sec := &provider.Secret{Key: "db/password", Value: "s3cr3t", Version: "1"}
	if err := m.PutSecret(ctx, sec); err != nil {
		t.Fatalf("PutSecret: %v", err)
	}

	got, err := m.GetSecret(ctx, "db/password")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if got.Value != sec.Value {
		t.Errorf("expected value %q, got %q", sec.Value, got.Value)
	}
}

func TestMock_Delete(t *testing.T) {
	m := mock.New(map[string]*provider.Secret{
		"key1": {Key: "key1", Value: "val1"},
	})
	ctx := context.Background()

	if err := m.DeleteSecret(ctx, "key1"); err != nil {
		t.Fatalf("DeleteSecret: %v", err)
	}
	if _, err := m.GetSecret(ctx, "key1"); err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

func TestMock_ListSecrets(t *testing.T) {
	m := mock.New(map[string]*provider.Secret{
		"app/db": {Key: "app/db", Value: "v1"},
		"app/api": {Key: "app/api", Value: "v2"},
		"other": {Key: "other", Value: "v3"},
	})
	ctx := context.Background()

	keys, err := m.ListSecrets(ctx, "app/")
	if err != nil {
		t.Fatalf("ListSecrets: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys with prefix 'app/', got %d", len(keys))
	}
}
