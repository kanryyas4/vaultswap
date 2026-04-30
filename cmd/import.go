package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultswap/internal/config"
	importpkg "github.com/yourusername/vaultswap/internal/import"
	"github.com/yourusername/vaultswap/internal/provider"
)

var (
	importDest   string
	importFile   string
	importFormat string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import secrets from a file into a provider",
	Long: `Import reads secrets from a JSON or dotenv file and writes them
to the specified destination provider alias.

Examples:
  vaultswap import --dest vault --file secrets.json --format json
  vaultswap import --dest aws --file .env --format dotenv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		providers := make(map[string]provider.Provider)
		for _, s := range cfg.Stores {
			p, err := provider.New(s.Provider, s.Options)
			if err != nil {
				return fmt.Errorf("initialising provider %q: %w", s.Alias, err)
			}
			providers[s.Alias] = p
		}

		im := importpkg.New(providers)
		if err := im.Import(importDest, importFile, importFormat); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return err
		}

		fmt.Printf("Secrets imported successfully into %q\n", importDest)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
	importCmd.Flags().StringVar(&importDest, "dest", "", "Destination provider alias (required)")
	importCmd.Flags().StringVar(&importFile, "file", "", "Path to the secrets file (required)")
	importCmd.Flags().StringVar(&importFormat, "format", "json", "File format: json or dotenv")
	_ = importCmd.MarkFlagRequired("dest")
	_ = importCmd.MarkFlagRequired("file")
}
