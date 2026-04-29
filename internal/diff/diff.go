// Package diff provides functionality for comparing secrets between providers.
package diff

import (
	"context"
	"fmt"

	"github.com/vaultswap/internal/provider"
)

// Result holds the diff output between a source and destination provider.
type Result struct {
	OnlyInSource []string
	OnlyInDest   []string
	Diverged     []string
	InSync       []string
}

// Differ compares secrets between two providers.
type Differ struct {
	source provider.Provider
	dest   provider.Provider
}

// New creates a new Differ for the given source and destination providers.
func New(source, dest provider.Provider) *Differ {
	return &Differ{source: source, dest: dest}
}

// Compare lists all secrets from both providers and classifies them.
func (d *Differ) Compare(ctx context.Context) (*Result, error) {
	srcKeys, err := d.source.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("diff: list source secrets: %w", err)
	}

	dstKeys, err := d.dest.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("diff: list dest secrets: %w", err)
	}

	dstSet := make(map[string]struct{}, len(dstKeys))
	for _, k := range dstKeys {
		dstSet[k] = struct{}{}
	}

	srcSet := make(map[string]struct{}, len(srcKeys))
	for _, k := range srcKeys {
		srcSet[k] = struct{}{}
	}

	res := &Result{}

	for _, k := range srcKeys {
		if _, ok := dstSet[k]; !ok {
			res.OnlyInSource = append(res.OnlyInSource, k)
			continue
		}
		srcVal, err := d.source.GetSecret(ctx, k)
		if err != nil {
			return nil, fmt.Errorf("diff: get source secret %q: %w", k, err)
		}
		dstVal, err := d.dest.GetSecret(ctx, k)
		if err != nil {
			return nil, fmt.Errorf("diff: get dest secret %q: %w", k, err)
		}
		if srcVal != dstVal {
			res.Diverged = append(res.Diverged, k)
		} else {
			res.InSync = append(res.InSync, k)
		}
	}

	for _, k := range dstKeys {
		if _, ok := srcSet[k]; !ok {
			res.OnlyInDest = append(res.OnlyInDest, k)
		}
	}

	return res, nil
}
