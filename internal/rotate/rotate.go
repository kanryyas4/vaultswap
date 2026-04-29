package rotate

import (
	"context"
	"fmt"

	"github.com/vaultswap/internal/provider"
)

// Options holds configuration for a rotation operation.
type Options struct {
	Alias     string
	SecretKey string
	NewValue  string
	BackupKey string // optional: key to store old value before overwriting
}

// Result captures the outcome of a single rotation.
type Result struct {
	Alias     string
	SecretKey string
	OldValue  string
	Success   bool
	Err       error
}

// Rotator performs secret rotation against a named provider.
type Rotator struct {
	providers map[string]provider.Provider
}

// New creates a Rotator backed by the supplied provider map.
func New(providers map[string]provider.Provider) *Rotator {
	return &Rotator{providers: providers}
}

// Rotate fetches the current value, optionally backs it up, then writes the
// new value to the target provider.
func (r *Rotator) Rotate(ctx context.Context, opts Options) Result {
	res := Result{Alias: opts.Alias, SecretKey: opts.SecretKey}

	p, ok := r.providers[opts.Alias]
	if !ok {
		res.Err = fmt.Errorf("rotate: unknown alias %q", opts.Alias)
		return res
	}

	old, err := p.GetSecret(ctx, opts.SecretKey)
	if err != nil {
		res.Err = fmt.Errorf("rotate: get current secret: %w", err)
		return res
	}
	res.OldValue = old

	if opts.BackupKey != "" {
		if err := p.PutSecret(ctx, opts.BackupKey, old); err != nil {
			res.Err = fmt.Errorf("rotate: backup secret: %w", err)
			return res
		}
	}

	if err := p.PutSecret(ctx, opts.SecretKey, opts.NewValue); err != nil {
		res.Err = fmt.Errorf("rotate: put new secret: %w", err)
		return res
	}

	res.Success = true
	return res
}
