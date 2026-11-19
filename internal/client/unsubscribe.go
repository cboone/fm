package client

import (
	"errors"
	"slices"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/email"

	"github.com/cboone/fm/internal/unsubscribe"
)

// ErrSenderSelectionChanged means a stable newest-message selection was unavailable.
var ErrSenderSelectionChanged = errors.New("sender selection changed; preview again")

// ErrSenderResponseInvalid identifies missing or malformed protocol metadata.
var ErrSenderResponseInvalid = errors.New("invalid sender metadata response")

// SenderLookupError preserves the query/get operation and typed cause.
type SenderLookupError struct {
	Operation string
	Err       error
}

func (e *SenderLookupError) Error() string { return "sender lookup failed" }

// Unwrap exposes the cause without including server payloads in Error.
func (e *SenderLookupError) Unwrap() error { return e.Err }

// senderResponse validates the complete decoded shape before selection consumes
// it. A malformed later sender cannot abort reporting of earlier POST results.
func (c *Client) senderResponse(req *jmap.Request, operation string) (*jmap.Response, error) {
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if err := validateSenderResponse(resp, operation, c.accountID); err != nil {
		return nil, err
	}
	return resp, nil
}

func validateSenderResponse(resp *jmap.Response, operation string, accountID jmap.ID) error {
	if resp == nil || len(resp.Responses) != 1 || resp.Responses[0] == nil {
		return ErrSenderResponseInvalid
	}
	switch r := resp.Responses[0].Args.(type) {
	case *email.QueryResponse:
		if operation != "query" || r == nil || accountID == "" || r.Account != accountID {
			return ErrSenderResponseInvalid
		}
		for _, id := range r.IDs {
			if id == "" {
				return ErrSenderResponseInvalid
			}
		}
	case *email.GetResponse:
		if operation != "get" || r == nil || accountID == "" || r.Account != accountID {
			return ErrSenderResponseInvalid
		}
		for _, id := range r.NotFound {
			if id == "" {
				return ErrSenderResponseInvalid
			}
		}
		for _, message := range r.List {
			if message == nil || message.ID == "" || message.ReceivedAt == nil {
				return ErrSenderResponseInvalid
			}
			if slices.Contains(message.From, nil) || slices.Contains(message.To, nil) || slices.Contains(message.CC, nil) {
				return ErrSenderResponseInvalid
			}
			for _, header := range message.Headers {
				if header == nil || !validSenderHeaderName(header.Name) {
					return ErrSenderResponseInvalid
				}
			}
		}
	case *jmap.MethodError:
		if r == nil || r.Type == "" {
			return ErrSenderResponseInvalid
		}
	default:
		return ErrSenderResponseInvalid
	}
	return nil
}

func validSenderHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for _, ch := range name {
		if ch < 33 || ch > 126 || ch == ':' {
			return false
		}
	}
	return true
}

type senderQueryCursor struct {
	position int64
	state    string
}

func (c *Client) senderQueryPage(query *email.Query, cursor *senderQueryCursor) ([]jmap.ID, error) {
	query.Position = cursor.position
	req := &jmap.Request{}
	req.Invoke(query)
	resp, err := c.senderResponse(req, "query")
	if err != nil {
		return nil, err
	}
	switch page := resp.Responses[0].Args.(type) {
	case *email.QueryResponse:
		if page.Position != uint64(cursor.position) || (cursor.position > 0 && page.QueryState != cursor.state) {
			return nil, ErrSenderSelectionChanged
		}
		if uint64(len(page.IDs)) > query.Limit {
			return nil, ErrSenderResponseInvalid
		}
		cursor.position += int64(len(page.IDs))
		cursor.state = page.QueryState
		return page.IDs, nil
	case *jmap.MethodError:
		return nil, page
	}
	return nil, ErrSenderResponseInvalid
}

func (c *Client) senderMessages(ids []jmap.ID, properties []string) ([]*email.Email, error) {
	req := &jmap.Request{}
	req.Invoke(&email.Get{Account: c.accountID, IDs: ids, Properties: properties})
	resp, err := c.senderResponse(req, "get")
	if err != nil {
		return nil, err
	}
	get, ok := resp.Responses[0].Args.(*email.GetResponse)
	if !ok {
		return nil, resp.Responses[0].Args.(*jmap.MethodError)
	}
	if len(get.List) != len(ids) || len(get.NotFound) != 0 {
		return nil, ErrSenderSelectionChanged
	}
	messages := make(map[jmap.ID]*email.Email, len(ids))
	for _, message := range get.List {
		messages[message.ID] = message
	}
	for i, id := range ids {
		message := messages[id]
		if message == nil {
			return nil, ErrSenderSelectionChanged
		}
		get.List[i] = message
		delete(messages, id)
	}
	return get.List, nil
}

// NewestSenderEmail pages past fuzzy From matches and reads only metadata and
// ordered headers. Query state changes abort selection instead of skipping mail.
func (c *Client) NewestSenderEmail(sender string) (*unsubscribe.Candidate, error) {
	cursor := senderQueryCursor{}
	query := &email.Query{
		Account: c.accountID, Filter: &email.FilterCondition{From: sender},
		Sort:  []*email.SortComparator{{Property: "receivedAt", IsAscending: false}},
		Limit: uint64(c.senderPageSize()),
	}
	for {
		ids, err := c.senderQueryPage(query, &cursor)
		if err != nil {
			return nil, &SenderLookupError{Operation: "query", Err: err}
		}
		if len(ids) == 0 {
			return nil, nil
		}
		messages, err := c.senderMessages(ids, []string{"id", "from", "receivedAt", "headers"})
		if err != nil {
			return nil, &SenderLookupError{Operation: "get", Err: err}
		}
		for _, message := range messages {
			matches := false
			for _, from := range message.From {
				if unsubscribe.EqualSenderAddress(from.Email, sender) {
					matches = true
				}
			}
			if !matches {
				continue
			}
			candidate := &unsubscribe.Candidate{ID: string(message.ID), ReceivedAt: safeTime(message.ReceivedAt)}
			for _, from := range message.From {
				candidate.From = append(candidate.From, from.Email)
			}
			for _, h := range message.Headers {
				candidate.Headers = append(candidate.Headers, unsubscribe.Header{Name: h.Name, Value: h.Value})
			}
			return candidate, nil
		}
	}
}
