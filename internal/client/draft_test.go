package client

import (
	"fmt"
	"strings"
	"testing"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"

	"github.com/cboone/fm/internal/types"
)

const testDraftsID = "mb-drafts"

// testClientForDraft creates a Client with a Drafts mailbox and the given doFunc.
func testClientForDraft(doFunc func(*jmap.Request) (*jmap.Response, error)) *Client {
	return &Client{
		accountID: "test-account",
		jmap: &jmap.Client{
			Session: &jmap.Session{
				Username: "user@fastmail.com",
			},
		},
		mailboxCache: []*mailbox.Mailbox{
			{ID: testDraftsID, Name: "Drafts", Role: mailbox.RoleDrafts},
			{ID: "mb-inbox", Name: "Inbox", Role: mailbox.RoleInbox},
		},
		doFunc: doFunc,
	}
}

// mockDraftCreateSuccess returns a doFunc that simulates successful draft creation.
// It validates the Email/set request and returns the created email ID.
func mockDraftCreateSuccess(createdID string) func(*jmap.Request) (*jmap.Response, error) {
	return func(req *jmap.Request) (*jmap.Response, error) {
		setReq := req.Calls[0].Args.(*email.Set)
		var createKey jmap.ID
		for k := range setReq.Create {
			createKey = k
		}
		return &jmap.Response{Responses: []*jmap.Invocation{
			{Name: "Email/set", CallID: "0", Args: &email.SetResponse{
				Created: map[jmap.ID]*email.Email{
					createKey: {ID: jmap.ID(createdID)},
				},
			}},
		}}, nil
	}
}

// --- New draft tests ---

func TestCreateDraft_NewBasic(t *testing.T) {
	c := testClientForDraft(mockDraftCreateSuccess("M-new-draft"))

	result, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com", Name: "Alice"}},
		Subject: "Hello",
		Body:    "Hi Alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "M-new-draft" {
		t.Errorf("expected ID M-new-draft, got %s", result.ID)
	}
	if result.Mode != "new" {
		t.Errorf("expected mode new, got %s", result.Mode)
	}
	if result.Subject != "Hello" {
		t.Errorf("expected subject Hello, got %s", result.Subject)
	}
	if len(result.To) != 1 || result.To[0].Email != "alice@example.com" {
		t.Errorf("unexpected To: %v", result.To)
	}
	if result.Mailbox == nil || result.Mailbox.ID != testDraftsID {
		t.Errorf("expected drafts mailbox, got: %v", result.Mailbox)
	}
	if len(result.From) != 1 || result.From[0].Email != "user@fastmail.com" {
		t.Errorf("expected From derived from session, got: %v", result.From)
	}
	if result.InReplyTo != "" {
		t.Errorf("expected empty InReplyTo for new draft, got %s", result.InReplyTo)
	}
}

func TestCreateDraft_NewHTML(t *testing.T) {
	var capturedSet *email.Set
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		capturedSet = req.Calls[0].Args.(*email.Set)
		return mockDraftCreateSuccess("M-html")(req)
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "HTML",
		Body:    "<p>Hello</p>",
		HTML:    true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, draft := range capturedSet.Create {
		if len(draft.HTMLBody) == 0 {
			t.Error("expected HTMLBody to be set for HTML mode")
		}
		if len(draft.TextBody) != 0 {
			t.Error("expected TextBody to be empty for HTML mode")
		}
	}
}

func TestCreateDraft_NewPlaintext(t *testing.T) {
	var capturedSet *email.Set
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		capturedSet = req.Calls[0].Args.(*email.Set)
		return mockDraftCreateSuccess("M-text")(req)
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "Text",
		Body:    "Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, draft := range capturedSet.Create {
		if len(draft.TextBody) == 0 {
			t.Error("expected TextBody to be set for plaintext mode")
		}
		if len(draft.HTMLBody) != 0 {
			t.Error("expected HTMLBody to be empty for plaintext mode")
		}
	}
}

func TestCreateDraft_NewWithCC(t *testing.T) {
	c := testClientForDraft(mockDraftCreateSuccess("M-cc"))

	result, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		CC:      []types.Address{{Email: "bob@example.com"}},
		Subject: "FYI",
		Body:    "Info",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.CC) != 1 || result.CC[0].Email != "bob@example.com" {
		t.Errorf("unexpected CC: %v", result.CC)
	}
}

// --- Reply tests ---

func TestCreateDraft_Reply(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			// Fetch original email.
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:        "M-orig",
						From:      []*mail.Address{{Name: "Sender", Email: "sender@example.com"}},
						To:        []*mail.Address{{Name: "Me", Email: "user@fastmail.com"}},
						Subject:   "Original subject",
						MessageID: []string{"<msg-001@example.com>"},
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-reply")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReply,
		OriginalID: "M-orig",
		Body:       "Thanks!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Mode != "reply" {
		t.Errorf("expected mode reply, got %s", result.Mode)
	}
	if result.Subject != "Re: Original subject" {
		t.Errorf("expected 'Re: Original subject', got %s", result.Subject)
	}
	if len(result.To) != 1 || result.To[0].Email != "sender@example.com" {
		t.Errorf("expected To=sender@example.com, got: %v", result.To)
	}
	if result.InReplyTo != "<msg-001@example.com>" {
		t.Errorf("expected InReplyTo, got: %s", result.InReplyTo)
	}
}

func TestCreateDraft_ReplyUsesReplyTo(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:        "M-orig",
						From:      []*mail.Address{{Email: "sender@example.com"}},
						ReplyTo:   []*mail.Address{{Email: "replies@example.com"}},
						Subject:   "Test",
						MessageID: []string{"<msg-002@example.com>"},
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-reply2")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReply,
		OriginalID: "M-orig",
		Body:       "Got it",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.To) != 1 || result.To[0].Email != "replies@example.com" {
		t.Errorf("expected To=replies@example.com, got: %v", result.To)
	}
}

func TestCreateDraft_ReplySubjectAlreadyHasRe(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:      "M-orig",
						From:    []*mail.Address{{Email: "sender@example.com"}},
						Subject: "Re: Already replied",
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-re")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReply,
		OriginalID: "M-orig",
		Body:       "Again",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Subject != "Re: Already replied" {
		t.Errorf("expected no double Re:, got %s", result.Subject)
	}
}

func TestCreateDraft_ReplySubjectOverride(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:      "M-orig",
						From:    []*mail.Address{{Email: "sender@example.com"}},
						Subject: "Original",
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-override")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReply,
		OriginalID: "M-orig",
		Subject:    "Custom subject",
		Body:       "Custom",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Subject != "Custom subject" {
		t.Errorf("expected subject override, got %s", result.Subject)
	}
}

// --- Reply-all tests ---

func TestCreateDraft_ReplyAll(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:        "M-orig",
						From:      []*mail.Address{{Email: "sender@example.com"}},
						To:        []*mail.Address{{Email: "user@fastmail.com"}, {Email: "charlie@example.com"}},
						CC:        []*mail.Address{{Email: "dave@example.com"}},
						Subject:   "Group discussion",
						MessageID: []string{"<msg-003@example.com>"},
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-replyall")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReplyAll,
		OriginalID: "M-orig",
		Body:       "My reply to all",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// To should be sender (from Reply-To fallback to From).
	if len(result.To) != 1 || result.To[0].Email != "sender@example.com" {
		t.Errorf("expected To=sender@example.com, got: %v", result.To)
	}

	// CC should include charlie and dave but not self (user@fastmail.com) and not sender (already in To).
	ccEmails := make(map[string]bool)
	for _, a := range result.CC {
		ccEmails[a.Email] = true
	}
	if !ccEmails["charlie@example.com"] {
		t.Error("expected charlie@example.com in CC")
	}
	if !ccEmails["dave@example.com"] {
		t.Error("expected dave@example.com in CC")
	}
	if ccEmails["user@fastmail.com"] {
		t.Error("self should be excluded from CC")
	}
	if ccEmails["sender@example.com"] {
		t.Error("sender should not be in CC (already in To)")
	}
}

func TestCreateDraft_ReplyAllDeduplication(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:   "M-orig",
						From: []*mail.Address{{Email: "sender@example.com"}},
						To:   []*mail.Address{{Email: "user@fastmail.com"}, {Email: "bob@example.com"}},
						CC:   []*mail.Address{{Email: "bob@example.com"}}, // duplicate
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-dedup")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReplyAll,
		OriginalID: "M-orig",
		Body:       "Dedup test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// bob should appear only once in CC.
	bobCount := 0
	for _, a := range result.CC {
		if a.Email == "bob@example.com" {
			bobCount++
		}
	}
	if bobCount != 1 {
		t.Errorf("expected bob once in CC, got %d", bobCount)
	}
}

func TestCreateDraft_ReplyAllWithUserCC(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:   "M-orig",
						From: []*mail.Address{{Email: "sender@example.com"}},
						To:   []*mail.Address{{Email: "user@fastmail.com"}},
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-usercc")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReplyAll,
		OriginalID: "M-orig",
		CC:         []types.Address{{Email: "extra@example.com"}},
		Body:       "With extra CC",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ccEmails := make(map[string]bool)
	for _, a := range result.CC {
		ccEmails[a.Email] = true
	}
	if !ccEmails["extra@example.com"] {
		t.Error("expected user-supplied CC extra@example.com")
	}
}

// --- Forward tests ---

func TestCreateDraft_Forward(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:       "M-orig",
						From:     []*mail.Address{{Email: "sender@example.com"}},
						Subject:  "FYI",
						TextBody: []*email.BodyPart{{PartID: "1"}},
						BodyValues: map[string]*email.BodyValue{
							"1": {Value: "Original body text"},
						},
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-fwd")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeForward,
		OriginalID: "M-orig",
		To:         []types.Address{{Email: "alice@example.com"}},
		Body:       "See below",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Subject != "Fwd: FYI" {
		t.Errorf("expected 'Fwd: FYI', got %s", result.Subject)
	}
	if result.InReplyTo != "" {
		t.Error("forward should not have InReplyTo")
	}
}

func TestCreateDraft_ForwardBodyIncludesOriginal(t *testing.T) {
	var capturedSet *email.Set
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:       "M-orig",
						From:     []*mail.Address{{Email: "sender@example.com"}},
						Subject:  "Info",
						TextBody: []*email.BodyPart{{PartID: "1"}},
						BodyValues: map[string]*email.BodyValue{
							"1": {Value: "Original content here"},
						},
					}},
				}},
			}}, nil
		}
		capturedSet = req.Calls[0].Args.(*email.Set)
		return mockDraftCreateSuccess("M-fwd2")(req)
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeForward,
		OriginalID: "M-orig",
		To:         []types.Address{{Email: "alice@example.com"}},
		Body:       "FYI",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, draft := range capturedSet.Create {
		bodyContent := draft.BodyValues["body"].Value
		if !strings.Contains(bodyContent, "FYI") {
			t.Error("expected user body to be included")
		}
		if !strings.Contains(bodyContent, "Original content here") {
			t.Error("expected original body to be included in forward")
		}
		if !strings.Contains(bodyContent, "Forwarded message") {
			t.Error("expected forwarded message separator")
		}
	}
}

func TestCreateDraft_ForwardSubjectAlreadyHasFwd(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:      "M-orig",
						From:    []*mail.Address{{Email: "sender@example.com"}},
						Subject: "Fwd: Already forwarded",
					}},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-nodoublefwd")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeForward,
		OriginalID: "M-orig",
		To:         []types.Address{{Email: "alice@example.com"}},
		Body:       "Again",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Subject != "Fwd: Already forwarded" {
		t.Errorf("expected no double Fwd:, got %s", result.Subject)
	}
}

// --- Threading header tests ---

func TestCreateDraft_ReplyThreadingHeaders(t *testing.T) {
	var capturedSet *email.Set
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:         "M-orig",
						From:       []*mail.Address{{Email: "sender@example.com"}},
						Subject:    "Thread",
						MessageID:  []string{"<msg-100@example.com>"},
						References: []string{"<msg-099@example.com>"},
					}},
				}},
			}}, nil
		}
		capturedSet = req.Calls[0].Args.(*email.Set)
		return mockDraftCreateSuccess("M-threaded")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReply,
		OriginalID: "M-orig",
		Body:       "In thread",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.InReplyTo != "<msg-100@example.com>" {
		t.Errorf("expected InReplyTo=<msg-100@example.com>, got %s", result.InReplyTo)
	}

	for _, draft := range capturedSet.Create {
		if len(draft.InReplyTo) != 1 || draft.InReplyTo[0] != "<msg-100@example.com>" {
			t.Errorf("expected InReplyTo on draft, got: %v", draft.InReplyTo)
		}
		if len(draft.References) != 2 {
			t.Fatalf("expected 2 references, got %d", len(draft.References))
		}
		if draft.References[0] != "<msg-099@example.com>" {
			t.Errorf("expected first ref <msg-099@example.com>, got %s", draft.References[0])
		}
		if draft.References[1] != "<msg-100@example.com>" {
			t.Errorf("expected second ref <msg-100@example.com>, got %s", draft.References[1])
		}
	}
}

func TestCreateDraft_ReplyNoMessageID(t *testing.T) {
	var capturedSet *email.Set
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if callCount == 1 {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
					List: []*email.Email{{
						ID:      "M-orig",
						From:    []*mail.Address{{Email: "sender@example.com"}},
						Subject: "No MsgID",
					}},
				}},
			}}, nil
		}
		capturedSet = req.Calls[0].Args.(*email.Set)
		return mockDraftCreateSuccess("M-nomsgid")(req)
	})

	result, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReply,
		OriginalID: "M-orig",
		Body:       "No thread",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.InReplyTo != "" {
		t.Errorf("expected empty InReplyTo, got %s", result.InReplyTo)
	}

	for _, draft := range capturedSet.Create {
		if len(draft.InReplyTo) != 0 {
			t.Errorf("expected no InReplyTo on draft, got: %v", draft.InReplyTo)
		}
		if len(draft.References) != 0 {
			t.Errorf("expected no References on draft, got: %v", draft.References)
		}
	}
}

// --- From behavior tests ---

func TestCreateDraft_FromDerivedFromSession(t *testing.T) {
	var capturedSet *email.Set
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		capturedSet = req.Calls[0].Args.(*email.Set)
		return mockDraftCreateSuccess("M-from")(req)
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "From test",
		Body:    "Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, draft := range capturedSet.Create {
		if len(draft.From) != 1 || draft.From[0].Email != "user@fastmail.com" {
			t.Errorf("expected From=user@fastmail.com, got: %v", draft.From)
		}
	}
}

func TestCreateDraft_NoFromWhenUsernameNotEmail(t *testing.T) {
	var capturedSet *email.Set
	c := &Client{
		accountID: "test-account",
		jmap: &jmap.Client{
			Session: &jmap.Session{
				Username: "admin-user",
			},
		},
		mailboxCache: []*mailbox.Mailbox{
			{ID: testDraftsID, Name: "Drafts", Role: mailbox.RoleDrafts},
		},
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			capturedSet = req.Calls[0].Args.(*email.Set)
			return mockDraftCreateSuccess("M-nofrom")(req)
		},
	}

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "No From",
		Body:    "Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, draft := range capturedSet.Create {
		if len(draft.From) != 0 {
			t.Errorf("expected no From when username is not an email, got: %v", draft.From)
		}
	}
}

// --- Safety validation tests ---

func TestCreateDraft_MailboxIDsAndKeywords(t *testing.T) {
	var capturedSet *email.Set
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		capturedSet = req.Calls[0].Args.(*email.Set)
		return mockDraftCreateSuccess("M-safe")(req)
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "Safety check",
		Body:    "Body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, draft := range capturedSet.Create {
		if !draft.MailboxIDs[jmap.ID(testDraftsID)] {
			t.Error("expected draft to be in Drafts mailbox")
		}
		if len(draft.MailboxIDs) != 1 {
			t.Errorf("expected exactly 1 mailbox, got %d", len(draft.MailboxIDs))
		}
		if !draft.Keywords["$draft"] {
			t.Error("expected $draft keyword")
		}
		if !draft.Keywords["$seen"] {
			t.Error("expected $seen keyword")
		}
	}
}

// --- Error case tests ---

func TestCreateDraft_NoDraftsMailbox(t *testing.T) {
	c := &Client{
		accountID:    "test-account",
		mailboxCache: []*mailbox.Mailbox{{ID: "mb-inbox", Name: "Inbox", Role: mailbox.RoleInbox}},
	}

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "Test",
		Body:    "Body",
	})
	if err == nil {
		t.Fatal("expected error when Drafts mailbox missing")
	}
	if !strings.Contains(err.Error(), "drafts") {
		t.Errorf("expected error about drafts, got: %v", err)
	}
}

func TestCreateDraft_OriginalNotFound(t *testing.T) {
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		return &jmap.Response{Responses: []*jmap.Invocation{
			{Name: "Email/get", CallID: "0", Args: &email.GetResponse{
				NotFound: []jmap.ID{"M-missing"},
			}},
		}}, nil
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:       DraftModeReply,
		OriginalID: "M-missing",
		Body:       "Reply",
	})
	if err == nil {
		t.Fatal("expected error for missing original")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestCreateDraft_ServerRejectsCreate(t *testing.T) {
	callCount := 0
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		callCount++
		if _, ok := req.Calls[0].Args.(*email.Set); ok {
			desc := "server says no"
			setReq := req.Calls[0].Args.(*email.Set)
			var createKey jmap.ID
			for k := range setReq.Create {
				createKey = k
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Email/set", CallID: "0", Args: &email.SetResponse{
					NotCreated: map[jmap.ID]*jmap.SetError{
						createKey: {Type: "invalidProperties", Description: &desc},
					},
				}},
			}}, nil
		}
		return mockDraftCreateSuccess("M-fail")(req)
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "Fail",
		Body:    "Body",
	})
	if err == nil {
		t.Fatal("expected error for NotCreated")
	}
	if !strings.Contains(err.Error(), "server says no") {
		t.Errorf("expected server error message, got: %v", err)
	}
}

func TestCreateDraft_DoError(t *testing.T) {
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		return nil, fmt.Errorf("network failure")
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "Net fail",
		Body:    "Body",
	})
	if err == nil {
		t.Fatal("expected error for Do failure")
	}
}

func TestCreateDraft_MethodError(t *testing.T) {
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		return &jmap.Response{Responses: []*jmap.Invocation{
			{Name: "error", CallID: "0", Args: &jmap.MethodError{Type: "serverFail"}},
		}}, nil
	})

	_, err := c.CreateDraft(DraftOptions{
		Mode:    DraftModeNew,
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "Method err",
		Body:    "Body",
	})
	if err == nil {
		t.Fatal("expected error for MethodError")
	}
}

// --- Helper tests ---

func TestReplySubject(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Hello", "Re: Hello"},
		{"Re: Hello", "Re: Hello"},
		{"re: hello", "re: hello"},
		{"RE: HELLO", "RE: HELLO"},
		{"", "Re: "},
		{"  Re: spaced  ", "Re: spaced"},
	}
	for _, tc := range tests {
		got := replySubject(tc.input)
		if got != tc.want {
			t.Errorf("replySubject(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestForwardSubject(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Hello", "Fwd: Hello"},
		{"Fwd: Hello", "Fwd: Hello"},
		{"fwd: hello", "fwd: hello"},
		{"FWD: HELLO", "FWD: HELLO"},
	}
	for _, tc := range tests {
		got := forwardSubject(tc.input)
		if got != tc.want {
			t.Errorf("forwardSubject(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestDedup(t *testing.T) {
	tests := []struct {
		input []string
		want  []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{[]string{}, nil},
		{nil, nil},
	}
	for _, tc := range tests {
		got := dedup(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("dedup(%v) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("dedup(%v)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestAppendDedup(t *testing.T) {
	base := []*mail.Address{
		{Email: "alice@example.com"},
		{Email: "bob@example.com"},
	}
	additional := []*mail.Address{
		{Email: "Bob@Example.com"}, // duplicate (case-insensitive)
		{Email: "charlie@example.com"},
	}
	result := appendDedup(base, additional)
	if len(result) != 3 {
		t.Fatalf("expected 3 addresses, got %d", len(result))
	}
	if result[2].Email != "charlie@example.com" {
		t.Errorf("expected charlie, got %s", result[2].Email)
	}
}

func TestUnknownDraftMode(t *testing.T) {
	c := testClientForDraft(mockDraftCreateSuccess("M-bad"))

	_, err := c.CreateDraft(DraftOptions{
		Mode:    "bogus",
		To:      []types.Address{{Email: "alice@example.com"}},
		Subject: "Bad",
		Body:    "Body",
	})
	if err == nil {
		t.Fatal("expected error for unknown draft mode")
	}
}
