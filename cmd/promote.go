package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultswap/internal/config"
	"github.com/yourusername/vaultswap/internal/promote"
	"github.com/yourusername/vaultswap/internal/provider"
)

func init() {
	var (
		cfgFile      string
		sourcePrefix string
		destPrefix   string
		dryRun       bool
		overwrite    bool
	)

	cmd := &cobra.Command{
		Use:   "promote <source-alias> <dest-alias>",
		Short: "Promote secrets from one environment to another",
		Long: `Copies secrets from the source provider to the destination provider.
An optional source prefix filter and destination prefix can be applied
to rewrite keys during promotion (e.g. staging/ -> prod/).`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			providers := make(map[string]provider.Provider)
			for _, b := range cfg.Backends {
				p, err := provider.New(b.Provider, b.Options)
				if err != nil {
					return fmt.Errorf("init provider %q: %w", b.Alias, err)
				}
				providers[b.Alias] = p
			}

			promoter := promote.New(providers)
			res, err := promoter.Promote(cmd.Context(), promote.Options{
				SourceAlias:  args[0],
				DestAlias:    args[1],
				SourcePrefix: sourcePrefix,
				DestPrefix:   destPrefix,
				DryRun:       dryRun,
				Overwrite:    overwrite,
			})
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Fprintln(os.Stdout, "[dry-run] secrets that would be promoted:")
			}
			for _, k := range res.Promoted {
				fmt.Fprintf(os.Stdout, "  promoted: %s\n", k)
			}
			for _, k := range res.Skipped {
				fmt.Fprintf(os.Stdout, "  skipped (exists): %s\n", k)
			}
			fmt.Fprintf(os.Stdout, "done: %d promoted, %d skipped\n", len(res.Promoted), len(res.Skipped))
			return nil
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "vaultswap.yaml", "path to config file")
	cmd.Flags().StringVar(&sourcePrefix, "source-prefix", "", "only promote keys with this prefix (stripped before writing)")
	cmd.Flags().StringVar(&destPrefix, "dest-prefix", "", "prepend this prefix to destination keys")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing secrets in destination")

	rootCmd.AddCommand(cmd)
}
