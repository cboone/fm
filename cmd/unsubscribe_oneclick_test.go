package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
	"github.com/cboone/fm/internal/output"
	"github.com/cboone/fm/internal/types"
	"github.com/cboone/fm/internal/unsubscribe"
)

type unsubscribeTransport func(*http.Request) (*http.Response, error)

func (f unsubscribeTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestOneClickSenders(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		valid bool
	}{
		{[]string{"--one-click", "--sender", "a@example.com", "--sender", "A@example.com", "--sender", "b@example.com"}, true},
		{[]string{"--sender", "a@example.com"}, false},
		{[]string{"--one-click"}, false},
		{[]string{"--one-click", "--sender", "a@liſt.example"}, false},
		{[]string{"--one-click", "--sender", "a@K.example"}, false},
		{[]string{"--one-click", "--sender", "a@lİst.example"}, false},
		{[]string{"--one-click", "--sender", "Name <a@example.com>"}, false},
		{[]string{"--one-click", "--sender", "a@example.com,b@example.com"}, false},
		{[]string{"--one-click", "--sender", "*@example.com"}, false},
		{[]string{"--one-click", "--sender", "a@example.com", "--from", ""}, false},
		{[]string{"--one-click", "--sender", "a@example.com", "--draft"}, false},
		{[]string{"--one-click", "--sender", "a@example.com", "--confirm", "--dry-run"}, false},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("one-click", false, "")
			cmd.Flags().Bool("draft", false, "")
			cmd.Flags().Bool("confirm", false, "")
			cmd.Flags().Bool("dry-run", false, "")
			cmd.Flags().StringArray("sender", nil, "")
			addFilterFlags(cmd)
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatal(err)
			}
			got, err := oneClickSenders(cmd, nil)
			if (err == nil) != tc.valid {
				t.Fatalf("got %v, %v", got, err)
			}
			if tc.valid && len(got) != 2 {
				t.Fatalf("deduplicated senders = %v", got)
			}
			if tc.valid {
				if _, err := oneClickSenders(cmd, []string{"M1"}); err == nil {
					t.Error("accepted ID")
				}
			}
		})
	}
}

func TestOneClickSafeFailureReasons(t *testing.T) {
	description := "secret"
	for _, tc := range []struct {
		err      error
		category string
	}{
		{client.ErrNoPrimaryAccount, "no primary mail account"},
		{&url.Error{Op: "Get", URL: "https://secret.invalid/private?token=secret", Err: errors.New("secret")}, "connection failure"},
		{errors.New("secret"), "session discovery failed"},
	} {
		failure := oneClickSessionFailure(tc.err)
		if strings.Contains(failure.Message+failure.Hint, "secret") || !strings.Contains(failure.Message, tc.category) || failure.Hint == "" {
			t.Fatalf("unsafe or unhelpful failure: %v", failure)
		}
	}
	for _, tc := range []struct {
		err      error
		category string
	}{
		{client.ErrSenderSelectionChanged, "selection changed"},
		{&url.Error{Op: "Post", URL: "https://secret.invalid/private?token=secret", Err: errors.New("secret")}, "connectivity"},
		{errors.New("secret"), "unclassified transport or protocol failure"},
		{&client.SenderLookupError{Operation: "query", Err: &jmap.MethodError{Type: "serverFail", Description: &description}}, "sender query failed: JMAP serverFail"},
		{&client.SenderLookupError{Operation: "get", Err: &jmap.MethodError{Type: "forbidden", Description: &description}}, "sender metadata read failed: JMAP forbidden"},
		{&jmap.RequestError{Status: 503, Detail: "secret"}, "HTTP 503"},
		{&jmap.MethodError{Type: "secret"}, "JMAP method error"},
		{client.ErrSenderResponseInvalid, "invalid metadata response"},
	} {
		result, failed := oneClickResult(context.Background(), "a@example.com", false, func(string) (*unsubscribe.Candidate, error) { return nil, tc.err }, nil)
		if !failed || strings.Contains(result.Reason, "secret") || !strings.Contains(result.Reason, tc.category) {
			t.Fatalf("unsafe or unhelpful lookup: %v", result)
		}
	}
}

func TestOneClickMailboxSetupStages(t *testing.T) {
	oldTransport := http.DefaultTransport
	t.Cleanup(func() {
		http.DefaultTransport = oldTransport
		config.Set("credential_command", nil)
		config.Set("session_url", nil)
		config.Set("account_id", nil)
	})
	for _, tc := range []struct {
		name, command, body        string
		status                     int
		credential, missingAccount bool
		requests                   int
	}{
		{name: "credential", command: "exit 99", credential: true},
		{name: "HTTP rejection", command: "echo test-token", status: 401, body: `{"status":401,"detail":"secret"}`, requests: 1},
		{name: "missing account", command: "echo test-token", status: 200, body: `{"capabilities":{},"primaryAccounts":{},"accounts":{},"apiUrl":"http://fixture.invalid/api"}`, missingAccount: true, requests: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			config.Set("credential_command", tc.command)
			config.Set("session_url", "http://fixture.invalid/session")
			config.Set("account_id", "")
			http.DefaultTransport = unsubscribeTransport(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "fixture.invalid" || r.URL.Path != "/session" {
					t.Fatal("test refuses unexpected host/path")
				}
				requests++
				return &http.Response{StatusCode: tc.status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})
			_, err := openOneClickMailbox()
			var setupErr *oneClickSetupError
			if !errors.As(err, &setupErr) || setupErr.credential != tc.credential || requests != tc.requests {
				t.Fatalf("setup stage lost: err=%v requests=%d", err, requests)
			}
			if errors.Is(err, client.ErrNoPrimaryAccount) != tc.missingAccount {
				t.Fatalf("account cause lost: %v", err)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("setup error exposed server payload")
			}
		})
	}
}

type oneClickMailboxFunc func(string) (*unsubscribe.Candidate, error)

func (f oneClickMailboxFunc) NewestSenderEmail(sender string) (*unsubscribe.Candidate, error) {
	return f(sender)
}

func TestOneClickCommandConsent(t *testing.T) {
	oldOpen, oldHTTP := oneClickOpenMailbox, oneClickHTTPClient
	t.Cleanup(func() { oneClickOpenMailbox, oneClickHTTPClient = oldOpen, oldHTTP; resetCommandFlags(rootCmd) })
	for _, tc := range []struct {
		name       string
		flags      []string
		change     func(*unsubscribe.Candidate)
		status     string
		posts      int
		postStatus int
	}{
		{name: "default preview", status: "previewed"},
		{name: "explicit preview", flags: []string{"--dry-run"}, status: "previewed"},
		{name: "confirm", flags: []string{"--confirm"}, status: "http_status", posts: 1},
		{name: "HTTP failure", flags: []string{"--confirm"}, status: "http_status", posts: 1, postStatus: 503},
		{name: "deduplicated confirm", flags: []string{"--confirm", "--sender", "a@example.com"}, status: "http_status", posts: 1},
		{name: "unknown authentication", flags: []string{"--confirm"}, change: func(m *unsubscribe.Candidate) { m.Headers = m.Headers[1:] }, status: "skipped"},
		{name: "unsafe target", flags: []string{"--confirm"}, change: func(m *unsubscribe.Candidate) { m.Headers[3].Value = "<https://127.0.0.1/fixture>" }, status: "skipped"},
		{name: "ambiguous sender", flags: []string{"--confirm"}, change: func(m *unsubscribe.Candidate) { m.From = append(m.From, "other@example.com") }, status: "skipped"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opens, lookups, posts := 0, 0, 0
			postStatus := tc.postStatus
			if postStatus == 0 {
				postStatus = 204
			}
			oneClickOpenMailbox = func() (oneClickMailbox, error) {
				opens++
				return oneClickMailboxFunc(func(sender string) (*unsubscribe.Candidate, error) {
					lookups++
					if sender != "a@example.com" {
						t.Fatal("unexpected sender")
					}
					m := oneClickMessage()
					if tc.change != nil {
						tc.change(m)
					}
					return m, nil
				}), nil
			}
			oneClickHTTPClient = func() *http.Client {
				c := unsubscribe.NewHTTPClient()
				c.Transport = unsubscribeTransport(func(r *http.Request) (*http.Response, error) {
					if r.URL.Host != "list.example" {
						t.Fatal("test refuses unexpected host")
					}
					posts++
					return &http.Response{StatusCode: postStatus, Body: io.NopCloser(strings.NewReader("secret-response")), Header: make(http.Header)}, nil
				})
				return c
			}
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			args := []string{"--config", config, "--format", "json", "--credential-command", "exit 99", "--session-url", "https://api.fastmail.com/jmap/session", "unsubscribe", "--one-click", "--sender", "a@example.com"}
			stdout, stderr, err := runCLICommand(t, append(args, tc.flags...))
			if (err != nil) != (postStatus == 503) || (postStatus == 503 && !strings.Contains(stderr, "partial_failure")) || (postStatus != 503 && stderr != "") {
				t.Fatalf("command failed: %v %s", err, stderr)
			}
			var wire []map[string]any
			if err := json.Unmarshal([]byte(stdout), &wire); err != nil || len(wire) != 1 {
				t.Fatalf("JSON result array: %v %s", err, stdout)
			}
			for key, value := range wire[0] {
				switch key {
				case "sender", "status", "reason", "email_id", "received_at", "host":
					if _, ok := value.(string); !ok {
						t.Fatalf("JSON string field %s: %T", key, value)
					}
				case "http_status":
					if value != float64(postStatus) || tc.posts != 1 {
						t.Fatalf("JSON HTTP status: %v", value)
					}
				default:
					t.Fatalf("unexpected JSON key %q", key)
				}
			}
			if wire[0]["sender"] != "a@example.com" || wire[0]["status"] != tc.status {
				t.Fatalf("JSON required fields: %v", wire)
			}
			if tc.posts == 1 && wire[0]["http_status"] != float64(postStatus) {
				t.Fatal("JSON http_status omitted")
			}
			if tc.status != "skipped" && (wire[0]["email_id"] != "M1" || wire[0]["host"] != "list.example" || wire[0]["received_at"] != "2026-11-18T12:00:00Z") {
				t.Fatal("JSON metadata keys omitted")
			}
			if tc.status == "skipped" && wire[0]["reason"] == nil {
				t.Fatal("JSON skip reason omitted")
			}
			var results []types.OneClickResult
			if err := json.Unmarshal([]byte(stdout), &results); err != nil {
				t.Fatal(err)
			}
			if len(results) != 1 || results[0].Status != tc.status || posts != tc.posts || opens != 1 || lookups != 1 {
				t.Fatalf("consent forwarding: results=%v posts=%d opens=%d lookups=%d", results, posts, opens, lookups)
			}
			if results[0].EmailID != "M1" && tc.status != "skipped" {
				t.Fatal("message metadata omitted")
			}
			if tc.status != "skipped" && (results[0].Host != "list.example" || results[0].ReceivedAt != "2026-11-18T12:00:00Z") {
				t.Fatal("preview metadata omitted")
			}
			if tc.posts == 1 && results[0].HTTPStatus != postStatus {
				t.Fatal("HTTP status omitted")
			}
			if strings.Contains(stdout+stderr, "secret") {
				t.Fatal("command exposed capability or response body")
			}
		})
	}
}

type oneClickFailWriter struct{}

func (oneClickFailWriter) Write(p []byte) (int, error) {
	if strings.Contains(string(p), "message=") {
		return 0, errors.New("synthetic writer failure")
	}
	return len(p), nil
}

func TestOneClickTextWriteFailure(t *testing.T) {
	rows := []types.OneClickResult{{Sender: "a@example.com", Status: "previewed", EmailID: "M1"}}
	if err := output.New("text").Format(oneClickFailWriter{}, rows); err == nil {
		t.Fatal("text metadata write failure was ignored")
	}
}

func TestOneClickTextOutput(t *testing.T) {
	rows := []types.OneClickResult{
		{Sender: "p@example.com", Status: "previewed", EmailID: "M1", ReceivedAt: "2026-11-18T12:00:00Z", Host: "list.example"},
		{Sender: "s@example.com", Status: "skipped", Reason: "unverified: no DKIM pass"},
		{Sender: "h@example.com", Status: "http_status", HTTPStatus: 204, Host: "list.example"},
		{Sender: "e@example.com", Status: "error", Reason: "outcome unknown"},
	}
	want := "p@example.com: previewed message=M1 received=2026-11-18T12:00:00Z host=list.example\ns@example.com: skipped (unverified: no DKIM pass)\nh@example.com: http_status HTTP 204 host=list.example\ne@example.com: error (outcome unknown)\n"
	var buf bytes.Buffer
	if err := output.New("text").Format(&buf, rows); err != nil {
		t.Fatal(err)
	}
	if buf.String() != want {
		t.Fatalf("text output = %q; want %q", buf.String(), want)
	}
}

func TestOneClickCommandBatch(t *testing.T) {
	oldOpen, oldHTTP := oneClickOpenMailbox, oneClickHTTPClient
	t.Cleanup(func() { oneClickOpenMailbox, oneClickHTTPClient = oldOpen, oldHTTP; resetCommandFlags(rootCmd) })
	for _, tc := range []struct {
		name          string
		firstStatus   int
		lookupFailure bool
	}{
		{name: "two accepted", firstStatus: 204},
		{name: "continue after HTTP failure", firstStatus: 503},
		{name: "report before later metadata error", firstStatus: 204, lookupFailure: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var lookedUp, attempted []string
			oneClickOpenMailbox = func() (oneClickMailbox, error) {
				return oneClickMailboxFunc(func(sender string) (*unsubscribe.Candidate, error) {
					lookedUp = append(lookedUp, sender)
					if sender != "a@example.com" && sender != "b@example.com" {
						t.Fatal("unexpected sender")
					}
					if sender == "b@example.com" && tc.lookupFailure {
						return nil, &client.SenderLookupError{Operation: "get", Err: client.ErrSenderResponseInvalid}
					}
					m := oneClickMessage()
					m.ID = "M-" + sender[:1]
					m.From = []string{sender}
					m.Headers[3].Value = "<https://" + sender[:1] + ".list.example/secret?token=" + sender[:1] + ">"
					return m, nil
				}), nil
			}
			oneClickHTTPClient = func() *http.Client {
				c := unsubscribe.NewHTTPClient()
				c.Transport = unsubscribeTransport(func(r *http.Request) (*http.Response, error) {
					if r.URL.Host != "a.list.example" && r.URL.Host != "b.list.example" {
						t.Fatal("test refuses unexpected host")
					}
					attempted = append(attempted, r.URL.Host)
					status := 204
					if r.URL.Host == "a.list.example" {
						status = tc.firstStatus
					}
					return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("secret-response")), Header: make(http.Header)}, nil
				})
				return c
			}
			config := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(config, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			args := []string{"--config", config, "--format", "json", "--credential-command", "exit 99", "--session-url", "https://api.fastmail.com/jmap/session", "unsubscribe", "--one-click", "--confirm", "--sender", "a@example.com", "--sender", "b@example.com"}
			stdout, stderr, err := runCLICommand(t, args)
			wantFailure := tc.firstStatus == 503 || tc.lookupFailure
			if (err != nil) != wantFailure || (wantFailure && !strings.Contains(stderr, "partial_failure")) || (!wantFailure && stderr != "") {
				t.Fatalf("batch exit: err=%v stderr=%s", err, stderr)
			}
			var rows []types.OneClickResult
			if err := json.Unmarshal([]byte(stdout), &rows); err != nil {
				t.Fatal(err)
			}
			wantAttempts := "a.list.example,b.list.example"
			if tc.lookupFailure {
				wantAttempts = "a.list.example"
			}
			if len(rows) != 2 || strings.Join(lookedUp, ",") != "a@example.com,b@example.com" || strings.Join(attempted, ",") != wantAttempts {
				t.Fatalf("batch did not process every sender: rows=%v lookups=%v attempts=%v", rows, lookedUp, attempted)
			}
			if rows[0].Sender != "a@example.com" || rows[0].Status != "http_status" || rows[0].HTTPStatus != tc.firstStatus || rows[1].Sender != "b@example.com" {
				t.Fatalf("ordered HTTP results lost: %v", rows)
			}
			if tc.lookupFailure {
				if rows[1].Status != "error" || !strings.Contains(rows[1].Reason, "invalid metadata response") {
					t.Fatal("later malformed response not reported")
				}
			} else if rows[1].Status != "http_status" || rows[1].HTTPStatus != 204 {
				t.Fatal("later sender was not confirmed")
			}
			if strings.Contains(stdout+stderr, "secret") {
				t.Fatal("batch exposed capability or response body")
			}
		})
	}
}

func oneClickMessage() *unsubscribe.Candidate {
	return &unsubscribe.Candidate{ID: "M1", From: []string{"a@example.com"}, ReceivedAt: time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC), Headers: []unsubscribe.Header{
		{Name: "Authentication-Results", Value: "mx1.messagingengine.com; dkim=pass header.d=example.com header.b=abc"},
		{Name: "Received", Value: "from smtp.example.com by mx1.messagingengine.com with ESMTP; date"},
		{Name: "DKIM-Signature", Value: "d=example.com; b=abcdef; h=from:list-unsubscribe:list-unsubscribe-post"},
		{Name: "List-Unsubscribe", Value: "<https://list.example/secret?token=secret>"},
		{Name: "List-Unsubscribe-Post", Value: "List-Unsubscribe=One-Click"},
	}}
}

func TestOneClickResults(t *testing.T) {
	for _, confirm := range []bool{false, true} {
		calls := 0
		c := unsubscribe.NewHTTPClient()
		c.Transport = unsubscribeTransport(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Host != "list.example" {
				t.Fatal("test refuses unexpected host")
			}
			if !confirm {
				t.Fatal("preview sent a request")
			}
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("secret")), Header: make(http.Header)}, nil
		})
		selectEmail := func(sender string) (*unsubscribe.Candidate, error) {
			switch sender {
			case "a@example.com":
				return oneClickMessage(), nil
			case "missing@example.com":
				return nil, nil
			case "error@example.com":
				return nil, errors.New("secret")
			case "ambiguous@example.com":
				m := oneClickMessage()
				m.From = append(m.From, "ambiguous@example.com")
				return m, nil
			case "unknown@example.com":
				m := oneClickMessage()
				m.From = []string{sender}
				m.Headers = m.Headers[1:]
				return m, nil
			default:
				t.Fatal("unexpected sender")
				return nil, nil
			}
		}
		results, failed := oneClickResults(context.Background(), []string{"a@example.com", "missing@example.com", "error@example.com", "ambiguous@example.com", "unknown@example.com"}, confirm, selectEmail, c)
		if !failed || len(results) != 5 || results[1].Status != "skipped" || results[2].Status != "error" || results[3].Status != "skipped" || results[4].Status != "skipped" {
			t.Fatalf("unexpected results %v, failed=%v", results, failed)
		}
		if confirm {
			if calls != 1 || results[0].HTTPStatus != 503 || results[0].Status != "http_status" {
				t.Fatalf("confirm: calls=%d results=%v", calls, results)
			}
		} else if calls != 0 || results[0].Status != "previewed" {
			t.Fatalf("preview: calls=%d results=%v", calls, results)
		}
		for _, format := range []string{"json", "text"} {
			var buf bytes.Buffer
			if err := output.New(format).Format(&buf, results); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(buf.String(), "secret") || !strings.Contains(buf.String(), "a@example.com") || !strings.Contains(buf.String(), results[0].Status) {
				t.Fatalf("bad %s output: %s", format, buf.String())
			}
		}
	}
}

func TestOneClickHTTPBatchFailure(t *testing.T) {
	for _, status := range []int{200, 204, 302, 400, 429, 503, 0} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			calls := 0
			c := unsubscribe.NewHTTPClient()
			c.Transport = unsubscribeTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host != "list.example" {
					t.Fatal("unexpected host")
				}
				if status == 0 {
					return nil, errors.New("synthetic uncertain failure")
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
			})
			results, failed := oneClickResults(context.Background(), []string{"a@example.com"}, true, func(string) (*unsubscribe.Candidate, error) { return oneClickMessage(), nil }, c)
			wantFailure := status < 200 || status >= 300
			if failed != wantFailure || len(results) != 1 || results[0].HTTPStatus != status || calls != 1 {
				t.Fatalf("status=%d failed=%v results=%v calls=%d", status, failed, results, calls)
			}
		})
	}
}

func TestOneClickFromASCII(t *testing.T) {
	message := oneClickMessage()
	message.From = []string{"a@liſt.example"}
	for i := range message.Headers {
		message.Headers[i].Value = strings.ReplaceAll(message.Headers[i].Value, "example.com", "list.example")
	}
	selectEmail := func(string) (*unsubscribe.Candidate, error) { return message, nil }
	result, failed := oneClickResult(context.Background(), "a@list.example", false, selectEmail, nil)
	if failed || result.Status != "skipped" || result.Reason != "ambiguous or nonmatching From address" {
		t.Errorf("Unicode From result=%+v failed=%v", result, failed)
	}
}
