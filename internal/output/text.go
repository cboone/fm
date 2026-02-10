package output

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/cboone/jm/internal/types"
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
	for _, mb := range mailboxes {
		role := ""
		if mb.Role != "" {
			role = fmt.Sprintf(" [%s]", mb.Role)
		}
		fmt.Fprintf(w, "%-40s %s  total:%-6d unread:%-6d %s\n",
			mb.Name, mb.ID, mb.TotalEmails, mb.UnreadEmails, role)
	}
	return nil
}

func (f *TextFormatter) formatEmailList(w io.Writer, result types.EmailListResult) error {
	fmt.Fprintf(w, "Total: %d (showing %d from offset %d)\n\n", result.Total, len(result.Emails), result.Offset)
	for _, e := range result.Emails {
		unread := " "
		if e.IsUnread {
			unread = "*"
		}
		from := ""
		if len(e.From) > 0 {
			from = formatAddr(e.From[0])
		}
		fmt.Fprintf(w, "%s %-30s  %-50s  %s\n", unread, from, e.Subject, e.ReceivedAt.Format("2006-01-02 15:04"))
		fmt.Fprintf(w, "  ID: %s\n", e.ID)
		if e.Snippet != "" {
			fmt.Fprintf(w, "  ...%s\n", e.Snippet)
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
	if len(r.Archived) > 0 {
		fmt.Fprintf(w, "Archived: %s\n", strings.Join(r.Archived, ", "))
	}
	if len(r.MarkedSpam) > 0 {
		fmt.Fprintf(w, "Marked as spam: %s\n", strings.Join(r.MarkedSpam, ", "))
	}
	if len(r.MarkedAsRead) > 0 {
		fmt.Fprintf(w, "Marked as read: %s\n", strings.Join(r.MarkedAsRead, ", "))
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
