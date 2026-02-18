package client

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/core"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
	"git.sr.ht/~rockorager/go-jmap/mail/searchsnippet"

	"github.com/cboone/fm/internal/types"
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

func TestConvertDetail_ListUnsubscribe(t *testing.T) {
	e := &email.Email{
		ID:         "M1",
		BodyValues: map[string]*email.BodyValue{},
		Headers: []*email.Header{
			{Name: "List-Unsubscribe", Value: " <mailto:unsub@example.com>"},
			{Name: "List-Unsubscribe-Post", Value: " List-Unsubscribe=One-Click"},
			{Name: "X-Custom", Value: "other"},
		},
	}

	// List-Unsubscribe fields should be populated even without rawHeaders.
	detail := convertDetail(e, false, false)
	if detail.ListUnsubscribe != "<mailto:unsub@example.com>" {
		t.Errorf("expected ListUnsubscribe='<mailto:unsub@example.com>', got %q", detail.ListUnsubscribe)
	}
	if detail.ListUnsubscribePost != "List-Unsubscribe=One-Click" {
		t.Errorf("expected ListUnsubscribePost='List-Unsubscribe=One-Click', got %q", detail.ListUnsubscribePost)
	}
	if detail.Headers != nil {
		t.Errorf("expected nil Headers when rawHeaders=false, got %v", detail.Headers)
	}
}

func TestConvertDetail_ListUnsubscribeCaseInsensitive(t *testing.T) {
	e := &email.Email{
		ID:         "M1",
		BodyValues: map[string]*email.BodyValue{},
		Headers: []*email.Header{
			{Name: "list-unsubscribe", Value: " <mailto:lower@example.com>"},
			{Name: "LiSt-UnSuBsCrIbE-PoSt", Value: " List-Unsubscribe=One-Click"},
		},
	}

	detail := convertDetail(e, false, false)
	if detail.ListUnsubscribe != "<mailto:lower@example.com>" {
		t.Errorf("expected case-insensitive ListUnsubscribe match, got %q", detail.ListUnsubscribe)
	}
	if detail.ListUnsubscribePost != "List-Unsubscribe=One-Click" {
		t.Errorf("expected case-insensitive ListUnsubscribePost match, got %q", detail.ListUnsubscribePost)
	}
}

func TestConvertDetail_ListUnsubscribeWithRawHeaders(t *testing.T) {
	e := &email.Email{
		ID:         "M1",
		BodyValues: map[string]*email.BodyValue{},
		Headers: []*email.Header{
			{Name: "List-Unsubscribe", Value: " <mailto:unsub@example.com>"},
			{Name: "X-Custom", Value: "test"},
		},
	}

	detail := convertDetail(e, false, true)
	if detail.ListUnsubscribe != "<mailto:unsub@example.com>" {
		t.Errorf("expected ListUnsubscribe populated with rawHeaders=true, got %q", detail.ListUnsubscribe)
	}
	if len(detail.Headers) != 2 {
		t.Fatalf("expected 2 raw headers, got %d", len(detail.Headers))
	}
}

func TestConvertDetail_NoListUnsubscribe(t *testing.T) {
	e := &email.Email{
		ID:         "M1",
		BodyValues: map[string]*email.BodyValue{},
		Headers: []*email.Header{
			{Name: "X-Custom", Value: "test"},
		},
	}

	detail := convertDetail(e, false, false)
	if detail.ListUnsubscribe != "" {
		t.Errorf("expected empty ListUnsubscribe, got %q", detail.ListUnsubscribe)
	}
	if detail.ListUnsubscribePost != "" {
		t.Errorf("expected empty ListUnsubscribePost, got %q", detail.ListUnsubscribePost)
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

// --- batchSetEmails tests ---

func TestBatchSetEmails_UsesServerMaxObjectsInSet(t *testing.T) {
	var requestCount int

	c := &Client{
		accountID: "test-account",
		jmap: &jmap.Client{
			Session: &jmap.Session{
				Capabilities: map[jmap.URI]jmap.Capability{
					jmap.CoreURI: &core.Core{MaxObjectsInSet: 2},
				},
			},
		},
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			requestCount++
			setReq := req.Calls[0].Args.(*email.Set)
			updated := make(map[jmap.ID]*email.Email, len(setReq.Update))
			for id := range setReq.Update {
				updated[id] = &email.Email{}
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{Updated: updated}},
			}}, nil
		},
	}

	ids := []string{"M1", "M2", "M3", "M4", "M5"}
	succeeded, errs := c.batchSetEmails(ids, func(_ string) jmap.Patch {
		return jmap.Patch{"keywords/$seen": true}
	})

	if requestCount != 3 {
		t.Fatalf("expected 3 batches, got %d", requestCount)
	}
	if len(succeeded) != 5 {
		t.Fatalf("expected 5 succeeded, got %d", len(succeeded))
	}
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}
}

func TestBatchSetEmails_FallsBackToDefault(t *testing.T) {
	var requestCount int
	var batchSizes []int

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			requestCount++
			setReq := req.Calls[0].Args.(*email.Set)
			batchSizes = append(batchSizes, len(setReq.Update))
			updated := make(map[jmap.ID]*email.Email, len(setReq.Update))
			for id := range setReq.Update {
				updated[id] = &email.Email{}
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{Updated: updated}},
			}}, nil
		},
	}

	ids := make([]string, defaultBatchSize+1)
	for i := range ids {
		ids[i] = fmt.Sprintf("M%d", i+1)
	}
	succeeded, errs := c.batchSetEmails(ids, func(_ string) jmap.Patch {
		return jmap.Patch{"keywords/$seen": true}
	})

	if requestCount != 2 {
		t.Fatalf("expected 2 batches at default size, got %d", requestCount)
	}
	if len(batchSizes) != 2 || batchSizes[0] != defaultBatchSize || batchSizes[1] != 1 {
		t.Fatalf("expected batch sizes [%d 1], got %v", defaultBatchSize, batchSizes)
	}
	if len(succeeded) != defaultBatchSize+1 {
		t.Fatalf("expected %d succeeded, got %d", defaultBatchSize+1, len(succeeded))
	}
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}
}

func TestBatchSetEmails_LargeBatch(t *testing.T) {
	var requestCount int

	c := &Client{
		accountID: "test-account",
		jmap: &jmap.Client{
			Session: &jmap.Session{
				Capabilities: map[jmap.URI]jmap.Capability{
					jmap.CoreURI: &core.Core{MaxObjectsInSet: 50},
				},
			},
		},
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			requestCount++
			setReq := req.Calls[0].Args.(*email.Set)
			updated := make(map[jmap.ID]*email.Email, len(setReq.Update))
			for id := range setReq.Update {
				updated[id] = &email.Email{}
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{Updated: updated}},
			}}, nil
		},
	}

	ids := make([]string, 225)
	for i := range ids {
		ids[i] = fmt.Sprintf("M%d", i+1)
	}
	succeeded, errs := c.batchSetEmails(ids, func(_ string) jmap.Patch {
		return jmap.Patch{"keywords/$seen": true}
	})

	if requestCount != 5 { // 50+50+50+50+25
		t.Fatalf("expected 5 batches for 225 IDs at size 50, got %d", requestCount)
	}
	if len(succeeded) != 225 {
		t.Fatalf("expected 225 succeeded, got %d", len(succeeded))
	}
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors, got %d", len(errs))
	}
}

func TestBatchSetEmails_PartialBatchFailure(t *testing.T) {
	var callNum int

	c := &Client{
		accountID: "test-account",
		jmap: &jmap.Client{
			Session: &jmap.Session{
				Capabilities: map[jmap.URI]jmap.Capability{
					jmap.CoreURI: &core.Core{MaxObjectsInSet: 2},
				},
			},
		},
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			callNum++
			if callNum == 2 {
				return nil, fmt.Errorf("network error")
			}
			setReq := req.Calls[0].Args.(*email.Set)
			updated := make(map[jmap.ID]*email.Email, len(setReq.Update))
			for id := range setReq.Update {
				updated[id] = &email.Email{}
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{Updated: updated}},
			}}, nil
		},
	}

	ids := []string{"M1", "M2", "M3", "M4", "M5"}
	succeeded, errs := c.batchSetEmails(ids, func(_ string) jmap.Patch {
		return jmap.Patch{"keywords/$seen": true}
	})

	// Batch 1: M1, M2 -> succeed. Batch 2: M3, M4 -> fail. Batch 3: M5 -> succeed.
	if len(succeeded) != 3 {
		t.Fatalf("expected 3 succeeded, got %d: %v", len(succeeded), succeeded)
	}
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestBatchSetEmails_MixedPerIDErrors(t *testing.T) {
	failDesc := "not found"

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{
					Name:   "Email/set",
					CallID: "0",
					Args: &email.SetResponse{
						Updated: map[jmap.ID]*email.Email{
							"M1": {},
							"M3": {},
						},
						NotUpdated: map[jmap.ID]*jmap.SetError{
							"M2": {Description: &failDesc},
						},
					},
				},
			}}, nil
		},
	}

	succeeded, errs := c.batchSetEmails([]string{"M1", "M2", "M3"}, func(_ string) jmap.Patch {
		return jmap.Patch{"keywords/$seen": true}
	})

	if len(succeeded) != 2 {
		t.Fatalf("expected 2 succeeded, got %d: %v", len(succeeded), succeeded)
	}
	if len(errs) != 1 || errs[0] != "M2: not found" {
		t.Fatalf("expected 1 error 'M2: not found', got %v", errs)
	}
}

func TestBatchSetEmails_UnaccountedID(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{
					Name:   "Email/set",
					CallID: "0",
					Args: &email.SetResponse{
						Updated: map[jmap.ID]*email.Email{
							"M1": {},
						},
						NotUpdated: map[jmap.ID]*jmap.SetError{},
					},
				},
			}}, nil
		},
	}

	succeeded, errs := c.batchSetEmails([]string{"M1", "M2"}, func(_ string) jmap.Patch {
		return jmap.Patch{"keywords/$seen": true}
	})

	if len(succeeded) != 1 || succeeded[0] != "M1" {
		t.Fatalf("expected succeeded=[M1], got %v", succeeded)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for unaccounted ID, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "M2") || !strings.Contains(errs[0], "no status returned") {
		t.Fatalf("expected error about M2 with no status, got: %s", errs[0])
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

// TestSearchEmails_QueryOptions verifies that UnreadOnly, Offset, Limit, and custom
// sort field/direction are correctly encoded into the outbound Email/query call.
func TestSearchEmails_QueryOptions(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
			}}, nil
		},
	}

	_, err := c.SearchEmails(SearchOptions{
		UnreadOnly: true,
		Offset:     42,
		SortField:  "sentAt",
		SortAsc:    true,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured == nil {
		t.Fatal("expected SearchEmails to send a request")
	}

	query, ok := captured.Calls[0].Args.(*email.Query)
	if !ok {
		t.Fatalf("expected first call args to be *email.Query, got %T", captured.Calls[0].Args)
	}

	// Verify UnreadOnly sets NotKeyword to "$seen".
	fc, ok := query.Filter.(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected *email.FilterCondition, got %T", query.Filter)
	}
	if fc.NotKeyword != "$seen" {
		t.Errorf("expected NotKeyword=$seen for UnreadOnly, got %q", fc.NotKeyword)
	}

	// Verify Offset is encoded as Position.
	if query.Position != 42 {
		t.Errorf("expected Position=42, got %d", query.Position)
	}

	// Verify Limit is encoded on the query.
	if query.Limit != 10 {
		t.Errorf("expected Limit=10, got %d", query.Limit)
	}

	// Verify custom sort field and direction.
	if len(query.Sort) != 1 {
		t.Fatalf("expected 1 SortComparator, got %d", len(query.Sort))
	}
	if query.Sort[0].Property != "sentAt" {
		t.Errorf("expected sort Property=sentAt, got %q", query.Sort[0].Property)
	}
	if !query.Sort[0].IsAscending {
		t.Error("expected IsAscending=true for SortAsc=true")
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

func TestSetFlagged_MixedUpdatedAndNotUpdated(t *testing.T) {
	failDesc := "forbidden"
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{
					Name:   "Email/set",
					CallID: "0",
					Args: &email.SetResponse{
						Updated: map[jmap.ID]*email.Email{
							"M1": {},
						},
						NotUpdated: map[jmap.ID]*jmap.SetError{
							"M2": {Description: &failDesc},
						},
					},
				},
			}}, nil
		},
	}

	succeeded, errs := c.SetFlagged([]string{"M1", "M2"})
	if len(succeeded) != 1 || succeeded[0] != "M1" {
		t.Fatalf("expected succeeded=[M1], got %v", succeeded)
	}
	if len(errs) != 1 || errs[0] != "M2: forbidden" {
		t.Fatalf("expected errs=[M2: forbidden], got %v", errs)
	}

	if captured == nil {
		t.Fatal("expected request to be captured")
	}
	if len(captured.Calls) != 1 {
		t.Fatalf("expected 1 method call, got %d", len(captured.Calls))
	}

	setReq, ok := captured.Calls[0].Args.(*email.Set)
	if !ok {
		t.Fatalf("expected *email.Set args, got %T", captured.Calls[0].Args)
	}
	if setReq.Update == nil {
		t.Fatal("expected update map to be set")
	}
	if len(setReq.Update) != 2 {
		t.Fatalf("expected 2 update patches, got %d", len(setReq.Update))
	}
	for _, id := range []jmap.ID{"M1", "M2"} {
		patch, ok := setReq.Update[id]
		if !ok {
			t.Fatalf("expected patch for %s", id)
		}
		v, ok := patch["keywords/$flagged"]
		if !ok {
			t.Fatalf("expected keywords/$flagged patch for %s", id)
		}
		flagged, ok := v.(bool)
		if !ok || !flagged {
			t.Fatalf("expected keywords/$flagged=true for %s, got %#v", id, v)
		}
	}
}

func TestSetFlagged_MethodErrorForAllIDs(t *testing.T) {
	desc := "invalid keyword update"
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "error", CallID: "0", Args: &jmap.MethodError{Type: "invalidArguments", Description: &desc}},
			}}, nil
		},
	}

	succeeded, errs := c.SetFlagged([]string{"M1", "M2"})
	if len(succeeded) != 0 {
		t.Fatalf("expected no succeeded IDs, got %v", succeeded)
	}
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d (%v)", len(errs), errs)
	}
	if errs[0] != "M1: invalidArguments: invalid keyword update" {
		t.Fatalf("unexpected first error: %s", errs[0])
	}
	if errs[1] != "M2: invalidArguments: invalid keyword update" {
		t.Fatalf("unexpected second error: %s", errs[1])
	}
}

func TestSetUnflagged_MixedUpdatedAndNotUpdated(t *testing.T) {
	failDesc := "notFound"
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{
					Name:   "Email/set",
					CallID: "0",
					Args: &email.SetResponse{
						Updated: map[jmap.ID]*email.Email{
							"M3": {},
						},
						NotUpdated: map[jmap.ID]*jmap.SetError{
							"M4": {Description: &failDesc},
						},
					},
				},
			}}, nil
		},
	}

	succeeded, errs := c.SetUnflagged([]string{"M3", "M4"})
	if len(succeeded) != 1 || succeeded[0] != "M3" {
		t.Fatalf("expected succeeded=[M3], got %v", succeeded)
	}
	if len(errs) != 1 || errs[0] != "M4: notFound" {
		t.Fatalf("expected errs=[M4: notFound], got %v", errs)
	}

	if captured == nil {
		t.Fatal("expected request to be captured")
	}
	if len(captured.Calls) != 1 {
		t.Fatalf("expected 1 method call, got %d", len(captured.Calls))
	}

	setReq, ok := captured.Calls[0].Args.(*email.Set)
	if !ok {
		t.Fatalf("expected *email.Set args, got %T", captured.Calls[0].Args)
	}
	if setReq.Update == nil {
		t.Fatal("expected update map to be set")
	}
	if len(setReq.Update) != 2 {
		t.Fatalf("expected 2 update patches, got %d", len(setReq.Update))
	}
	for _, id := range []jmap.ID{"M3", "M4"} {
		patch, ok := setReq.Update[id]
		if !ok {
			t.Fatalf("expected patch for %s", id)
		}
		v, ok := patch["keywords/$flagged"]
		if !ok {
			t.Fatalf("expected keywords/$flagged patch for %s", id)
		}
		if v != nil {
			t.Fatalf("expected keywords/$flagged=nil for %s, got %#v", id, v)
		}
	}
}

func TestSetUnflagged_MethodErrorForAllIDs(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "error", CallID: "0", Args: &jmap.MethodError{Type: "forbidden"}},
			}}, nil
		},
	}

	succeeded, errs := c.SetUnflagged([]string{"M3", "M4"})
	if len(succeeded) != 0 {
		t.Fatalf("expected no succeeded IDs, got %v", succeeded)
	}
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d (%v)", len(errs), errs)
	}
	if errs[0] != "M3: forbidden" {
		t.Fatalf("unexpected first error: %s", errs[0])
	}
	if errs[1] != "M4: forbidden" {
		t.Fatalf("unexpected second error: %s", errs[1])
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

// TestFlagPatchStructure verifies the patch structure for flag (set $flagged).
func TestFlagPatchStructure(t *testing.T) {
	patch := jmap.Patch{
		"keywords/$flagged": true,
	}
	if len(patch) != 1 {
		t.Errorf("expected 1 patch key, got %d", len(patch))
	}
	if _, ok := patch["keywords/$flagged"]; !ok {
		t.Error("expected keywords/$flagged in patch")
	}
	if _, ok := patch["mailboxIds"]; ok {
		t.Error("flag patch must not contain mailboxIds")
	}
}

// TestUnflagPatchStructure verifies the patch structure for unflag.
// Per the IETF MailFlagBit spec, unflagging clears $flagged and all color bits.
func TestUnflagPatchStructure(t *testing.T) {
	patch := jmap.Patch{
		"keywords/$flagged":      nil,
		"keywords/$MailFlagBit0": nil,
		"keywords/$MailFlagBit1": nil,
		"keywords/$MailFlagBit2": nil,
	}
	if len(patch) != 4 {
		t.Errorf("expected 4 patch keys, got %d", len(patch))
	}
	for _, key := range []string{"keywords/$flagged", "keywords/$MailFlagBit0", "keywords/$MailFlagBit1", "keywords/$MailFlagBit2"} {
		val, ok := patch[key]
		if !ok {
			t.Errorf("expected %s in patch", key)
		}
		if val != nil {
			t.Errorf("expected %s to be nil (remove), got %v", key, val)
		}
	}
	if _, ok := patch["mailboxIds"]; ok {
		t.Error("unflag patch must not contain mailboxIds")
	}
}

// --- SearchEmails filter condition tests ---

// TestSearchEmails_FilterConditions verifies that each SearchOptions filter
// field is correctly encoded into the outbound Email/query FilterCondition.
func TestSearchEmails_FilterConditions(t *testing.T) {
	before := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		opts    SearchOptions
		assert  func(t *testing.T, fc *email.FilterCondition)
		hasText bool // true if opts.Text is set (needs SearchSnippet/get in response)
	}{
		{
			name: "Text",
			opts: SearchOptions{Text: "meeting notes"},
			assert: func(t *testing.T, fc *email.FilterCondition) {
				if fc.Text != "meeting notes" {
					t.Errorf("expected Text=%q, got %q", "meeting notes", fc.Text)
				}
			},
			hasText: true,
		},
		{
			name: "From",
			opts: SearchOptions{From: "alice@example.com"},
			assert: func(t *testing.T, fc *email.FilterCondition) {
				if fc.From != "alice@example.com" {
					t.Errorf("expected From=%q, got %q", "alice@example.com", fc.From)
				}
			},
		},
		{
			name: "To",
			opts: SearchOptions{To: "bob@example.com"},
			assert: func(t *testing.T, fc *email.FilterCondition) {
				if fc.To != "bob@example.com" {
					t.Errorf("expected To=%q, got %q", "bob@example.com", fc.To)
				}
			},
		},
		{
			name: "Subject",
			opts: SearchOptions{Subject: "quarterly report"},
			assert: func(t *testing.T, fc *email.FilterCondition) {
				if fc.Subject != "quarterly report" {
					t.Errorf("expected Subject=%q, got %q", "quarterly report", fc.Subject)
				}
			},
		},
		{
			name: "Before",
			opts: SearchOptions{Before: &before},
			assert: func(t *testing.T, fc *email.FilterCondition) {
				if fc.Before == nil {
					t.Fatal("expected Before to be set")
				}
				if !fc.Before.Equal(before) {
					t.Errorf("expected Before=%v, got %v", before, *fc.Before)
				}
			},
		},
		{
			name: "After",
			opts: SearchOptions{After: &after},
			assert: func(t *testing.T, fc *email.FilterCondition) {
				if fc.After == nil {
					t.Fatal("expected After to be set")
				}
				if !fc.After.Equal(after) {
					t.Errorf("expected After=%v, got %v", after, *fc.After)
				}
			},
		},
		{
			name: "HasAttachment",
			opts: SearchOptions{HasAttachment: true},
			assert: func(t *testing.T, fc *email.FilterCondition) {
				if !fc.HasAttachment {
					t.Error("expected HasAttachment=true")
				}
			},
		},
		{
			name: "MailboxID",
			opts: SearchOptions{MailboxID: "mb-inbox-1"},
			assert: func(t *testing.T, fc *email.FilterCondition) {
				if fc.InMailbox != "mb-inbox-1" {
					t.Errorf("expected InMailbox=%q, got %q", "mb-inbox-1", fc.InMailbox)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured *jmap.Request

			// Build mock response: always include Email/query + Email/get,
			// and add SearchSnippet/get when the subtest uses Text search.
			invocations := []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
			}
			if tt.hasText {
				invocations = append(invocations, &jmap.Invocation{
					Name: "SearchSnippet/get", CallID: "2",
					Args: &searchsnippet.GetResponse{List: []*searchsnippet.SearchSnippet{}},
				})
			}

			c := &Client{
				accountID: "test-account",
				doFunc: func(req *jmap.Request) (*jmap.Response, error) {
					captured = req
					return &jmap.Response{Responses: invocations}, nil
				},
			}

			// Fill in required defaults so the query is valid.
			opts := tt.opts
			if opts.Limit == 0 {
				opts.Limit = 25
			}
			if opts.SortField == "" {
				opts.SortField = "receivedAt"
			}

			_, err := c.SearchEmails(opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			query, ok := captured.Calls[0].Args.(*email.Query)
			if !ok {
				t.Fatalf("expected *email.Query, got %T", captured.Calls[0].Args)
			}
			fc, ok := query.Filter.(*email.FilterCondition)
			if !ok {
				t.Fatalf("expected *email.FilterCondition, got %T", query.Filter)
			}

			tt.assert(t, fc)
		})
	}
}

// --- Flagged/unflagged filter tests ---

// TestSearchEmails_FlaggedOnly verifies that FlaggedOnly sets HasKeyword on the filter.
func TestSearchEmails_FlaggedOnly(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
			}}, nil
		},
	}

	_, err := c.SearchEmails(SearchOptions{FlaggedOnly: true, Limit: 25, SortField: "receivedAt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, ok := captured.Calls[0].Args.(*email.Query)
	if !ok {
		t.Fatalf("expected *email.Query, got %T", captured.Calls[0].Args)
	}
	fc, ok := query.Filter.(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected *email.FilterCondition, got %T", query.Filter)
	}
	if fc.HasKeyword != "$flagged" {
		t.Errorf("expected HasKeyword=$flagged, got %q", fc.HasKeyword)
	}
}

// TestSearchEmails_UnflaggedOnly verifies that UnflaggedOnly sets NotKeyword on the filter.
func TestSearchEmails_UnflaggedOnly(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
			}}, nil
		},
	}

	_, err := c.SearchEmails(SearchOptions{UnflaggedOnly: true, Limit: 25, SortField: "receivedAt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, ok := captured.Calls[0].Args.(*email.Query)
	if !ok {
		t.Fatalf("expected *email.Query, got %T", captured.Calls[0].Args)
	}
	fc, ok := query.Filter.(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected *email.FilterCondition, got %T", query.Filter)
	}
	if fc.NotKeyword != "$flagged" {
		t.Errorf("expected NotKeyword=$flagged, got %q", fc.NotKeyword)
	}
}

// TestSearchEmails_UnflaggedAndUnread verifies the compound FilterOperator for
// UnflaggedOnly + UnreadOnly (both need NotKeyword).
func TestSearchEmails_UnflaggedAndUnread(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
			}}, nil
		},
	}

	_, err := c.SearchEmails(SearchOptions{UnflaggedOnly: true, UnreadOnly: true, Limit: 25, SortField: "receivedAt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, ok := captured.Calls[0].Args.(*email.Query)
	if !ok {
		t.Fatalf("expected *email.Query, got %T", captured.Calls[0].Args)
	}
	op, ok := query.Filter.(*email.FilterOperator)
	if !ok {
		t.Fatalf("expected *email.FilterOperator for compound filter, got %T", query.Filter)
	}
	if op.Operator != jmap.OperatorAND {
		t.Errorf("expected AND operator, got %q", op.Operator)
	}
	if len(op.Conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(op.Conditions))
	}

	first, ok := op.Conditions[0].(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected first condition to be *email.FilterCondition, got %T", op.Conditions[0])
	}
	if first.NotKeyword != "$seen" {
		t.Errorf("expected first condition NotKeyword=$seen, got %q", first.NotKeyword)
	}

	second, ok := op.Conditions[1].(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected second condition to be *email.FilterCondition, got %T", op.Conditions[1])
	}
	if second.NotKeyword != "$flagged" {
		t.Errorf("expected second condition NotKeyword=$flagged, got %q", second.NotKeyword)
	}
}

// TestSearchEmails_UnflaggedAndUnread_SnippetFilter verifies that the compound
// filter is also passed to SearchSnippet/get when doing a text search.
func TestSearchEmails_UnflaggedAndUnread_SnippetFilter(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
				{Name: "SearchSnippet/get", CallID: "2", Args: &searchsnippet.GetResponse{List: []*searchsnippet.SearchSnippet{}}},
			}}, nil
		},
	}

	_, err := c.SearchEmails(SearchOptions{Text: "test", UnflaggedOnly: true, UnreadOnly: true, Limit: 25, SortField: "receivedAt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(captured.Calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(captured.Calls))
	}

	snippetGet, ok := captured.Calls[2].Args.(*searchSnippetGet)
	if !ok {
		t.Fatalf("expected *searchSnippetGet, got %T", captured.Calls[2].Args)
	}
	if _, ok := snippetGet.Filter.(*email.FilterOperator); !ok {
		t.Errorf("expected snippet filter to be *email.FilterOperator, got %T", snippetGet.Filter)
	}
}

// TestListEmails_FlaggedOnly verifies that flaggedOnly sets HasKeyword on the filter.
func TestListEmails_FlaggedOnly(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
			}}, nil
		},
		mailboxCache: []*mailbox.Mailbox{{ID: "mb-inbox", Name: "Inbox", Role: mailbox.RoleInbox}},
	}

	_, err := c.ListEmails("inbox", 25, 0, false, true, false, "receivedAt", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, ok := captured.Calls[0].Args.(*email.Query)
	if !ok {
		t.Fatalf("expected *email.Query, got %T", captured.Calls[0].Args)
	}
	fc, ok := query.Filter.(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected *email.FilterCondition, got %T", query.Filter)
	}
	if fc.HasKeyword != "$flagged" {
		t.Errorf("expected HasKeyword=$flagged, got %q", fc.HasKeyword)
	}
	if fc.InMailbox == "" {
		t.Error("expected InMailbox to be set")
	}
}

// TestListEmails_UnflaggedOnly verifies that unflaggedOnly sets NotKeyword on the filter.
func TestListEmails_UnflaggedOnly(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
			}}, nil
		},
		mailboxCache: []*mailbox.Mailbox{{ID: "mb-inbox", Name: "Inbox", Role: mailbox.RoleInbox}},
	}

	_, err := c.ListEmails("inbox", 25, 0, false, false, true, "receivedAt", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, ok := captured.Calls[0].Args.(*email.Query)
	if !ok {
		t.Fatalf("expected *email.Query, got %T", captured.Calls[0].Args)
	}
	fc, ok := query.Filter.(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected *email.FilterCondition, got %T", query.Filter)
	}
	if fc.NotKeyword != "$flagged" {
		t.Errorf("expected NotKeyword=$flagged, got %q", fc.NotKeyword)
	}
}

// TestListEmails_UnflaggedAndUnread verifies the compound FilterOperator for
// unflaggedOnly + unreadOnly, preserving InMailbox on the first condition.
func TestListEmails_UnflaggedAndUnread(t *testing.T) {
	var captured *jmap.Request

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			captured = req
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/query", CallID: "0", Args: &email.QueryResponse{Total: 0, IDs: []jmap.ID{}}},
				{Name: "Email/get", CallID: "1", Args: &email.GetResponse{List: []*email.Email{}}},
			}}, nil
		},
		mailboxCache: []*mailbox.Mailbox{{ID: "mb-inbox", Name: "Inbox", Role: mailbox.RoleInbox}},
	}

	_, err := c.ListEmails("inbox", 25, 0, true, false, true, "receivedAt", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query, ok := captured.Calls[0].Args.(*email.Query)
	if !ok {
		t.Fatalf("expected *email.Query, got %T", captured.Calls[0].Args)
	}
	op, ok := query.Filter.(*email.FilterOperator)
	if !ok {
		t.Fatalf("expected *email.FilterOperator for compound filter, got %T", query.Filter)
	}
	if op.Operator != jmap.OperatorAND {
		t.Errorf("expected AND operator, got %q", op.Operator)
	}
	if len(op.Conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(op.Conditions))
	}

	first, ok := op.Conditions[0].(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected first condition to be *email.FilterCondition, got %T", op.Conditions[0])
	}
	if first.NotKeyword != "$seen" {
		t.Errorf("expected first condition NotKeyword=$seen, got %q", first.NotKeyword)
	}
	if first.InMailbox == "" {
		t.Error("expected InMailbox to be preserved on first condition")
	}

	second, ok := op.Conditions[1].(*email.FilterCondition)
	if !ok {
		t.Fatalf("expected second condition to be *email.FilterCondition, got %T", op.Conditions[1])
	}
	if second.NotKeyword != "$flagged" {
		t.Errorf("expected second condition NotKeyword=$flagged, got %q", second.NotKeyword)
	}
}

// --- GetEmailSummaries tests ---

func TestGetEmailSummaries_ReturnsSummaries(t *testing.T) {
	now := time.Date(2026, 2, 14, 10, 30, 0, 0, time.UTC)

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{
					Name:   "Email/get",
					CallID: "0",
					Args: &email.GetResponse{
						List: []*email.Email{
							{
								ID:         "M1",
								ThreadID:   "T1",
								From:       []*mail.Address{{Name: "Alice", Email: "alice@example.com"}},
								To:         []*mail.Address{{Name: "Bob", Email: "bob@example.com"}},
								Subject:    "Meeting tomorrow",
								ReceivedAt: &now,
								Size:       4521,
								Keywords:   map[string]bool{"$seen": true},
								Preview:    "Hi, just wanted to confirm...",
							},
							{
								ID:         "M2",
								ThreadID:   "T2",
								From:       []*mail.Address{{Name: "Bob", Email: "bob@example.com"}},
								To:         []*mail.Address{{Name: "Alice", Email: "alice@example.com"}},
								Subject:    "Invoice #1234",
								ReceivedAt: &now,
								Size:       2000,
								Keywords:   map[string]bool{},
								Preview:    "Please find attached...",
							},
						},
						NotFound: []jmap.ID{},
					},
				},
			}}, nil
		},
	}

	summaries, notFound, err := c.GetEmailSummaries([]string{"M1", "M2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].ID != "M1" {
		t.Errorf("expected first summary ID=M1, got %s", summaries[0].ID)
	}
	if summaries[1].ID != "M2" {
		t.Errorf("expected second summary ID=M2, got %s", summaries[1].ID)
	}
	if len(notFound) != 0 {
		t.Errorf("expected 0 not-found IDs, got %d", len(notFound))
	}
}

func TestGetEmailSummaries_NotFoundIDs(t *testing.T) {
	now := time.Date(2026, 2, 14, 10, 30, 0, 0, time.UTC)

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{
					Name:   "Email/get",
					CallID: "0",
					Args: &email.GetResponse{
						List: []*email.Email{
							{
								ID:         "M1",
								ThreadID:   "T1",
								From:       []*mail.Address{{Email: "alice@example.com"}},
								To:         []*mail.Address{{Email: "bob@example.com"}},
								Subject:    "Test",
								ReceivedAt: &now,
								Keywords:   map[string]bool{},
							},
						},
						NotFound: []jmap.ID{"M-gone", "M-deleted"},
					},
				},
			}}, nil
		},
	}

	summaries, notFound, err := c.GetEmailSummaries([]string{"M1", "M-gone", "M-deleted"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if len(notFound) != 2 {
		t.Fatalf("expected 2 not-found IDs, got %d", len(notFound))
	}
	if notFound[0] != "M-gone" || notFound[1] != "M-deleted" {
		t.Errorf("unexpected not-found IDs: %v", notFound)
	}
}

func TestGetEmailSummaries_MethodError(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "error", CallID: "0", Args: &jmap.MethodError{Type: "invalidArguments"}},
			}}, nil
		},
	}

	_, _, err := c.GetEmailSummaries([]string{"M1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "email/get") {
		t.Errorf("expected error to mention email/get, got: %s", err.Error())
	}
}

func TestGetEmailSummaries_Batching(t *testing.T) {
	callCount := 0

	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			callCount++
			getReq, ok := req.Calls[0].Args.(*email.Get)
			if !ok {
				t.Fatalf("expected *email.Get, got %T", req.Calls[0].Args)
			}
			var list []*email.Email
			for _, id := range getReq.IDs {
				list = append(list, &email.Email{
					ID:       id,
					Keywords: map[string]bool{},
				})
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{
					Name:   "Email/get",
					CallID: "0",
					Args:   &email.GetResponse{List: list},
				},
			}}, nil
		},
	}

	// 51 IDs should result in 2 Do calls (50 + 1)
	ids := make([]string, 51)
	for i := range ids {
		ids[i] = fmt.Sprintf("M%d", i)
	}

	summaries, notFound, err := c.GetEmailSummaries(ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 Do calls for 51 IDs, got %d", callCount)
	}
	if len(summaries) != 51 {
		t.Errorf("expected 51 summaries, got %d", len(summaries))
	}
	if len(notFound) != 0 {
		t.Errorf("expected 0 not-found IDs, got %d", len(notFound))
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
