package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/types"
)

var markReadCmd = &cobra.Command{
	Use:   "mark-read [email-id...]",
	Short: "Mark emails as read (set the $seen keyword)",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateIDsOrFilters(cmd, args); err != nil {
			return err
		}

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
			return dryRunPreview(c, ids, "mark-read", nil)
		}

		succeeded, errors := c.MarkAsRead(ids)

		result := types.MoveResult{
			Matched:      len(ids),
			Processed:    len(succeeded) + len(errors),
			Failed:       len(errors),
			MarkedAsRead: succeeded,
			Errors:       errors,
		}

		if err := formatter().Format(os.Stdout, result); err != nil {
			return err
		}

		if len(errors) > 0 {
			return exitError("partial_failure", "one or more emails failed to mark as read", "")
		}

		return nil
	},
}

func init() {
	markReadCmd.Flags().BoolP("dry-run", "n", false, "preview affected emails without making changes")
	addFilterFlags(markReadCmd)
	rootCmd.AddCommand(markReadCmd)
}
