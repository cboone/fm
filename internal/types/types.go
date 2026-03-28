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
	CC         []Address `json:"cc,omitempty"`
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
	ID                  string       `json:"id"`
	ThreadID            string       `json:"thread_id"`
	From                []Address    `json:"from"`
	To                  []Address    `json:"to"`
	CC                  []Address    `json:"cc"`
	BCC                 []Address    `json:"bcc,omitempty"`
	ReplyTo             []Address    `json:"reply_to,omitempty"`
	Subject             string       `json:"subject"`
	SentAt              *time.Time   `json:"sent_at,omitempty"`
	ReceivedAt          time.Time    `json:"received_at"`
	IsUnread            bool         `json:"is_unread"`
	IsFlagged           bool         `json:"is_flagged"`
	Body                string       `json:"body"`
	ListUnsubscribe     string       `json:"list_unsubscribe,omitempty"`
	ListUnsubscribePost string       `json:"list_unsubscribe_post,omitempty"`
	Attachments         []Attachment `json:"attachments"`
	Headers             []Header     `json:"headers,omitempty"`
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
	Matched      int              `json:"matched"`
	Processed    int              `json:"processed"`
	Failed       int              `json:"failed"`
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

// DryRunResult previews the emails that would be affected by a mutating command.
type DryRunResult struct {
	Operation   string           `json:"operation"`
	Count       int              `json:"count"`
	Emails      []EmailSummary   `json:"emails"`
	NotFound    []string         `json:"not_found,omitempty"`
	Destination *DestinationInfo `json:"destination,omitempty"`
}

// SenderStat is an aggregated count for a single sender address.
type SenderStat struct {
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Count    int      `json:"count"`
	Subjects []string `json:"subjects,omitempty"`
}

// StatsResult wraps an aggregated sender distribution.
type StatsResult struct {
	Total   uint64       `json:"total"`
	Senders []SenderStat `json:"senders"`
}

// DomainStat is an aggregated count for a single sender domain.
type DomainStat struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

// SummaryResult wraps a triage-oriented inbox summary with sender and domain aggregation.
type SummaryResult struct {
	Total       uint64       `json:"total"`
	Unread      uint64       `json:"unread"`
	TopSenders  []SenderStat `json:"top_senders"`
	TopDomains  []DomainStat `json:"top_domains"`
	Newsletters []SenderStat `json:"newsletters,omitempty"`
}

// DraftResult reports the outcome of a draft creation.
type DraftResult struct {
	ID        string           `json:"id"`
	Mode      string           `json:"mode"`
	Mailbox   *DestinationInfo `json:"mailbox"`
	From      []Address        `json:"from,omitempty"`
	To        []Address        `json:"to"`
	CC        []Address        `json:"cc,omitempty"`
	Subject   string           `json:"subject"`
	InReplyTo string           `json:"in_reply_to,omitempty"`
}

// SieveScriptInfo is a summary view of a sieve script for list output.
type SieveScriptInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

// SieveScriptListResult wraps a list of sieve scripts.
type SieveScriptListResult struct {
	Total   int               `json:"total"`
	Scripts []SieveScriptInfo `json:"scripts"`
}

// SieveScriptDetail is a full view of a sieve script including content.
type SieveScriptDetail struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BlobID   string `json:"blob_id"`
	IsActive bool   `json:"is_active"`
	Content  string `json:"content"`
}

// SieveCreateResult reports the outcome of sieve script creation.
type SieveCreateResult struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BlobID   string `json:"blob_id"`
	IsActive bool   `json:"is_active"`
	Content  string `json:"content"`
}

// SieveDeleteResult reports the outcome of sieve script deletion.
type SieveDeleteResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SieveActivateResult reports the outcome of activation or deactivation.
type SieveActivateResult struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	IsActive bool   `json:"is_active"`
}

// SieveValidateResult reports the outcome of script validation.
type SieveValidateResult struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Content string `json:"content"`
}

// SieveDryRunResult previews a sieve mutation without executing it.
type SieveDryRunResult struct {
	Operation string `json:"operation"`
	Script    string `json:"script,omitempty"`
	Content   string `json:"content,omitempty"`
	Valid     *bool  `json:"valid,omitempty"`
}

// UnsubscribeResult reports the outcome of an unsubscribe inspection or draft creation.
type UnsubscribeResult struct {
	EmailID   string `json:"email_id"`
	Mechanism string `json:"mechanism"`
	Mailto    string `json:"mailto,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Body      string `json:"body,omitempty"`
	URL       string `json:"url,omitempty"`
	OneClick  bool   `json:"one_click"`
	DraftID   string `json:"draft_id,omitempty"`
}

// AppError is a structured error for JSON output.
type AppError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}
