package cmd

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestExactSenderCLIRejectsEmptyAddress(t *testing.T) {
	for _, command := range []string{"search", "archive"} {
		t.Run(command, func(t *testing.T) {
			server := newJMAPMockServer(t, nil, nil, nil)
			args := commandArgsForServer(t, server.server.URL,
				command, "--mailbox", "inbox", "--from-exact", "", "--credential-command", "exit 99")
			if command == "archive" {
				args = append(args, "--dry-run")
			}
			stdout, stderr, err := runCLICommand(t, args)
			var result struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}
			if decodeErr := json.Unmarshal([]byte(stderr), &result); decodeErr != nil ||
				!errors.Is(err, ErrSilent) || stdout != "" || result.Error != "general_error" ||
				result.Message != "--from-exact requires a plain email address" {
				t.Errorf("empty address: stdout=%q stderr=%q err=%v, want argument rejection before authentication", stdout, stderr, err)
			}
		})
	}
}

func TestExactSenderCLI(t *testing.T) {
	for _, command := range []string{"search", "archive"} {
		for _, tc := range []struct {
			flag  string
			value string
			want  []string
		}{
			{"from-exact", "NOTICES@EXAMPLE.COM", []string{"exact"}},
			{"from-domain", "Example.COM", []string{"exact", "domain"}},
		} {
			t.Run(command+"/"+tc.flag, func(t *testing.T) {
				addresses := []struct{ id, email, name string }{
					{"fuzzy", "someone@sub.example.com", "notices@example.com"},
					{"exact", "Notices@example.com", ""},
					{"domain", "other@example.com", ""},
				}
				var messages []map[string]any
				for _, addr := range addresses {
					messages = append(messages, map[string]any{
						"id": addr.id, "receivedAt": "2026-02-14T10:30:00Z",
						"from": []map[string]string{{"email": addr.email, "name": addr.name}},
					})
				}
				server := newJMAPMockServer(t, []map[string]any{
					{"id": "mb-inbox", "name": "Inbox", "role": "inbox"},
					{"id": "mb-archive", "name": "Archive", "role": "archive"},
				}, messages, nil)
				flags := []string{command, "--mailbox", "inbox", "--" + tc.flag, tc.value}
				color := command == "search" && tc.flag == "from-domain"
				if color {
					flags = append(flags, "--flagged", "--color", "green")
				}
				if command == "archive" {
					flags = append(flags, "--dry-run")
				}
				args := commandArgsForServer(t, server.server.URL, flags...)
				stdout, stderr, err := runCLICommand(t, args)
				if err != nil {
					t.Fatalf("%v: %v\n%s", flags, err, stderr)
				}
				var result struct {
					Count  int `json:"count"`
					Total  int `json:"total"`
					Emails []struct {
						ID string `json:"id"`
					} `json:"emails"`
				}
				if err := json.Unmarshal([]byte(stdout), &result); err != nil {
					t.Fatalf("decode %s result: %v\n%s", command, err, stdout)
				}
				count := result.Total
				if command == "archive" {
					count = result.Count
				}
				var got []string
				for _, message := range result.Emails {
					got = append(got, message.ID)
				}
				if !slices.Equal(got, tc.want) || count != len(tc.want) {
					t.Errorf("%s returned IDs %v and count %d, want %v and %d", command, got, count, tc.want, len(tc.want))
				}
				if calls := server.count("Email/set"); calls != 0 {
					t.Errorf("%s made %d mutation calls", command, calls)
				}
				if color {
					assertExactSenderColorFilter(t, server, tc.value)
				}
			})
		}
	}
}

func assertExactSenderColorFilter(t *testing.T, server *jmapMockServer, domain string) {
	t.Helper()
	server.mu.Lock()
	defer server.mu.Unlock()
	if len(server.queryFilters) != 2 {
		t.Fatalf("got %d Email/query calls, want one candidate page and one empty page", len(server.queryFilters))
	}
	want := map[string]any{
		"operator": "AND",
		"conditions": []any{
			map[string]any{"from": domain, "hasKeyword": "$flagged", "inMailbox": "mb-inbox"},
			map[string]any{"hasKeyword": "$mailflagbit0"},
			map[string]any{"hasKeyword": "$mailflagbit1"},
			map[string]any{"notKeyword": "$mailflagbit2"},
		},
	}
	for _, raw := range server.queryFilters {
		var got map[string]any
		if err := json.Unmarshal(raw, &got); err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("Email/query filter = %s, error = %v, want %v", raw, err, want)
		}
	}
}
