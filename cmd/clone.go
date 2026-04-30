package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultswap/internal/clone"
	"github.com/yourusername/vaultswap/internal/config"
	"github.com/yourusername/vaultswap/internal/provider"
)

func init() {
	var cfgFile string
	var overwrite bool

	cloneCmd := &cobra.Command{
		Use:   "clone <src-alias> <dst-alias>",
		Short: "Clone all secrets from one provider to another",
		Long: `Clone copies every secret from the source provider to the destination.
By default existing keys at the destination are skipped; use --overwrite to replace them.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcAlias := args[0]
			dstAlias := args[1]

			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			providers := make(map[string]interface {
				ListSecrets(ctx interface{ Done() <-chan struct{} }) ([]string, error)
				GetSecret(ctx interface{ Done() <-chan struct{} }, key string) (string, error)
				PutSecret(ctx interface{ Done() <-chan struct{} }, key, value string) error
				DeleteSecret(ctx interface{ Done() <-chan struct{} }, key string) error
			})
			_ = providers

			providerMap := make(map[string]provider.Provider)
			for _, p := range cfg.Providers {
				prov, err := provider.New(p.Type, p.Options)
				if err != nil {
					return fmt.Errorf("init provider %q: %w", p.Alias, err)
				}
				providerMap[p.Alias] = prov
			}

			c := clone.New(providerMap)
			results, err := c.Clone(cmd.Context(), srcAlias, dstAlias, overwrite)
			if err != nil {
				return err
			}

			if len(results) == 0 {
				fmt.Fprintln(os.Stdout, "No secrets cloned (all skipped or source empty).")
				return nil
			}
			for _, r := range results {
				action := "cloned"
				if r.Overwrote {
					action = "overwritten"
				}
				fmt.Fprintf(os.Stdout, "  %s %s\n", r.Key, action)
			}
			fmt.Fprintf(os.Stdout, "\nTotal: %d secret(s) cloned from %q to %q.\n", len(results), srcAlias, dstAlias)
			return nil
		},
	}

	cloneCmd.Flags().StringVarP(&cfgFile, "config", "c", "vaultswap.yaml", "Path to config file")
	cloneCmd.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite existing secrets at destination")

	rootCmd.AddCommand(cloneCmd)
}
