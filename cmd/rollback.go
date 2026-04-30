package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultswap/internal/config"
	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/rollback"
)

var (
	rollbackAlias  string
	rollbackKeys   []string
	rollbackDryRun bool
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Restore secrets to a previously captured snapshot",
	Long: `Capture the current state of an alias, optionally filter by key,
and immediately restore it (useful for scripted undo after a failed operation).`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		ctx := context.Background()
		rb := rollback.New(providers)

		snap, err := rb.Capture(ctx, rollbackAlias)
		if err != nil {
			return err
		}

		// Filter to requested keys when --keys is provided.
		if len(rollbackKeys) > 0 {
			filtered := make(map[string]string, len(rollbackKeys))
			for _, k := range rollbackKeys {
				if v, ok := snap.Secrets[k]; ok {
					filtered[k] = v
				}
			}
			snap.Secrets = filtered
		}

		if rollbackDryRun {
			fmt.Fprintf(os.Stdout, "dry-run: would restore %d secret(s) into %q\n",
				len(snap.Secrets), snap.Alias)
			return nil
		}

		if err := rb.Restore(ctx, snap); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "restored %d secret(s) into %q\n", len(snap.Secrets), snap.Alias)
		return nil
	},
}

func init() {
	rollbackCmd.Flags().StringVar(&rollbackAlias, "alias", "", "provider alias to snapshot and restore (required)")
	rollbackCmd.Flags().StringSliceVar(&rollbackKeys, "keys", nil, "comma-separated list of keys to restore (default: all)")
	rollbackCmd.Flags().BoolVar(&rollbackDryRun, "dry-run", false, "print what would be restored without writing")
	_ = rollbackCmd.MarkFlagRequired("alias")
	rootCmd.AddCommand(rollbackCmd)
}
