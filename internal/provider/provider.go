package provider

import "context"

// Secret represents a key-value secret with optional metadata.
type Secret struct {
	Key     string
	Value   string
	Version string
}

// Provider defines the interface for interacting with a secret manager backend.
type Provider interface {
	// Name returns the provider identifier (e.g. "aws", "vault", "gcp").
	Name() string

	// GetSecret retrieves a secret by key.
	GetSecret(ctx context.Context, key string) (*Secret, error)

	// PutSecret creates or updates a secret.
	PutSecret(ctx context.Context, secret *Secret) error

	// DeleteSecret removes a secret by key.
	DeleteSecret(ctx context.Context, key string) error

	// ListSecrets returns all secret keys available under an optional prefix.
	ListSecrets(ctx context.Context, prefix string) ([]string, error)
}

// Registry holds registered provider factories keyed by provider type name.
var registry = map[string]Factory{}

// Factory is a function that constructs a Provider from a map of options.
type Factory func(opts map[string]string) (Provider, error)

// Register adds a provider factory to the global registry.
func Register(name string, f Factory) {
	registry[name] = f
}

// New instantiates a provider by name using the global registry.
func New(name string, opts map[string]string) (Provider, error) {
	f, ok := registry[name]
	if !ok {
		return nil, &ErrUnknownProvider{Name: name}
	}
	return f(opts)
}

// ErrUnknownProvider is returned when no factory is registered for a given name.
type ErrUnknownProvider struct {
	Name string
}

func (e *ErrUnknownProvider) Error() string {
	return "unknown provider: " + e.Name
}
