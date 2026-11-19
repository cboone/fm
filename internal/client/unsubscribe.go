package client

import (
	"errors"

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
			for _, from := range message.From {
				if from == nil {
					return ErrSenderResponseInvalid
				}
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

// NewestSenderEmail pages past fuzzy From matches and reads only metadata and
// ordered headers. Query state changes abort selection instead of skipping mail.
func (c *Client) NewestSenderEmail(sender string) (*unsubscribe.Candidate, error) {
	var position int64
	var state string
	for {
		req := &jmap.Request{}
		req.Invoke(&email.Query{
			Account: c.accountID, Filter: &email.FilterCondition{From: sender},
			Sort:     []*email.SortComparator{{Property: "receivedAt", IsAscending: false}},
			Position: position, Limit: 50,
		})
		resp, err := c.senderResponse(req, "query")
		if err != nil {
			return nil, &SenderLookupError{Operation: "query", Err: err}
		}
		var page *email.QueryResponse
		for _, inv := range resp.Responses {
			switch r := inv.Args.(type) {
			case *email.QueryResponse:
				page = r
			case *jmap.MethodError:
				return nil, &SenderLookupError{Operation: "query", Err: r}
			}
		}
		if page.Position != uint64(position) || (position > 0 && page.QueryState != state) {
			return nil, &SenderLookupError{Operation: "query", Err: ErrSenderSelectionChanged}
		}
		state = page.QueryState
		if len(page.IDs) == 0 {
			return nil, nil
		}
		req = &jmap.Request{}
		req.Invoke(&email.Get{Account: c.accountID, IDs: page.IDs, Properties: []string{"id", "from", "receivedAt", "headers"}})
		resp, err = c.senderResponse(req, "get")
		if err != nil {
			return nil, &SenderLookupError{Operation: "get", Err: err}
		}
		messages := make(map[jmap.ID]*email.Email)
		for _, inv := range resp.Responses {
			switch r := inv.Args.(type) {
			case *email.GetResponse:
				for _, message := range r.List {
					messages[message.ID] = message
				}
			case *jmap.MethodError:
				return nil, &SenderLookupError{Operation: "get", Err: r}
			}
		}
		for _, id := range page.IDs {
			message := messages[id]
			if message == nil {
				return nil, &SenderLookupError{Operation: "get", Err: ErrSenderSelectionChanged}
			}
			matches := false
			for _, from := range message.From {
				if unsubscribe.EqualSenderAddress(from.Email, sender) {
					matches = true
				}
			}
			if !matches {
				continue
			}
			candidate := &unsubscribe.Candidate{ID: string(id), ReceivedAt: safeTime(message.ReceivedAt)}
			for _, from := range message.From {
				candidate.From = append(candidate.From, from.Email)
			}
			for _, h := range message.Headers {
				candidate.Headers = append(candidate.Headers, unsubscribe.Header{Name: h.Name, Value: h.Value})
			}
			return candidate, nil
		}
		position += int64(len(page.IDs))
	}
}
