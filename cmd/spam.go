package cmd

import (
	"os"

	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
	"github.com/spf13/cobra"

	"github.com/cboone/jm/internal/types"
)

var spamCmd = &cobra.Command{
	Use:   "spam <email-id> [email-id...]",
	Short: "Move emails to the Junk/Spam mailbox",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in JMAP_TOKEN or config file")
		}

		junkMB, err := c.GetMailboxByRole(mailbox.RoleJunk)
		if err != nil {
			return exitError("not_found", "junk mailbox not found: "+err.Error(), "")
		}

		succeeded, errors := c.MarkAsSpam(args, junkMB.ID)

		result := types.MoveResult{
			MarkedSpam: succeeded,
			Errors:     errors,
			Destination: &types.DestinationInfo{
				ID:   string(junkMB.ID),
				Name: junkMB.Name,
			},
		}

		if err := formatter().Format(os.Stdout, result); err != nil {
			return err
		}

		if len(errors) > 0 {
			return exitError("partial_failure", "one or more emails failed to mark as spam", "")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(spamCmd)
}
