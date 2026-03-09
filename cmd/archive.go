package cmd

import (
	"os"

	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/types"
)

var archiveCmd = &cobra.Command{
	Use:   "archive [email-id...]",
	Short: "Move emails to the Archive mailbox",
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

		archiveMB, err := c.GetMailboxByRole(mailbox.RoleArchive)
		if err != nil {
			return exitError("not_found", "archive mailbox not found: "+err.Error(), "")
		}

		ids, err := resolveEmailIDs(cmd, args, c)
		if err != nil {
			return err
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunPreview(c, ids, "archive", &types.DestinationInfo{
				ID:   string(archiveMB.ID),
				Name: archiveMB.Name,
			})
		}

		succeeded, errors := c.MoveEmails(ids, archiveMB.ID)

		result := types.MoveResult{
			Matched:   len(ids),
			Processed: len(succeeded) + len(errors),
			Failed:    len(errors),
			Archived:  succeeded,
			Errors:    errors,
			Destination: &types.DestinationInfo{
				ID:   string(archiveMB.ID),
				Name: archiveMB.Name,
			},
		}

		if err := formatter().Format(os.Stdout, result); err != nil {
			return err
		}

		if len(errors) > 0 {
			return exitError("partial_failure", "one or more emails failed to archive", "")
		}

		return nil
	},
}

func init() {
	archiveCmd.Flags().BoolP("dry-run", "n", false, "preview affected emails without making changes")
	addFilterFlags(archiveCmd)
	rootCmd.AddCommand(archiveCmd)
}
