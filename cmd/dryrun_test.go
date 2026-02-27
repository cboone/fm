package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type jmapMockServer struct {
	server *httptest.Server

	mailboxes []map[string]any
	emails    []map[string]any
	notFound  []string

	mu           sync.Mutex
	methodCounts map[string]int
}

func newJMAPMockServer(t *testing.T, mailboxes []map[string]any, emails []map[string]any, notFound []string) *jmapMockServer {
	t.Helper()

	m := &jmapMockServer{
		mailboxes:    mailboxes,
		emails:       emails,
		notFound:     notFound,
		methodCounts: make(map[string]int),
	}

	baseURL := ""
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/session":
			writeJSON(w, map[string]any{
				"capabilities": map[string]any{
					"urn:ietf:params:jmap:core": map[string]any{},
					"urn:ietf:params:jmap:mail": map[string]any{},
				},
				"accounts": map[string]any{
					"A1": map[string]any{
						"name":       "test@example.com",
						"isPersonal": true,
					},
				},
				"primaryAccounts": map[string]any{
					"urn:ietf:params:jmap:mail": "A1",
				},
				"username":       "test@example.com",
				"apiUrl":         baseURL + "/api",
				"downloadUrl":    baseURL + "/download/{accountId}/{blobId}/{name}?type={type}",
				"uploadUrl":      baseURL + "/upload/{accountId}",
				"eventSourceUrl": baseURL + "/events",
				"state":          "state-1",
			})
			return
		case r.Method == http.MethodPost && r.URL.Path == "/api":
			var req struct {
				MethodCalls [][]json.RawMessage `json:"methodCalls"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			resp := struct {
				MethodResponses [][]any `json:"methodResponses"`
				SessionState    string  `json:"sessionState"`
			}{SessionState: "state-1"}

			for _, call := range req.MethodCalls {
				if len(call) != 3 {
					http.Error(w, "invalid invocation", http.StatusBadRequest)
					return
				}

				var name string
				if err := json.Unmarshal(call[0], &name); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				var callID string
				if err := json.Unmarshal(call[2], &callID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				m.mu.Lock()
				m.methodCounts[name]++
				m.mu.Unlock()

				switch name {
				case "Mailbox/get":
					resp.MethodResponses = append(resp.MethodResponses, []any{
						"Mailbox/get",
						map[string]any{
							"accountId": "A1",
							"state":     "state-1",
							"list":      m.mailboxes,
						},
						callID,
					})
				case "Email/get":
					args := map[string]any{
						"accountId": "A1",
						"state":     "state-1",
						"list":      m.emails,
					}
					if len(m.notFound) > 0 {
						args["notFound"] = m.notFound
					}
					resp.MethodResponses = append(resp.MethodResponses, []any{"Email/get", args, callID})
				case "Email/query":
					ids := make([]string, len(m.emails))
					for i, e := range m.emails {
						ids[i] = e["id"].(string)
					}
					resp.MethodResponses = append(resp.MethodResponses, []any{
						"Email/query",
						map[string]any{
							"accountId":  "A1",
							"queryState": "q-1",
							"total":      uint64(len(m.emails)),
							"ids":        ids,
							"position":   0,
						},
						callID,
					})
				case "Email/set":
					// Parse the request to extract IDs and mark them all as updated.
					var setArgs map[string]json.RawMessage
					_ = json.Unmarshal(call[1], &setArgs)
					updated := map[string]any{}
					if raw, ok := setArgs["update"]; ok {
						var updateMap map[string]json.RawMessage
						if json.Unmarshal(raw, &updateMap) == nil {
							for id := range updateMap {
								updated[id] = nil
							}
						}
					}
					resp.MethodResponses = append(resp.MethodResponses, []any{
						"Email/set",
						map[string]any{"accountId": "A1", "updated": updated},
						callID,
					})
				default:
					resp.MethodResponses = append(resp.MethodResponses, []any{
						"error",
						map[string]any{"type": "serverFail", "description": "unexpected method: " + name},
						callID,
					})
				}
			}

			writeJSON(w, resp)
			return
		default:
			http.NotFound(w, r)
			return
		}
	}))
	baseURL = m.server.URL

	t.Cleanup(func() {
		m.server.Close()
	})

	return m
}

func (m *jmapMockServer) count(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.methodCounts[method]
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// resetCommandFlags resets the Changed state and value for all flags
// (local and persistent) on a command and its subcommands so that
// consecutive test runs don't leak flag state.
func resetCommandFlags(cmd *cobra.Command) {
	resetFlagSet := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			f.Changed = false
			_ = f.Value.Set(f.DefValue)
		})
	}
	resetFlagSet(cmd.Flags())
	resetFlagSet(cmd.PersistentFlags())
	for _, sub := range cmd.Commands() {
		resetCommandFlags(sub)
	}
}

func runCLICommand(t *testing.T, args []string) (stdout string, stderr string, err error) {
	t.Helper()

	resetCommandFlags(rootCmd)

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}

	os.Stdout = outW
	os.Stderr = errW

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	outDone := make(chan struct{})
	errDone := make(chan struct{})

	go func() {
		_, _ = io.Copy(&outBuf, outR)
		close(outDone)
	}()
	go func() {
		_, _ = io.Copy(&errBuf, errR)
		close(errDone)
	}()

	defer func() {
		_ = outW.Close()
		_ = errW.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		<-outDone
		<-errDone
		_ = outR.Close()
		_ = errR.Close()
		stdout = outBuf.String()
		stderr = errBuf.String()
	}()

	rootCmd.SetArgs(args)
	err = rootCmd.Execute()
	return
}

func commandArgsForServer(t *testing.T, serverURL string, commandArgs ...string) []string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(""), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	base := []string{
		"--config", configPath,
		"--session-url", serverURL + "/session",
		"--credential-command", "echo test-token",
		"--account-id", "A1",
		"--format", "json",
	}

	return append(base, commandArgs...)
}

func TestArchiveDryRun_DoesNotCallMutation(t *testing.T) {
	server := newJMAPMockServer(t,
		[]map[string]any{{"id": "mb-archive", "name": "Archive", "role": "archive"}},
		[]map[string]any{{
			"id":         "M1",
			"threadId":   "T1",
			"from":       []map[string]any{{"name": "Alice", "email": "alice@example.com"}},
			"to":         []map[string]any{{"email": "me@example.com"}},
			"subject":    "Dry run test",
			"receivedAt": "2026-02-14T10:30:00Z",
			"keywords":   map[string]bool{},
			"preview":    "Preview text",
		}},
		nil,
	)

	args := commandArgsForServer(t, server.server.URL, "archive", "--dry-run", "M1")
	stdout, stderr, err := runCLICommand(t, args)
	if err != nil {
		t.Fatalf("expected success, got: %v\nstderr=%s", err, stderr)
	}

	if !strings.Contains(stdout, `"operation": "archive"`) {
		t.Fatalf("expected dry-run operation in stdout, got: %s", stdout)
	}
	if !strings.Contains(stdout, `"count": 1`) {
		t.Fatalf("expected dry-run count in stdout, got: %s", stdout)
	}
	if server.count("Mailbox/get") != 1 {
		t.Fatalf("expected Mailbox/get once, got %d", server.count("Mailbox/get"))
	}
	if server.count("Email/get") != 1 {
		t.Fatalf("expected Email/get once, got %d", server.count("Email/get"))
	}
	if server.count("Email/set") != 0 {
		t.Fatalf("expected Email/set not to be called, got %d", server.count("Email/set"))
	}
}

func TestMoveDryRun_StillEnforcesTrashSafety(t *testing.T) {
	server := newJMAPMockServer(t,
		[]map[string]any{{"id": "mb-trash", "name": "Trash", "role": "trash"}},
		nil,
		nil,
	)

	args := commandArgsForServer(t, server.server.URL, "move", "--dry-run", "M1", "--to", "Trash")
	_, stderr, err := runCLICommand(t, args)
	if !errors.Is(err, ErrSilent) {
		t.Fatalf("expected ErrSilent, got: %v", err)
	}
	if !strings.Contains(stderr, `"error": "forbidden_operation"`) {
		t.Fatalf("expected forbidden_operation error, got: %s", stderr)
	}
	if server.count("Mailbox/get") != 1 {
		t.Fatalf("expected Mailbox/get once, got %d", server.count("Mailbox/get"))
	}
	if server.count("Email/get") != 0 {
		t.Fatalf("expected Email/get not to run, got %d", server.count("Email/get"))
	}
	if server.count("Email/set") != 0 {
		t.Fatalf("expected Email/set not to be called, got %d", server.count("Email/set"))
	}
}

func TestArchiveDryRun_NotFoundReturnsPartialFailure(t *testing.T) {
	server := newJMAPMockServer(t,
		[]map[string]any{{"id": "mb-archive", "name": "Archive", "role": "archive"}},
		[]map[string]any{{
			"id":         "M1",
			"threadId":   "T1",
			"subject":    "Dry run test",
			"receivedAt": "2026-02-14T10:30:00Z",
			"keywords":   map[string]bool{},
		}},
		[]string{"missing-id"},
	)

	args := commandArgsForServer(t, server.server.URL, "archive", "--dry-run", "M1", "missing-id")
	stdout, stderr, err := runCLICommand(t, args)
	if !errors.Is(err, ErrSilent) {
		t.Fatalf("expected ErrSilent for partial failure, got: %v", err)
	}
	if !strings.Contains(stdout, `"not_found": [`) {
		t.Fatalf("expected not_found in dry-run output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "missing-id") {
		t.Fatalf("expected missing ID in dry-run output, got: %s", stdout)
	}
	if !strings.Contains(stderr, `"error": "partial_failure"`) {
		t.Fatalf("expected partial_failure stderr output, got: %s", stderr)
	}
	if server.count("Email/set") != 0 {
		t.Fatalf("expected Email/set not to be called, got %d", server.count("Email/set"))
	}
}

func TestFlagDryRunShortFlag_DoesNotCallMutation(t *testing.T) {
	server := newJMAPMockServer(t,
		nil,
		[]map[string]any{{
			"id":         "M1",
			"threadId":   "T1",
			"subject":    "Flag dry run",
			"receivedAt": "2026-02-14T10:30:00Z",
			"keywords":   map[string]bool{},
		}},
		nil,
	)

	args := commandArgsForServer(t, server.server.URL, "flag", "-n", "M1")
	stdout, stderr, err := runCLICommand(t, args)
	if err != nil {
		t.Fatalf("expected success, got: %v\nstderr=%s", err, stderr)
	}
	if !strings.Contains(stdout, `"operation": "flag"`) {
		t.Fatalf("expected flag dry-run output, got: %s", stdout)
	}
	if server.count("Email/get") != 1 {
		t.Fatalf("expected Email/get once, got %d", server.count("Email/get"))
	}
	if server.count("Email/set") != 0 {
		t.Fatalf("expected Email/set not to be called, got %d", server.count("Email/set"))
	}
}

// --- Filter-based action tests ---

func TestArchiveWithFilters_QueriesAndMutates(t *testing.T) {
	server := newJMAPMockServer(t,
		[]map[string]any{
			{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
			{"id": "mb-archive", "name": "Archive", "role": "archive"},
		},
		[]map[string]any{{
			"id":         "M1",
			"threadId":   "T1",
			"from":       []map[string]any{{"name": "Alice", "email": "alice@example.com"}},
			"to":         []map[string]any{{"email": "me@example.com"}},
			"subject":    "Filter test",
			"receivedAt": "2026-02-14T10:30:00Z",
			"keywords":   map[string]bool{},
		}},
		nil,
	)

	args := commandArgsForServer(t, server.server.URL, "archive", "--mailbox", "inbox", "--unread")
	stdout, _, err := runCLICommand(t, args)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !strings.Contains(stdout, `"matched": 1`) {
		t.Fatalf("expected matched=1 in output, got: %s", stdout)
	}
	if server.count("Email/query") != 1 {
		t.Fatalf("expected Email/query once, got %d", server.count("Email/query"))
	}
	if server.count("Email/set") != 1 {
		t.Fatalf("expected Email/set once, got %d", server.count("Email/set"))
	}
}

func TestMarkReadWithFilters_DryRun(t *testing.T) {
	server := newJMAPMockServer(t,
		[]map[string]any{
			{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
		},
		[]map[string]any{{
			"id":         "M1",
			"threadId":   "T1",
			"subject":    "Dry run filter",
			"receivedAt": "2026-02-14T10:30:00Z",
			"keywords":   map[string]bool{},
		}},
		nil,
	)

	args := commandArgsForServer(t, server.server.URL, "mark-read", "--mailbox", "inbox", "--unread", "--dry-run")
	stdout, _, err := runCLICommand(t, args)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !strings.Contains(stdout, `"operation": "mark-read"`) {
		t.Fatalf("expected dry-run operation in stdout, got: %s", stdout)
	}
	if server.count("Email/query") != 1 {
		t.Fatalf("expected Email/query once, got %d", server.count("Email/query"))
	}
	if server.count("Email/get") != 1 {
		t.Fatalf("expected Email/get once (for summaries), got %d", server.count("Email/get"))
	}
	if server.count("Email/set") != 0 {
		t.Fatalf("expected Email/set not to be called, got %d", server.count("Email/set"))
	}
}

func TestArchive_IDsAndFiltersMutuallyExclusive(t *testing.T) {
	server := newJMAPMockServer(t, nil, nil, nil)

	args := commandArgsForServer(t, server.server.URL, "archive", "M1", "--from", "alice@test.com")
	_, stderr, err := runCLICommand(t, args)
	if !errors.Is(err, ErrSilent) {
		t.Fatalf("expected ErrSilent, got: %v", err)
	}
	if !strings.Contains(stderr, `"error": "general_error"`) {
		t.Fatalf("expected general_error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "cannot combine email IDs with filter flags") {
		t.Fatalf("expected mutual exclusivity message, got: %s", stderr)
	}
}

func TestArchive_NoIDsNoFilters(t *testing.T) {
	server := newJMAPMockServer(t, nil, nil, nil)

	args := commandArgsForServer(t, server.server.URL, "archive")
	_, stderr, err := runCLICommand(t, args)
	if !errors.Is(err, ErrSilent) {
		t.Fatalf("expected ErrSilent, got: %v", err)
	}
	if !strings.Contains(stderr, `"error": "general_error"`) {
		t.Fatalf("expected general_error, got: %s", stderr)
	}
	if !strings.Contains(stderr, "no emails specified") {
		t.Fatalf("expected 'no emails specified' message, got: %s", stderr)
	}
}

func TestMoveWithFilters_StillRequiresToFlag(t *testing.T) {
	server := newJMAPMockServer(t,
		[]map[string]any{{"id": "mb-inbox", "name": "Inbox", "role": "inbox"}},
		nil,
		nil,
	)

	args := commandArgsForServer(t, server.server.URL, "move", "--mailbox", "inbox", "--unread")
	_, stderr, err := runCLICommand(t, args)
	if !errors.Is(err, ErrSilent) {
		t.Fatalf("expected ErrSilent, got: %v", err)
	}
	if !strings.Contains(stderr, `"error": "general_error"`) {
		t.Fatalf("expected general_error for missing --to, got: %s", stderr)
	}
	if !strings.Contains(stderr, "required flag") {
		t.Fatalf("expected 'required flag' message, got: %s", stderr)
	}
}
