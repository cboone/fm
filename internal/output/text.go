package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/cboone/fm/internal/types"
)

const (
	// maxFromWidth is the maximum rune count for the sender column in email lists.
	maxFromWidth = 40
	// maxSubjectWidth is the maximum rune count for the subject column in email lists.
	maxSubjectWidth = 80
)

// TextFormatter outputs data as human-readable text.
type TextFormatter struct{}

func (f *TextFormatter) Format(w io.Writer, v any) error {
	switch val := v.(type) {
	case types.SessionInfo:
		return f.formatSession(w, val)
	case []types.MailboxInfo:
		return f.formatMailboxes(w, val)
	case types.EmailListResult:
		return f.formatEmailList(w, val)
	case types.EmailDetail:
		return f.formatEmailDetail(w, val)
	case types.ThreadView:
		return f.formatThreadView(w, val)
	case types.MoveResult:
		return f.formatMoveResult(w, val)
	case types.DryRunResult:
		return f.formatDryRunResult(w, val)
	default:
		// Fall back to JSON formatter for unknown types.
		return (&JSONFormatter{}).Format(w, v)
	}
}

func (f *TextFormatter) FormatError(w io.Writer, code string, message string, hint string) error {
	fmt.Fprintf(w, "Error [%s]: %s\n", code, message)
	if hint != "" {
		fmt.Fprintf(w, "Hint: %s\n", hint)
	}
	return nil
}

func (f *TextFormatter) formatSession(w io.Writer, s types.SessionInfo) error {
	fmt.Fprintf(w, "Username: %s\n", s.Username)
	fmt.Fprintf(w, "Capabilities: %s\n", strings.Join(s.Capabilities, ", "))
	ids := make([]string, 0, len(s.Accounts))
	for id := range s.Accounts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		acct := s.Accounts[id]
		personal := ""
		if acct.IsPersonal {
			personal = " (personal)"
		}
		fmt.Fprintf(w, "Account: %s - %s%s\n", id, acct.Name, personal)
	}
	return nil
}

func (f *TextFormatter) formatMailboxes(w io.Writer, mailboxes []types.MailboxInfo) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, mb := range mailboxes {
		role := ""
		if mb.Role != "" {
			role = fmt.Sprintf("[%s]", mb.Role)
		}
		fmt.Fprintf(tw, "%s\t%s\ttotal:%d\tunread:%d\t%s\n",
			mb.Name, mb.ID, mb.TotalEmails, mb.UnreadEmails, role)
	}
	return tw.Flush()
}

func (f *TextFormatter) formatEmailList(w io.Writer, result types.EmailListResult) error {
	fmt.Fprintf(w, "Total: %d (showing %d from offset %d)\n\n", result.Total, len(result.Emails), result.Offset)

	// First pass: build display strings with truncation and track max column widths.
	type displayRow struct {
		unread  string
		from    string
		subject string
		date    string
	}

	rows := make([]displayRow, len(result.Emails))
	maxFrom := 0
	maxSubject := 0

	for i, e := range result.Emails {
		unread := " "
		if e.IsUnread {
			unread = "*"
		}
		from := ""
		if len(e.From) > 0 {
			from = truncate(formatAddr(e.From[0]), maxFromWidth)
		}
		subject := truncate(e.Subject, maxSubjectWidth)

		rows[i] = displayRow{unread, from, subject, e.ReceivedAt.Format("2006-01-02 15:04")}

		fromWidth := utf8.RuneCountInString(from)
		if fromWidth > maxFrom {
			maxFrom = fromWidth
		}
		subjectWidth := utf8.RuneCountInString(subject)
		if subjectWidth > maxSubject {
			maxSubject = subjectWidth
		}
	}

	// Second pass: print with computed widths for aligned columns.
	fmtStr := fmt.Sprintf("%%s %%-%ds  %%-%ds  %%s\n", maxFrom, maxSubject)
	for i, r := range rows {
		fmt.Fprintf(w, fmtStr, r.unread, r.from, r.subject, r.date)
		fmt.Fprintf(w, "  ID: %s\n", result.Emails[i].ID)
		if result.Emails[i].Snippet != "" {
			fmt.Fprintf(w, "  ...%s\n", result.Emails[i].Snippet)
		}
	}
	return nil
}

func (f *TextFormatter) formatEmailDetail(w io.Writer, e types.EmailDetail) error {
	fmt.Fprintf(w, "Subject: %s\n", e.Subject)
	fmt.Fprintf(w, "From: %s\n", formatAddrs(e.From))
	fmt.Fprintf(w, "To: %s\n", formatAddrs(e.To))
	if len(e.CC) > 0 {
		fmt.Fprintf(w, "CC: %s\n", formatAddrs(e.CC))
	}
	fmt.Fprintf(w, "Date: %s\n", e.ReceivedAt.Format("2006-01-02 15:04:05 -0700"))
	if e.ListUnsubscribe != "" {
		fmt.Fprintf(w, "List-Unsubscribe: %s\n", e.ListUnsubscribe)
	}
	if e.ListUnsubscribePost != "" {
		fmt.Fprintf(w, "List-Unsubscribe-Post: %s\n", e.ListUnsubscribePost)
	}
	fmt.Fprintf(w, "ID: %s\n", e.ID)
	fmt.Fprintln(w, strings.Repeat("-", 72))
	fmt.Fprintln(w, e.Body)
	if len(e.Attachments) > 0 {
		fmt.Fprintln(w, strings.Repeat("-", 72))
		fmt.Fprintf(w, "Attachments (%d):\n", len(e.Attachments))
		for _, a := range e.Attachments {
			fmt.Fprintf(w, "  - %s (%s, %d bytes)\n", a.Name, a.Type, a.Size)
		}
	}
	return nil
}

func (f *TextFormatter) formatThreadView(w io.Writer, tv types.ThreadView) error {
	fmt.Fprintf(w, "Thread (%d messages):\n\n", len(tv.Thread))
	for i, te := range tv.Thread {
		marker := "  "
		if te.ID == tv.Email.ID {
			marker = "> "
		}
		from := ""
		if len(te.From) > 0 {
			from = formatAddr(te.From[0])
		}
		fmt.Fprintf(w, "%s[%d] %s - %s (%s)\n", marker, i+1, from, te.Subject, te.ReceivedAt.Format("2006-01-02 15:04"))
		if te.ID != tv.Email.ID {
			fmt.Fprintf(w, "      %s\n", te.Preview)
		}
	}
	fmt.Fprintln(w)
	return f.formatEmailDetail(w, tv.Email)
}

func (f *TextFormatter) formatMoveResult(w io.Writer, r types.MoveResult) error {
	fmt.Fprintf(w, "Matched: %d, Processed: %d, Failed: %d\n", r.Matched, r.Processed, r.Failed)
	if len(r.Archived) > 0 {
		fmt.Fprintf(w, "Archived: %s\n", strings.Join(r.Archived, ", "))
	}
	if len(r.MarkedSpam) > 0 {
		fmt.Fprintf(w, "Marked as spam: %s\n", strings.Join(r.MarkedSpam, ", "))
	}
	if len(r.MarkedAsRead) > 0 {
		fmt.Fprintf(w, "Marked as read: %s\n", strings.Join(r.MarkedAsRead, ", "))
	}
	if len(r.Flagged) > 0 {
		fmt.Fprintf(w, "Flagged: %s\n", strings.Join(r.Flagged, ", "))
	}
	if len(r.Unflagged) > 0 {
		fmt.Fprintf(w, "Unflagged: %s\n", strings.Join(r.Unflagged, ", "))
	}
	if len(r.Moved) > 0 {
		fmt.Fprintf(w, "Moved: %s\n", strings.Join(r.Moved, ", "))
	}
	if r.Destination != nil {
		fmt.Fprintf(w, "Destination: %s (%s)\n", r.Destination.Name, r.Destination.ID)
	}
	if len(r.Errors) > 0 {
		fmt.Fprintf(w, "Errors:\n")
		for _, e := range r.Errors {
			fmt.Fprintf(w, "  - %s\n", e)
		}
	}
	return nil
}

func (f *TextFormatter) formatDryRunResult(w io.Writer, r types.DryRunResult) error {
	fmt.Fprintf(w, "Dry run: would %s %d email(s)\n", r.Operation, r.Count)

	if len(r.Emails) > 0 {
		fmt.Fprintln(w)

		// Build display rows with truncation.
		type displayRow struct {
			id      string
			from    string
			subject string
			date    string
		}

		rows := make([]displayRow, len(r.Emails))
		maxID := 0
		maxFrom := 0
		maxSubject := 0

		for i, e := range r.Emails {
			from := ""
			if len(e.From) > 0 {
				from = truncate(formatAddr(e.From[0]), maxFromWidth)
			}
			subject := truncate(e.Subject, maxSubjectWidth)

			rows[i] = displayRow{e.ID, from, subject, e.ReceivedAt.Format("2006-01-02 15:04")}

			if idLen := utf8.RuneCountInString(e.ID); idLen > maxID {
				maxID = idLen
			}
			if fromLen := utf8.RuneCountInString(from); fromLen > maxFrom {
				maxFrom = fromLen
			}
			if subjLen := utf8.RuneCountInString(subject); subjLen > maxSubject {
				maxSubject = subjLen
			}
		}

		fmtStr := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds  %%s\n", maxID, maxFrom, maxSubject)
		for _, row := range rows {
			fmt.Fprintf(w, fmtStr, row.id, row.from, row.subject, row.date)
		}
	}

	if r.Destination != nil {
		fmt.Fprintf(w, "\nDestination: %s (%s)\n", r.Destination.Name, r.Destination.ID)
	}

	if len(r.NotFound) > 0 {
		fmt.Fprintf(w, "\nNot found: %s\n", strings.Join(r.NotFound, ", "))
	}

	return nil
}

func formatAddr(a types.Address) string {
	if a.Name != "" {
		return fmt.Sprintf("%s <%s>", a.Name, a.Email)
	}
	return a.Email
}

func formatAddrs(addrs []types.Address) string {
	parts := make([]string, len(addrs))
	for i, a := range addrs {
		parts[i] = formatAddr(a)
	}
	return strings.Join(parts, ", ")
}

// truncate shortens s to maxLen characters, replacing the end with "..."
// if truncation is needed. If maxLen < 4, it returns s unchanged.
func truncate(s string, maxLen int) string {
	if maxLen < 4 || utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}
