package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultswap/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd is the base command for vaultswap.
var rootCmd = &cobra.Command{
	Use:   "vaultswap",
	Short: "Sync and rotate secrets across multiple secret managers",
	Long: `vaultswap is a CLI tool for syncing and rotating secrets across
AWS Secrets Manager, HashiCorp Vault, and GCP Secret Manager.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config", "c",
		"vaultswap.yaml",
		"path to vaultswap config file",
	)
}
