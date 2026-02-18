package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
)

const recipientToUsage = "filter by recipient address/name"

// filterFlagNames lists all flags that addFilterFlags may register.
var filterFlagNames = []string{
	"mailbox", "from", "to", "subject",
	"before", "after", "has-attachment",
	"unread", "flagged", "unflagged",
}

// addFilterFlags registers shared search/filter flags on an action command.
// It skips --to if the command already defines that flag (e.g. move).
func addFilterFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("mailbox", "m", "", "restrict to a specific mailbox")
	cmd.Flags().String("from", "", "filter by sender address/name")
	if cmd.Flags().Lookup("to") == nil {
		cmd.Flags().String("to", "", recipientToUsage)
	}
	cmd.Flags().String("subject", "", "filter by subject text")
	cmd.Flags().String("before", "", "emails received before this date (RFC 3339 or YYYY-MM-DD)")
	cmd.Flags().String("after", "", "emails received after this date (RFC 3339 or YYYY-MM-DD)")
	cmd.Flags().Bool("has-attachment", false, "only emails with attachments")
	cmd.Flags().BoolP("unread", "u", false, "only unread messages")
	cmd.Flags().BoolP("flagged", "f", false, "only flagged messages")
	cmd.Flags().Bool("unflagged", false, "only unflagged messages")
}

// hasFilterFlags returns true if any filter flag has an effective value.
// It ignores no-op values such as --unread=false and --subject "".
// It also skips --to on commands where it is the destination flag
// (e.g. move) instead of a recipient filter.
func hasFilterFlags(cmd *cobra.Command) bool {
	for _, name := range filterFlagNames {
		f := cmd.Flags().Lookup(name)
		if f == nil || !cmd.Flags().Changed(name) {
			continue
		}

		switch name {
		case "mailbox", "from", "to", "subject", "before", "after":
			if name == "to" && !isRecipientToFilterFlag(cmd) {
				continue
			}
			value, _ := cmd.Flags().GetString(name)
			if strings.TrimSpace(value) != "" {
				return true
			}
		case "has-attachment", "unread", "flagged", "unflagged":
			value, _ := cmd.Flags().GetBool(name)
			if value {
				return true
			}
		}
	}
	return false
}

func isRecipientToFilterFlag(cmd *cobra.Command) bool {
	f := cmd.Flags().Lookup("to")
	return f != nil && f.Usage == recipientToUsage
}

// parseFilterOptions reads filter flags from the command and builds SearchOptions.
func parseFilterOptions(cmd *cobra.Command, c *client.Client) (client.SearchOptions, error) {
	opts := client.SearchOptions{}

	if from, _ := cmd.Flags().GetString("from"); strings.TrimSpace(from) != "" {
		opts.From = from
	}
	if isRecipientToFilterFlag(cmd) {
		if to, _ := cmd.Flags().GetString("to"); strings.TrimSpace(to) != "" {
			opts.To = to
		}
	}
	if subject, _ := cmd.Flags().GetString("subject"); strings.TrimSpace(subject) != "" {
		opts.Subject = subject
	}
	opts.HasAttachment, _ = cmd.Flags().GetBool("has-attachment")
	opts.UnreadOnly, _ = cmd.Flags().GetBool("unread")
	opts.FlaggedOnly, _ = cmd.Flags().GetBool("flagged")
	opts.UnflaggedOnly, _ = cmd.Flags().GetBool("unflagged")

	if opts.FlaggedOnly && opts.UnflaggedOnly {
		return client.SearchOptions{}, exitError("general_error", "--flagged and --unflagged are mutually exclusive", "")
	}

	if beforeStr, _ := cmd.Flags().GetString("before"); strings.TrimSpace(beforeStr) != "" {
		beforeStr = strings.TrimSpace(beforeStr)
		t, err := parseDate(beforeStr)
		if err != nil {
			return client.SearchOptions{}, exitError("general_error", "invalid --before date: "+err.Error(),
				"Use RFC 3339 format (e.g. 2026-01-15T00:00:00Z) or a bare date (e.g. 2026-01-15)")
		}
		opts.Before = &t
	}

	if afterStr, _ := cmd.Flags().GetString("after"); strings.TrimSpace(afterStr) != "" {
		afterStr = strings.TrimSpace(afterStr)
		t, err := parseDate(afterStr)
		if err != nil {
			return client.SearchOptions{}, exitError("general_error", "invalid --after date: "+err.Error(),
				"Use RFC 3339 format (e.g. 2026-01-15T00:00:00Z) or a bare date (e.g. 2026-01-15)")
		}
		opts.After = &t
	}

	if mailboxName, _ := cmd.Flags().GetString("mailbox"); strings.TrimSpace(mailboxName) != "" {
		mailboxName = strings.TrimSpace(mailboxName)
		mailboxID, err := c.ResolveMailboxID(mailboxName)
		if err != nil {
			return client.SearchOptions{}, exitError("not_found", err.Error(), "")
		}
		opts.MailboxID = string(mailboxID)
	}

	return opts, nil
}

// validateIDsOrFilters ensures exactly one of email IDs or filter flags is provided.
// It also checks for mutually exclusive filter flags early, before authentication.
func validateIDsOrFilters(cmd *cobra.Command, args []string) error {
	hasIDs := len(args) > 0
	hasFilters := hasFilterFlags(cmd)

	if hasIDs && hasFilters {
		return exitError("general_error", "cannot combine email IDs with filter flags",
			"Use either email IDs or filter flags, not both")
	}
	if !hasIDs && !hasFilters {
		return exitError("general_error", "no emails specified",
			"Provide email IDs as arguments or use filter flags (e.g. --mailbox inbox --unread)")
	}

	// Check mutually exclusive flags early (before client creation).
	flagged, _ := cmd.Flags().GetBool("flagged")
	unflagged, _ := cmd.Flags().GetBool("unflagged")
	if flagged && unflagged {
		return exitError("general_error", "--flagged and --unflagged are mutually exclusive", "")
	}

	return nil
}

// resolveEmailIDs returns email IDs from args or queries them using filter flags.
func resolveEmailIDs(cmd *cobra.Command, args []string, c *client.Client) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	opts, err := parseFilterOptions(cmd, c)
	if err != nil {
		return nil, err
	}

	ids, err := c.QueryEmailIDs(opts)
	if err != nil {
		return nil, exitError("jmap_error", err.Error(), "")
	}

	if len(ids) == 0 {
		return nil, exitError("not_found", "no emails matched the given filters", "")
	}

	return ids, nil
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
