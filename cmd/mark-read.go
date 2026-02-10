package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/jm/internal/types"
)

var markReadCmd = &cobra.Command{
	Use:   "mark-read <email-id> [email-id...]",
	Short: "Mark emails as read (set the $seen keyword)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in JMAP_TOKEN or config file")
		}

		succeeded, errors := c.MarkAsRead(args)

		result := types.MoveResult{
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
	rootCmd.AddCommand(markReadCmd)
}
