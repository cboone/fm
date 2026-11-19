package client

import (
	"fmt"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/core"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/searchsnippet"

	"github.com/cboone/fm/internal/types"
	"github.com/cboone/fm/internal/unsubscribe"
)

func (opts SearchOptions) hasExactSender() bool {
	return opts.FromExact != "" || opts.FromDomain != ""
}

func (opts SearchOptions) matchesSender(address string) bool {
	at := strings.LastIndexByte(address, '@')
	if at <= 0 || at == len(address)-1 {
		return false
	}
	return (opts.FromExact == "" || unsubscribe.EqualSenderAddress(address, opts.FromExact)) &&
		(opts.FromDomain == "" || unsubscribe.EqualFoldASCII(address[at+1:], opts.FromDomain))
}

func buildExactSenderFilter(opts SearchOptions) email.Filter {
	term := opts.FromExact
	if term == "" {
		term = opts.FromDomain
	}
	if opts.From == "" {
		opts.From = term
		return buildSearchFilter(opts)
	}
	return &email.FilterOperator{
		Operator:   jmap.OperatorAND,
		Conditions: []email.Filter{buildSearchFilter(opts), &email.FilterCondition{From: term}},
	}
}

func (c *Client) senderPageSize() int {
	if c.jmap != nil && c.jmap.Session != nil {
		capability, _ := c.jmap.Session.Capabilities[jmap.CoreURI].(*core.Core)
		if capability != nil && capability.MaxObjectsInGet > 0 && capability.MaxObjectsInGet < defaultBatchSize {
			return int(capability.MaxObjectsInGet)
		}
	}
	return defaultBatchSize
}

// exactSenderIDs counts matches and keeps only the requested window. A zero
// limit keeps all matches. firstOnly stops at the first confirmed match.
func (c *Client) exactSenderIDs(opts SearchOptions, offset, limit uint64, firstOnly bool) ([]jmap.ID, uint64, error) {
	query := &email.Query{Account: c.accountID, Filter: buildExactSenderFilter(opts), Limit: uint64(c.senderPageSize())}
	if opts.SortField != "" {
		query.Sort = []*email.SortComparator{{Property: opts.SortField, IsAscending: opts.SortAsc}}
	}
	cursor := senderQueryCursor{}
	var matches []jmap.ID
	var total uint64
	for {
		ids, err := c.senderQueryPage(query, &cursor)
		if err != nil {
			return nil, 0, fmt.Errorf("exact sender query: %w", err)
		}
		if len(ids) == 0 {
			return matches, total, nil
		}
		messages, err := c.senderMessages(ids, []string{"id", "from", "receivedAt"})
		if err != nil {
			return nil, 0, fmt.Errorf("exact sender candidates: %w", err)
		}
		for _, message := range messages {
			for _, from := range message.From {
				if !opts.matchesSender(from.Email) {
					continue
				}
				if total >= offset && (limit == 0 || uint64(len(matches)) < limit) {
					matches = append(matches, message.ID)
				}
				total++
				if firstOnly {
					return matches, total, nil
				}
				break
			}
		}
	}
}

func (c *Client) searchExactSender(opts SearchOptions) (types.EmailListResult, error) {
	if opts.Offset < 0 {
		return types.EmailListResult{}, fmt.Errorf("search offset must be non-negative")
	}
	if opts.SortField == "" {
		opts.SortField = "receivedAt"
	}
	ids, total, err := c.exactSenderIDs(opts, uint64(opts.Offset), opts.Limit, false)
	if err != nil {
		return types.EmailListResult{}, err
	}
	result := types.EmailListResult{Total: total, Offset: opts.Offset, Emails: []types.EmailSummary{}}
	size := c.senderPageSize()
	for start := 0; start < len(ids); start += size {
		batch := ids[start:min(start+size, len(ids))]
		messages, err := c.senderMessages(batch, summaryProperties)
		if err != nil {
			return types.EmailListResult{}, fmt.Errorf("exact sender summaries: %w", err)
		}
		summaries := convertSummaries(messages)
		if opts.Text != "" {
			if err := c.attachSenderSnippets(batch, buildSearchFilter(opts), summaries); err != nil {
				return types.EmailListResult{}, err
			}
		}
		result.Emails = append(result.Emails, summaries...)
	}
	return result, nil
}

func (c *Client) attachSenderSnippets(ids []jmap.ID, filter email.Filter, summaries []types.EmailSummary) error {
	req := &jmap.Request{}
	req.Invoke(&searchSnippetGet{searchsnippet.Get{Account: c.accountID, EmailIDs: ids, Filter: filter}})
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("exact sender snippets: %w", err)
	}
	if resp == nil || len(resp.Responses) != 1 || resp.Responses[0] == nil {
		return nil
	}
	snippets, ok := resp.Responses[0].Args.(*searchsnippet.GetResponse)
	if !ok || snippets == nil {
		return nil
	}
	attachSnippets(summaries, snippets.List)
	return nil
}
