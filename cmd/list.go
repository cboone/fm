package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List emails in a mailbox",
	RunE: func(cmd *cobra.Command, args []string) error {
		mailboxName, _ := cmd.Flags().GetString("mailbox")
		limit, _ := cmd.Flags().GetUint64("limit")
		if limit == 0 {
			return exitError("general_error", "--limit must be at least 1", "")
		}
		offset, _ := cmd.Flags().GetInt64("offset")
		if offset < 0 {
			return exitError("general_error", "--offset must be non-negative", "")
		}
		unread, _ := cmd.Flags().GetBool("unread")
		flagged, _ := cmd.Flags().GetBool("flagged")
		unflagged, _ := cmd.Flags().GetBool("unflagged")
		if flagged && unflagged {
			return exitError("general_error", "--flagged and --unflagged are mutually exclusive", "")
		}
		sort, _ := cmd.Flags().GetString("sort")

		sortField, sortAsc, err := parseSort(sort)
		if err != nil {
			return exitError("general_error", err.Error(), "Supported sort fields: receivedAt, sentAt, from, subject")
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.ListEmails(client.ListOptions{
			MailboxNameOrID: mailboxName,
			Limit:           limit,
			Offset:          offset,
			UnreadOnly:      unread,
			FlaggedOnly:     flagged,
			UnflaggedOnly:   unflagged,
			SortField:       sortField,
			SortAsc:         sortAsc,
		})
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
	listCmd.Flags().BoolP("flagged", "f", false, "only show flagged messages")
	listCmd.Flags().Bool("unflagged", false, "only show unflagged messages")
	listCmd.Flags().StringP("sort", "s", "receivedAt desc", "sort order (receivedAt, sentAt, from, subject) with asc/desc")
	rootCmd.AddCommand(listCmd)
}

var validSortFields = map[string]string{
	"receivedat": "receivedAt",
	"sentat":     "sentAt",
	"from":       "from",
	"subject":    "subject",
}

func parseSort(s string) (field string, ascending bool, err error) {
	s = strings.ReplaceAll(s, ":", " ")
	parts := strings.Fields(s)
	field = "receivedAt"
	ascending = false

	if len(parts) >= 1 {
		normalized, ok := validSortFields[strings.ToLower(parts[0])]
		if !ok {
			return "", false, fmt.Errorf("unsupported sort field %q", parts[0])
		}
		field = normalized
	}
	if len(parts) >= 2 {
		switch strings.ToLower(parts[1]) {
		case "asc":
			ascending = true
		case "desc":
			// already default
		default:
			return "", false, fmt.Errorf("unsupported sort direction %q (use asc or desc)", parts[1])
		}
	}
	return field, ascending, nil
}
