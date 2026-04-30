// Package export provides functionality to export secrets from a provider
// to a local file in various formats (json, dotenv).
package export

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Format represents the output format for exported secrets.
type Format string

const (
	FormatJSON   Format = "json"
	FormatDotenv Format = "dotenv"
)

// Exporter writes secrets from a provider to a file.
type Exporter struct {
	providers map[string]provider.Provider
}

// New creates a new Exporter with the given provider map.
func New(providers map[string]provider.Provider) *Exporter {
	return &Exporter{providers: providers}
}

// Export retrieves all secrets from the named provider alias and writes
// them to destPath in the requested format.
func (e *Exporter) Export(alias, destPath string, format Format) error {
	p, ok := e.providers[alias]
	if !ok {
		return fmt.Errorf("export: unknown provider alias %q", alias)
	}

	keys, err := p.ListSecrets()
	if err != nil {
		return fmt.Errorf("export: list secrets: %w", err)
	}

	secrets := make(map[string]string, len(keys))
	for _, k := range keys {
		v, err := p.GetSecret(k)
		if err != nil {
			return fmt.Errorf("export: get secret %q: %w", k, err)
		}
		secrets[k] = v
	}

	var data []byte
	switch format {
	case FormatJSON:
		data, err = json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return fmt.Errorf("export: marshal json: %w", err)
		}
	case FormatDotenv:
		var sb strings.Builder
		for k, v := range secrets {
			fmt.Fprintf(&sb, "%s=%s\n", k, v)
		}
		data = []byte(sb.String())
	default:
		return fmt.Errorf("export: unsupported format %q", format)
	}

	if err := os.WriteFile(destPath, data, 0o600); err != nil {
		return fmt.Errorf("export: write file: %w", err)
	}
	return nil
}
