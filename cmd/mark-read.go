package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/types"
)

var markReadCmd = &cobra.Command{
	Use:   "mark-read <email-id> [email-id...]",
	Short: "Mark emails as read (set the $seen keyword)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in FM_TOKEN or config file")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunPreview(c, args, "mark-read", nil)
		}

		succeeded, errors := c.MarkAsRead(args)

		result := types.MoveResult{
			Matched:      len(args),
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
	rootCmd.AddCommand(markReadCmd)
}
