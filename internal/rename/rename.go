// Package rename provides functionality for renaming (re-keying) a secret
// from one key name to another within the same provider or across providers.
package rename

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Renamer holds the providers map and audit logger interface.
type Renamer struct {
	providers map[string]provider.Provider
}

// New creates a new Renamer with the given providers map.
func New(providers map[string]provider.Provider) *Renamer {
	return &Renamer{providers: providers}
}

// Options configures a rename operation.
type Options struct {
	SourceAlias string
	DestAlias   string
	OldKey      string
	NewKey      string
	DeleteOld   bool
}

// Run performs the rename operation: reads the secret at OldKey from
// SourceAlias, writes it to NewKey in DestAlias, and optionally deletes
// the original.
func (r *Renamer) Run(ctx context.Context, opts Options) error {
	src, ok := r.providers[opts.SourceAlias]
	if !ok {
		return fmt.Errorf("rename: unknown source alias %q", opts.SourceAlias)
	}

	dst, ok := r.providers[opts.DestAlias]
	if !ok {
		return fmt.Errorf("rename: unknown dest alias %q", opts.DestAlias)
	}

	value, err := src.GetSecret(ctx, opts.OldKey)
	if err != nil {
		return fmt.Errorf("rename: get %q from %q: %w", opts.OldKey, opts.SourceAlias, err)
	}

	if err := dst.PutSecret(ctx, opts.NewKey, value); err != nil {
		return fmt.Errorf("rename: put %q into %q: %w", opts.NewKey, opts.DestAlias, err)
	}

	if opts.DeleteOld {
		if err := src.DeleteSecret(ctx, opts.OldKey); err != nil {
			return fmt.Errorf("rename: delete old key %q from %q: %w", opts.OldKey, opts.SourceAlias, err)
		}
	}

	return nil
}
