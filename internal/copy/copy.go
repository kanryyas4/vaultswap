// Package copy provides functionality to copy a single secret from one
// provider to another, optionally overwriting an existing destination key.
package copy

import (
	"context"
	"fmt"

	"github.com/vaultswap/vaultswap/internal/audit"
	"github.com/vaultswap/vaultswap/internal/provider"
)

// Options configures a Copy operation.
type Options struct {
	SourceAlias string
	SourceKey   string
	DestAlias   string
	DestKey     string
	Overwrite   bool
}

// Copier copies a secret between two providers.
type Copier struct {
	providers map[string]provider.Provider
	auditor   *audit.Auditor
}

// New creates a Copier backed by the given provider map and auditor.
func New(providers map[string]provider.Provider, a *audit.Auditor) *Copier {
	return &Copier{providers: providers, auditor: a}
}

// Copy reads a secret from the source provider and writes it to the destination.
func (c *Copier) Copy(ctx context.Context, opts Options) error {
	src, ok := c.providers[opts.SourceAlias]
	if !ok {
		return fmt.Errorf("copy: unknown source alias %q", opts.SourceAlias)
	}

	dst, ok := c.providers[opts.DestAlias]
	if !ok {
		return fmt.Errorf("copy: unknown destination alias %q", opts.DestAlias)
	}

	value, err := src.GetSecret(ctx, opts.SourceKey)
	if err != nil {
		c.auditor.Log(opts.SourceAlias, opts.SourceKey, "copy_read", err)
		return fmt.Errorf("copy: read source secret %q: %w", opts.SourceKey, err)
	}

	destKey := opts.DestKey
	if destKey == "" {
		destKey = opts.SourceKey
	}

	if !opts.Overwrite {
		if _, err := dst.GetSecret(ctx, destKey); err == nil {
			return fmt.Errorf("copy: destination key %q already exists (use --overwrite to replace)", destKey)
		}
	}

	if err := dst.PutSecret(ctx, destKey, value); err != nil {
		c.auditor.Log(opts.DestAlias, destKey, "copy_write", err)
		return fmt.Errorf("copy: write destination secret %q: %w", destKey, err)
	}

	c.auditor.Log(opts.DestAlias, destKey, "copy_write", nil)
	return nil
}
