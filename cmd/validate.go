package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultswap/internal/config"
	"github.com/yourusername/vaultswap/internal/provider"
	"github.com/yourusername/vaultswap/internal/validate"
)

func init() {
	var cfgFile string
	var aliases []string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate secrets against defined rules",
		Long: `Check that secrets in one or more providers satisfy
required-key, non-empty, and regex-pattern rules defined in config.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			providers := make(map[string]provider.Provider)
			for _, b := range cfg.Backends {
				if len(aliases) > 0 && !containsAlias(aliases, b.Alias) {
					continue
				}
				p, err := provider.New(b.Provider, b.Options)
				if err != nil {
					return fmt.Errorf("init provider %q: %w", b.Alias, err)
				}
				providers[b.Alias] = p
			}

			var rules []validate.Rule
			for _, r := range cfg.Validate {
				rules = append(rules, validate.Rule{
					Key:      r.Key,
					Required: r.Required,
					NonEmpty: r.NonEmpty,
					Pattern:  r.Pattern,
				})
			}

			v := validate.New(providers, rules)
			results, err := v.Run()
			if err != nil {
				return err
			}

			failed := false
			for _, r := range results {
				status := "PASS"
				if !r.Passed {
					status = "FAIL"
					failed = true
				}
				fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s / %s: %s\n", status, r.Alias, r.Key, r.Message)
			}

			if failed {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "vaultswap.yaml", "path to config file")
	cmd.Flags().StringSliceVarP(&aliases, "aliases", "a", nil, "limit validation to these backend aliases")
	rootCmd.AddCommand(cmd)
}

func containsAlias(list []string, alias string) bool {
	for _, a := range list {
		if a == alias {
			return true
		}
	}
	return false
}
