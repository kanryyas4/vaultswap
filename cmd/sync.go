package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultswap/internal/audit"
	"github.com/yourusername/vaultswap/internal/config"
	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/sync"
)

var (
	syncSource  string
	syncDest    string
	syncDryRun  bool
	syncAudit   string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync secrets from a source provider to a destination provider",
	Long: `Sync copies all secrets from the source provider alias to the destination
provider alias. Secrets already present in the destination are overwritten
with the source value.

Use --dry-run to preview changes without writing anything.`,
	Example: `  vaultswap sync --source aws-prod --dest vault-staging
  vaultswap sync --source gcp-main --dest aws-backup --dry-run`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringVar(&syncSource, "source", "", "alias of the source provider (required)")
	syncCmd.Flags().StringVar(&syncDest, "dest", "", "alias of the destination provider (required)")
	syncCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "preview changes without writing to the destination")
	syncCmd.Flags().StringVar(&syncAudit, "audit-log", "", "optional path to append audit log entries")

	_ = syncCmd.MarkFlagRequired("source")
	_ = syncCmd.MarkFlagRequired("dest")
}

func runSync(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	providers := make(map[string]provider.Provider)
	for _, p := range cfg.Providers {
		prov, err := provider.New(p.Provider, p.Options)
		if err != nil {
			return fmt.Errorf("initialising provider %q: %w", p.Alias, err)
		}
		providers[p.Alias] = prov
	}

	var auditor *audit.Auditor
	if syncAudit != "" {
		f, err := os.OpenFile(syncAudit, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("opening audit log: %w", err)
		}
		defer f.Close()
		auditor = audit.NewFile(f)
	} else {
		auditor = audit.New(cmd.OutOrStdout())
	}

	syncer := sync.New(providers, auditor)

	result, err := syncer.Sync(cmd.Context(), syncSource, syncDest, syncDryRun)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if syncDryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] would sync %d secret(s) from %q to %q\n",
			result.Count, syncSource, syncDest)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "synced %d secret(s) from %q to %q\n",
			result.Count, syncSource, syncDest)
	}

	return nil
}
