// Package purge provides functionality to delete secrets matching a given
// prefix or explicit list from a target provider.
package purge

import (
	"context"
	"fmt"
	"strings"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Options controls the behaviour of a purge run.
type Options struct {
	// Prefix filters secrets to those whose key starts with Prefix.
	// Ignored when Keys is non-empty.
	Prefix string
	// Keys is an explicit list of secret keys to delete.
	Keys []string
	// DryRun reports what would be deleted without performing any writes.
	DryRun bool
}

// Result summarises the outcome of a purge operation.
type Result struct {
	Deleted []string
	Skipped []string
}

// Purger deletes secrets from a provider.
type Purger struct {
	providers map[string]provider.Provider
}

// New creates a Purger backed by the supplied provider map.
func New(providers map[string]provider.Provider) *Purger {
	return &Purger{providers: providers}
}

// Run executes the purge against the provider identified by alias.
func (p *Purger) Run(ctx context.Context, alias string, opts Options) (*Result, error) {
	prov, ok := p.providers[alias]
	if !ok {
		return nil, fmt.Errorf("purge: unknown alias %q", alias)
	}

	keys, err := p.resolveKeys(ctx, prov, opts)
	if err != nil {
		return nil, err
	}

	res := &Result{}
	for _, key := range keys {
		if opts.DryRun {
			res.Skipped = append(res.Skipped, key)
			continue
		}
		if err := prov.Delete(ctx, key); err != nil {
			return res, fmt.Errorf("purge: delete %q: %w", key, err)
		}
		res.Deleted = append(res.Deleted, key)
	}
	return res, nil
}

func (p *Purger) resolveKeys(ctx context.Context, prov provider.Provider, opts Options) ([]string, error) {
	if len(opts.Keys) > 0 {
		return opts.Keys, nil
	}
	all, err := prov.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("purge: list secrets: %w", err)
	}
	if opts.Prefix == "" {
		return all, nil
	}
	var filtered []string
	for _, k := range all {
		if strings.HasPrefix(k, opts.Prefix) {
			filtered = append(filtered, k)
		}
	}
	return filtered, nil
}
