package types

import "time"

// Address is a simplified email address for output.
type Address struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Attachment is a simplified attachment descriptor for output.
type Attachment struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size uint64 `json:"size"`
}

// MailboxInfo is a simplified mailbox for output.
type MailboxInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Role         string `json:"role,omitempty"`
	TotalEmails  uint64 `json:"total_emails"`
	UnreadEmails uint64 `json:"unread_emails"`
	ParentID     string `json:"parent_id,omitempty"`
}

// EmailSummary is a brief view of an email for list/search results.
type EmailSummary struct {
	ID         string    `json:"id"`
	ThreadID   string    `json:"thread_id"`
	From       []Address `json:"from"`
	To         []Address `json:"to"`
	Subject    string    `json:"subject"`
	ReceivedAt time.Time `json:"received_at"`
	Size       uint64    `json:"size"`
	IsUnread   bool      `json:"is_unread"`
	IsFlagged  bool      `json:"is_flagged"`
	Preview    string    `json:"preview"`
	Snippet    string    `json:"snippet,omitempty"`
}

// EmailListResult wraps a paginated email list.
type EmailListResult struct {
	Total  uint64         `json:"total"`
	Offset int64          `json:"offset"`
	Emails []EmailSummary `json:"emails"`
}

// EmailDetail is a full view of a single email.
type EmailDetail struct {
	ID          string       `json:"id"`
	ThreadID    string       `json:"thread_id"`
	From        []Address    `json:"from"`
	To          []Address    `json:"to"`
	CC          []Address    `json:"cc"`
	BCC         []Address    `json:"bcc,omitempty"`
	ReplyTo     []Address    `json:"reply_to,omitempty"`
	Subject     string       `json:"subject"`
	SentAt      *time.Time   `json:"sent_at,omitempty"`
	ReceivedAt  time.Time    `json:"received_at"`
	IsUnread    bool         `json:"is_unread"`
	IsFlagged   bool         `json:"is_flagged"`
	Body        string       `json:"body"`
	Attachments []Attachment `json:"attachments"`
	Headers     []Header     `json:"headers,omitempty"`
}

// Header is a raw email header.
type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ThreadEmail is a condensed view of an email within a thread.
type ThreadEmail struct {
	ID         string    `json:"id"`
	From       []Address `json:"from"`
	To         []Address `json:"to"`
	Subject    string    `json:"subject"`
	ReceivedAt time.Time `json:"received_at"`
	Preview    string    `json:"preview"`
	IsUnread   bool      `json:"is_unread"`
}

// ThreadView wraps a full email with surrounding thread context.
type ThreadView struct {
	Email  EmailDetail   `json:"email"`
	Thread []ThreadEmail `json:"thread"`
}

// SessionInfo is a simplified session for output.
type SessionInfo struct {
	Username     string                 `json:"username"`
	Accounts     map[string]AccountInfo `json:"accounts"`
	Capabilities []string               `json:"capabilities"`
}

// AccountInfo is a simplified account for output.
type AccountInfo struct {
	Name       string `json:"name"`
	IsPersonal bool   `json:"is_personal"`
}

// MoveResult reports the outcome of a move/archive/spam/mark-read/flag/unflag operation.
type MoveResult struct {
	Moved        []string         `json:"moved,omitempty"`
	Archived     []string         `json:"archived,omitempty"`
	MarkedSpam   []string         `json:"marked_as_spam,omitempty"`
	MarkedAsRead []string         `json:"marked_as_read,omitempty"`
	Flagged      []string         `json:"flagged,omitempty"`
	Unflagged    []string         `json:"unflagged,omitempty"`
	Destination  *DestinationInfo `json:"destination,omitempty"`
	Errors       []string         `json:"errors"`
}

// DestinationInfo identifies the target mailbox of a move.
type DestinationInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AppError is a structured error for JSON output.
type AppError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}
