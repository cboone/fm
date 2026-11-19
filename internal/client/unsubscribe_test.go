package client

import (
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"testing"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
)

func TestNewestSenderEmail(t *testing.T) {
	for _, scenario := range []string{"paged", "ambiguous", "none", "changed", "missing", "query error", "get error"} {
		t.Run(scenario, func(t *testing.T) {
			queries := 0
			c := &Client{accountID: "test"}
			c.doFunc = func(req *jmap.Request) (*jmap.Response, error) {
				if len(req.Calls) != 1 {
					t.Fatal("unexpected request")
				}
				var response jmap.MethodResponse
				switch call := req.Calls[0].Args.(type) {
				case *email.Query:
					queries++
					if scenario == "query error" {
						return nil, errors.New("synthetic")
					}
					if call.Filter.(*email.FilterCondition).From != "sender@example.com" || len(call.Sort) != 1 || call.Sort[0].Property != "receivedAt" || call.Sort[0].IsAscending {
						t.Fatal("incorrect selection query")
					}
					page := &email.QueryResponse{Account: "test", QueryState: "stable", Position: uint64(call.Position), IDs: []jmap.ID{"fuzzy-new", "fuzzy-old"}}
					if queries > 1 {
						if call.Position != 2 {
							t.Fatalf("position = %d", call.Position)
						}
						page.IDs = []jmap.ID{"exact-new", "exact-old"}
						if scenario == "changed" {
							page.QueryState = "changed"
						}
					}
					if scenario == "none" {
						page.IDs = nil
					}
					response = page
				case *email.Get:
					if scenario == "get error" {
						return nil, errors.New("synthetic")
					}
					if !reflect.DeepEqual(call.Properties, []string{"id", "from", "receivedAt", "headers"}) {
						t.Fatalf("unexpected properties: %v", call.Properties)
					}
					get := &email.GetResponse{Account: "test"}
					for _, id := range call.IDs {
						from := []*mail.Address{{Name: "sender@example.com", Email: "other@example.com"}}
						if queries > 1 {
							from = []*mail.Address{{Email: "sender@example.com"}}
						}
						if scenario == "ambiguous" && queries > 1 {
							from = append(from, &mail.Address{Email: "other@example.com"})
						}
						// Get response order deliberately differs from query order.
						receivedAt := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
						get.List = append([]*email.Email{{ID: id, From: from, ReceivedAt: &receivedAt, Headers: []*email.Header{{Name: "List-Unsubscribe", Value: "one"}, {Name: "List-Unsubscribe", Value: "two"}}}}, get.List...)
					}
					if scenario == "missing" {
						get.List = nil
					}
					response = get
				default:
					t.Fatalf("test refuses unexpected JMAP call %T", call)
				}
				return &jmap.Response{Responses: []*jmap.Invocation{{Args: response}}}, nil
			}
			got, err := c.NewestSenderEmail("sender@example.com")
			switch scenario {
			case "changed", "missing", "query error", "get error":
				if err == nil {
					t.Fatal("missing expected error")
				}
				if (scenario == "changed" || scenario == "missing") && !errors.Is(err, ErrSenderSelectionChanged) {
					t.Fatalf("selection cause lost: %v", err)
				}
			case "none":
				if err != nil || got != nil {
					t.Fatalf("got %v, %v", got, err)
				}
			default:
				if err != nil || got == nil || got.ID != "exact-new" || len(got.Headers) != 2 {
					t.Fatalf("got %v, %v", got, err)
				}
				if scenario == "ambiguous" && len(got.From) != 2 {
					t.Fatal("ambiguity discarded")
				}
			}
		})
	}
}

func TestSenderLookupNetworkErrors(t *testing.T) {
	for _, operation := range []string{"query", "get"} {
		cause := &url.Error{Op: "Post", URL: "http://fixture.invalid/private", Err: errors.New("synthetic")}
		c := &Client{accountID: "test"}
		c.doFunc = func(req *jmap.Request) (*jmap.Response, error) {
			if _, ok := req.Calls[0].Args.(*email.Query); ok && operation == "get" {
				return &jmap.Response{Responses: []*jmap.Invocation{{Args: &email.QueryResponse{Account: "test", IDs: []jmap.ID{"M1"}, QueryState: "stable"}}}}, nil
			}
			return nil, cause
		}
		_, err := c.NewestSenderEmail("sender@example.com")
		var lookupErr *SenderLookupError
		if !errors.As(err, &lookupErr) || lookupErr.Operation != operation || !errors.Is(err, cause) {
			t.Fatalf("network operation/cause lost: %v", err)
		}
	}
}

func TestSenderLookupInvalidResponses(t *testing.T) {
	cases := []struct {
		name, operation string
		change          func(*jmap.Response) *jmap.Response
	}{
		{name: "nil response", change: func(*jmap.Response) *jmap.Response { return nil }},
		{name: "empty responses", change: func(*jmap.Response) *jmap.Response { return &jmap.Response{} }},
		{name: "multiple responses", change: func(r *jmap.Response) *jmap.Response { r.Responses = append(r.Responses, r.Responses[0]); return r }},
		{name: "nil invocation", change: func(r *jmap.Response) *jmap.Response { r.Responses[0] = nil; return r }},
		{name: "null wire invocation", change: func(*jmap.Response) *jmap.Response {
			var decoded jmap.Response
			if err := json.Unmarshal([]byte(`{"methodResponses":[null]}`), &decoded); err != nil {
				t.Fatal(err)
			}
			return &decoded
		}},
		{name: "null wire args", change: func(r *jmap.Response) *jmap.Response {
			method := "Email/query"
			if _, ok := r.Responses[0].Args.(*email.GetResponse); ok {
				method = "Email/get"
			}
			var decoded jmap.Response
			if err := json.Unmarshal([]byte(`{"methodResponses":[["`+method+`",null,"0"]]}`), &decoded); err != nil {
				t.Fatal(err)
			}
			return &decoded
		}},
		{name: "wrong account", change: func(r *jmap.Response) *jmap.Response {
			switch args := r.Responses[0].Args.(type) {
			case *email.QueryResponse:
				args.Account = "other"
			case *email.GetResponse:
				args.Account = "other"
			}
			return r
		}},
		{name: "nil args", change: func(r *jmap.Response) *jmap.Response { r.Responses[0].Args = nil; return r }},
		{name: "nil query args", change: func(r *jmap.Response) *jmap.Response { r.Responses[0].Args = (*email.QueryResponse)(nil); return r }},
		{name: "nil get args", change: func(r *jmap.Response) *jmap.Response { r.Responses[0].Args = (*email.GetResponse)(nil); return r }},
		{name: "nil method error", change: func(r *jmap.Response) *jmap.Response { r.Responses[0].Args = (*jmap.MethodError)(nil); return r }},
		{name: "empty method error", change: func(r *jmap.Response) *jmap.Response { r.Responses[0].Args = &jmap.MethodError{}; return r }},
		{name: "wrong method response", change: func(r *jmap.Response) *jmap.Response { r.Responses[0].Args = &email.SetResponse{}; return r }},
		{name: "wrong typed operation", change: func(r *jmap.Response) *jmap.Response {
			if _, ok := r.Responses[0].Args.(*email.QueryResponse); ok {
				r.Responses[0].Args = &email.GetResponse{Account: "test"}
			} else {
				r.Responses[0].Args = &email.QueryResponse{Account: "test"}
			}
			return r
		}},
		{name: "empty query ID", operation: "query", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.QueryResponse).IDs = []jmap.ID{""}
			return r
		}},
		{name: "nil message", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0] = nil
			return r
		}},
		{name: "empty message ID", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0].ID = ""
			return r
		}},
		{name: "nil date", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0].ReceivedAt = nil
			return r
		}},
		{name: "nil address", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0].From[0] = nil
			return r
		}},
		{name: "nil header", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0].Headers[0] = nil
			return r
		}},
		{name: "empty header name", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0].Headers[0].Name = ""
			return r
		}},
		{name: "invalid header name", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0].Headers[0].Name = "Bad:Header"
			return r
		}},
		{name: "empty not-found ID", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).NotFound = []jmap.ID{""}
			return r
		}},
		{name: "header whitespace", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0].Headers[0].Name = "Bad Header"
			return r
		}},
		{name: "header non-ASCII", operation: "get", change: func(r *jmap.Response) *jmap.Response {
			r.Responses[0].Args.(*email.GetResponse).List[0].Headers[0].Name = "BadéHeader"
			return r
		}},
	}
	for _, operation := range []string{"query", "get"} {
		for _, tc := range cases {
			if tc.operation != "" && tc.operation != operation {
				continue
			}
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				defer func() {
					if recover() != nil {
						t.Fatal("invalid metadata was processed before rejection")
					}
				}()
				c := &Client{accountID: "test"}
				c.doFunc = func(req *jmap.Request) (*jmap.Response, error) {
					if _, ok := req.Calls[0].Args.(*email.Query); ok {
						r := &jmap.Response{Responses: []*jmap.Invocation{{Args: &email.QueryResponse{Account: "test", IDs: []jmap.ID{"M1"}, QueryState: "stable"}}}}
						if operation == "query" {
							return tc.change(r), nil
						}
						return r, nil
					}
					if operation == "query" {
						t.Fatal("invalid query reached a later request")
					}
					receivedAt := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
					r := &jmap.Response{Responses: []*jmap.Invocation{{Args: &email.GetResponse{Account: "test", List: []*email.Email{{ID: "M1", ReceivedAt: &receivedAt, From: []*mail.Address{{Email: "sender@example.com"}}, Headers: []*email.Header{{Name: "From", Value: "sender@example.com"}}}}}}}}
					return tc.change(r), nil
				}
				_, err := c.NewestSenderEmail("sender@example.com")
				var lookupErr *SenderLookupError
				if !errors.Is(err, ErrSenderResponseInvalid) || !errors.As(err, &lookupErr) || lookupErr.Operation != operation {
					t.Fatalf("invalid response operation/cause lost: %v", err)
				}
			})
		}
	}
}

func TestSenderLookupTypedErrors(t *testing.T) {
	for _, operation := range []string{"query", "get"} {
		t.Run(operation, func(t *testing.T) {
			description := "synthetic private server payload"
			cause := &jmap.MethodError{Type: "serverFail", Description: &description}
			c := &Client{accountID: "test"}
			c.doFunc = func(req *jmap.Request) (*jmap.Response, error) {
				var response jmap.MethodResponse = cause
				if _, ok := req.Calls[0].Args.(*email.Query); ok && operation == "get" {
					response = &email.QueryResponse{Account: "test", IDs: []jmap.ID{"M1"}, QueryState: "stable"}
				}
				return &jmap.Response{Responses: []*jmap.Invocation{{Args: response}}}, nil
			}
			_, err := c.NewestSenderEmail("sender@example.com")
			var lookupErr *SenderLookupError
			var methodErr *jmap.MethodError
			if !errors.As(err, &lookupErr) || lookupErr.Operation != operation || !errors.As(err, &methodErr) || methodErr != cause {
				t.Fatalf("typed cause/operation lost: %v", err)
			}
		})
	}
}
