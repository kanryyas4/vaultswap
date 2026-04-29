package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultswap/vaultswap/internal/audit"
	"github.com/vaultswap/vaultswap/internal/config"
	"github.com/vaultswap/vaultswap/internal/copy"
	"github.com/vaultswap/vaultswap/internal/provider"
)

var (
	copySourceAlias string
	copySourceKey   string
	copyDestAlias   string
	copyDestKey     string
	copyOverwrite   bool
	copyAuditFile   string
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy a single secret from one provider to another",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		providers := make(map[string]provider.Provider, len(cfg.Providers))
		for _, p := range cfg.Providers {
			prov, err := provider.New(p.Provider, p.Options)
			if err != nil {
				return fmt.Errorf("init provider %q: %w", p.Alias, err)
			}
			providers[p.Alias] = prov
		}

		var auditor *audit.Auditor
		if copyAuditFile != "" {
			auditor, err = audit.NewFile(copyAuditFile)
			if err != nil {
				return fmt.Errorf("open audit log: %w", err)
			}
		} else {
			auditor = audit.New()
		}

		c := copy.New(providers, auditor)
		if err := c.Copy(cmd.Context(), copy.Options{
			SourceAlias: copySourceAlias,
			SourceKey:   copySourceKey,
			DestAlias:   copyDestAlias,
			DestKey:     copyDestKey,
			Overwrite:   copyOverwrite,
		}); err != nil {
			fmt.Fprintln(os.Stderr, "copy failed:", err)
			return err
		}

		fmt.Printf("copied %q (%s) → %q (%s)\n", copySourceKey, copySourceAlias, copyDestKey, copyDestAlias)
		return nil
	},
}

func init() {
	copyCmd.Flags().StringVar(&copySourceAlias, "src", "", "source provider alias (required)")
	copyCmd.Flags().StringVar(&copySourceKey, "src-key", "", "secret key to copy from source (required)")
	copyCmd.Flags().StringVar(&copyDestAlias, "dst", "", "destination provider alias (required)")
	copyCmd.Flags().StringVar(&copyDestKey, "dst-key", "", "secret key at destination (defaults to src-key)")
	copyCmd.Flags().BoolVar(&copyOverwrite, "overwrite", false, "overwrite destination key if it already exists")
	copyCmd.Flags().StringVar(&copyAuditFile, "audit-log", "", "path to append audit log entries")
	_ = copyCmd.MarkFlagRequired("src")
	_ = copyCmd.MarkFlagRequired("src-key")
	_ = copyCmd.MarkFlagRequired("dst")
	rootCmd.AddCommand(copyCmd)
}
