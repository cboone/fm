package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cboone/jm/internal/types"
)

func TestTextFormatter_Session(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	s := types.SessionInfo{
		Username: "user@fastmail.com",
		Accounts: map[string]types.AccountInfo{
			"abc123": {Name: "user@fastmail.com", IsPersonal: true},
		},
		Capabilities: []string{"urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"},
	}

	if err := f.Format(&buf, s); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "user@fastmail.com") {
		t.Errorf("expected username in output, got: %s", out)
	}
	if !strings.Contains(out, "(personal)") {
		t.Errorf("expected '(personal)' in output, got: %s", out)
	}
	if !strings.Contains(out, "urn:ietf:params:jmap:core") {
		t.Errorf("expected capability in output, got: %s", out)
	}
}

func TestTextFormatter_Mailboxes(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	mailboxes := []types.MailboxInfo{
		{ID: "mb1", Name: "Inbox", Role: "inbox", TotalEmails: 100, UnreadEmails: 5},
		{ID: "mb2", Name: "Archive", Role: "archive", TotalEmails: 5000, UnreadEmails: 0},
	}

	if err := f.Format(&buf, mailboxes); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Inbox") {
		t.Errorf("expected 'Inbox' in output, got: %s", out)
	}
	if !strings.Contains(out, "Archive") {
		t.Errorf("expected 'Archive' in output, got: %s", out)
	}
	if !strings.Contains(out, "[inbox]") {
		t.Errorf("expected '[inbox]' role in output, got: %s", out)
	}
}

func TestTextFormatter_EmailList(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	result := types.EmailListResult{
		Total:  42,
		Offset: 0,
		Emails: []types.EmailSummary{
			{
				ID:         "M1",
				From:       []types.Address{{Name: "Alice", Email: "alice@test.com"}},
				To:         []types.Address{{Name: "Bob", Email: "bob@test.com"}},
				Subject:    "Meeting tomorrow",
				ReceivedAt: now,
				IsUnread:   true,
			},
			{
				ID:         "M2",
				From:       []types.Address{{Name: "", Email: "noreply@test.com"}},
				To:         []types.Address{{Name: "Bob", Email: "bob@test.com"}},
				Subject:    "Read message",
				ReceivedAt: now,
				IsUnread:   false,
			},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Total: 42") {
		t.Errorf("expected 'Total: 42' in output, got: %s", out)
	}
	if !strings.Contains(out, "Meeting tomorrow") {
		t.Errorf("expected subject in output, got: %s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected sender name in output, got: %s", out)
	}
	// Unread marker: '*' for unread, ' ' for read.
	if !strings.Contains(out, "* ") {
		t.Errorf("expected unread marker '*' in output, got: %s", out)
	}
	if !strings.Contains(out, "ID: M1") {
		t.Errorf("expected email ID in output, got: %s", out)
	}
}

func TestTextFormatter_EmailListWithSnippet(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	result := types.EmailListResult{
		Total: 1,
		Emails: []types.EmailSummary{
			{
				ID:         "M1",
				From:       []types.Address{{Email: "test@test.com"}},
				To:         []types.Address{{Email: "me@test.com"}},
				Subject:    "Test",
				ReceivedAt: now,
				Snippet:    "matching text here",
			},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "matching text here") {
		t.Errorf("expected snippet in output, got: %s", out)
	}
}

func TestTextFormatter_EmailDetail(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	detail := types.EmailDetail{
		ID:         "M1",
		From:       []types.Address{{Name: "Alice", Email: "alice@test.com"}},
		To:         []types.Address{{Name: "Bob", Email: "bob@test.com"}},
		CC:         []types.Address{{Name: "Charlie", Email: "charlie@test.com"}},
		Subject:    "Important Meeting",
		ReceivedAt: now,
		Body:       "Hello Bob,\n\nPlease review the attached document.",
		Attachments: []types.Attachment{
			{Name: "report.pdf", Type: "application/pdf", Size: 24000},
		},
	}

	if err := f.Format(&buf, detail); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Subject: Important Meeting") {
		t.Errorf("expected subject header, got: %s", out)
	}
	if !strings.Contains(out, "From: Alice <alice@test.com>") {
		t.Errorf("expected from header, got: %s", out)
	}
	if !strings.Contains(out, "CC: Charlie <charlie@test.com>") {
		t.Errorf("expected CC header, got: %s", out)
	}
	if !strings.Contains(out, "Hello Bob") {
		t.Errorf("expected body content, got: %s", out)
	}
	if !strings.Contains(out, "report.pdf") {
		t.Errorf("expected attachment name, got: %s", out)
	}
	if !strings.Contains(out, "Attachments (1)") {
		t.Errorf("expected attachment count, got: %s", out)
	}
}

func TestTextFormatter_EmailDetailNoCC(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	detail := types.EmailDetail{
		ID:          "M1",
		From:        []types.Address{{Email: "alice@test.com"}},
		To:          []types.Address{{Email: "bob@test.com"}},
		CC:          []types.Address{},
		Subject:     "Test",
		ReceivedAt:  now,
		Body:        "body",
		Attachments: []types.Attachment{},
	}

	if err := f.Format(&buf, detail); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, "CC:") {
		t.Errorf("expected no CC line for empty CC, got: %s", out)
	}
	if strings.Contains(out, "Attachments") {
		t.Errorf("expected no Attachments section for empty attachments, got: %s", out)
	}
}

func TestTextFormatter_ThreadView(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	earlier := time.Date(2026, 2, 3, 15, 0, 0, 0, time.UTC)

	tv := types.ThreadView{
		Email: types.EmailDetail{
			ID:          "M2",
			From:        []types.Address{{Name: "Bob", Email: "bob@test.com"}},
			To:          []types.Address{{Name: "Alice", Email: "alice@test.com"}},
			Subject:     "Re: Meeting",
			ReceivedAt:  now,
			Body:        "Sounds good!",
			Attachments: []types.Attachment{},
		},
		Thread: []types.ThreadEmail{
			{
				ID:         "M1",
				From:       []types.Address{{Name: "Alice", Email: "alice@test.com"}},
				To:         []types.Address{{Name: "Bob", Email: "bob@test.com"}},
				Subject:    "Meeting",
				ReceivedAt: earlier,
				Preview:    "Can we meet tomorrow?",
			},
			{
				ID:         "M2",
				From:       []types.Address{{Name: "Bob", Email: "bob@test.com"}},
				To:         []types.Address{{Name: "Alice", Email: "alice@test.com"}},
				Subject:    "Re: Meeting",
				ReceivedAt: now,
			},
		},
	}

	if err := f.Format(&buf, tv); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Thread (2 messages)") {
		t.Errorf("expected thread count, got: %s", out)
	}
	if !strings.Contains(out, "> ") {
		t.Errorf("expected '> ' marker for current email, got: %s", out)
	}
	if !strings.Contains(out, "Can we meet tomorrow?") {
		t.Errorf("expected preview for non-current thread email, got: %s", out)
	}
	if !strings.Contains(out, "Sounds good!") {
		t.Errorf("expected full body for current email, got: %s", out)
	}
}

func TestTextFormatter_MoveResult(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.MoveResult{
		Archived: []string{"M1", "M2"},
		Destination: &types.DestinationInfo{
			ID: "mb-archive", Name: "Archive",
		},
		Errors: []string{},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Archived: M1, M2") {
		t.Errorf("expected archived IDs, got: %s", out)
	}
	if !strings.Contains(out, "Archive") {
		t.Errorf("expected destination name, got: %s", out)
	}
}

func TestTextFormatter_MoveResultWithErrors(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.MoveResult{
		Moved:  []string{"M1"},
		Errors: []string{"M2: not found"},
		Destination: &types.DestinationInfo{
			ID: "mb-receipts", Name: "Receipts",
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Moved: M1") {
		t.Errorf("expected moved IDs, got: %s", out)
	}
	if !strings.Contains(out, "M2: not found") {
		t.Errorf("expected error detail, got: %s", out)
	}
}

func TestTextFormatter_MoveResultSpam(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.MoveResult{
		MarkedSpam: []string{"M1"},
		Errors:     []string{},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Marked as spam: M1") {
		t.Errorf("expected spam IDs, got: %s", out)
	}
}

func TestTextFormatter_MoveResultMarkedAsRead(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.MoveResult{
		MarkedAsRead: []string{"M1", "M2"},
		Errors:       []string{},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Marked as read: M1, M2") {
		t.Errorf("expected marked-as-read IDs, got: %s", out)
	}
}

func TestTextFormatter_ErrorWithHint(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	if err := f.FormatError(&buf, "auth_failed", "bad token", "set JMAP_TOKEN"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Hint: set JMAP_TOKEN") {
		t.Errorf("expected hint in output, got: %s", out)
	}
}

func TestTextFormatter_ErrorWithoutHint(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	if err := f.FormatError(&buf, "jmap_error", "server error", ""); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, "Hint:") {
		t.Errorf("expected no hint line when hint is empty, got: %s", out)
	}
}

// --- formatAddr / formatAddrs tests ---

func TestFormatAddr_WithName(t *testing.T) {
	a := types.Address{Name: "Alice", Email: "alice@test.com"}
	result := formatAddr(a)
	if result != "Alice <alice@test.com>" {
		t.Errorf("expected 'Alice <alice@test.com>', got: %s", result)
	}
}

func TestFormatAddr_WithoutName(t *testing.T) {
	a := types.Address{Name: "", Email: "alice@test.com"}
	result := formatAddr(a)
	if result != "alice@test.com" {
		t.Errorf("expected 'alice@test.com', got: %s", result)
	}
}

func TestFormatAddrs_Multiple(t *testing.T) {
	addrs := []types.Address{
		{Name: "Alice", Email: "alice@test.com"},
		{Name: "", Email: "bob@test.com"},
	}
	result := formatAddrs(addrs)
	if result != "Alice <alice@test.com>, bob@test.com" {
		t.Errorf("expected formatted addresses, got: %s", result)
	}
}

func TestFormatAddrs_Empty(t *testing.T) {
	result := formatAddrs([]types.Address{})
	if result != "" {
		t.Errorf("expected empty string, got: %s", result)
	}
}

// --- New factory test ---

func TestNew_JSON(t *testing.T) {
	f := New("json")
	if _, ok := f.(*JSONFormatter); !ok {
		t.Errorf("expected JSONFormatter, got %T", f)
	}
}

func TestNew_Text(t *testing.T) {
	f := New("text")
	if _, ok := f.(*TextFormatter); !ok {
		t.Errorf("expected TextFormatter, got %T", f)
	}
}

func TestNew_Default(t *testing.T) {
	f := New("unknown")
	if _, ok := f.(*JSONFormatter); !ok {
		t.Errorf("expected JSONFormatter as default, got %T", f)
	}
}

func TestNew_Empty(t *testing.T) {
	f := New("")
	if _, ok := f.(*JSONFormatter); !ok {
		t.Errorf("expected JSONFormatter for empty string, got %T", f)
	}
}
