package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cboone/fm/internal/types"
	"github.com/mattn/go-runewidth"
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

func TestTextFormatter_EmailDetailWithListUnsubscribe(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	detail := types.EmailDetail{
		ID:                  "M1",
		From:                []types.Address{{Email: "news@example.com"}},
		To:                  []types.Address{{Email: "user@example.com"}},
		Subject:             "Newsletter",
		ReceivedAt:          now,
		Body:                "content",
		ListUnsubscribe:     "<mailto:unsub@example.com>, <https://example.com/unsub>",
		ListUnsubscribePost: "List-Unsubscribe=One-Click",
		Attachments:         []types.Attachment{},
	}

	if err := f.Format(&buf, detail); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "List-Unsubscribe: <mailto:unsub@example.com>, <https://example.com/unsub>") {
		t.Errorf("expected List-Unsubscribe header in output, got: %s", out)
	}
	if !strings.Contains(out, "List-Unsubscribe-Post: List-Unsubscribe=One-Click") {
		t.Errorf("expected List-Unsubscribe-Post header in output, got: %s", out)
	}
}

func TestTextFormatter_EmailDetailNoListUnsubscribe(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	detail := types.EmailDetail{
		ID:          "M1",
		From:        []types.Address{{Email: "alice@test.com"}},
		To:          []types.Address{{Email: "bob@test.com"}},
		Subject:     "Test",
		ReceivedAt:  now,
		Body:        "body",
		Attachments: []types.Attachment{},
	}

	if err := f.Format(&buf, detail); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "List-Unsubscribe:") || strings.HasPrefix(line, "List-Unsubscribe-Post:") {
			t.Errorf("expected no List-Unsubscribe headers when empty, got line: %s", line)
		}
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
		Matched:   2,
		Processed: 2,
		Failed:    0,
		Archived:  []string{"M1", "M2"},
		Destination: &types.DestinationInfo{
			ID: "mb-archive", Name: "Archive",
		},
		Errors: []string{},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Matched: 2, Processed: 2, Failed: 0") {
		t.Errorf("expected summary line, got: %s", out)
	}
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
		Matched:   2,
		Processed: 2,
		Failed:    1,
		Moved:     []string{"M1"},
		Errors:    []string{"M2: not found"},
		Destination: &types.DestinationInfo{
			ID: "mb-receipts", Name: "Receipts",
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Matched: 2, Processed: 2, Failed: 1") {
		t.Errorf("expected summary line, got: %s", out)
	}
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

func TestTextFormatter_MoveResultFlagged(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.MoveResult{
		Flagged: []string{"M1", "M2"},
		Errors:  []string{},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Flagged: M1, M2") {
		t.Errorf("expected flagged IDs, got: %s", out)
	}
}

func TestTextFormatter_MoveResultUnflagged(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.MoveResult{
		Unflagged: []string{"M3"},
		Errors:    []string{},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Unflagged: M3") {
		t.Errorf("expected unflagged IDs, got: %s", out)
	}
}

func TestTextFormatter_DryRunResult_WithDestination(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 14, 10, 30, 0, 0, time.UTC)
	result := types.DryRunResult{
		Operation: "archive",
		Count:     2,
		Emails: []types.EmailSummary{
			{
				ID:         "M1",
				From:       []types.Address{{Name: "Alice", Email: "alice@example.com"}},
				Subject:    "Meeting tomorrow",
				ReceivedAt: now,
			},
			{
				ID:         "M2",
				From:       []types.Address{{Name: "Bob", Email: "bob@example.com"}},
				Subject:    "Invoice #1234",
				ReceivedAt: now,
			},
		},
		Destination: &types.DestinationInfo{
			ID:   "mb-archive-id",
			Name: "Archive",
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Dry run: would archive 2 email(s)") {
		t.Errorf("expected header line, got: %s", out)
	}
	if !strings.Contains(out, "M1") {
		t.Errorf("expected M1 in output, got: %s", out)
	}
	if !strings.Contains(out, "M2") {
		t.Errorf("expected M2 in output, got: %s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected sender name in output, got: %s", out)
	}
	if !strings.Contains(out, "Meeting tomorrow") {
		t.Errorf("expected subject in output, got: %s", out)
	}
	if !strings.Contains(out, "Destination: Archive (mb-archive-id)") {
		t.Errorf("expected destination line, got: %s", out)
	}
	if strings.Contains(out, "Not found:") {
		t.Errorf("expected no Not found line, got: %s", out)
	}
}

func TestTextFormatter_DryRunResult_NoDestination(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 14, 10, 30, 0, 0, time.UTC)
	result := types.DryRunResult{
		Operation: "mark-read",
		Count:     1,
		Emails: []types.EmailSummary{
			{
				ID:         "M1",
				From:       []types.Address{{Email: "alice@example.com"}},
				Subject:    "Test",
				ReceivedAt: now,
			},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Dry run: would mark-read 1 email(s)") {
		t.Errorf("expected header line, got: %s", out)
	}
	if strings.Contains(out, "Destination:") {
		t.Errorf("expected no destination line for mark-read, got: %s", out)
	}
}

func TestTextFormatter_DryRunResult_WithNotFound(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.DryRunResult{
		Operation: "flag",
		Count:     0,
		Emails:    []types.EmailSummary{},
		NotFound:  []string{"M4", "M5"},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Dry run: would flag 0 email(s)") {
		t.Errorf("expected header line, got: %s", out)
	}
	if !strings.Contains(out, "Not found: M4, M5") {
		t.Errorf("expected not-found line, got: %s", out)
	}
}

func TestTextFormatter_ErrorWithHint(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	if err := f.FormatError(&buf, "auth_failed", "bad token", "set FM_TOKEN"); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Hint: set FM_TOKEN") {
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

// --- formatStats tests ---

func TestTextFormatter_StatsResult_Basic(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.StatsResult{
		Total: 25,
		Senders: []types.SenderStat{
			{Email: "newsletter@example.com", Name: "Example Newsletter", Count: 15},
			{Email: "alice@example.com", Name: "Alice Smith", Count: 8},
			{Email: "bob@example.com", Name: "", Count: 2},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Total: 25 emails from 3 senders") {
		t.Errorf("expected header line, got: %s", out)
	}
	if !strings.Contains(out, "newsletter@example.com") {
		t.Errorf("expected sender email, got: %s", out)
	}
	if !strings.Contains(out, "Example Newsletter") {
		t.Errorf("expected sender name, got: %s", out)
	}
	if !strings.Contains(out, "Alice Smith") {
		t.Errorf("expected sender name, got: %s", out)
	}
	// bob has no name, should just show email.
	if !strings.Contains(out, "bob@example.com") {
		t.Errorf("expected bob email, got: %s", out)
	}
}

func TestTextFormatter_StatsResult_WithSubjects(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.StatsResult{
		Total: 10,
		Senders: []types.SenderStat{
			{
				Email:    "news@example.com",
				Name:     "News",
				Count:    5,
				Subjects: []string{"Weekly Digest #42", "Weekly Digest #41"},
			},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Weekly Digest #42") {
		t.Errorf("expected subject line, got: %s", out)
	}
	if !strings.Contains(out, "Weekly Digest #41") {
		t.Errorf("expected subject line, got: %s", out)
	}
}

func TestTextFormatter_StatsResult_Empty(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.StatsResult{
		Total:   0,
		Senders: []types.SenderStat{},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Total: 0 emails from 0 senders") {
		t.Errorf("expected empty header line, got: %s", out)
	}
}

func TestTextFormatter_StatsResult_CountAlignment(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	result := types.StatsResult{
		Total: 115,
		Senders: []types.SenderStat{
			{Email: "bulk@example.com", Count: 100},
			{Email: "rare@example.com", Count: 3},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	lines := strings.Split(out, "\n")
	// Find the two sender lines (containing "@").
	var senderLines []string
	for _, line := range lines {
		if strings.Contains(line, "@") {
			senderLines = append(senderLines, line)
		}
	}
	if len(senderLines) != 2 {
		t.Fatalf("expected 2 sender lines, got %d\nOutput:\n%s", len(senderLines), out)
	}
	// The email column should start at the same position.
	idx1 := strings.Index(senderLines[0], "bulk@")
	idx2 := strings.Index(senderLines[1], "rare@")
	if idx1 != idx2 {
		t.Errorf("email columns not aligned: line 1 at %d, line 2 at %d\nOutput:\n%s", idx1, idx2, out)
	}
}

// --- truncate tests ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{"short string unchanged", "hello", 10, "hello"},
		{"exact length unchanged", "hello", 5, "hello"},
		{"truncated with ellipsis", "hello world", 8, "hello..."},
		{"maxWidth too small returns unchanged", "hello", 3, "hello"},
		{"empty string", "", 10, ""},
		{"wide chars truncated by display width", strings.Repeat("界", 10), 10, strings.Repeat("界", 3) + "..."},
		{"wide chars fitting exactly", "界界", 4, "界界"},
		{"wide chars just over limit", "界界界", 5, "界..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxWidth)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.want)
			}
		})
	}
}

// --- alignment tests ---

func TestTextFormatter_MailboxesAlignment(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	mailboxes := []types.MailboxInfo{
		{ID: "mb1", Name: "Inbox", Role: "inbox", TotalEmails: 100, UnreadEmails: 5},
		{ID: "mb2", Name: "A Very Long Mailbox Name That Exceeds Forty Characters Easily", Role: "", TotalEmails: 50, UnreadEmails: 0},
	}

	if err := f.Format(&buf, mailboxes); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// With tabwriter, the ID column should start at the same position in both lines.
	idx1 := strings.Index(lines[0], "mb1")
	idx2 := strings.Index(lines[1], "mb2")
	if idx1 == -1 || idx2 == -1 {
		t.Fatalf("expected mailbox IDs in output for alignment check, got:\n%s", out)
	}
	if idx1 != idx2 {
		t.Errorf("ID columns not aligned: line 1 at %d, line 2 at %d\nOutput:\n%s", idx1, idx2, out)
	}
}

func TestTextFormatter_EmailListTruncation(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	longName := strings.Repeat("A", 60)
	longSubject := strings.Repeat("B", 100)
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)

	result := types.EmailListResult{
		Total: 1,
		Emails: []types.EmailSummary{
			{
				ID:         "M1",
				From:       []types.Address{{Name: longName, Email: "a@test.com"}},
				To:         []types.Address{{Email: "b@test.com"}},
				Subject:    longSubject,
				ReceivedAt: now,
			},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, longName) {
		t.Error("expected long sender name to be truncated")
	}
	if !strings.Contains(out, "...") {
		t.Error("expected ellipsis in truncated output")
	}
	if strings.Contains(out, longSubject) {
		t.Error("expected long subject to be truncated")
	}
}

func TestTextFormatter_EmailListAlignment(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	result := types.EmailListResult{
		Total: 2,
		Emails: []types.EmailSummary{
			{
				ID:         "M1",
				From:       []types.Address{{Name: "Al", Email: "al@test.com"}},
				To:         []types.Address{{Email: "b@test.com"}},
				Subject:    "Short",
				ReceivedAt: now,
			},
			{
				ID:         "M2",
				From:       []types.Address{{Name: "Alexander Hamilton", Email: "alexander@test.com"}},
				To:         []types.Address{{Email: "b@test.com"}},
				Subject:    "A much longer subject line here",
				ReceivedAt: now,
			},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	lines := strings.Split(out, "\n")

	// Find the two main email lines (contain a date timestamp).
	var mainLines []string
	for _, line := range lines {
		if strings.Contains(line, "2026-02-04 10:30") {
			mainLines = append(mainLines, line)
		}
	}
	if len(mainLines) != 2 {
		t.Fatalf("expected 2 main email lines, got %d\nOutput:\n%s", len(mainLines), out)
	}

	// The date column should start at the same position in both lines.
	idx1 := strings.Index(mainLines[0], "2026-02-04")
	idx2 := strings.Index(mainLines[1], "2026-02-04")
	if idx1 == -1 || idx2 == -1 {
		t.Fatalf("expected '2026-02-04' in output for alignment check, got:\n%s", out)
	}
	if idx1 != idx2 {
		t.Errorf("date columns not aligned: line 1 at %d, line 2 at %d\nOutput:\n%s", idx1, idx2, out)
	}
}

func TestTextFormatter_EmailListAlignmentWideCharacters(t *testing.T) {
	f := &TextFormatter{}
	var buf bytes.Buffer

	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	result := types.EmailListResult{
		Total: 2,
		Emails: []types.EmailSummary{
			{
				ID:         "M1",
				From:       []types.Address{{Name: strings.Repeat("界", 50), Email: "wide@test.com"}},
				To:         []types.Address{{Email: "b@test.com"}},
				Subject:    "Subj",
				ReceivedAt: now,
			},
			{
				ID:         "M2",
				From:       []types.Address{{Name: "Alice", Email: "alice@test.com"}},
				To:         []types.Address{{Email: "b@test.com"}},
				Subject:    "Normal subject",
				ReceivedAt: now,
			},
		},
	}

	if err := f.Format(&buf, result); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(buf.String(), "\n")
	var mainLines []string
	for _, line := range lines {
		if strings.Contains(line, "2026-02-04 10:30") {
			mainLines = append(mainLines, line)
		}
	}
	if len(mainLines) != 2 {
		t.Fatalf("expected 2 main email lines, got %d\nOutput:\n%s", len(mainLines), buf.String())
	}

	// The date column should start at the same display-column position in both lines,
	// even though line 1 contains double-width CJK characters and line 2 does not.
	idx1 := strings.Index(mainLines[0], "2026-02-04")
	idx2 := strings.Index(mainLines[1], "2026-02-04")
	if idx1 == -1 || idx2 == -1 {
		t.Fatalf("expected '2026-02-04' in output for alignment check, got:\n%s\n%s", mainLines[0], mainLines[1])
	}
	width1 := runewidth.StringWidth(mainLines[0][:idx1])
	width2 := runewidth.StringWidth(mainLines[1][:idx2])
	if width1 != width2 {
		t.Errorf("date columns not visually aligned: line 1 at display col %d, line 2 at display col %d\nLine 1: %s\nLine 2: %s",
			width1, width2, mainLines[0], mainLines[1])
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
