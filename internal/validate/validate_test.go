package validate_test

import (
	"testing"

	mockprovider "github.com/yourusername/vaultswap/internal/provider/mock"
	"github.com/yourusername/vaultswap/internal/validate"
	"github.com/yourusername/vaultswap/internal/provider"
)

func makeProviders(data map[string]string) map[string]provider.Provider {
	p := mockprovider.New()
	for k, v := range data {
		_ = p.PutSecret(k, v)
	}
	return map[string]provider.Provider{"mock": p}
}

func TestValidate_RequiredPresent(t *testing.T) {
	providers := makeProviders(map[string]string{"DB_PASS": "secret"})
	v := validate.New(providers, []validate.Rule{{Key: "DB_PASS", Required: true}})
	results, err := v.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || !results[0].Passed {
		t.Errorf("expected pass, got %+v", results)
	}
}

func TestValidate_RequiredMissing(t *testing.T) {
	providers := makeProviders(map[string]string{})
	v := validate.New(providers, []validate.Rule{{Key: "DB_PASS", Required: true}})
	results, err := v.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Passed {
		t.Errorf("expected failure for missing required key")
	}
}

func TestValidate_NonEmpty(t *testing.T) {
	providers := makeProviders(map[string]string{"TOKEN": ""})
	v := validate.New(providers, []validate.Rule{{Key: "TOKEN", Required: true, NonEmpty: true}})
	results, err := v.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Passed {
		t.Errorf("expected failure for empty value")
	}
}

func TestValidate_PatternMatch(t *testing.T) {
	providers := makeProviders(map[string]string{"PORT": "8080"})
	v := validate.New(providers, []validate.Rule{{Key: "PORT", Pattern: `^\d+$`}})
	results, err := v.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !results[0].Passed {
		t.Errorf("expected pass for numeric port")
	}
}

func TestValidate_PatternMismatch(t *testing.T) {
	providers := makeProviders(map[string]string{"PORT": "abc"})
	v := validate.New(providers, []validate.Rule{{Key: "PORT", Pattern: `^\d+$`}})
	results, err := v.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Passed {
		t.Errorf("expected failure for non-numeric port")
	}
}

func TestValidate_InvalidPattern(t *testing.T) {
	providers := makeProviders(map[string]string{"KEY": "val"})
	v := validate.New(providers, []validate.Rule{{Key: "KEY", Pattern: `[invalid`}})
	_, err := v.Run()
	if err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}
