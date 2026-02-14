package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMailboxInfo_JSON(t *testing.T) {
	mb := MailboxInfo{
		ID:           "mb1",
		Name:         "Inbox",
		Role:         "inbox",
		TotalEmails:  100,
		UnreadEmails: 5,
		ParentID:     "",
	}

	data, err := json.Marshal(mb)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result["id"] != "mb1" {
		t.Errorf("expected id=mb1, got %v", result["id"])
	}
	if result["name"] != "Inbox" {
		t.Errorf("expected name=Inbox, got %v", result["name"])
	}
	// ParentID should be omitted when empty.
	if _, ok := result["parent_id"]; ok {
		t.Errorf("expected parent_id to be omitted when empty, got %v", result["parent_id"])
	}
}

func TestMailboxInfo_JSON_WithParent(t *testing.T) {
	mb := MailboxInfo{
		ID:       "mb2",
		Name:     "Subfolder",
		ParentID: "mb1",
	}

	data, err := json.Marshal(mb)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result["parent_id"] != "mb1" {
		t.Errorf("expected parent_id=mb1, got %v", result["parent_id"])
	}
}

func TestEmailSummary_JSON(t *testing.T) {
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	s := EmailSummary{
		ID:         "M1",
		ThreadID:   "T1",
		From:       []Address{{Name: "Alice", Email: "alice@test.com"}},
		To:         []Address{{Name: "Bob", Email: "bob@test.com"}},
		Subject:    "Test",
		ReceivedAt: now,
		Size:       1234,
		IsUnread:   true,
		IsFlagged:  false,
		Preview:    "Preview...",
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result["id"] != "M1" {
		t.Errorf("expected id=M1, got %v", result["id"])
	}
	if result["thread_id"] != "T1" {
		t.Errorf("expected thread_id=T1, got %v", result["thread_id"])
	}
	if result["is_unread"] != true {
		t.Errorf("expected is_unread=true, got %v", result["is_unread"])
	}
	if result["is_flagged"] != false {
		t.Errorf("expected is_flagged=false, got %v", result["is_flagged"])
	}
	// Snippet should be omitted when empty.
	if _, ok := result["snippet"]; ok {
		t.Error("expected snippet to be omitted when empty")
	}
}

func TestEmailSummary_JSON_WithSnippet(t *testing.T) {
	s := EmailSummary{
		ID:      "M1",
		Snippet: "matching text",
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result["snippet"] != "matching text" {
		t.Errorf("expected snippet, got %v", result["snippet"])
	}
}

func TestEmailDetail_JSON(t *testing.T) {
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	detail := EmailDetail{
		ID:         "M1",
		ThreadID:   "T1",
		From:       []Address{{Name: "Alice", Email: "alice@test.com"}},
		To:         []Address{{Name: "Bob", Email: "bob@test.com"}},
		CC:         []Address{},
		Subject:    "Test",
		ReceivedAt: now,
		IsUnread:   false,
		Body:       "Hello",
		Attachments: []Attachment{
			{Name: "file.pdf", Type: "application/pdf", Size: 2048},
		},
	}

	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result["body"] != "Hello" {
		t.Errorf("expected body=Hello, got %v", result["body"])
	}

	attachments, ok := result["attachments"].([]interface{})
	if !ok {
		t.Fatal("expected attachments array")
	}
	if len(attachments) != 1 {
		t.Errorf("expected 1 attachment, got %d", len(attachments))
	}
}

func TestEmailDetail_JSON_NullSentAt(t *testing.T) {
	detail := EmailDetail{
		ID:          "M1",
		SentAt:      nil,
		Attachments: []Attachment{},
	}

	data, err := json.Marshal(detail)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if _, ok := result["sent_at"]; ok {
		// sent_at should be null or omitted.
		if result["sent_at"] != nil {
			t.Errorf("expected null sent_at, got %v", result["sent_at"])
		}
	}
}

func TestMoveResult_JSON_Archive(t *testing.T) {
	r := MoveResult{
		Archived: []string{"M1", "M2"},
		Destination: &DestinationInfo{
			ID: "mb-archive", Name: "Archive",
		},
		Errors: []string{},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	archived, ok := result["archived"].([]interface{})
	if !ok {
		t.Fatal("expected archived array")
	}
	if len(archived) != 2 {
		t.Errorf("expected 2 archived, got %d", len(archived))
	}
	// moved and marked_as_spam should be omitted.
	if _, ok := result["moved"]; ok {
		t.Error("expected moved to be omitted for archive result")
	}
	if _, ok := result["marked_as_spam"]; ok {
		t.Error("expected marked_as_spam to be omitted for archive result")
	}
}

func TestMoveResult_JSON_Spam(t *testing.T) {
	r := MoveResult{
		MarkedSpam: []string{"M1"},
		Errors:     []string{},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	spam, ok := result["marked_as_spam"].([]interface{})
	if !ok {
		t.Fatal("expected marked_as_spam array")
	}
	if len(spam) != 1 {
		t.Errorf("expected 1 spam, got %d", len(spam))
	}
}

func TestMoveResult_JSON_Flagged(t *testing.T) {
	r := MoveResult{
		Flagged: []string{"M1", "M2"},
		Errors:  []string{},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	flagged, ok := result["flagged"].([]interface{})
	if !ok {
		t.Fatal("expected flagged array")
	}
	if len(flagged) != 2 {
		t.Errorf("expected 2 flagged, got %d", len(flagged))
	}
	// Other action fields should be omitted.
	if _, ok := result["moved"]; ok {
		t.Error("expected moved to be omitted for flag result")
	}
	if _, ok := result["unflagged"]; ok {
		t.Error("expected unflagged to be omitted for flag result")
	}
}

func TestMoveResult_JSON_Unflagged(t *testing.T) {
	r := MoveResult{
		Unflagged: []string{"M3"},
		Errors:    []string{},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	unflagged, ok := result["unflagged"].([]interface{})
	if !ok {
		t.Fatal("expected unflagged array")
	}
	if len(unflagged) != 1 {
		t.Errorf("expected 1 unflagged, got %d", len(unflagged))
	}
	if _, ok := result["flagged"]; ok {
		t.Error("expected flagged to be omitted for unflag result")
	}
}

func TestMoveResult_JSON_WithErrors(t *testing.T) {
	r := MoveResult{
		Moved:  []string{"M1"},
		Errors: []string{"M2: not found"},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	errors, ok := result["errors"].([]interface{})
	if !ok {
		t.Fatal("expected errors array")
	}
	if len(errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(errors))
	}
	if errors[0] != "M2: not found" {
		t.Errorf("expected error message, got %v", errors[0])
	}
}

func TestAppError_JSON(t *testing.T) {
	e := AppError{
		Error:   "auth_failed",
		Message: "bad token",
		Hint:    "check JMAP_TOKEN",
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result["error"] != "auth_failed" {
		t.Errorf("expected error=auth_failed, got %v", result["error"])
	}
	if result["hint"] != "check JMAP_TOKEN" {
		t.Errorf("expected hint, got %v", result["hint"])
	}
}

func TestAppError_JSON_EmptyHint(t *testing.T) {
	e := AppError{
		Error:   "jmap_error",
		Message: "server error",
		Hint:    "",
	}

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	// Hint has omitempty, so it should be absent when empty.
	if _, ok := result["hint"]; ok {
		t.Error("expected hint to be omitted when empty (omitempty tag)")
	}
}

func TestThreadView_JSON(t *testing.T) {
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	tv := ThreadView{
		Email: EmailDetail{
			ID:          "M2",
			Subject:     "Re: Test",
			ReceivedAt:  now,
			Body:        "Reply",
			Attachments: []Attachment{},
		},
		Thread: []ThreadEmail{
			{
				ID:         "M1",
				Subject:    "Test",
				ReceivedAt: now,
				Preview:    "Original message",
			},
			{
				ID:         "M2",
				Subject:    "Re: Test",
				ReceivedAt: now,
			},
		},
	}

	data, err := json.Marshal(tv)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	thread, ok := result["thread"].([]interface{})
	if !ok {
		t.Fatal("expected thread array")
	}
	if len(thread) != 2 {
		t.Errorf("expected 2 thread emails, got %d", len(thread))
	}

	emailObj, ok := result["email"].(map[string]interface{})
	if !ok {
		t.Fatal("expected email object")
	}
	if emailObj["body"] != "Reply" {
		t.Errorf("expected body=Reply, got %v", emailObj["body"])
	}
}

func TestAddress_JSON(t *testing.T) {
	a := Address{Name: "Alice", Email: "alice@test.com"}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result["name"] != "Alice" {
		t.Errorf("expected name=Alice, got %v", result["name"])
	}
	if result["email"] != "alice@test.com" {
		t.Errorf("expected email=alice@test.com, got %v", result["email"])
	}
}

func TestEmailListResult_JSON(t *testing.T) {
	r := EmailListResult{
		Total:  42,
		Offset: 10,
		Emails: []EmailSummary{},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}

	if result["total"] != float64(42) {
		t.Errorf("expected total=42, got %v", result["total"])
	}
	if result["offset"] != float64(10) {
		t.Errorf("expected offset=10, got %v", result["offset"])
	}
}

// Roundtrip test: marshal then unmarshal should produce equivalent data.
func TestEmailSummary_Roundtrip(t *testing.T) {
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	original := EmailSummary{
		ID:         "M1",
		ThreadID:   "T1",
		From:       []Address{{Name: "Alice", Email: "alice@test.com"}},
		To:         []Address{{Email: "bob@test.com"}},
		Subject:    "Test",
		ReceivedAt: now,
		Size:       1234,
		IsUnread:   true,
		IsFlagged:  false,
		Preview:    "Preview...",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var roundtripped EmailSummary
	if err := json.Unmarshal(data, &roundtripped); err != nil {
		t.Fatal(err)
	}

	if roundtripped.ID != original.ID {
		t.Errorf("ID mismatch: %s != %s", roundtripped.ID, original.ID)
	}
	if roundtripped.Subject != original.Subject {
		t.Errorf("Subject mismatch: %s != %s", roundtripped.Subject, original.Subject)
	}
	if roundtripped.IsUnread != original.IsUnread {
		t.Errorf("IsUnread mismatch: %v != %v", roundtripped.IsUnread, original.IsUnread)
	}
	if len(roundtripped.From) != len(original.From) {
		t.Errorf("From length mismatch: %d != %d", len(roundtripped.From), len(original.From))
	}
}
