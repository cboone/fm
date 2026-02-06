package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List emails in a mailbox",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in JMAP_TOKEN or config file")
		}

		mailboxName, _ := cmd.Flags().GetString("mailbox")
		limit, _ := cmd.Flags().GetUint64("limit")
		if limit == 0 {
			return exitError("general_error", "--limit must be at least 1", "")
		}
		offset, _ := cmd.Flags().GetInt64("offset")
		unread, _ := cmd.Flags().GetBool("unread")
		sort, _ := cmd.Flags().GetString("sort")

		sortField, sortAsc := parseSort(sort)

		result, err := c.ListEmails(mailboxName, limit, offset, unread, sortField, sortAsc)
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	listCmd.Flags().StringP("mailbox", "m", "inbox", "mailbox name or ID")
	listCmd.Flags().Uint64P("limit", "l", 25, "maximum number of results")
	listCmd.Flags().Int64P("offset", "o", 0, "pagination offset")
	listCmd.Flags().BoolP("unread", "u", false, "only show unread messages")
	listCmd.Flags().StringP("sort", "s", "receivedAt desc", "sort order (receivedAt, sentAt, from, subject) with asc/desc")
	rootCmd.AddCommand(listCmd)
}

func parseSort(s string) (field string, ascending bool) {
	parts := strings.Fields(s)
	field = "receivedAt"
	ascending = false

	if len(parts) >= 1 {
		field = parts[0]
	}
	if len(parts) >= 2 && strings.EqualFold(parts[1], "asc") {
		ascending = true
	}
	return field, ascending
}
