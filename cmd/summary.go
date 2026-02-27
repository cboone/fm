package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show inbox triage summary with sender and domain aggregation",
	Args:  cobra.NoArgs,
	Long: `Aggregate emails by sender and domain, count unread messages, and optionally
detect newsletters. Provides a single-pass triage overview of a mailbox.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mailboxName, _ := cmd.Flags().GetString("mailbox")
		unread, _ := cmd.Flags().GetBool("unread")
		flagged, _ := cmd.Flags().GetBool("flagged")
		unflagged, _ := cmd.Flags().GetBool("unflagged")
		limit, _ := cmd.Flags().GetInt("limit")
		subjects, _ := cmd.Flags().GetBool("subjects")
		newsletters, _ := cmd.Flags().GetBool("newsletters")

		if flagged && unflagged {
			return exitError("general_error", "--flagged and --unflagged are mutually exclusive", "")
		}
		if limit < 1 {
			return exitError("general_error", "--limit must be at least 1", "")
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		mailboxID, err := c.ResolveMailboxID(mailboxName)
		if err != nil {
			return exitError("not_found", err.Error(), "")
		}

		result, err := c.AggregateSummary(client.SummaryOptions{
			MailboxID:     string(mailboxID),
			UnreadOnly:    unread,
			FlaggedOnly:   flagged,
			UnflaggedOnly: unflagged,
			Limit:         limit,
			Subjects:      subjects,
			Newsletters:   newsletters,
		})
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	summaryCmd.Flags().StringP("mailbox", "m", "inbox", "mailbox name or ID")
	summaryCmd.Flags().BoolP("unread", "u", false, "only count unread messages")
	summaryCmd.Flags().BoolP("flagged", "f", false, "only count flagged messages")
	summaryCmd.Flags().Bool("unflagged", false, "only count unflagged messages")
	summaryCmd.Flags().IntP("limit", "l", 10, "number of top senders/domains to show")
	summaryCmd.Flags().Bool("subjects", false, "include sample subjects per sender")
	summaryCmd.Flags().Bool("newsletters", false, "detect newsletters via List-Id/List-Unsubscribe headers")
	rootCmd.AddCommand(summaryCmd)
}
