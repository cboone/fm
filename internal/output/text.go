package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/cboone/fm/internal/types"
	"github.com/mattn/go-runewidth"
)

const (
	// maxFromWidth is the maximum display column width for the sender column in email lists.
	maxFromWidth = 40
	// maxSubjectWidth is the maximum display column width for the subject column in email lists.
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
	case types.StatsResult:
		return f.formatStats(w, val)
	case types.SummaryResult:
		return f.formatSummary(w, val)
	case types.DryRunResult:
		return f.formatDryRunResult(w, val)
	case types.DraftResult:
		return f.formatDraftResult(w, val)
	case types.SieveScriptListResult:
		return f.formatSieveScriptList(w, val)
	case types.SieveScriptDetail:
		return f.formatSieveScriptDetail(w, val)
	case types.SieveCreateResult:
		return f.formatSieveCreateResult(w, val)
	case types.SieveDeleteResult:
		return f.formatSieveDeleteResult(w, val)
	case types.SieveActivateResult:
		return f.formatSieveActivateResult(w, val)
	case types.SieveValidateResult:
		return f.formatSieveValidateResult(w, val)
	case types.SieveDryRunResult:
		return f.formatSieveDryRunResult(w, val)
	default:
		// Fall back to JSON formatter for unknown types.
		return (&JSONFormatter{}).Format(w, v)
	}
}

func (f *TextFormatter) FormatError(w io.Writer, code string, message string, hint string) error {
	_, _ = fmt.Fprintf(w, "Error [%s]: %s\n", code, message)
	if hint != "" {
		_, _ = fmt.Fprintf(w, "Hint: %s\n", hint)
	}
	return nil
}

func (f *TextFormatter) formatSession(w io.Writer, s types.SessionInfo) error {
	_, _ = fmt.Fprintf(w, "Username: %s\n", s.Username)
	_, _ = fmt.Fprintf(w, "Capabilities: %s\n", strings.Join(s.Capabilities, ", "))
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
		_, _ = fmt.Fprintf(w, "Account: %s - %s%s\n", id, acct.Name, personal)
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
		_, _ = fmt.Fprintf(tw, "%s\t%s\ttotal:%d\tunread:%d\t%s\n",
			mb.Name, mb.ID, mb.TotalEmails, mb.UnreadEmails, role)
	}
	return tw.Flush()
}

func (f *TextFormatter) formatEmailList(w io.Writer, result types.EmailListResult) error {
	_, _ = fmt.Fprintf(w, "Total: %d (showing %d from offset %d)\n\n", result.Total, len(result.Emails), result.Offset)

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

		fromWidth := runewidth.StringWidth(from)
		if fromWidth > maxFrom {
			maxFrom = fromWidth
		}
		subjectWidth := runewidth.StringWidth(subject)
		if subjectWidth > maxSubject {
			maxSubject = subjectWidth
		}
	}

	// Second pass: print with computed widths for aligned columns.
	for i, r := range rows {
		_, _ = fmt.Fprintf(w, "%s %s  %s  %s\n", r.unread,
			runewidth.FillRight(r.from, maxFrom),
			runewidth.FillRight(r.subject, maxSubject),
			r.date)
		_, _ = fmt.Fprintf(w, "  ID: %s\n", result.Emails[i].ID)
		if result.Emails[i].Snippet != "" {
			_, _ = fmt.Fprintf(w, "  ...%s\n", result.Emails[i].Snippet)
		}
	}
	return nil
}

func (f *TextFormatter) formatEmailDetail(w io.Writer, e types.EmailDetail) error {
	_, _ = fmt.Fprintf(w, "Subject: %s\n", e.Subject)
	_, _ = fmt.Fprintf(w, "From: %s\n", formatAddrs(e.From))
	_, _ = fmt.Fprintf(w, "To: %s\n", formatAddrs(e.To))
	if len(e.CC) > 0 {
		_, _ = fmt.Fprintf(w, "CC: %s\n", formatAddrs(e.CC))
	}
	_, _ = fmt.Fprintf(w, "Date: %s\n", e.ReceivedAt.Format("2006-01-02 15:04:05 -0700"))
	if e.ListUnsubscribe != "" {
		_, _ = fmt.Fprintf(w, "List-Unsubscribe: %s\n", e.ListUnsubscribe)
	}
	if e.ListUnsubscribePost != "" {
		_, _ = fmt.Fprintf(w, "List-Unsubscribe-Post: %s\n", e.ListUnsubscribePost)
	}
	_, _ = fmt.Fprintf(w, "ID: %s\n", e.ID)
	_, _ = fmt.Fprintln(w, strings.Repeat("-", 72))
	_, _ = fmt.Fprintln(w, e.Body)
	if len(e.Attachments) > 0 {
		_, _ = fmt.Fprintln(w, strings.Repeat("-", 72))
		_, _ = fmt.Fprintf(w, "Attachments (%d):\n", len(e.Attachments))
		for _, a := range e.Attachments {
			_, _ = fmt.Fprintf(w, "  - %s (%s, %d bytes)\n", a.Name, a.Type, a.Size)
		}
	}
	return nil
}

func (f *TextFormatter) formatThreadView(w io.Writer, tv types.ThreadView) error {
	_, _ = fmt.Fprintf(w, "Thread (%d messages):\n\n", len(tv.Thread))
	for i, te := range tv.Thread {
		marker := "  "
		if te.ID == tv.Email.ID {
			marker = "> "
		}
		from := ""
		if len(te.From) > 0 {
			from = formatAddr(te.From[0])
		}
		_, _ = fmt.Fprintf(w, "%s[%d] %s - %s (%s)\n", marker, i+1, from, te.Subject, te.ReceivedAt.Format("2006-01-02 15:04"))
		if te.ID != tv.Email.ID {
			_, _ = fmt.Fprintf(w, "      %s\n", te.Preview)
		}
	}
	_, _ = fmt.Fprintln(w)
	return f.formatEmailDetail(w, tv.Email)
}

func (f *TextFormatter) formatMoveResult(w io.Writer, r types.MoveResult) error {
	_, _ = fmt.Fprintf(w, "Matched: %d, Processed: %d, Failed: %d\n", r.Matched, r.Processed, r.Failed)
	if len(r.Archived) > 0 {
		_, _ = fmt.Fprintf(w, "Archived: %s\n", strings.Join(r.Archived, ", "))
	}
	if len(r.MarkedSpam) > 0 {
		_, _ = fmt.Fprintf(w, "Marked as spam: %s\n", strings.Join(r.MarkedSpam, ", "))
	}
	if len(r.MarkedAsRead) > 0 {
		_, _ = fmt.Fprintf(w, "Marked as read: %s\n", strings.Join(r.MarkedAsRead, ", "))
	}
	if len(r.Flagged) > 0 {
		_, _ = fmt.Fprintf(w, "Flagged: %s\n", strings.Join(r.Flagged, ", "))
	}
	if len(r.Unflagged) > 0 {
		_, _ = fmt.Fprintf(w, "Unflagged: %s\n", strings.Join(r.Unflagged, ", "))
	}
	if len(r.Moved) > 0 {
		_, _ = fmt.Fprintf(w, "Moved: %s\n", strings.Join(r.Moved, ", "))
	}
	if r.Destination != nil {
		_, _ = fmt.Fprintf(w, "Destination: %s (%s)\n", r.Destination.Name, r.Destination.ID)
	}
	if len(r.Errors) > 0 {
		_, _ = fmt.Fprintf(w, "Errors:\n")
		for _, e := range r.Errors {
			_, _ = fmt.Fprintf(w, "  - %s\n", e)
		}
	}
	return nil
}

func (f *TextFormatter) formatDryRunResult(w io.Writer, r types.DryRunResult) error {
	_, _ = fmt.Fprintf(w, "Dry run: would %s %d email(s)\n", r.Operation, r.Count)

	if len(r.Emails) > 0 {
		_, _ = fmt.Fprintln(w)

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

			if idLen := runewidth.StringWidth(e.ID); idLen > maxID {
				maxID = idLen
			}
			if fromLen := runewidth.StringWidth(from); fromLen > maxFrom {
				maxFrom = fromLen
			}
			if subjLen := runewidth.StringWidth(subject); subjLen > maxSubject {
				maxSubject = subjLen
			}
		}

		for _, row := range rows {
			_, _ = fmt.Fprintf(w, "  %s  %s  %s  %s\n",
				runewidth.FillRight(row.id, maxID),
				runewidth.FillRight(row.from, maxFrom),
				runewidth.FillRight(row.subject, maxSubject),
				row.date)
		}
	}

	if r.Destination != nil {
		_, _ = fmt.Fprintf(w, "\nDestination: %s (%s)\n", r.Destination.Name, r.Destination.ID)
	}

	if len(r.NotFound) > 0 {
		_, _ = fmt.Fprintf(w, "\nNot found: %s\n", strings.Join(r.NotFound, ", "))
	}

	return nil
}

func (f *TextFormatter) formatStats(w io.Writer, r types.StatsResult) error {
	_, _ = fmt.Fprintf(w, "Total: %d emails from %d senders\n", r.Total, len(r.Senders))

	if len(r.Senders) == 0 {
		return nil
	}

	_, _ = fmt.Fprintln(w)

	// Compute max count width for right-alignment.
	maxCount := 0
	for _, s := range r.Senders {
		if s.Count > maxCount {
			maxCount = s.Count
		}
	}
	countWidth := len(fmt.Sprintf("%d", maxCount))

	for _, s := range r.Senders {
		if s.Name != "" {
			_, _ = fmt.Fprintf(w, "%*d  %s  %s\n", countWidth, s.Count, s.Email, s.Name)
		} else {
			_, _ = fmt.Fprintf(w, "%*d  %s\n", countWidth, s.Count, s.Email)
		}
		for _, subj := range s.Subjects {
			_, _ = fmt.Fprintf(w, "%*s  %s\n", countWidth, "", "  "+subj)
		}
	}
	return nil
}

func (f *TextFormatter) formatSummary(w io.Writer, r types.SummaryResult) error {
	_, _ = fmt.Fprintf(w, "Total: %d emails (%d unread)\n", r.Total, r.Unread)

	if len(r.TopSenders) > 0 {
		_, _ = fmt.Fprintln(w, "\nTop senders:")
		maxCount := 0
		for _, s := range r.TopSenders {
			if s.Count > maxCount {
				maxCount = s.Count
			}
		}
		countWidth := len(fmt.Sprintf("%d", maxCount))
		for _, s := range r.TopSenders {
			if s.Name != "" {
				_, _ = fmt.Fprintf(w, "%*d  %s  %s\n", countWidth, s.Count, s.Email, s.Name)
			} else {
				_, _ = fmt.Fprintf(w, "%*d  %s\n", countWidth, s.Count, s.Email)
			}
			for _, subj := range s.Subjects {
				_, _ = fmt.Fprintf(w, "%*s  %s\n", countWidth, "", "  "+subj)
			}
		}
	}

	if len(r.TopDomains) > 0 {
		_, _ = fmt.Fprintln(w, "\nTop domains:")
		maxCount := 0
		for _, d := range r.TopDomains {
			if d.Count > maxCount {
				maxCount = d.Count
			}
		}
		countWidth := len(fmt.Sprintf("%d", maxCount))
		for _, d := range r.TopDomains {
			_, _ = fmt.Fprintf(w, "%*d  %s\n", countWidth, d.Count, d.Domain)
		}
	}

	if len(r.Newsletters) > 0 {
		_, _ = fmt.Fprintln(w, "\nNewsletters / mailing lists:")
		maxCount := 0
		for _, s := range r.Newsletters {
			if s.Count > maxCount {
				maxCount = s.Count
			}
		}
		countWidth := len(fmt.Sprintf("%d", maxCount))
		for _, s := range r.Newsletters {
			if s.Name != "" {
				_, _ = fmt.Fprintf(w, "%*d  %s  %s\n", countWidth, s.Count, s.Email, s.Name)
			} else {
				_, _ = fmt.Fprintf(w, "%*d  %s\n", countWidth, s.Count, s.Email)
			}
			for _, subj := range s.Subjects {
				_, _ = fmt.Fprintf(w, "%*s  %s\n", countWidth, "", "  "+subj)
			}
		}
	}

	return nil
}

func (f *TextFormatter) formatDraftResult(w io.Writer, r types.DraftResult) error {
	_, _ = fmt.Fprintf(w, "Draft created: %s\n", r.ID)
	_, _ = fmt.Fprintf(w, "Mode: %s\n", r.Mode)
	if len(r.From) > 0 {
		_, _ = fmt.Fprintf(w, "From: %s\n", formatAddrs(r.From))
	}
	_, _ = fmt.Fprintf(w, "To: %s\n", formatAddrs(r.To))
	if len(r.CC) > 0 {
		_, _ = fmt.Fprintf(w, "CC: %s\n", formatAddrs(r.CC))
	}
	_, _ = fmt.Fprintf(w, "Subject: %s\n", r.Subject)
	if r.Mailbox != nil {
		_, _ = fmt.Fprintf(w, "Mailbox: %s (%s)\n", r.Mailbox.Name, r.Mailbox.ID)
	}
	if r.InReplyTo != "" {
		_, _ = fmt.Fprintf(w, "In-Reply-To: %s\n", r.InReplyTo)
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

func (f *TextFormatter) formatSieveScriptList(w io.Writer, r types.SieveScriptListResult) error {
	_, _ = fmt.Fprintf(w, "Total: %d script(s)\n", r.Total)
	if r.Total == 0 {
		return nil
	}

	_, _ = fmt.Fprintln(w)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "ID\tName\tActive\n")
	for _, s := range r.Scripts {
		active := ""
		if s.IsActive {
			active = "*"
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", s.ID, s.Name, active)
	}
	return tw.Flush()
}

func (f *TextFormatter) formatSieveScriptDetail(w io.Writer, s types.SieveScriptDetail) error {
	active := "no"
	if s.IsActive {
		active = "yes"
	}
	_, _ = fmt.Fprintf(w, "ID: %s\n", s.ID)
	_, _ = fmt.Fprintf(w, "Name: %s\n", s.Name)
	_, _ = fmt.Fprintf(w, "Active: %s\n", active)
	_, _ = fmt.Fprintf(w, "Blob: %s\n", s.BlobID)
	_, _ = fmt.Fprintln(w, strings.Repeat("-", 72))
	_, _ = fmt.Fprint(w, s.Content)
	return nil
}

func (f *TextFormatter) formatSieveCreateResult(w io.Writer, r types.SieveCreateResult) error {
	active := "no"
	if r.IsActive {
		active = "yes"
	}
	_, _ = fmt.Fprintf(w, "Created sieve script: %s\n", r.ID)
	_, _ = fmt.Fprintf(w, "Name: %s\n", r.Name)
	_, _ = fmt.Fprintf(w, "Active: %s\n", active)
	return nil
}

func (f *TextFormatter) formatSieveDeleteResult(w io.Writer, r types.SieveDeleteResult) error {
	_, _ = fmt.Fprintf(w, "Deleted sieve script: %s\n", r.ID)
	return nil
}

func (f *TextFormatter) formatSieveActivateResult(w io.Writer, r types.SieveActivateResult) error {
	if r.IsActive {
		_, _ = fmt.Fprintf(w, "Activated sieve script: %s\n", r.ID)
	} else {
		_, _ = fmt.Fprintln(w, "Deactivated active sieve script")
	}
	return nil
}

func (f *TextFormatter) formatSieveValidateResult(w io.Writer, r types.SieveValidateResult) error {
	if r.Valid {
		_, _ = fmt.Fprintln(w, "Valid: yes")
	} else {
		_, _ = fmt.Fprintln(w, "Valid: no")
		_, _ = fmt.Fprintf(w, "Error: %s\n", r.Error)
	}
	return nil
}

func (f *TextFormatter) formatSieveDryRunResult(w io.Writer, r types.SieveDryRunResult) error {
	_, _ = fmt.Fprintf(w, "Dry run: would %s", r.Operation)
	if r.Script != "" {
		_, _ = fmt.Fprintf(w, " script %q", r.Script)
	}
	_, _ = fmt.Fprintln(w)

	if r.Content != "" {
		_, _ = fmt.Fprintln(w, strings.Repeat("-", 72))
		_, _ = fmt.Fprint(w, r.Content)
	}

	if r.Valid != nil {
		if *r.Valid {
			_, _ = fmt.Fprintln(w, "\nValidation: passed")
		} else {
			_, _ = fmt.Fprintln(w, "\nValidation: failed")
		}
	}

	return nil
}

// truncate shortens s to maxWidth display columns, replacing the end with
// "..." if truncation is needed. If maxWidth < 4, it returns s unchanged.
func truncate(s string, maxWidth int) string {
	if maxWidth < 4 || runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	return runewidth.Truncate(s, maxWidth, "...")
}
