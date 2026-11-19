package client

import (
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/core"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/searchsnippet"
)

func TestExactSenderSearch(t *testing.T) {
	for _, tc := range []struct {
		name   string
		exact  string
		domain string
		offset int64
		total  uint64
		want   []string
		term   string
	}{
		{name: "address", exact: "sales@EXAMPLE.COM", total: 2, want: []string{"first"}, term: "sales@EXAMPLE.COM"},
		{name: "domain", domain: "EXAMPLE.COM", offset: 1, total: 3, want: []string{"other"}, term: "EXAMPLE.COM"},
		{name: "both", exact: "sales@EXAMPLE.COM", domain: "example.com", offset: 1, total: 2, want: []string{"last"}, term: "sales@EXAMPLE.COM"},
		{name: "different From entries", exact: "sales@EXAMPLE.COM", domain: "else.test", want: []string{}, term: "sales@EXAMPLE.COM"},
		{name: "domain rejects Unicode fold", domain: "k.example.com", want: []string{}, term: "k.example.com"},
		{name: "address rejects Unicode domain fold", exact: "sales@k.example.com", want: []string{}, term: "sales@k.example.com"},
		{name: "past end", domain: "EXAMPLE.COM", offset: 99, total: 3, want: []string{}, term: "EXAMPLE.COM"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, fixture := exactSenderFixture(t)
			fixture.checkQuery = func(q *email.Query) {
				want := &email.FilterOperator{Operator: jmap.OperatorAND, Conditions: []email.Filter{
					&email.FilterCondition{Text: "invoice", From: "Sales", InMailbox: "inbox", NotKeyword: "$seen"},
					&email.FilterCondition{From: tc.term},
				}}
				if !reflect.DeepEqual(q.Filter, want) {
					t.Errorf("candidate filter = %#v, want %#v", q.Filter, want)
				}
				if len(q.Sort) != 1 || q.Sort[0].Property != "subject" || !q.Sort[0].IsAscending {
					t.Errorf("candidate sort = %#v, want subject asc", q.Sort)
				}
			}
			result, err := c.SearchEmails(SearchOptions{
				Text: "invoice", From: "Sales", MailboxID: "inbox", UnreadOnly: true,
				FromExact: tc.exact, FromDomain: tc.domain, Offset: tc.offset, Limit: 1,
				SortField: "subject", SortAsc: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(result.Emails))
			for _, summary := range result.Emails {
				got = append(got, summary.ID)
				if summary.Snippet != "invoice excerpt for "+summary.ID {
					t.Errorf("snippet for %s = %q", summary.ID, summary.Snippet)
				}
			}
			if !slices.Equal(got, tc.want) || result.Total != tc.total || result.Offset != tc.offset {
				t.Errorf("search = %v/%d at %d, want %v/%d at %d", got, result.Total, result.Offset, tc.want, tc.total, tc.offset)
			}
			if !slices.Equal(fixture.summaryIDs, tc.want) {
				t.Errorf("summaries fetched for %v, want %v", fixture.summaryIDs, tc.want)
			}
		})
	}
}

func TestExactSenderSearchSummaryBatches(t *testing.T) {
	c, fixture := exactSenderFixture(t)
	result, err := c.SearchEmails(SearchOptions{
		Text: "invoice", From: "Sales", FromDomain: "example.com", Limit: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"first", "other", "last"}
	var got []string
	for _, summary := range result.Emails {
		got = append(got, summary.ID)
		if summary.Snippet != "invoice excerpt for "+summary.ID {
			t.Errorf("snippet for %s = %q", summary.ID, summary.Snippet)
		}
	}
	if !slices.Equal(got, want) || result.Total != 3 || fixture.summaryBatches != 2 {
		t.Errorf("search = %v/%d in %d summary batches, want %v/3 in 2 batches", got, result.Total, fixture.summaryBatches, want)
	}
}

func TestExactSenderActionIDs(t *testing.T) {
	c, fixture := exactSenderFixture(t)
	fixture.checkQuery = func(q *email.Query) {
		want := &email.FilterCondition{From: "example.com", Subject: "newsletter", HasKeyword: "$flagged"}
		if !reflect.DeepEqual(q.Filter, want) || len(q.Sort) != 0 {
			t.Errorf("action query = %#v, want filter %#v and default sort", q, want)
		}
	}
	got, err := c.QueryEmailIDs(SearchOptions{
		FromDomain: "example.com", Subject: "newsletter", FlaggedOnly: true, Limit: 1, Offset: 2, SortField: "subject",
	})
	want := []string{"first", "other", "last"}
	if err != nil || !slices.Equal(got, want) || len(fixture.summaryIDs) != 0 {
		t.Errorf("action IDs = %v, error = %v, summary fetches = %v, want %v and no fetches", got, err, fixture.summaryIDs, want)
	}
}

func TestExactSenderFirstID(t *testing.T) {
	c, fixture := exactSenderFixture(t)
	fixture.checkQuery = func(q *email.Query) {
		if len(q.Sort) != 1 || q.Sort[0].Property != "receivedAt" || q.Sort[0].IsAscending {
			t.Errorf("first ID sort = %#v, want receivedAt desc", q.Sort)
		}
	}
	got, err := c.QueryFirstEmailID(SearchOptions{FromExact: "sales@example.com", SortField: "subject", SortAsc: true})
	if err != nil || got != "first" || !slices.Equal(fixture.positions, []int64{0, 2}) {
		t.Errorf("first ID = %q, error = %v, positions = %v, want first after two pages", got, err, fixture.positions)
	}
}

func TestExactSenderSelectionErrors(t *testing.T) {
	for _, scenario := range []string{"state changed", "wrong position", "missing message", "wrong summary"} {
		t.Run(scenario, func(t *testing.T) {
			c, fixture := exactSenderFixture(t)
			fixture.failure = scenario
			opts := SearchOptions{FromDomain: "example.com", Limit: 1}
			result, err := c.SearchEmails(opts)
			if !errors.Is(err, ErrSenderSelectionChanged) || result.Total != 0 || len(result.Emails) != 0 {
				t.Errorf("search = %#v, %v, want selection error and no partial results", result, err)
			}
			if scenario == "wrong summary" {
				return
			}
			fixture.positions = nil
			ids, err := c.QueryEmailIDs(opts)
			if !errors.Is(err, ErrSenderSelectionChanged) || len(ids) != 0 {
				t.Errorf("action IDs = %v, %v, want selection error and no IDs", ids, err)
			}
		})
	}
}

type senderFixture struct {
	checkQuery     func(*email.Query)
	failure        string
	positions      []int64
	summaryIDs     []string
	summaryBatches int
	lastSummaryIDs []jmap.ID
}

func exactSenderFixture(t *testing.T) (*Client, *senderFixture) {
	t.Helper()
	fixture := &senderFixture{}
	c := &Client{accountID: "synthetic", jmap: &jmap.Client{
		Session: &jmap.Session{Capabilities: map[jmap.URI]jmap.Capability{jmap.CoreURI: &core.Core{MaxObjectsInGet: 2}}},
	}}
	ids := []jmap.ID{"named", "subdomain", "first", "other", "suffix", "last", "unicode", "missing-from"}
	addresses := map[jmap.ID][]*mail.Address{
		"named":     {{Name: "sales@example.com", Email: "sales@else.test"}},
		"subdomain": {{Email: "sales@sub.example.com"}},
		"first":     {{Email: "sales@else.test"}, {Email: "SaLeS@Example.COM"}},
		"other":     {{Email: "service@EXAMPLE.com"}},
		"suffix":    {{Email: "sales@notexample.com"}},
		"last":      {{Email: "sales@example.com"}, {Email: "sales@example.com"}},
		"unicode":   {{Email: "sales@K.example.com"}}, // The Kelvin sign folds to ASCII k.
	}
	c.doFunc = func(req *jmap.Request) (*jmap.Response, error) {
		if len(req.Calls) != 1 {
			t.Fatalf("unexpected number of synthetic calls: %d", len(req.Calls))
		}
		var response jmap.MethodResponse
		switch call := req.Calls[0].Args.(type) {
		case *email.Query:
			if call.Limit != 2 || call.Position < 0 || call.Position > int64(len(ids)) {
				t.Fatalf("candidate query position = %d, limit = %d", call.Position, call.Limit)
			}
			if fixture.checkQuery != nil {
				fixture.checkQuery(call)
			}
			fixture.positions = append(fixture.positions, call.Position)
			page := &email.QueryResponse{Account: c.accountID, Position: uint64(call.Position), QueryState: "stable",
				IDs: ids[call.Position:min(call.Position+2, int64(len(ids)))], Total: uint64(len(ids))}
			if call.Position == 4 && fixture.failure == "state changed" {
				page.QueryState = "changed"
			}
			if call.Position == 4 && fixture.failure == "wrong position" {
				page.Position++
			}
			response = page
		case *email.Get:
			isSummary := slices.Equal(call.Properties, summaryProperties)
			if len(call.IDs) > 2 || (!isSummary && !slices.Equal(call.Properties, []string{"id", "from", "receivedAt"})) {
				t.Fatalf("unexpected synthetic get: %#v", call)
			}
			if isSummary {
				fixture.summaryBatches++
				fixture.lastSummaryIDs = call.IDs
				for _, id := range call.IDs {
					fixture.summaryIDs = append(fixture.summaryIDs, string(id))
				}
			}
			get := &email.GetResponse{Account: c.accountID}
			for i := len(call.IDs) - 1; i >= 0; i-- {
				id := call.IDs[i]
				if isSummary {
					if fixture.failure == "wrong summary" {
						id = "named"
					}
				} else if fixture.failure == "missing message" && id == "suffix" {
					get.NotFound = append(get.NotFound, id)
					continue
				}
				at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				get.List = append(get.List, &email.Email{ID: id, From: addresses[id], ReceivedAt: &at})
			}
			response = get
		case *searchSnippetGet:
			if f, ok := call.Filter.(*email.FilterCondition); !ok || f.Text != "invoice" || f.From != "Sales" {
				t.Fatalf("unexpected snippet filter: %#v", call.Filter)
			}
			if !slices.Equal(call.EmailIDs, fixture.lastSummaryIDs) {
				t.Fatalf("unexpected snippet batch: %v", call.EmailIDs)
			}
			snippets := &searchsnippet.GetResponse{Account: c.accountID}
			for i := len(call.EmailIDs) - 1; i >= 0; i-- {
				id := call.EmailIDs[i]
				snippets.List = append(snippets.List, &searchsnippet.SearchSnippet{Email: id, Preview: "invoice excerpt for " + string(id)})
			}
			response = snippets
		default:
			t.Fatalf("synthetic fixture refuses %T", call)
		}
		return &jmap.Response{Responses: []*jmap.Invocation{{Args: response}}}, nil
	}
	return c, fixture
}
