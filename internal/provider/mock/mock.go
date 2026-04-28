// Package mock provides an in-memory Provider implementation for testing.
package mock

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Mock is a simple in-memory secret store.
type Mock struct {
	store map[string]*provider.Secret
}

// New creates a new Mock provider, optionally pre-populated with secrets.
func New(initial map[string]*provider.Secret) *Mock {
	s := make(map[string]*provider.Secret, len(initial))
	for k, v := range initial {
		s[k] = v
	}
	return &Mock{store: s}
}

func (m *Mock) Name() string { return "mock" }

func (m *Mock) GetSecret(_ context.Context, key string) (*provider.Secret, error) {
	s, ok := m.store[key]
	if !ok {
		return nil, fmt.Errorf("secret %q not found", key)
	}
	return s, nil
}

func (m *Mock) PutSecret(_ context.Context, secret *provider.Secret) error {
	if secret.Key == "" {
		return fmt.Errorf("secret key must not be empty")
	}
	m.store[secret.Key] = secret
	return nil
}

func (m *Mock) DeleteSecret(_ context.Context, key string) error {
	if _, ok := m.store[key]; !ok {
		return fmt.Errorf("secret %q not found", key)
	}
	delete(m.store, key)
	return nil
}

func (m *Mock) ListSecrets(_ context.Context, prefix string) ([]string, error) {
	keys := make([]string, 0, len(m.store))
	for k := range m.store {
		if prefix == "" || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}
	return keys, nil
}
