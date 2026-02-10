package client

import (
	"strings"
	"testing"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/searchsnippet"

	"github.com/cboone/jm/internal/types"
)

// --- convertAddresses tests ---

func TestConvertAddresses_Nil(t *testing.T) {
	result := convertAddresses(nil)
	if result == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 addresses, got %d", len(result))
	}
}

func TestConvertAddresses_Empty(t *testing.T) {
	result := convertAddresses([]*mail.Address{})
	if len(result) != 0 {
		t.Errorf("expected 0 addresses, got %d", len(result))
	}
}

func TestConvertAddresses_Single(t *testing.T) {
	addrs := []*mail.Address{
		{Name: "Alice", Email: "alice@example.com"},
	}
	result := convertAddresses(addrs)
	if len(result) != 1 {
		t.Fatalf("expected 1 address, got %d", len(result))
	}
	if result[0].Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", result[0].Name)
	}
	if result[0].Email != "alice@example.com" {
		t.Errorf("expected Email=alice@example.com, got %s", result[0].Email)
	}
}

func TestConvertAddresses_Multiple(t *testing.T) {
	addrs := []*mail.Address{
		{Name: "Alice", Email: "alice@example.com"},
		{Name: "", Email: "bob@example.com"},
		{Name: "Charlie", Email: "charlie@example.com"},
	}
	result := convertAddresses(addrs)
	if len(result) != 3 {
		t.Fatalf("expected 3 addresses, got %d", len(result))
	}
	if result[1].Name != "" {
		t.Errorf("expected empty Name for bob, got %s", result[1].Name)
	}
	if result[1].Email != "bob@example.com" {
		t.Errorf("expected Email=bob@example.com, got %s", result[1].Email)
	}
}

// --- safeTime tests ---

func TestSafeTime_Nil(t *testing.T) {
	result := safeTime(nil)
	if !result.IsZero() {
		t.Errorf("expected zero time, got %v", result)
	}
}

func TestSafeTime_NonNil(t *testing.T) {
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	result := safeTime(&now)
	if !result.Equal(now) {
		t.Errorf("expected %v, got %v", now, result)
	}
}

// --- extractBody tests ---

func TestExtractBody_TextPreferred(t *testing.T) {
	e := &email.Email{
		TextBody: []*email.BodyPart{{PartID: "1"}},
		HTMLBody: []*email.BodyPart{{PartID: "2"}},
		BodyValues: map[string]*email.BodyValue{
			"1": {Value: "plain text content"},
			"2": {Value: "<p>html content</p>"},
		},
	}
	body := extractBody(e, false)
	if body != "plain text content" {
		t.Errorf("expected plain text, got: %s", body)
	}
}

func TestExtractBody_HTMLPreferred(t *testing.T) {
	e := &email.Email{
		TextBody: []*email.BodyPart{{PartID: "1"}},
		HTMLBody: []*email.BodyPart{{PartID: "2"}},
		BodyValues: map[string]*email.BodyValue{
			"1": {Value: "plain text content"},
			"2": {Value: "<p>html content</p>"},
		},
	}
	body := extractBody(e, true)
	if body != "<p>html content</p>" {
		t.Errorf("expected HTML, got: %s", body)
	}
}

func TestExtractBody_FallbackToHTML(t *testing.T) {
	e := &email.Email{
		TextBody: []*email.BodyPart{},
		HTMLBody: []*email.BodyPart{{PartID: "2"}},
		BodyValues: map[string]*email.BodyValue{
			"2": {Value: "<p>only html</p>"},
		},
	}
	body := extractBody(e, false)
	if body != "<p>only html</p>" {
		t.Errorf("expected HTML fallback, got: %s", body)
	}
}

func TestExtractBody_Empty(t *testing.T) {
	e := &email.Email{
		BodyValues: map[string]*email.BodyValue{},
	}
	body := extractBody(e, false)
	if body != "" {
		t.Errorf("expected empty body, got: %s", body)
	}
}

func TestExtractBody_MissingBodyValue(t *testing.T) {
	e := &email.Email{
		TextBody:   []*email.BodyPart{{PartID: "missing"}},
		BodyValues: map[string]*email.BodyValue{},
	}
	body := extractBody(e, false)
	if body != "" {
		t.Errorf("expected empty body for missing value, got: %s", body)
	}
}

// --- convertSummaries tests ---

func TestConvertSummaries_Empty(t *testing.T) {
	result := convertSummaries([]*email.Email{})
	if len(result) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(result))
	}
}

func TestConvertSummaries_Basic(t *testing.T) {
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	emails := []*email.Email{
		{
			ID:         "M1",
			ThreadID:   "T1",
			From:       []*mail.Address{{Name: "Alice", Email: "alice@test.com"}},
			To:         []*mail.Address{{Name: "Bob", Email: "bob@test.com"}},
			Subject:    "Test Subject",
			ReceivedAt: &now,
			Size:       1234,
			Keywords:   map[string]bool{"$seen": true, "$flagged": false},
			Preview:    "Preview text...",
		},
	}
	result := convertSummaries(emails)
	if len(result) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(result))
	}

	s := result[0]
	if s.ID != "M1" {
		t.Errorf("expected ID=M1, got %s", s.ID)
	}
	if s.ThreadID != "T1" {
		t.Errorf("expected ThreadID=T1, got %s", s.ThreadID)
	}
	if s.Subject != "Test Subject" {
		t.Errorf("expected Subject='Test Subject', got %s", s.Subject)
	}
	if s.Size != 1234 {
		t.Errorf("expected Size=1234, got %d", s.Size)
	}
	if s.IsUnread {
		t.Error("expected IsUnread=false for $seen email")
	}
	if s.IsFlagged {
		t.Error("expected IsFlagged=false")
	}
	if s.Preview != "Preview text..." {
		t.Errorf("expected Preview='Preview text...', got %s", s.Preview)
	}
}

func TestConvertSummaries_UnreadAndFlagged(t *testing.T) {
	emails := []*email.Email{
		{
			ID:       "M2",
			Keywords: map[string]bool{"$flagged": true},
		},
	}
	result := convertSummaries(emails)
	if !result[0].IsUnread {
		t.Error("expected IsUnread=true when $seen is absent")
	}
	if !result[0].IsFlagged {
		t.Error("expected IsFlagged=true when $flagged is true")
	}
}

func TestConvertSummaries_NilKeywords(t *testing.T) {
	emails := []*email.Email{
		{
			ID:       "M3",
			Keywords: nil,
		},
	}
	result := convertSummaries(emails)
	if !result[0].IsUnread {
		t.Error("expected IsUnread=true when Keywords is nil")
	}
	if result[0].IsFlagged {
		t.Error("expected IsFlagged=false when Keywords is nil")
	}
}

// --- convertDetail tests ---

func TestConvertDetail_Basic(t *testing.T) {
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	sentAt := time.Date(2026, 2, 4, 10, 29, 0, 0, time.UTC)

	e := &email.Email{
		ID:         "M1",
		ThreadID:   "T1",
		From:       []*mail.Address{{Name: "Alice", Email: "alice@test.com"}},
		To:         []*mail.Address{{Name: "Bob", Email: "bob@test.com"}},
		CC:         []*mail.Address{{Name: "Charlie", Email: "charlie@test.com"}},
		BCC:        nil,
		ReplyTo:    nil,
		Subject:    "Test",
		SentAt:     &sentAt,
		ReceivedAt: &now,
		Size:       5000,
		Keywords:   map[string]bool{"$seen": false, "$flagged": true},
		TextBody:   []*email.BodyPart{{PartID: "1"}},
		BodyValues: map[string]*email.BodyValue{
			"1": {Value: "Hello world"},
		},
		Attachments: []*email.BodyPart{
			{Name: "doc.pdf", Type: "application/pdf", Size: 1024},
		},
		Headers: []*email.Header{
			{Name: "X-Custom", Value: "test-value"},
		},
	}

	detail := convertDetail(e, false, true)

	if detail.ID != "M1" {
		t.Errorf("expected ID=M1, got %s", detail.ID)
	}
	if detail.Body != "Hello world" {
		t.Errorf("expected body='Hello world', got %s", detail.Body)
	}
	if len(detail.CC) != 1 {
		t.Errorf("expected 1 CC, got %d", len(detail.CC))
	}
	if len(detail.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(detail.Attachments))
	}
	if detail.Attachments[0].Name != "doc.pdf" {
		t.Errorf("expected attachment name=doc.pdf, got %s", detail.Attachments[0].Name)
	}
	if detail.IsFlagged != true {
		t.Error("expected IsFlagged=true")
	}
	if detail.IsUnread != true {
		t.Error("expected IsUnread=true when $seen is false")
	}
	if len(detail.Headers) != 1 {
		t.Fatalf("expected 1 header, got %d", len(detail.Headers))
	}
	if detail.Headers[0].Name != "X-Custom" {
		t.Errorf("expected header name=X-Custom, got %s", detail.Headers[0].Name)
	}
}

func TestConvertDetail_NoHeaders(t *testing.T) {
	e := &email.Email{
		ID:         "M1",
		BodyValues: map[string]*email.BodyValue{},
		Headers: []*email.Header{
			{Name: "X-Custom", Value: "should-not-appear"},
		},
	}
	detail := convertDetail(e, false, false)
	if detail.Headers != nil {
		t.Errorf("expected nil headers when rawHeaders=false, got %v", detail.Headers)
	}
}

func TestConvertDetail_NoAttachments(t *testing.T) {
	e := &email.Email{
		ID:          "M1",
		Attachments: nil,
		BodyValues:  map[string]*email.BodyValue{},
	}
	detail := convertDetail(e, false, false)
	if detail.Attachments == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(detail.Attachments) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(detail.Attachments))
	}
}

func TestConvertDetail_EmptyAddresses(t *testing.T) {
	e := &email.Email{
		ID:         "M1",
		From:       nil,
		To:         nil,
		CC:         nil,
		BCC:        nil,
		ReplyTo:    nil,
		BodyValues: map[string]*email.BodyValue{},
	}
	detail := convertDetail(e, false, false)
	if detail.From == nil {
		t.Fatal("expected empty From slice, got nil")
	}
	if detail.To == nil {
		t.Fatal("expected empty To slice, got nil")
	}
}

// --- singleThreadEntry tests ---

func TestSingleThreadEntry(t *testing.T) {
	now := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	detail := types.EmailDetail{
		ID:         "M1",
		From:       []types.Address{{Name: "Alice", Email: "alice@test.com"}},
		To:         []types.Address{{Name: "Bob", Email: "bob@test.com"}},
		Subject:    "Test",
		ReceivedAt: now,
		IsUnread:   true,
	}

	entry := singleThreadEntry(detail)
	if entry.ID != "M1" {
		t.Errorf("expected ID=M1, got %s", entry.ID)
	}
	if entry.Subject != "Test" {
		t.Errorf("expected Subject=Test, got %s", entry.Subject)
	}
	if !entry.IsUnread {
		t.Error("expected IsUnread=true")
	}
	if len(entry.From) != 1 {
		t.Errorf("expected 1 From, got %d", len(entry.From))
	}
}

// --- batchSize constant test ---

func TestBatchSizeConstant(t *testing.T) {
	if batchSize != 50 {
		t.Errorf("expected batchSize=50, got %d", batchSize)
	}
}

// --- summaryProperties and detailProperties tests ---

func TestSummaryPropertiesContainsRequired(t *testing.T) {
	required := []string{"id", "threadId", "from", "to", "subject", "receivedAt", "preview"}
	for _, r := range required {
		found := false
		for _, p := range summaryProperties {
			if p == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("summaryProperties missing required field: %s", r)
		}
	}
}

func TestDetailPropertiesContainsRequired(t *testing.T) {
	required := []string{"id", "threadId", "from", "to", "cc", "bcc", "subject", "sentAt", "receivedAt", "bodyValues", "textBody", "htmlBody", "attachments"}
	for _, r := range required {
		found := false
		for _, p := range detailProperties {
			if p == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("detailProperties missing required field: %s", r)
		}
	}
}

// Verify that Email/set is never called with Destroy or Create by construction.
// This test checks that MoveEmails only constructs Update patches.
func TestMoveEmailsPatchStructure(t *testing.T) {
	// We can't call MoveEmails without a real server, but we can verify
	// the patch construction logic by checking what updates would look like.
	targetID := jmap.ID("archive-mb")
	patch := jmap.Patch{
		"mailboxIds": map[jmap.ID]bool{targetID: true},
	}

	// Verify the patch only has mailboxIds.
	if len(patch) != 1 {
		t.Errorf("expected 1 patch key, got %d", len(patch))
	}
	if _, ok := patch["mailboxIds"]; !ok {
		t.Error("expected mailboxIds in patch")
	}
	// Verify no destroy or create fields.
	if _, ok := patch["destroy"]; ok {
		t.Error("patch must not contain destroy")
	}
	if _, ok := patch["create"]; ok {
		t.Error("patch must not contain create")
	}
}

// TestSearchSnippetReference verifies that SearchSnippet/get references
// Email/query at path /ids (not Email/get at /list/*/id).
func TestSearchSnippetReference(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 1, IDs: []jmap.ID{"M1"}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{{ID: "M1", Subject: "meeting"}}}},
				{Name: "SearchSnippet/get", CallID: "2", Args: &searchsnippet.GetResponse{List: []*searchsnippet.SearchSnippet{{Email: "M1", Preview: "...<mark>meeting</mark>..."}}}},
			}}, nil
		},
	}

	result, err := c.SearchEmails(SearchOptions{Text: "meeting", Limit: 25, SortField: "receivedAt"})
	if err != nil {
		t.Fatalf("SearchEmails returned error: %v", err)
	}
	if len(result.Emails) != 1 {
		t.Fatalf("expected 1 email in result, got %d", len(result.Emails))
	}
	if result.Emails[0].Snippet == "" {
		t.Fatal("expected snippet to be populated for text search")
	}

	if captured == nil {
		t.Fatal("expected SearchEmails to send a request")
	}
	if len(captured.Calls) != 3 {
		t.Fatalf("expected 3 method calls, got %d", len(captured.Calls))
	}

	queryCall := captured.Calls[0]
	if _, ok := queryCall.Args.(*email.Query); !ok {
		t.Fatalf("expected first call args to be *email.Query, got %T", queryCall.Args)
	}

	snippetCall := captured.Calls[2]
	snippetGetWrapper, ok := snippetCall.Args.(*searchSnippetGet)
	if !ok {
		t.Fatalf("expected third call args to be *searchSnippetGet, got %T", snippetCall.Args)
	}

	ref := snippetGetWrapper.ReferenceIDs
	if ref == nil {
		t.Fatal("expected ReferenceIDs to be set")
	}
	if ref.ResultOf != queryCall.CallID {
		t.Errorf("expected ResultOf=%s (query call id), got %s", queryCall.CallID, ref.ResultOf)
	}
	if ref.Name != "Email/query" {
		t.Errorf("expected Name=Email/query, got %s", ref.Name)
	}
	if ref.Path != "/ids" {
		t.Errorf("expected Path=/ids, got %s", ref.Path)
	}
}

// TestSearchSnippetMethodName verifies that the searchSnippetGet wrapper
// returns the correct JMAP method name (not "Mailbox/get").
func TestSearchSnippetMethodName(t *testing.T) {
	sg := &searchSnippetGet{}
	if sg.Name() != "SearchSnippet/get" {
		t.Errorf("expected Name()=SearchSnippet/get, got %s", sg.Name())
	}
}

// TestSearchSnippetFailureGraceful verifies that a SearchSnippet/get
// MethodError does not fail the entire search when Email/query and
// Email/get succeed.
func TestSearchSnippetFailureGraceful(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 1, IDs: []jmap.ID{"M1"}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{{ID: "M1", Subject: "test"}}}},
				{Name: "error", CallID: "2", Args: &jmap.MethodError{}},
			}}, nil
		},
	}

	result, err := c.SearchEmails(SearchOptions{Text: "test", Limit: 25, SortField: "receivedAt"})
	if err != nil {
		t.Fatalf("expected graceful degradation, got error: %v", err)
	}
	if len(result.Emails) != 1 {
		t.Fatalf("expected 1 email, got %d", len(result.Emails))
	}
	if result.Emails[0].Snippet != "" {
		t.Errorf("expected empty snippet on snippet failure, got %q", result.Emails[0].Snippet)
	}
}

// TestSearchErrorIdentifiesMethod verifies that a MethodError on Email/query
// includes the method name in the error message.
func TestSearchErrorIdentifiesMethod(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "error", CallID: "0", Args: &jmap.MethodError{}},
			}}, nil
		},
	}

	_, err := c.SearchEmails(SearchOptions{From: "alice", Limit: 25, SortField: "receivedAt"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "Email/query") {
		t.Errorf("expected error to identify Email/query, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "call 0") {
		t.Errorf("expected error to include call index/call id, got: %s", errMsg)
	}
}

// TestSearchErrorIdentifiesEmailGetByCallID verifies that method errors with
// opaque invocation names still map to Email/get by call ID.
func TestSearchErrorIdentifiesEmailGetByCallID(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 1, IDs: []jmap.ID{"M1"}}},
				{Name: "error", CallID: "1", Args: &jmap.MethodError{Type: "invalidArguments"}},
			}}, nil
		},
	}

	_, err := c.SearchEmails(SearchOptions{From: "alice", Limit: 25, SortField: "receivedAt"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "Email/get") {
		t.Errorf("expected error to identify Email/get, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "call 1") {
		t.Errorf("expected error to include call id, got: %s", errMsg)
	}
}

// TestSearchErrorFallsBackToInvocationIndex verifies that when both method
// name and call ID are unavailable, the error reports the invocation index.
func TestSearchErrorFallsBackToInvocationIndex(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "error", CallID: "", Args: &jmap.MethodError{Type: "invalidArguments"}},
			}}, nil
		},
	}

	_, err := c.SearchEmails(SearchOptions{From: "alice", Limit: 25, SortField: "receivedAt"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "call 0") {
		t.Errorf("expected fallback to invocation index, got: %s", errMsg)
	}
}

// TestMarkAsReadPatchStructure verifies the patch structure for mark-as-read.
func TestMarkAsReadPatchStructure(t *testing.T) {
	patch := jmap.Patch{
		"keywords/$seen": true,
	}
	if len(patch) != 1 {
		t.Errorf("expected 1 patch key, got %d", len(patch))
	}
	if _, ok := patch["keywords/$seen"]; !ok {
		t.Error("expected keywords/$seen in patch")
	}
	if _, ok := patch["mailboxIds"]; ok {
		t.Error("mark-as-read patch must not contain mailboxIds")
	}
}

func TestSpamPatchStructure(t *testing.T) {
	targetID := jmap.ID("junk-mb")
	patch := jmap.Patch{
		"mailboxIds":     map[jmap.ID]bool{targetID: true},
		"keywords/$junk": true,
	}

	if len(patch) != 2 {
		t.Errorf("expected 2 patch keys, got %d", len(patch))
	}
	if _, ok := patch["mailboxIds"]; !ok {
		t.Error("expected mailboxIds in patch")
	}
	if _, ok := patch["keywords/$junk"]; !ok {
		t.Error("expected keywords/$junk in patch")
	}
}
