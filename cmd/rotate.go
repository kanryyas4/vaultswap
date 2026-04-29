package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultswap/internal/config"
	"github.com/vaultswap/internal/provider"
	"github.com/vaultswap/internal/rotate"
)

var (
	rotateAlias     string
	rotateSecretKey string
	rotateNewValue  string
	rotateBackupKey string
)

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate a secret in a target provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		providers := make(map[string]interface {
			GetSecret(context.Context, string) (string, error)
			PutSecret(context.Context, string, string) error
			DeleteSecret(context.Context, string) error
			ListSecrets(context.Context) ([]string, error)
		})

		for _, s := range cfg.Sources {
			p, err := provider.New(s.Provider, s.Options)
			if err != nil {
				return fmt.Errorf("init provider %q: %w", s.Alias, err)
			}
			providers[s.Alias] = p
		}

		r := rotate.New(providers)
		res := r.Rotate(context.Background(), rotate.Options{
			Alias:     rotateAlias,
			SecretKey: rotateSecretKey,
			NewValue:  rotateNewValue,
			BackupKey: rotateBackupKey,
		})

		if !res.Success {
			fmt.Fprintf(os.Stderr, "rotation failed: %v\n", res.Err)
			return res.Err
		}

		fmt.Printf("rotated %q on %q (old value backed up: %v)\n",
			rotateSecretKey, rotateAlias, rotateBackupKey != "")
		return nil
	},
}

func init() {
	rotateCmd.Flags().StringVarP(&rotateAlias, "alias", "a", "", "provider alias (required)")
	rotateCmd.Flags().StringVarP(&rotateSecretKey, "key", "k", "", "secret key to rotate (required)")
	rotateCmd.Flags().StringVarP(&rotateNewValue, "value", "v", "", "new secret value (required)")
	rotateCmd.Flags().StringVar(&rotateBackupKey, "backup-key", "", "key to store old value before overwriting")
	_ = rotateCmd.MarkFlagRequired("alias")
	_ = rotateCmd.MarkFlagRequired("key")
	_ = rotateCmd.MarkFlagRequired("value")
	RootCmd.AddCommand(rotateCmd)
}
