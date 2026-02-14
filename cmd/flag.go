package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/jm/internal/types"
)

var flagCmd = &cobra.Command{
	Use:   "flag <email-id> [email-id...]",
	Short: "Flag emails (set the $flagged keyword)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in JMAP_TOKEN or config file")
		}

		succeeded, errors := c.SetFlagged(args)

		result := types.MoveResult{
			Flagged: succeeded,
			Errors:  errors,
		}

		if err := formatter().Format(os.Stdout, result); err != nil {
			return err
		}

		if len(errors) > 0 {
			return exitError("partial_failure", "one or more emails failed to flag", "")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(flagCmd)
}
