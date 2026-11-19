package client

import (
	"bytes"
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
	"git.sr.ht/~rockorager/go-jmap/mail/thread"

	"github.com/cboone/fm/internal/output"
	"github.com/cboone/fm/internal/types"
)

func TestFlagColorQuery(t *testing.T) {
	for _, tc := range []struct {
		color    FlagColor
		keywords []string
	}{
		{FlagColorRed, nil},
		{FlagColorOrange, []string{"$mailflagbit0"}},
		{FlagColorYellow, []string{"$mailflagbit1"}},
		{FlagColorGreen, []string{"$mailflagbit0", "$mailflagbit1"}},
		{FlagColorBlue, []string{"$mailflagbit2"}},
		{FlagColorPurple, []string{"$mailflagbit0", "$mailflagbit2"}},
		{FlagColorGray, []string{"$mailflagbit1", "$mailflagbit2"}},
	} {
		for _, command := range []string{"list", "search"} {
			t.Run(command+"/"+tc.color.String(), func(t *testing.T) {
				base := &email.FilterCondition{
					InMailbox: "inbox-id", HasKeyword: "$flagged", NotKeyword: "$seen", Subject: "travel",
				}
				want := &email.FilterOperator{Operator: jmap.OperatorAND, Conditions: []email.Filter{base}}
				for _, keyword := range []string{"$mailflagbit0", "$mailflagbit1", "$mailflagbit2"} {
					if slices.Contains(tc.keywords, keyword) {
						want.Conditions = append(want.Conditions, &email.FilterCondition{HasKeyword: keyword})
					} else {
						want.Conditions = append(want.Conditions, &email.FilterCondition{NotKeyword: keyword})
					}
				}
				message := &email.Email{ID: "message", Keywords: map[string]bool{"$flagged": true}}
				for _, keyword := range tc.keywords {
					message.Keywords[keyword] = true
				}
				c := &Client{
					mailboxCache: []*mailbox.Mailbox{{ID: "inbox-id", Role: mailbox.RoleInbox}},
					doFunc: func(req *jmap.Request) (*jmap.Response, error) {
						query := req.Calls[0].Args.(*email.Query)
						if !reflect.DeepEqual(query.Filter, want) {
							gotJSON, _ := json.Marshal(query.Filter)
							wantJSON, _ := json.Marshal(want)
							t.Fatalf("query filter = %s, want %s", gotJSON, wantJSON)
						}
						get := req.Calls[1].Args.(*email.Get)
						if !slices.Contains(get.Properties, "keywords") {
							t.Fatal("Email/get must request keywords")
						}
						return &jmap.Response{Responses: []*jmap.Invocation{
							{Args: &email.QueryResponse{IDs: []jmap.ID{"message"}, Total: 1}},
							{Args: &email.GetResponse{List: []*email.Email{message}}},
						}}, nil
					},
				}
				var result types.EmailListResult
				var err error
				if command == "list" {
					result, err = c.ListEmails(ListOptions{
						MailboxNameOrID: "inbox", Subject: "travel", UnreadOnly: true, FlagColor: &tc.color,
					})
				} else {
					result, err = c.SearchEmails(SearchOptions{
						MailboxID: "inbox-id", Subject: "travel", UnreadOnly: true, FlagColor: &tc.color,
					})
				}
				if err != nil || len(result.Emails) != 1 {
					t.Fatalf("%s = %+v, %v", command, result, err)
				}
				assertFormattedFlagColor(t, result, "emails", tc.color.String())
			})
		}
	}
}

func TestFlagColorRead(t *testing.T) {
	for _, tc := range []struct {
		name     string
		keywords map[string]bool
		want     string
	}{
		{"unflagged", map[string]bool{"$mailflagbit0": true}, ""},
		{"red", map[string]bool{"$flagged": true}, "red"},
		{"orange", map[string]bool{"$flagged": true, "$mailflagbit0": true}, "orange"},
		{"reserved", map[string]bool{"$flagged": true, "$mailflagbit0": true, "$mailflagbit1": true, "$mailflagbit2": true}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message := &email.Email{ID: "message", ThreadID: "thread", Keywords: tc.keywords}
			c := flagColorThreadClient(t, message)
			summaries := types.EmailListResult{
				Emails: convertSummaries([]*email.Email{message}),
			}
			assertFormattedFlagColor(t, summaries, "emails", tc.want)
			detail, err := c.ReadEmail("message", false, false)
			if err != nil {
				t.Fatal(err)
			}
			assertFormattedFlagColor(t, detail, "", tc.want)
			view, err := c.ReadThread("message", false, false)
			if err != nil {
				t.Fatal(err)
			}
			assertFormattedFlagColor(t, view, "thread", tc.want)
		})
	}
}

func TestFlagColorThreadContext(t *testing.T) {
	c := flagColorThreadClient(t,
		&email.Email{ID: "selected", ThreadID: "thread", Keywords: map[string]bool{"$flagged": true}},
		&email.Email{ID: "previous", Keywords: map[string]bool{"$flagged": true, "$mailflagbit0": true}},
		&email.Email{ID: "unflagged", Keywords: map[string]bool{"$mailflagbit0": true}},
	)
	view, err := c.ReadThread("selected", false, false)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := output.New("json").Format(&buf, view); err != nil {
		t.Fatal(err)
	}
	var result struct {
		Thread []struct {
			ID        string          `json:"id"`
			FlagColor json.RawMessage `json:"flag_color"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	got := make(map[string]string)
	for _, message := range result.Thread {
		got[message.ID] = string(message.FlagColor)
	}
	want := map[string]string{"selected": `"red"`, "previous": `"orange"`, "unflagged": "null"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("thread colors = %v, want %v", got, want)
	}

	buf.Reset()
	if err := output.New("text").Format(&buf, view); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\n      Flag color: orange\n") {
		t.Errorf("expected %q in %q", "Flag color: orange", buf.String())
	}
}

func flagColorThreadClient(t *testing.T, selected *email.Email, context ...*email.Email) *Client {
	t.Helper()
	messages := append([]*email.Email{selected}, context...)
	ids := make([]jmap.ID, 0, len(messages))
	for _, message := range messages {
		ids = append(ids, message.ID)
	}
	return &Client{doFunc: func(req *jmap.Request) (*jmap.Response, error) {
		switch call := req.Calls[0].Args.(type) {
		case *email.Get:
			if !slices.Contains(call.Properties, "keywords") {
				t.Fatal("Email/get must request keywords")
			}
			var list []*email.Email
			switch {
			case slices.Equal(call.IDs, ids):
				list = messages
			case slices.Equal(call.IDs, []jmap.ID{selected.ID}):
				list = []*email.Email{selected}
			default:
				t.Fatalf("unexpected Email/get IDs: %v", call.IDs)
			}
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Args: &email.GetResponse{List: list}},
			}}, nil
		case *thread.Get:
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Args: &thread.GetResponse{List: []*thread.Thread{{ID: "thread", EmailIDs: ids}}}},
			}}, nil
		default:
			t.Fatalf("unexpected JMAP call: %T", call)
			return nil, nil
		}
	}}
}

func assertFormattedFlagColor(t *testing.T, result any, collection, want string) {
	t.Helper()
	var buf bytes.Buffer
	if err := output.New("json").Format(&buf, result); err != nil {
		t.Fatal(err)
	}
	var item map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if collection != "" {
		var items []map[string]json.RawMessage
		if err := json.Unmarshal(item[collection], &items); err != nil || len(items) != 1 {
			t.Fatalf("json %s = %s, %v", collection, item[collection], err)
		}
		item = items[0]
	}
	wantJSON := "null"
	if want != "" {
		wantJSON = `"` + want + `"`
	}
	if got := string(item["flag_color"]); got != wantJSON {
		t.Errorf("flag_color = %s, want %s", got, wantJSON)
	}
	buf.Reset()
	if err := output.New("text").Format(&buf, result); err != nil {
		t.Fatal(err)
	}
	label := "Flag color:"
	if want == "" {
		if strings.Contains(buf.String(), label) {
			t.Errorf("unexpected color line in %q", buf.String())
		}
	} else if !strings.Contains(buf.String(), label+" "+want+"\n") {
		t.Errorf("expected %q in %q", label+" "+want, buf.String())
	}
}
