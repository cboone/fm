package client

import (
	"fmt"
	"sort"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/searchsnippet"
	"git.sr.ht/~rockorager/go-jmap/mail/thread"

	"github.com/cboone/jm/internal/types"
)

const batchSize = 50

// summaryProperties are the Email/get properties used for list and search results.
var summaryProperties = []string{
	"id", "threadId", "mailboxIds", "from", "to",
	"subject", "receivedAt", "size", "keywords", "preview",
}

// detailProperties are the Email/get properties used for full email reads.
var detailProperties = []string{
	"id", "threadId", "mailboxIds", "from", "to", "cc", "bcc",
	"replyTo", "subject", "sentAt", "receivedAt", "size", "keywords",
	"bodyValues", "textBody", "htmlBody", "attachments", "headers",
}

// ListEmails queries emails in a mailbox and returns summaries.
func (c *Client) ListEmails(mailboxNameOrID string, limit uint64, offset int64, unreadOnly bool, sortField string, sortAsc bool) (types.EmailListResult, error) {
	mailboxID, err := c.ResolveMailboxID(mailboxNameOrID)
	if err != nil {
		return types.EmailListResult{}, err
	}

	filter := &email.FilterCondition{
		InMailbox: mailboxID,
	}
	if unreadOnly {
		filter.NotKeyword = "$seen"
	}

	req := &jmap.Request{}
	queryCallID := req.Invoke(&email.Query{
		Account:        c.accountID,
		Filter:         filter,
		Sort:           []*email.SortComparator{{Property: sortField, IsAscending: sortAsc}},
		Position:       offset,
		Limit:          limit,
		CalculateTotal: true,
	})

	req.Invoke(&email.Get{
		Account:    c.accountID,
		Properties: summaryProperties,
		ReferenceIDs: &jmap.ResultReference{
			ResultOf: queryCallID,
			Name:     "Email/query",
			Path:     "/ids",
		},
	})

	resp, err := c.Do(req)
	if err != nil {
		return types.EmailListResult{}, fmt.Errorf("email query: %w", err)
	}

	result := types.EmailListResult{Offset: offset}
	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *email.QueryResponse:
			result.Total = r.Total
		case *email.GetResponse:
			result.Emails = convertSummaries(r.List)
		case *jmap.MethodError:
			return types.EmailListResult{}, fmt.Errorf("email query: %s", r.Error())
		}
	}

	return result, nil
}

// ReadEmail retrieves the full content of an email.
func (c *Client) ReadEmail(emailID string, preferHTML bool, rawHeaders bool) (types.EmailDetail, error) {
	props := make([]string, len(detailProperties))
	copy(props, detailProperties)
	if !rawHeaders {
		// Remove "headers" from properties.
		filtered := props[:0]
		for _, p := range props {
			if p != "headers" {
				filtered = append(filtered, p)
			}
		}
		props = filtered
	}

	req := &jmap.Request{}
	get := &email.Get{
		Account:    c.accountID,
		IDs:        []jmap.ID{jmap.ID(emailID)},
		Properties: props,
		BodyProperties: []string{
			"partId", "blobId", "size", "name", "type", "charset", "disposition",
		},
	}
	// Always fetch both so extractBody can fall back between HTML and text.
	get.FetchHTMLBodyValues = true
	get.FetchTextBodyValues = true
	req.Invoke(get)

	resp, err := c.Do(req)
	if err != nil {
		return types.EmailDetail{}, fmt.Errorf("email/get: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *email.GetResponse:
			if len(r.NotFound) > 0 {
				return types.EmailDetail{}, fmt.Errorf("email %s: %w", emailID, ErrNotFound)
			}
			if len(r.List) == 0 {
				return types.EmailDetail{}, fmt.Errorf("email %s: %w", emailID, ErrNotFound)
			}
			return convertDetail(r.List[0], preferHTML, rawHeaders), nil
		case *jmap.MethodError:
			return types.EmailDetail{}, fmt.Errorf("email/get: %s", r.Error())
		}
	}

	return types.EmailDetail{}, fmt.Errorf("email/get: unexpected response")
}

// ReadThread retrieves the full thread for an email using Thread/get.
func (c *Client) ReadThread(emailID string, preferHTML bool, rawHeaders bool) (types.ThreadView, error) {
	detail, err := c.ReadEmail(emailID, preferHTML, rawHeaders)
	if err != nil {
		return types.ThreadView{}, err
	}

	if detail.ThreadID == "" {
		return types.ThreadView{
			Email:  detail,
			Thread: []types.ThreadEmail{singleThreadEntry(detail)},
		}, nil
	}

	req := &jmap.Request{}
	req.Invoke(&thread.Get{
		Account:    c.accountID,
		IDs:        []jmap.ID{jmap.ID(detail.ThreadID)},
		Properties: []string{"id", "emailIds"},
	})

	resp, err := c.Do(req)
	if err != nil {
		return types.ThreadView{}, fmt.Errorf("thread/get: %w", err)
	}

	var threadEmailIDs []jmap.ID
	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *thread.GetResponse:
			if len(r.NotFound) > 0 || len(r.List) == 0 {
				return types.ThreadView{
					Email:  detail,
					Thread: []types.ThreadEmail{singleThreadEntry(detail)},
				}, nil
			}
			threadEmailIDs = r.List[0].EmailIDs
		case *jmap.MethodError:
			return types.ThreadView{}, fmt.Errorf("thread/get: %s", r.Error())
		}
	}

	if len(threadEmailIDs) == 0 {
		return types.ThreadView{
			Email:  detail,
			Thread: []types.ThreadEmail{singleThreadEntry(detail)},
		}, nil
	}

	req = &jmap.Request{}
	req.Invoke(&email.Get{
		Account:    c.accountID,
		IDs:        threadEmailIDs,
		Properties: []string{"id", "threadId", "from", "to", "subject", "receivedAt", "preview", "keywords"},
	})

	resp, err = c.Do(req)
	if err != nil {
		return types.ThreadView{}, fmt.Errorf("thread email/get: %w", err)
	}

	var threadEmails []types.ThreadEmail
	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *email.GetResponse:
			for _, e := range r.List {
				threadEmails = append(threadEmails, types.ThreadEmail{
					ID:         string(e.ID),
					From:       convertAddresses(e.From),
					To:         convertAddresses(e.To),
					Subject:    e.Subject,
					ReceivedAt: safeTime(e.ReceivedAt),
					Preview:    e.Preview,
					IsUnread:   !e.Keywords["$seen"],
				})
			}
		case *jmap.MethodError:
			return types.ThreadView{}, fmt.Errorf("thread email/get: %s", r.Error())
		}
	}

	if len(threadEmails) == 0 {
		threadEmails = []types.ThreadEmail{singleThreadEntry(detail)}
	}

	sort.Slice(threadEmails, func(i, j int) bool {
		return threadEmails[i].ReceivedAt.Before(threadEmails[j].ReceivedAt)
	})

	return types.ThreadView{
		Email:  detail,
		Thread: threadEmails,
	}, nil
}

func singleThreadEntry(d types.EmailDetail) types.ThreadEmail {
	return types.ThreadEmail{
		ID:         d.ID,
		From:       d.From,
		To:         d.To,
		Subject:    d.Subject,
		ReceivedAt: d.ReceivedAt,
		IsUnread:   d.IsUnread,
	}
}

// SearchEmails performs a filtered search across emails.
func (c *Client) SearchEmails(opts SearchOptions) (types.EmailListResult, error) {
	filter := &email.FilterCondition{}
	if opts.Text != "" {
		filter.Text = opts.Text
	}
	if opts.From != "" {
		filter.From = opts.From
	}
	if opts.To != "" {
		filter.To = opts.To
	}
	if opts.Subject != "" {
		filter.Subject = opts.Subject
	}
	if opts.Before != nil {
		filter.Before = opts.Before
	}
	if opts.After != nil {
		filter.After = opts.After
	}
	if opts.HasAttachment {
		filter.HasAttachment = true
	}
	if opts.MailboxID != "" {
		filter.InMailbox = jmap.ID(opts.MailboxID)
	}

	req := &jmap.Request{}
	queryCallID := req.Invoke(&email.Query{
		Account:        c.accountID,
		Filter:         filter,
		Sort:           []*email.SortComparator{{Property: "receivedAt", IsAscending: false}},
		Limit:          opts.Limit,
		CalculateTotal: true,
	})

	getCallID := req.Invoke(&email.Get{
		Account:    c.accountID,
		Properties: summaryProperties,
		ReferenceIDs: &jmap.ResultReference{
			ResultOf: queryCallID,
			Name:     "Email/query",
			Path:     "/ids",
		},
	})

	// Request search snippets if doing a text search.
	hasTextSearch := opts.Text != ""
	if hasTextSearch {
		req.Invoke(&searchsnippet.Get{
			Account: c.accountID,
			Filter:  filter,
			ReferenceIDs: &jmap.ResultReference{
				ResultOf: getCallID,
				Name:     "Email/get",
				Path:     "/list/*/id",
			},
		})
	}

	resp, err := c.Do(req)
	if err != nil {
		return types.EmailListResult{}, fmt.Errorf("search: %w", err)
	}

	result := types.EmailListResult{}
	snippets := make(map[string]string)

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *email.QueryResponse:
			result.Total = r.Total
		case *email.GetResponse:
			result.Emails = convertSummaries(r.List)
		case *searchsnippet.GetResponse:
			for _, s := range r.List {
				if s.Preview != "" {
					snippets[string(s.Email)] = s.Preview
				}
			}
		case *jmap.MethodError:
			return types.EmailListResult{}, fmt.Errorf("search: %s", r.Error())
		}
	}

	// Attach snippets to emails.
	for i := range result.Emails {
		if s, ok := snippets[result.Emails[i].ID]; ok {
			result.Emails[i].Snippet = s
		}
	}

	return result, nil
}

// SearchOptions holds search filter parameters.
type SearchOptions struct {
	Text          string
	MailboxID     string
	From          string
	To            string
	Subject       string
	Before        *time.Time
	After         *time.Time
	HasAttachment bool
	Limit         uint64
}

// MoveEmails moves emails to a target mailbox by updating their mailboxIds.
// It structurally cannot destroy emails or create new ones.
func (c *Client) MoveEmails(emailIDs []string, targetMailboxID jmap.ID) (succeeded []string, errors []string) {
	succeeded = []string{}
	errors = []string{}
	for start := 0; start < len(emailIDs); start += batchSize {
		end := start + batchSize
		if end > len(emailIDs) {
			end = len(emailIDs)
		}
		batch := emailIDs[start:end]

		updates := make(map[jmap.ID]jmap.Patch)
		for _, id := range batch {
			updates[jmap.ID(id)] = jmap.Patch{
				"mailboxIds": map[jmap.ID]bool{targetMailboxID: true},
			}
		}

		req := &jmap.Request{}
		req.Invoke(&email.Set{
			Account: c.accountID,
			Update:  updates,
		})

		resp, err := c.Do(req)
		if err != nil {
			for _, id := range batch {
				errors = append(errors, fmt.Sprintf("%s: %v", id, err))
			}
			continue
		}

		for _, inv := range resp.Responses {
			switch r := inv.Args.(type) {
			case *email.SetResponse:
				for _, idStr := range batch {
					jid := jmap.ID(idStr)
					if _, ok := r.Updated[jid]; ok {
						succeeded = append(succeeded, idStr)
					} else if setErr, ok := r.NotUpdated[jid]; ok {
						desc := "unknown error"
						if setErr.Description != nil {
							desc = *setErr.Description
						}
						errors = append(errors, fmt.Sprintf("%s: %s", idStr, desc))
					}
				}
			case *jmap.MethodError:
				for _, id := range batch {
					errors = append(errors, fmt.Sprintf("%s: %s", id, r.Error()))
				}
			}
		}
	}
	return succeeded, errors
}

// MarkAsSpam moves emails to junk and sets the $junk keyword.
func (c *Client) MarkAsSpam(emailIDs []string, junkMailboxID jmap.ID) (succeeded []string, errors []string) {
	succeeded = []string{}
	errors = []string{}
	for start := 0; start < len(emailIDs); start += batchSize {
		end := start + batchSize
		if end > len(emailIDs) {
			end = len(emailIDs)
		}
		batch := emailIDs[start:end]

		updates := make(map[jmap.ID]jmap.Patch)
		for _, id := range batch {
			updates[jmap.ID(id)] = jmap.Patch{
				"mailboxIds":     map[jmap.ID]bool{junkMailboxID: true},
				"keywords/$junk": true,
			}
		}

		req := &jmap.Request{}
		req.Invoke(&email.Set{
			Account: c.accountID,
			Update:  updates,
		})

		resp, err := c.Do(req)
		if err != nil {
			for _, id := range batch {
				errors = append(errors, fmt.Sprintf("%s: %v", id, err))
			}
			continue
		}

		for _, inv := range resp.Responses {
			switch r := inv.Args.(type) {
			case *email.SetResponse:
				for _, idStr := range batch {
					jid := jmap.ID(idStr)
					if _, ok := r.Updated[jid]; ok {
						succeeded = append(succeeded, idStr)
					} else if setErr, ok := r.NotUpdated[jid]; ok {
						desc := "unknown error"
						if setErr.Description != nil {
							desc = *setErr.Description
						}
						errors = append(errors, fmt.Sprintf("%s: %s", idStr, desc))
					}
				}
			case *jmap.MethodError:
				for _, id := range batch {
					errors = append(errors, fmt.Sprintf("%s: %s", id, r.Error()))
				}
			}
		}
	}
	return succeeded, errors
}

// --- Conversion helpers ---

func convertAddresses(addrs []*mail.Address) []types.Address {
	if addrs == nil {
		return []types.Address{}
	}
	out := make([]types.Address, len(addrs))
	for i, a := range addrs {
		out[i] = types.Address{Name: a.Name, Email: a.Email}
	}
	return out
}

func convertSummaries(emails []*email.Email) []types.EmailSummary {
	out := make([]types.EmailSummary, len(emails))
	for i, e := range emails {
		out[i] = types.EmailSummary{
			ID:         string(e.ID),
			ThreadID:   string(e.ThreadID),
			From:       convertAddresses(e.From),
			To:         convertAddresses(e.To),
			Subject:    e.Subject,
			ReceivedAt: safeTime(e.ReceivedAt),
			Size:       e.Size,
			IsUnread:   !e.Keywords["$seen"],
			IsFlagged:  e.Keywords["$flagged"],
			Preview:    e.Preview,
		}
	}
	return out
}

func convertDetail(e *email.Email, preferHTML bool, rawHeaders bool) types.EmailDetail {
	body := extractBody(e, preferHTML)

	var attachments []types.Attachment
	for _, a := range e.Attachments {
		attachments = append(attachments, types.Attachment{
			Name: a.Name,
			Type: a.Type,
			Size: a.Size,
		})
	}
	if attachments == nil {
		attachments = []types.Attachment{}
	}

	detail := types.EmailDetail{
		ID:          string(e.ID),
		ThreadID:    string(e.ThreadID),
		From:        convertAddresses(e.From),
		To:          convertAddresses(e.To),
		CC:          convertAddresses(e.CC),
		BCC:         convertAddresses(e.BCC),
		ReplyTo:     convertAddresses(e.ReplyTo),
		Subject:     e.Subject,
		SentAt:      e.SentAt,
		ReceivedAt:  safeTime(e.ReceivedAt),
		IsUnread:    !e.Keywords["$seen"],
		IsFlagged:   e.Keywords["$flagged"],
		Body:        body,
		Attachments: attachments,
	}

	if rawHeaders {
		for _, h := range e.Headers {
			detail.Headers = append(detail.Headers, types.Header{
				Name:  h.Name,
				Value: h.Value,
			})
		}
	}

	return detail
}

func extractBody(e *email.Email, preferHTML bool) string {
	if preferHTML {
		for _, part := range e.HTMLBody {
			if bv, ok := e.BodyValues[part.PartID]; ok {
				return bv.Value
			}
		}
	}
	for _, part := range e.TextBody {
		if bv, ok := e.BodyValues[part.PartID]; ok {
			return bv.Value
		}
	}
	// Fall back to HTML if text is empty.
	for _, part := range e.HTMLBody {
		if bv, ok := e.BodyValues[part.PartID]; ok {
			return bv.Value
		}
	}
	return ""
}

func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
