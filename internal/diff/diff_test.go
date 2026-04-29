package diff_test

import (
	"context"
	"testing"

	"github.com/vaultswap/internal/diff"
	mockprovider "github.com/vaultswap/internal/provider/mock"
)

func makeProvider(secrets map[string]string) *mockprovider.Mock {
	m := mockprovider.New()
	ctx := context.Background()
	for k, v := range secrets {
		_ = m.PutSecret(ctx, k, v)
	}
	return m
}

func TestDiff_OnlyInSource(t *testing.T) {
	src := makeProvider(map[string]string{"foo": "bar", "baz": "qux"})
	dst := makeProvider(map[string]string{"foo": "bar"})

	res, err := diff.New(src, dst).Compare(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.OnlyInSource) != 1 || res.OnlyInSource[0] != "baz" {
		t.Errorf("expected OnlyInSource=[baz], got %v", res.OnlyInSource)
	}
}

func TestDiff_OnlyInDest(t *testing.T) {
	src := makeProvider(map[string]string{"foo": "bar"})
	dst := makeProvider(map[string]string{"foo": "bar", "extra": "val"})

	res, err := diff.New(src, dst).Compare(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.OnlyInDest) != 1 || res.OnlyInDest[0] != "extra" {
		t.Errorf("expected OnlyInDest=[extra], got %v", res.OnlyInDest)
	}
}

func TestDiff_Diverged(t *testing.T) {
	src := makeProvider(map[string]string{"key": "v1"})
	dst := makeProvider(map[string]string{"key": "v2"})

	res, err := diff.New(src, dst).Compare(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Diverged) != 1 || res.Diverged[0] != "key" {
		t.Errorf("expected Diverged=[key], got %v", res.Diverged)
	}
}

func TestDiff_InSync(t *testing.T) {
	src := makeProvider(map[string]string{"key": "same"})
	dst := makeProvider(map[string]string{"key": "same"})

	res, err := diff.New(src, dst).Compare(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.InSync) != 1 || res.InSync[0] != "key" {
		t.Errorf("expected InSync=[key], got %v", res.InSync)
	}
	if len(res.Diverged) != 0 {
		t.Errorf("expected no diverged, got %v", res.Diverged)
	}
}

func TestDiff_Empty(t *testing.T) {
	src := makeProvider(map[string]string{})
	dst := makeProvider(map[string]string{})

	res, err := diff.New(src, dst).Compare(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.OnlyInSource)+len(res.OnlyInDest)+len(res.Diverged)+len(res.InSync) != 0 {
		t.Error("expected all empty result sets")
	}
}
