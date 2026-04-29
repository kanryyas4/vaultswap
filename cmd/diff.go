package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vaultswap/internal/config"
	"github.com/vaultswap/internal/diff"
	"github.com/vaultswap/internal/provider"
)

var diffCmd = &cobra.Command{
	Use:   "diff <source-alias> <dest-alias>",
	Short: "Show differences between secrets in two providers",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgFile, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		srcAlias, dstAlias := args[0], args[1]

		var srcCfg, dstCfg *config.Backend
		for i := range cfg.Backends {
			switch cfg.Backends[i].Alias {
			case srcAlias:
				srcCfg = &cfg.Backends[i]
			case dstAlias:
				dstCfg = &cfg.Backends[i]
			}
		}
		if srcCfg == nil {
			return fmt.Errorf("unknown source alias: %q", srcAlias)
		}
		if dstCfg == nil {
			return fmt.Errorf("unknown dest alias: %q", dstAlias)
		}

		src, err := provider.New(srcCfg.Provider, srcCfg.Options)
		if err != nil {
			return fmt.Errorf("init source provider: %w", err)
		}
		dst, err := provider.New(dstCfg.Provider, dstCfg.Options)
		if err != nil {
			return fmt.Errorf("init dest provider: %w", err)
		}

		res, err := diff.New(src, dst).Compare(cmd.Context())
		if err != nil {
			return err
		}

		w := cmd.OutOrStdout()
		printGroup := func(label string, keys []string) {
			if len(keys) == 0 {
				return
			}
			fmt.Fprintf(w, "%s:\n", label)
			for _, k := range keys {
				fmt.Fprintf(w, "  - %s\n", k)
			}
		}
		printGroup("only in source", res.OnlyInSource)
		printGroup("only in dest", res.OnlyInDest)
		printGroup("diverged", res.Diverged)
		printGroup("in sync", res.InSync)

		if len(res.Diverged) > 0 || len(res.OnlyInSource) > 0 || len(res.OnlyInDest) > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().StringP("config", "c", "vaultswap.yaml", "path to config file")
}
