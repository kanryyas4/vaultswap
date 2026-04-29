package sync

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultswap/internal/provider"
)

// SecretPair describes a source and destination for a single secret sync.
type SecretPair struct {
	SourceAlias string
	SourceKey   string
	DestAlias   string
	DestKey     string
}

// Result holds the outcome of syncing a single secret pair.
type Result struct {
	Pair    SecretPair
	Success bool
	Err     error
}

// Syncer copies secrets between providers.
type Syncer struct {
	providers map[string]provider.Provider
}

// New creates a Syncer with the given provider map (alias -> Provider).
func New(providers map[string]provider.Provider) *Syncer {
	return &Syncer{providers: providers}
}

// Sync copies each SecretPair from source to destination provider.
// It returns a slice of Results, one per pair.
func (s *Syncer) Sync(ctx context.Context, pairs []SecretPair) []Result {
	results := make([]Result, 0, len(pairs))

	for _, p := range pairs {
		r := Result{Pair: p}

		src, ok := s.providers[p.SourceAlias]
		if !ok {
			r.Err = fmt.Errorf("unknown source alias %q", p.SourceAlias)
			results = append(results, r)
			continue
		}

		dst, ok := s.providers[p.DestAlias]
		if !ok {
			r.Err = fmt.Errorf("unknown destination alias %q", p.DestAlias)
			results = append(results, r)
			continue
		}

		value, err := src.GetSecret(ctx, p.SourceKey)
		if err != nil {
			r.Err = fmt.Errorf("get %q from %q: %w", p.SourceKey, p.SourceAlias, err)
			results = append(results, r)
			continue
		}

		if err := dst.PutSecret(ctx, p.DestKey, value); err != nil {
			r.Err = fmt.Errorf("put %q to %q: %w", p.DestKey, p.DestAlias, err)
			results = append(results, r)
			continue
		}

		r.Success = true
		log.Printf("synced %s/%s -> %s/%s", p.SourceAlias, p.SourceKey, p.DestAlias, p.DestKey)
		results = append(results, r)
	}

	return results
}
