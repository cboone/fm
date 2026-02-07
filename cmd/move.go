package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/jm/internal/client"
	"github.com/cboone/jm/internal/types"
)

var moveCmd = &cobra.Command{
	Use:   "move <email-id> [email-id...] --to <mailbox>",
	Short: "Move emails to a specified mailbox",
	Long: `Move one or more emails to a target mailbox (by name or ID).
Moving to Trash or Deleted Items is not permitted.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("to")

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in JMAP_TOKEN or config file")
		}

		targetMB, err := c.GetMailboxByNameOrID(target)
		if err != nil {
			return exitError("not_found", err.Error(), "")
		}

		// Safety check: refuse to move to trash.
		if err := client.ValidateTargetMailbox(targetMB); err != nil {
			return exitError("forbidden_operation", err.Error(),
				"Deletion is not permitted by this tool")
		}

		succeeded, errors := c.MoveEmails(args, targetMB.ID)

		result := types.MoveResult{
			Moved:  succeeded,
			Errors: errors,
			Destination: &types.DestinationInfo{
				ID:   string(targetMB.ID),
				Name: targetMB.Name,
			},
		}

		if err := formatter().Format(os.Stdout, result); err != nil {
			return err
		}

		if len(errors) > 0 {
			return exitError("partial_failure", "one or more emails failed to move", "")
		}

		return nil
	},
}

func init() {
	moveCmd.Flags().String("to", "", "target mailbox name or ID (required)")
	if err := moveCmd.MarkFlagRequired("to"); err != nil {
		panic(err)
	}
	rootCmd.AddCommand(moveCmd)
}
