package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
	"github.com/cboone/fm/internal/types"
)

var moveCmd = &cobra.Command{
	Use:   "move [email-id...] --to <mailbox>",
	Short: "Move emails to a specified mailbox",
	Long: `Move one or more emails to a target mailbox (by name or ID).
Moving to Trash or Deleted Items is not permitted.`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateIDsOrFilters(cmd, args); err != nil {
			return err
		}

		target, _ := cmd.Flags().GetString("to")
		if target == "" {
			return exitError("general_error", "required flag \"to\" not set",
				"Specify the destination mailbox with --to <mailbox>")
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
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

		ids, err := resolveEmailIDs(cmd, args, c)
		if err != nil {
			return err
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunPreview(c, ids, "move", &types.DestinationInfo{
				ID:   string(targetMB.ID),
				Name: targetMB.Name,
			})
		}

		succeeded, errors := c.MoveEmails(ids, targetMB.ID)

		result := types.MoveResult{
			Matched:   len(ids),
			Processed: len(succeeded) + len(errors),
			Failed:    len(errors),
			Moved:     succeeded,
			Errors:    errors,
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
	moveCmd.Flags().BoolP("dry-run", "n", false, "preview affected emails without making changes")
	addFilterFlags(moveCmd)
	rootCmd.AddCommand(moveCmd)
}
