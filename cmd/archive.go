package cmd

import (
	"os"

	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
	"github.com/spf13/cobra"

	"github.com/cboone/jm/internal/types"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <email-id> [email-id...]",
	Short: "Move emails to the Archive mailbox",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in JMAP_TOKEN or config file")
		}

		archiveMB, err := c.GetMailboxByRole(mailbox.RoleArchive)
		if err != nil {
			return exitError("not_found", "archive mailbox not found: "+err.Error(), "")
		}

		succeeded, errors := c.MoveEmails(args, archiveMB.ID)

		result := types.MoveResult{
			Archived: succeeded,
			Errors:   errors,
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
	rootCmd.AddCommand(archiveCmd)
}
