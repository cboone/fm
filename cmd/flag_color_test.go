package cmd

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFlagColorCLIListNull(t *testing.T) {
	for _, command := range []string{"list", "search"} {
		t.Run(command, func(t *testing.T) {
			server := newJMAPMockServer(t,
				[]map[string]any{{"id": "mb-inbox", "name": "Inbox", "role": "inbox"}},
				[]map[string]any{
					{"id": "unflagged", "keywords": map[string]bool{"$mailflagbit0": true}},
					{"id": "reserved", "keywords": map[string]bool{
						"$flagged": true, "$mailflagbit0": true, "$mailflagbit1": true, "$mailflagbit2": true,
					}},
				},
				nil,
			)
			args := commandArgsForServer(t, server.server.URL, command)
			stdout, stderr, err := runCLICommand(t, args)
			if err != nil {
				t.Fatalf("%s failed: %v\n%s", command, err, stderr)
			}
			var result struct {
				Emails []struct {
					ID        string          `json:"id"`
					FlagColor json.RawMessage `json:"flag_color"`
				} `json:"emails"`
			}
			if err := json.Unmarshal([]byte(stdout), &result); err != nil || len(result.Emails) != 2 {
				t.Fatalf("decode %s results: %q, %v", command, stdout, err)
			}
			for _, message := range result.Emails {
				if string(message.FlagColor) != "null" {
					t.Errorf("%s flag_color = %s, want null", message.ID, message.FlagColor)
				}
			}
		})
	}
}

func TestFlagColorCLIQuery(t *testing.T) {
	for _, command := range []string{"list", "search"} {
		for _, tc := range []struct {
			color string
			bit0  map[string]any
		}{
			{"red", map[string]any{"notKeyword": "$mailflagbit0"}},
			{"Orange", map[string]any{"hasKeyword": "$mailflagbit0"}},
		} {
			t.Run(command+"/"+tc.color, func(t *testing.T) {
				server := newJMAPMockServer(t,
					[]map[string]any{{"id": "mb-inbox", "name": "Inbox", "role": "inbox"}},
					nil, nil,
				)
				args := commandArgsForServer(t, server.server.URL, command, "--flagged", "--color", tc.color)
				_, stderr, err := runCLICommand(t, args)
				if err != nil {
					t.Fatalf("%s failed: %v\n%s", command, err, stderr)
				}

				server.mu.Lock()
				defer server.mu.Unlock()
				if len(server.queryFilters) != 1 {
					t.Fatalf("got %d Email/query filters, want 1", len(server.queryFilters))
				}
				var filter map[string]any
				if err := json.Unmarshal(server.queryFilters[0], &filter); err != nil {
					t.Fatalf("decode Email/query filter: %v", err)
				}
				base := map[string]any{"hasKeyword": "$flagged"}
				if command == "list" {
					base["inMailbox"] = "mb-inbox"
				}
				want := map[string]any{
					"operator": "AND",
					"conditions": []any{
						base, tc.bit0,
						map[string]any{"notKeyword": "$mailflagbit1"},
						map[string]any{"notKeyword": "$mailflagbit2"},
					},
				}
				if !reflect.DeepEqual(filter, want) {
					t.Errorf("Email/query filter = %v, want %v", filter, want)
				}
			})
		}
	}
}
