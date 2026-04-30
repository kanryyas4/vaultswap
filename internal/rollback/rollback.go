// Package rollback provides functionality to restore secrets to a previous
// snapshot captured before a rotate or sync operation.
package rollback

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Snapshot holds a point-in-time copy of secrets for a single provider alias.
type Snapshot struct {
	Alias   string
	Secrets map[string]string
}

// Rollbacker restores secrets from a snapshot into a target provider.
type Rollbacker struct {
	providers map[string]provider.Provider
}

// New creates a Rollbacker backed by the supplied provider map.
func New(providers map[string]provider.Provider) *Rollbacker {
	return &Rollbacker{providers: providers}
}

// Capture records the current state of all secrets in the named alias.
func (r *Rollbacker) Capture(ctx context.Context, alias string) (*Snapshot, error) {
	p, ok := r.providers[alias]
	if !ok {
		return nil, fmt.Errorf("rollback: unknown alias %q", alias)
	}

	keys, err := p.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("rollback: list secrets for %q: %w", alias, err)
	}

	snap := &Snapshot{
		Alias:   alias,
		Secrets: make(map[string]string, len(keys)),
	}
	for _, k := range keys {
		v, err := p.GetSecret(ctx, k)
		if err != nil {
			return nil, fmt.Errorf("rollback: get secret %q from %q: %w", k, alias, err)
		}
		snap.Secrets[k] = v
	}
	return snap, nil
}

// Restore writes all secrets from snap back into the target provider,
// overwriting any current values.
func (r *Rollbacker) Restore(ctx context.Context, snap *Snapshot) error {
	p, ok := r.providers[snap.Alias]
	if !ok {
		return fmt.Errorf("rollback: unknown alias %q", snap.Alias)
	}

	for k, v := range snap.Secrets {
		if err := p.PutSecret(ctx, k, v); err != nil {
			return fmt.Errorf("rollback: restore secret %q into %q: %w", k, snap.Alias, err)
		}
	}
	return nil
}
