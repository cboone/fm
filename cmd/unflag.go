package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/types"
)

var unflagCmd = &cobra.Command{
	Use:   "unflag [email-id...]",
	Short: "Unflag emails (remove the $flagged keyword)",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateIDsOrFilters(cmd, args); err != nil {
			return err
		}

		colorOnly, _ := cmd.Flags().GetBool("color")

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		ids, err := resolveEmailIDs(cmd, args, c)
		if err != nil {
			return err
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunPreview(c, ids, "unflag", nil)
		}

		var succeeded, errors []string
		if colorOnly {
			succeeded, errors = c.ClearFlagColor(ids)
		} else {
			succeeded, errors = c.SetUnflagged(ids)
		}

		result := types.MoveResult{
			Matched:   len(ids),
			Processed: len(succeeded) + len(errors),
			Failed:    len(errors),
			Unflagged: succeeded,
			Errors:    errors,
		}

		if err := formatter().Format(os.Stdout, result); err != nil {
			return err
		}

		if len(errors) > 0 {
			return exitError("partial_failure", "one or more emails failed to unflag", "")
		}

		return nil
	},
}

func init() {
	unflagCmd.Flags().BoolP("color", "c", false, "remove flag color only (keep the email flagged)")
	unflagCmd.Flags().BoolP("dry-run", "n", false, "preview affected emails without making changes")
	addFilterFlags(unflagCmd)
	rootCmd.AddCommand(unflagCmd)
}
