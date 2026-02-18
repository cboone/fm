package client

import (
	"fmt"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"

	"github.com/cboone/fm/internal/types"
)

// DraftMode indicates the type of draft composition.
type DraftMode string

const (
	DraftModeNew      DraftMode = "new"
	DraftModeReply    DraftMode = "reply"
	DraftModeReplyAll DraftMode = "reply-all"
	DraftModeForward  DraftMode = "forward"
)

// DraftOptions holds parameters for creating a draft email.
type DraftOptions struct {
	Mode       DraftMode
	To         []types.Address
	CC         []types.Address
	BCC        []types.Address
	Subject    string
	Body       string
	HTML       bool
	OriginalID string // email ID for reply/reply-all/forward
}

// replyProperties are the Email/get properties needed for composing a reply.
var replyProperties = []string{
	"id", "from", "to", "cc", "replyTo", "subject",
	"messageId", "references", "inReplyTo",
	"bodyValues", "textBody", "htmlBody",
}

// CreateDraft creates a draft email in the Drafts mailbox.
func (c *Client) CreateDraft(opts DraftOptions) (types.DraftResult, error) {
	draftsMB, err := c.GetMailboxByRole(mailbox.RoleDrafts)
	if err != nil {
		return types.DraftResult{}, fmt.Errorf("drafts mailbox not found: %w", err)
	}

	draftsID := draftsMB.ID

	// Derive From from the session username.
	var fromAddrs []*mail.Address
	if c.jmap != nil && c.jmap.Session != nil && c.jmap.Session.Username != "" {
		username := c.jmap.Session.Username
		if strings.Contains(username, "@") {
			fromAddrs = []*mail.Address{{Email: username}}
		}
	}

	var (
		toAddrs  []*mail.Address
		ccAddrs  []*mail.Address
		bccAddrs []*mail.Address
		subject  string
		body     string
		replyTo  []string
		refs     []string
	)

	switch opts.Mode {
	case DraftModeNew:
		toAddrs = toJMAPAddresses(opts.To)
		ccAddrs = toJMAPAddresses(opts.CC)
		bccAddrs = toJMAPAddresses(opts.BCC)
		subject = opts.Subject
		body = opts.Body

	case DraftModeReply, DraftModeReplyAll:
		orig, fetchErr := c.fetchOriginalForReply(opts.OriginalID)
		if fetchErr != nil {
			return types.DraftResult{}, fetchErr
		}

		// Compute To: use ReplyTo if present, else From.
		baseTo := orig.ReplyTo
		if len(baseTo) == 0 {
			baseTo = orig.From
		}
		toAddrs = appendDedup(baseTo, toJMAPAddresses(opts.To))

		if opts.Mode == DraftModeReplyAll {
			// CC = original To + CC, minus self, minus anyone already in To.
			selfEmail := ""
			if len(fromAddrs) > 0 {
				selfEmail = strings.ToLower(fromAddrs[0].Email)
			}
			toSet := addrSet(toAddrs)
			var candidates []*mail.Address
			candidates = append(candidates, orig.To...)
			candidates = append(candidates, orig.CC...)
			for _, a := range candidates {
				lower := strings.ToLower(a.Email)
				if lower == selfEmail {
					continue
				}
				if toSet[lower] {
					continue
				}
				ccAddrs = append(ccAddrs, a)
				toSet[lower] = true
			}
			ccAddrs = appendDedup(ccAddrs, toJMAPAddresses(opts.CC))
		}

		bccAddrs = toJMAPAddresses(opts.BCC)

		subject = replySubject(orig.Subject)
		if opts.Subject != "" {
			subject = opts.Subject
		}
		body = opts.Body

		// Threading headers.
		if len(orig.MessageID) > 0 {
			replyTo = orig.MessageID
			refs = dedup(append(orig.References, orig.MessageID...))
		}

	case DraftModeForward:
		orig, fetchErr := c.fetchOriginalForReply(opts.OriginalID)
		if fetchErr != nil {
			return types.DraftResult{}, fetchErr
		}

		toAddrs = toJMAPAddresses(opts.To)
		ccAddrs = toJMAPAddresses(opts.CC)
		bccAddrs = toJMAPAddresses(opts.BCC)

		subject = forwardSubject(orig.Subject)
		if opts.Subject != "" {
			subject = opts.Subject
		}

		// Prepend quoted original to body.
		origBody := extractBody(orig, false)
		body = opts.Body + "\n\n---------- Forwarded message ----------\n" + origBody

	default:
		return types.DraftResult{}, fmt.Errorf("unknown draft mode: %s", opts.Mode)
	}

	// Construct the draft email.
	draft := &email.Email{
		MailboxIDs: map[jmap.ID]bool{draftsID: true},
		Keywords:   map[string]bool{"$draft": true, "$seen": true},
		To:         toAddrs,
		CC:         ccAddrs,
		BCC:        bccAddrs,
		Subject:    subject,
	}

	if len(fromAddrs) > 0 {
		draft.From = fromAddrs
	}

	if len(replyTo) > 0 {
		draft.InReplyTo = replyTo
	}
	if len(refs) > 0 {
		draft.References = refs
	}

	// Set body content.
	bodyPartID := "body"
	if opts.HTML {
		draft.HTMLBody = []*email.BodyPart{{PartID: bodyPartID, Type: "text/html"}}
	} else {
		draft.TextBody = []*email.BodyPart{{PartID: bodyPartID, Type: "text/plain"}}
	}
	draft.BodyValues = map[string]*email.BodyValue{
		bodyPartID: {Value: body},
	}

	// Validate and execute.
	createID := jmap.ID("draft-0")
	set := &email.Set{
		Account: c.accountID,
		Create:  map[jmap.ID]*email.Email{createID: draft},
	}

	if err := ValidateSetForDraft(set, draftsID); err != nil {
		return types.DraftResult{}, err
	}

	req := &jmap.Request{}
	req.Invoke(set)

	resp, err := c.Do(req)
	if err != nil {
		return types.DraftResult{}, fmt.Errorf("draft creation: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *email.SetResponse:
			if created, ok := r.Created[createID]; ok {
				result := types.DraftResult{
					ID:      string(created.ID),
					Mode:    string(opts.Mode),
					Mailbox: &types.DestinationInfo{ID: string(draftsID), Name: draftsMB.Name},
					From:    convertAddresses(fromAddrs),
					To:      convertAddresses(toAddrs),
					CC:      convertAddresses(ccAddrs),
					Subject: subject,
				}
				if len(replyTo) > 0 {
					result.InReplyTo = strings.Join(replyTo, ", ")
				}
				return result, nil
			}
			if setErr, ok := r.NotCreated[createID]; ok {
				desc := "unknown error"
				if setErr.Description != nil {
					desc = *setErr.Description
				}
				return types.DraftResult{}, fmt.Errorf("draft creation failed: %s", desc)
			}
		case *jmap.MethodError:
			return types.DraftResult{}, fmt.Errorf("draft creation: %s", r.Error())
		}
	}

	return types.DraftResult{}, fmt.Errorf("draft creation: unexpected response")
}

// fetchOriginalForReply fetches the original email needed for reply/reply-all/forward.
func (c *Client) fetchOriginalForReply(emailID string) (*email.Email, error) {
	req := &jmap.Request{}
	get := &email.Get{
		Account:    c.accountID,
		IDs:        []jmap.ID{jmap.ID(emailID)},
		Properties: replyProperties,
	}
	get.FetchTextBodyValues = true
	get.FetchHTMLBodyValues = true
	req.Invoke(get)

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching original email: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *email.GetResponse:
			if len(r.NotFound) > 0 {
				return nil, fmt.Errorf("original email %s: %w", emailID, ErrNotFound)
			}
			if len(r.List) == 0 {
				return nil, fmt.Errorf("original email %s: %w", emailID, ErrNotFound)
			}
			return r.List[0], nil
		case *jmap.MethodError:
			return nil, fmt.Errorf("fetching original email: %s", r.Error())
		}
	}
	return nil, fmt.Errorf("fetching original email: unexpected response")
}

// --- address helpers ---

func toJMAPAddresses(addrs []types.Address) []*mail.Address {
	if len(addrs) == 0 {
		return nil
	}
	out := make([]*mail.Address, len(addrs))
	for i, a := range addrs {
		out[i] = &mail.Address{Name: a.Name, Email: a.Email}
	}
	return out
}

// appendDedup appends additional addresses to base, skipping duplicates by email (case-insensitive).
func appendDedup(base, additional []*mail.Address) []*mail.Address {
	seen := addrSet(base)
	result := make([]*mail.Address, len(base))
	copy(result, base)
	for _, a := range additional {
		lower := strings.ToLower(a.Email)
		if seen[lower] {
			continue
		}
		result = append(result, a)
		seen[lower] = true
	}
	return result
}

func addrSet(addrs []*mail.Address) map[string]bool {
	m := make(map[string]bool, len(addrs))
	for _, a := range addrs {
		m[strings.ToLower(a.Email)] = true
	}
	return m
}

func replySubject(original string) string {
	trimmed := strings.TrimSpace(original)
	if strings.HasPrefix(strings.ToLower(trimmed), "re:") {
		return trimmed
	}
	return "Re: " + trimmed
}

func forwardSubject(original string) string {
	trimmed := strings.TrimSpace(original)
	if strings.HasPrefix(strings.ToLower(trimmed), "fwd:") {
		return trimmed
	}
	return "Fwd: " + trimmed
}

// dedup returns a slice with duplicates removed, preserving order.
func dedup(items []string) []string {
	seen := make(map[string]bool, len(items))
	var result []string
	for _, s := range items {
		if seen[s] {
			continue
		}
		seen[s] = true
		result = append(result, s)
	}
	return result
}
