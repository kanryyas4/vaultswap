// Package importpkg provides functionality for importing secrets from
// external formats (JSON, dotenv) into a target secret manager provider.
package importpkg

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/vaultswap/internal/provider"
)

// Importer reads secrets from a file and writes them to a provider.
type Importer struct {
	providers map[string]provider.Provider
}

// New creates a new Importer with the given provider map.
func New(providers map[string]provider.Provider) *Importer {
	return &Importer{providers: providers}
}

// Import reads secrets from filePath (json or dotenv) and writes them
// to the provider identified by destAlias. Returns an error if the alias
// is unknown, the format is unsupported, or any write fails.
func (im *Importer) Import(destAlias, filePath, format string) error {
	dest, ok := im.providers[destAlias]
	if !ok {
		return fmt.Errorf("unknown destination alias: %q", destAlias)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	secrets, err := parse(format, data)
	if err != nil {
		return err
	}

	for key, value := range secrets {
		if err := dest.Put(key, value); err != nil {
			return fmt.Errorf("writing secret %q to %q: %w", key, destAlias, err)
		}
	}
	return nil
}

func parse(format string, data []byte) (map[string]string, error) {
	switch strings.ToLower(format) {
	case "json":
		var secrets map[string]string
		if err := json.Unmarshal(data, &secrets); err != nil {
			return nil, fmt.Errorf("parsing JSON: %w", err)
		}
		return secrets, nil
	case "dotenv":
		return parseDotenv(string(data)), nil
	default:
		return nil, fmt.Errorf("unsupported format: %q (use json or dotenv)", format)
	}
}

func parseDotenv(content string) map[string]string {
	secrets := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		if key != "" {
			secrets[key] = value
		}
	}
	return secrets
}
