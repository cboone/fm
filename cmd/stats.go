package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show sender aggregation stats for a mailbox",
	Long: `Aggregate emails by sender address and display per-sender counts.
Queries all matching emails in the mailbox and groups them by sender,
sorted by volume descending.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mailboxName, _ := cmd.Flags().GetString("mailbox")
		unread, _ := cmd.Flags().GetBool("unread")
		flagged, _ := cmd.Flags().GetBool("flagged")
		unflagged, _ := cmd.Flags().GetBool("unflagged")
		subjects, _ := cmd.Flags().GetBool("subjects")

		if flagged && unflagged {
			return exitError("general_error", "--flagged and --unflagged are mutually exclusive", "")
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in FM_TOKEN or config file")
		}

		mailboxID, err := c.ResolveMailboxID(mailboxName)
		if err != nil {
			return exitError("not_found", err.Error(), "")
		}

		result, err := c.AggregateEmailsBySender(client.StatsOptions{
			MailboxID:     string(mailboxID),
			UnreadOnly:    unread,
			FlaggedOnly:   flagged,
			UnflaggedOnly: unflagged,
			Subjects:      subjects,
		})
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	statsCmd.Flags().StringP("mailbox", "m", "inbox", "mailbox name or ID")
	statsCmd.Flags().BoolP("unread", "u", false, "only count unread messages")
	statsCmd.Flags().BoolP("flagged", "f", false, "only count flagged messages")
	statsCmd.Flags().Bool("unflagged", false, "only count unflagged messages")
	statsCmd.Flags().Bool("subjects", false, "include subject lines per sender")
	rootCmd.AddCommand(statsCmd)
}
