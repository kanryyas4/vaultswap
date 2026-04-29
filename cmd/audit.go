package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultswap/internal/audit"
)

var auditFile string

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Display the vaultswap audit log",
	Long: `Reads and pretty-prints the newline-delimited JSON audit log produced
by vaultswap sync and rotate commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(auditFile)
		if err != nil {
			return fmt.Errorf("open audit log: %w", err)
		}
		defer f.Close()

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tOPERATION\tPROVIDER\tKEY\tSUCCESS\tERROR")

		dec := json.NewDecoder(f)
		for dec.More() {
			var e audit.Entry
			if err := dec.Decode(&e); err != nil {
				return fmt.Errorf("decode entry: %w", err)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%v\t%s\n",
				e.Timestamp.Format("2006-01-02T15:04:05Z"),
				e.Operation,
				e.Provider,
				e.SecretKey,
				e.Success,
				e.Error,
			)
		}
		return w.Flush()
	},
}

func init() {
	auditCmd.Flags().StringVarP(
		&auditFile, "file", "f",
		"vaultswap-audit.jsonl",
		"path to the audit log file",
	)
	rootCmd.AddCommand(auditCmd)
}
