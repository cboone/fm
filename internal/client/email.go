package client

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/searchsnippet"
	"git.sr.ht/~rockorager/go-jmap/mail/thread"

	"github.com/cboone/fm/internal/types"
)

// searchSnippetGet wraps searchsnippet.Get to fix the incorrect method name
// in go-jmap v0.5.3 (returns "Mailbox/get" instead of "SearchSnippet/get").
type searchSnippetGet struct {
	searchsnippet.Get
}

func (m *searchSnippetGet) Name() string { return "SearchSnippet/get" }

func (m *searchSnippetGet) Requires() []jmap.URI { return []jmap.URI{mail.URI} }

// batchSetEmails executes Email/set in server-aware batches.
// patchFn builds the jmap.Patch for a single email ID.
func (c *Client) batchSetEmails(emailIDs []string, patchFn func(string) jmap.Patch) (succeeded, errors []string) {
	size := c.maxBatchSize()
	succeeded = []string{}
	errors = []string{}

	for start := 0; start < len(emailIDs); start += size {
		end := start + size
		if end > len(emailIDs) {
			end = len(emailIDs)
		}
		batch := emailIDs[start:end]

		updates := make(map[jmap.ID]jmap.Patch, len(batch))
		for _, id := range batch {
			updates[jmap.ID(id)] = patchFn(id)
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
					} else {
						errors = append(errors, fmt.Sprintf("%s: no status returned by server", idStr))
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
func (c *Client) ListEmails(mailboxNameOrID string, limit uint64, offset int64, unreadOnly bool, flaggedOnly bool, unflaggedOnly bool, sortField string, sortAsc bool) (types.EmailListResult, error) {
	mailboxID, err := c.ResolveMailboxID(mailboxNameOrID)
	if err != nil {
		return types.EmailListResult{}, err
	}

	fc := &email.FilterCondition{
		InMailbox: mailboxID,
	}
	if unreadOnly {
		fc.NotKeyword = "$seen"
	}
	if flaggedOnly {
		fc.HasKeyword = "$flagged"
	}

	// Build the final filter. When both unflaggedOnly and unreadOnly are set,
	// they each need a separate NotKeyword field, so we must use a compound
	// FilterOperator with AND. InMailbox stays on the first FilterCondition
	// inside the operator so the mailbox scope is preserved.
	var filter email.Filter = fc
	if unflaggedOnly {
		if unreadOnly {
			filter = &email.FilterOperator{
				Operator:   jmap.OperatorAND,
				Conditions: []email.Filter{fc, &email.FilterCondition{NotKeyword: "$flagged"}},
			}
		} else {
			fc.NotKeyword = "$flagged"
		}
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

// buildSearchFilter constructs an email.Filter from SearchOptions.
func buildSearchFilter(opts SearchOptions) email.Filter {
	fc := &email.FilterCondition{}
	if opts.Text != "" {
		fc.Text = opts.Text
	}
	if opts.From != "" {
		fc.From = opts.From
	}
	if opts.To != "" {
		fc.To = opts.To
	}
	if opts.Subject != "" {
		fc.Subject = opts.Subject
	}
	if opts.Before != nil {
		fc.Before = opts.Before
	}
	if opts.After != nil {
		fc.After = opts.After
	}
	if opts.HasAttachment {
		fc.HasAttachment = true
	}
	if opts.UnreadOnly {
		fc.NotKeyword = "$seen"
	}
	if opts.MailboxID != "" {
		fc.InMailbox = jmap.ID(opts.MailboxID)
	}
	if opts.FlaggedOnly {
		fc.HasKeyword = "$flagged"
	}

	// When both UnflaggedOnly and UnreadOnly are set, they each need a
	// separate NotKeyword field, so we must use a compound FilterOperator
	// with AND.
	var filter email.Filter = fc
	if opts.UnflaggedOnly {
		if opts.UnreadOnly {
			filter = &email.FilterOperator{
				Operator:   jmap.OperatorAND,
				Conditions: []email.Filter{fc, &email.FilterCondition{NotKeyword: "$flagged"}},
			}
		} else {
			fc.NotKeyword = "$flagged"
		}
	}

	return filter
}

// SearchEmails performs a filtered search across emails.
func (c *Client) SearchEmails(opts SearchOptions) (types.EmailListResult, error) {
	filter := buildSearchFilter(opts)

	sortField := opts.SortField
	if sortField == "" {
		sortField = "receivedAt"
	}

	// Track call IDs so we can identify which method failed in errors.
	callMethods := make(map[string]string) // callID -> method name

	req := &jmap.Request{}
	queryCallID := req.Invoke(&email.Query{
		Account:        c.accountID,
		Filter:         filter,
		Sort:           []*email.SortComparator{{Property: sortField, IsAscending: opts.SortAsc}},
		Position:       opts.Offset,
		Limit:          opts.Limit,
		CalculateTotal: true,
	})
	callMethods[queryCallID] = "Email/query"

	getCallID := req.Invoke(&email.Get{
		Account:    c.accountID,
		Properties: summaryProperties,
		ReferenceIDs: &jmap.ResultReference{
			ResultOf: queryCallID,
			Name:     "Email/query",
			Path:     "/ids",
		},
	})
	callMethods[getCallID] = "Email/get"

	// Request search snippets if doing a text search.
	hasTextSearch := opts.Text != ""
	var snippetCallID string
	if hasTextSearch {
		snippetCallID = req.Invoke(&searchSnippetGet{searchsnippet.Get{
			Account: c.accountID,
			Filter:  filter,
			ReferenceIDs: &jmap.ResultReference{
				ResultOf: queryCallID,
				Name:     "Email/query",
				Path:     "/ids",
			},
		}})
		callMethods[snippetCallID] = "SearchSnippet/get"
	}

	resp, err := c.Do(req)
	if err != nil {
		return types.EmailListResult{}, fmt.Errorf("search: %w", err)
	}

	result := types.EmailListResult{Offset: opts.Offset}
	snippets := make(map[string]string)

	for i, inv := range resp.Responses {
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
			method := callMethods[inv.CallID]
			if method == "" {
				method = inv.Name
			}
			// If snippets fail but query+get succeeded, degrade gracefully.
			if hasTextSearch && (method == "SearchSnippet/get" || inv.CallID == snippetCallID) {
				continue
			}
			if method == "" || method == "error" {
				method = "unknown"
			}

			callRef := inv.CallID
			if callRef == "" {
				callRef = fmt.Sprintf("%d", i)
			}

			if method == "unknown" {
				return types.EmailListResult{}, fmt.Errorf("search: call %s returned %s", callRef, r.Error())
			}
			return types.EmailListResult{}, fmt.Errorf("search: %s (call %s) returned %s", method, callRef, r.Error())
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
	UnreadOnly    bool
	FlaggedOnly   bool
	UnflaggedOnly bool
	Limit         uint64
	Offset        int64
	SortField     string
	SortAsc       bool
}

const defaultQueryPageSize = 250

// QueryEmailIDs runs Email/query with the given filter options and paginates
// through the full result set, returning all matching email IDs.
// It ignores Limit, Offset, SortField, and SortAsc from opts.
func (c *Client) QueryEmailIDs(opts SearchOptions) ([]string, error) {
	filter := buildSearchFilter(opts)

	var collected []string
	for {
		req := &jmap.Request{}
		req.Invoke(&email.Query{
			Account:        c.accountID,
			Filter:         filter,
			Position:       int64(len(collected)),
			Limit:          defaultQueryPageSize,
			CalculateTotal: true,
		})

		resp, err := c.Do(req)
		if err != nil {
			return nil, fmt.Errorf("email/query: %w", err)
		}

		var pageIDs []jmap.ID
		var total uint64
		for _, inv := range resp.Responses {
			switch r := inv.Args.(type) {
			case *email.QueryResponse:
				pageIDs = r.IDs
				total = r.Total
			case *jmap.MethodError:
				return nil, fmt.Errorf("email/query: %s", r.Error())
			}
		}

		for _, id := range pageIDs {
			collected = append(collected, string(id))
		}

		if uint64(len(collected)) >= total || len(pageIDs) == 0 {
			break
		}
	}

	return collected, nil
}

// StatsOptions holds parameters for sender aggregation.
type StatsOptions struct {
	MailboxID     string
	UnreadOnly    bool
	FlaggedOnly   bool
	UnflaggedOnly bool
	Subjects      bool
}

// statsProperties are the minimal Email/get properties for aggregation.
var statsProperties = []string{"id", "from", "subject"}

// AggregateEmailsBySender queries all matching emails and returns per-sender counts.
func (c *Client) AggregateEmailsBySender(opts StatsOptions) (types.StatsResult, error) {
	fc := &email.FilterCondition{
		InMailbox: jmap.ID(opts.MailboxID),
	}
	if opts.UnreadOnly {
		fc.NotKeyword = "$seen"
	}
	if opts.FlaggedOnly {
		fc.HasKeyword = "$flagged"
	}

	var filter email.Filter = fc
	if opts.UnflaggedOnly {
		if opts.UnreadOnly {
			filter = &email.FilterOperator{
				Operator:   jmap.OperatorAND,
				Conditions: []email.Filter{fc, &email.FilterCondition{NotKeyword: "$flagged"}},
			}
		} else {
			fc.NotKeyword = "$flagged"
		}
	}

	type senderAcc struct {
		count    int
		name     string
		subjects map[string]bool
	}
	accum := make(map[string]*senderAcc)
	var total uint64
	var position int64

	for {
		req := &jmap.Request{}
		queryCallID := req.Invoke(&email.Query{
			Account:        c.accountID,
			Filter:         filter,
			Sort:           []*email.SortComparator{{Property: "receivedAt", IsAscending: false}},
			Position:       position,
			Limit:          500,
			CalculateTotal: true,
		})

		req.Invoke(&email.Get{
			Account:    c.accountID,
			Properties: statsProperties,
			ReferenceIDs: &jmap.ResultReference{
				ResultOf: queryCallID,
				Name:     "Email/query",
				Path:     "/ids",
			},
		})

		resp, err := c.Do(req)
		if err != nil {
			return types.StatsResult{}, fmt.Errorf("stats query: %w", err)
		}

		var pageIDs []jmap.ID
		var emails []*email.Email
		for _, inv := range resp.Responses {
			switch r := inv.Args.(type) {
			case *email.QueryResponse:
				if position == 0 {
					total = r.Total
				}
				pageIDs = r.IDs
			case *email.GetResponse:
				emails = r.List
			case *jmap.MethodError:
				return types.StatsResult{}, fmt.Errorf("stats query: %s", r.Error())
			}
		}

		for _, e := range emails {
			if len(e.From) == 0 || e.From[0].Email == "" {
				continue
			}
			key := strings.ToLower(e.From[0].Email)
			acc, ok := accum[key]
			if !ok {
				acc = &senderAcc{subjects: make(map[string]bool)}
				accum[key] = acc
			}
			acc.count++
			if e.From[0].Name != "" && acc.name == "" {
				acc.name = e.From[0].Name
			}
			if opts.Subjects && e.Subject != "" {
				acc.subjects[e.Subject] = true
			}
		}

		position += int64(len(pageIDs))
		if uint64(position) >= total || len(pageIDs) == 0 {
			break
		}
	}

	senders := make([]types.SenderStat, 0, len(accum))
	for addr, acc := range accum {
		stat := types.SenderStat{
			Email: addr,
			Name:  acc.name,
			Count: acc.count,
		}
		if opts.Subjects && len(acc.subjects) > 0 {
			subjects := make([]string, 0, len(acc.subjects))
			for s := range acc.subjects {
				subjects = append(subjects, s)
			}
			sort.Strings(subjects)
			stat.Subjects = subjects
		}
		senders = append(senders, stat)
	}

	sort.Slice(senders, func(i, j int) bool {
		if senders[i].Count != senders[j].Count {
			return senders[i].Count > senders[j].Count
		}
		return senders[i].Email < senders[j].Email
	})

	return types.StatsResult{
		Total:   total,
		Senders: senders,
	}, nil
}

// MoveEmails moves emails to a target mailbox by updating their mailboxIds.
// It structurally cannot destroy emails or create new ones.
func (c *Client) MoveEmails(emailIDs []string, targetMailboxID jmap.ID) ([]string, []string) {
	return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
		return jmap.Patch{"mailboxIds": map[jmap.ID]bool{targetMailboxID: true}}
	})
}

// MarkAsSpam moves emails to junk and sets the $junk keyword.
func (c *Client) MarkAsSpam(emailIDs []string, junkMailboxID jmap.ID) ([]string, []string) {
	return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
		return jmap.Patch{
			"mailboxIds":     map[jmap.ID]bool{junkMailboxID: true},
			"keywords/$junk": true,
		}
	})
}

// MarkAsRead sets the $seen keyword on emails.
func (c *Client) MarkAsRead(emailIDs []string) ([]string, []string) {
	return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
		return jmap.Patch{"keywords/$seen": true}
	})
}

// SetFlagged sets the $flagged keyword on emails.
func (c *Client) SetFlagged(emailIDs []string) ([]string, []string) {
	return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
		return jmap.Patch{"keywords/$flagged": true}
	})
}

// SetFlaggedWithColor sets the $flagged keyword and the color bits on emails.
func (c *Client) SetFlaggedWithColor(emailIDs []string, color FlagColor) ([]string, []string) {
	return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
		p := jmap.Patch{"keywords/$flagged": true}
		for k, v := range color.Patch() {
			p[k] = v
		}
		return p
	})
}

// SetUnflagged removes the $flagged keyword and clears all color bits from emails.
// Per the IETF MailFlagBit spec, color bits should be cleared when unflagging.
func (c *Client) SetUnflagged(emailIDs []string) ([]string, []string) {
	return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
		p := jmap.Patch{"keywords/$flagged": nil}
		for k, v := range clearColorPatch() {
			p[k] = v
		}
		return p
	})
}

// ClearFlagColor removes the color bits from emails without changing the $flagged keyword.
func (c *Client) ClearFlagColor(emailIDs []string) ([]string, []string) {
	return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
		return clearColorPatch()
	})
}

// GetEmailSummaries fetches summaries for the given IDs using read-only Email/get.
// Returns found summaries, not-found IDs, and any error.
func (c *Client) GetEmailSummaries(ids []string) ([]types.EmailSummary, []string, error) {
	var allSummaries []types.EmailSummary
	var allNotFound []string

	size := c.maxBatchSize()
	for start := 0; start < len(ids); start += size {
		end := start + size
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[start:end]

		jmapIDs := make([]jmap.ID, len(batch))
		for i, id := range batch {
			jmapIDs[i] = jmap.ID(id)
		}

		req := &jmap.Request{}
		req.Invoke(&email.Get{
			Account:    c.accountID,
			IDs:        jmapIDs,
			Properties: summaryProperties,
		})

		resp, err := c.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("email/get: %w", err)
		}

		for _, inv := range resp.Responses {
			switch r := inv.Args.(type) {
			case *email.GetResponse:
				allSummaries = append(allSummaries, convertSummaries(r.List)...)
				for _, nf := range r.NotFound {
					allNotFound = append(allNotFound, string(nf))
				}
			case *jmap.MethodError:
				return nil, nil, fmt.Errorf("email/get: %s", r.Error())
			}
		}
	}

	return allSummaries, allNotFound, nil
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

	for _, h := range e.Headers {
		if strings.EqualFold(h.Name, "List-Unsubscribe") {
			detail.ListUnsubscribe = strings.TrimSpace(h.Value)
		} else if strings.EqualFold(h.Name, "List-Unsubscribe-Post") {
			detail.ListUnsubscribePost = strings.TrimSpace(h.Value)
		}
		if rawHeaders {
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
