// Package promote implements secret promotion across environments.
// It copies secrets from a source provider to a destination provider,
// applying an optional prefix transformation (e.g. staging/ -> prod/).
package promote

import (
	"context"
	"fmt"
	"strings"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Options controls the behaviour of a promotion run.
type Options struct {
	SourceAlias  string
	DestAlias    string
	SourcePrefix string // strip this prefix from source keys
	DestPrefix   string // prepend this prefix to dest keys
	DryRun       bool
	Overwrite    bool
}

// Result holds the outcome of a promotion run.
type Result struct {
	Promoted  []string
	Skipped   []string
	DryRun    bool
}

// Promoter moves secrets between providers with optional prefix rewriting.
type Promoter struct {
	providers map[string]provider.Provider
}

// New creates a Promoter backed by the supplied provider map.
func New(providers map[string]provider.Provider) *Promoter {
	return &Promoter{providers: providers}
}

// Promote executes the promotion described by opts.
func (p *Promoter) Promote(ctx context.Context, opts Options) (*Result, error) {
	src, ok := p.providers[opts.SourceAlias]
	if !ok {
		return nil, fmt.Errorf("promote: unknown source alias %q", opts.SourceAlias)
	}
	dst, ok := p.providers[opts.DestAlias]
	if !ok {
		return nil, fmt.Errorf("promote: unknown dest alias %q", opts.DestAlias)
	}

	keys, err := src.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("promote: list source secrets: %w", err)
	}

	result := &Result{DryRun: opts.DryRun}

	for _, key := range keys {
		if opts.SourcePrefix != "" && !strings.HasPrefix(key, opts.SourcePrefix) {
			continue
		}
		stripped := strings.TrimPrefix(key, opts.SourcePrefix)
		destKey := opts.DestPrefix + stripped

		if !opts.Overwrite {
			if _, err := dst.GetSecret(ctx, destKey); err == nil {
				result.Skipped = append(result.Skipped, destKey)
				continue
			}
		}

		value, err := src.GetSecret(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("promote: get %q: %w", key, err)
		}

		if !opts.DryRun {
			if err := dst.PutSecret(ctx, destKey, value); err != nil {
				return nil, fmt.Errorf("promote: put %q: %w", destKey, err)
			}
		}
		result.Promoted = append(result.Promoted, destKey)
	}

	return result, nil
}
