// Package clone provides functionality for cloning all secrets from one
// provider to another, optionally overwriting existing secrets at the destination.
package clone

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Cloner copies all secrets from a source provider to a destination provider.
type Cloner struct {
	providers map[string]provider.Provider
}

// New returns a new Cloner backed by the given provider map.
func New(providers map[string]provider.Provider) *Cloner {
	return &Cloner{providers: providers}
}

// Result holds the outcome of a clone operation.
type Result struct {
	Key       string
	Overwrote bool
}

// Clone copies all secrets from srcAlias to dstAlias.
// If overwrite is false, existing keys at the destination are skipped.
func (c *Cloner) Clone(ctx context.Context, srcAlias, dstAlias string, overwrite bool) ([]Result, error) {
	src, ok := c.providers[srcAlias]
	if !ok {
		return nil, fmt.Errorf("clone: unknown source alias %q", srcAlias)
	}
	dst, ok := c.providers[dstAlias]
	if !ok {
		return nil, fmt.Errorf("clone: unknown destination alias %q", dstAlias)
	}

	keys, err := src.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("clone: list source secrets: %w", err)
	}

	dstKeys, err := dst.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("clone: list destination secrets: %w", err)
	}
	existing := make(map[string]struct{}, len(dstKeys))
	for _, k := range dstKeys {
		existing[k] = struct{}{}
	}

	var results []Result
	for _, key := range keys {
		_, exists := existing[key]
		if exists && !overwrite {
			continue
		}
		val, err := src.GetSecret(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("clone: get %q from source: %w", key, err)
		}
		if err := dst.PutSecret(ctx, key, val); err != nil {
			return nil, fmt.Errorf("clone: put %q to destination: %w", key, err)
		}
		results = append(results, Result{Key: key, Overwrote: exists})
	}
	return results, nil
}
