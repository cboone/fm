package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/cboone/jm/internal/client"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search emails by text and filters",
	Long: `Search emails using full-text search and/or structured filters.
The optional [query] argument searches across subject, from, to, and body.
If omitted, only the provided flags/filters are used for matching.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := client.SearchOptions{}

		if len(args) > 0 {
			opts.Text = args[0]
		}

		opts.From, _ = cmd.Flags().GetString("from")
		opts.To, _ = cmd.Flags().GetString("to")
		opts.Subject, _ = cmd.Flags().GetString("subject")
		opts.HasAttachment, _ = cmd.Flags().GetBool("has-attachment")
		opts.UnreadOnly, _ = cmd.Flags().GetBool("unread")
		opts.Limit, _ = cmd.Flags().GetUint64("limit")
		if opts.Limit == 0 {
			return exitError("general_error", "--limit must be at least 1", "")
		}
		opts.Offset, _ = cmd.Flags().GetInt64("offset")
		if opts.Offset < 0 {
			return exitError("general_error", "--offset must be non-negative", "")
		}

		sortStr, _ := cmd.Flags().GetString("sort")
		sortField, sortAsc, err := parseSort(sortStr)
		if err != nil {
			return exitError("general_error", err.Error(), "Supported sort fields: receivedAt, sentAt, from, subject")
		}
		opts.SortField = sortField
		opts.SortAsc = sortAsc

		mailboxName, _ := cmd.Flags().GetString("mailbox")

		if beforeStr, _ := cmd.Flags().GetString("before"); beforeStr != "" {
			t, err := parseDate(beforeStr)
			if err != nil {
				return exitError("general_error", "invalid --before date: "+err.Error(),
					"Use RFC 3339 format (e.g. 2026-01-15T00:00:00Z) or a bare date (e.g. 2026-01-15)")
			}
			opts.Before = &t
		}

		if afterStr, _ := cmd.Flags().GetString("after"); afterStr != "" {
			t, err := parseDate(afterStr)
			if err != nil {
				return exitError("general_error", "invalid --after date: "+err.Error(),
					"Use RFC 3339 format (e.g. 2026-01-15T00:00:00Z) or a bare date (e.g. 2026-01-15)")
			}
			opts.After = &t
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in JMAP_TOKEN or config file")
		}

		if mailboxName != "" {
			mailboxID, err := c.ResolveMailboxID(mailboxName)
			if err != nil {
				return exitError("not_found", err.Error(), "")
			}
			opts.MailboxID = string(mailboxID)
		}

		result, err := c.SearchEmails(opts)
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

// parseDate parses a date string in RFC 3339 format or as a bare date (YYYY-MM-DD).
// Bare dates are treated as midnight UTC on that day.
func parseDate(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	t, err2 := time.Parse("2006-01-02", s)
	if err2 == nil {
		return t, nil
	}
	return time.Time{}, err
}

func init() {
	searchCmd.Flags().StringP("mailbox", "m", "", "restrict search to a specific mailbox")
	searchCmd.Flags().Uint64P("limit", "l", 25, "maximum results")
	searchCmd.Flags().Int64P("offset", "o", 0, "pagination offset")
	searchCmd.Flags().BoolP("unread", "u", false, "only show unread messages")
	searchCmd.Flags().StringP("sort", "s", "receivedAt desc", "sort order (receivedAt, sentAt, from, subject) with asc/desc")
	searchCmd.Flags().String("from", "", "filter by sender address/name")
	searchCmd.Flags().String("to", "", "filter by recipient address/name")
	searchCmd.Flags().String("subject", "", "filter by subject text")
	searchCmd.Flags().String("before", "", "emails received before this date (RFC 3339 or YYYY-MM-DD)")
	searchCmd.Flags().String("after", "", "emails received after this date (RFC 3339 or YYYY-MM-DD)")
	searchCmd.Flags().Bool("has-attachment", false, "only emails with attachments")
	rootCmd.AddCommand(searchCmd)
}
