package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultswap/internal/config"
	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/rename"
)

var (
	renameSourceAlias string
	renameDestAlias   string
	renameOldKey      string
	renameNewKey      string
	renameDeleteOld   bool
)

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename (re-key) a secret within or across providers",
	Long: `Read a secret by its current key from the source provider and write it
under a new key in the destination provider. Optionally delete the original.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		providers := make(map[string]provider.Provider, len(cfg.Providers))
		for _, pc := range cfg.Providers {
			p, err := provider.New(pc)
			if err != nil {
				return fmt.Errorf("init provider %q: %w", pc.Alias, err)
			}
			providers[pc.Alias] = p
		}

		r := rename.New(providers)
		if err := r.Run(cmd.Context(), rename.Options{
			SourceAlias: renameSourceAlias,
			DestAlias:   renameDestAlias,
			OldKey:      renameOldKey,
			NewKey:      renameNewKey,
			DeleteOld:   renameDeleteOld,
		}); err != nil {
			fmt.Fprintln(os.Stderr, "rename failed:", err)
			return err
		}

		fmt.Printf("renamed %q -> %q (src=%s dst=%s)\n", renameOldKey, renameNewKey, renameSourceAlias, renameDestAlias)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
	renameCmd.Flags().StringVarP(&renameSourceAlias, "source", "s", "", "source provider alias (required)")
	renameCmd.Flags().StringVarP(&renameDestAlias, "dest", "d", "", "destination provider alias (required)")
	renameCmd.Flags().StringVar(&renameOldKey, "old-key", "", "existing secret key name (required)")
	renameCmd.Flags().StringVar(&renameNewKey, "new-key", "", "new secret key name (required)")
	renameCmd.Flags().BoolVar(&renameDeleteOld, "delete-old", false, "delete the original key after renaming")
	_ = renameCmd.MarkFlagRequired("source")
	_ = renameCmd.MarkFlagRequired("dest")
	_ = renameCmd.MarkFlagRequired("old-key")
	_ = renameCmd.MarkFlagRequired("new-key")
}
