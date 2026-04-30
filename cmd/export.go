package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultswap/internal/config"
	"github.com/yourusername/vaultswap/internal/export"
	"github.com/yourusername/vaultswap/internal/provider"
)

var (
	exportFormat string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export <alias>",
	Short: "Export all secrets from a provider to a local file",
	Long: `Export fetches every secret from the specified provider alias
and writes them to a file in the chosen format (json or dotenv).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := args[0]

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		providers := make(map[string]provider.Provider, len(cfg.Providers))
		for _, pc := range cfg.Providers {
			p, err := provider.New(pc)
			if err != nil {
				return fmt.Errorf("init provider %q: %w", pc.Alias, err)
			}
			providers[pc.Alias] = p
		}

		if exportOutput == "" {
			exportOutput = alias + "." + exportFormat
		}

		ex := export.New(providers)
		if err := ex.Export(alias, exportOutput, export.Format(exportFormat)); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}

		fmt.Printf("Exported secrets from %q to %s\n", alias, exportOutput)
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Output format: json or dotenv")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (default: <alias>.<format>)")
	rootCmd.AddCommand(exportCmd)
}
