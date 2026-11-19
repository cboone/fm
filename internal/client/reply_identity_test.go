package client

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/identity"
)

func TestCreateDraft_ReplyIdentity(t *testing.T) {
	for _, mode := range []DraftMode{DraftModeReply, DraftModeReplyAll} {
		for _, tt := range []struct {
			name, from, want string
			to, cc           []*mail.Address
		}{
			{"To match", "", "alias@example.com", []*mail.Address{{Email: "ALIAS@EXAMPLE.COM"}}, nil},
			{"Cc match", "", "alias@example.com", []*mail.Address{{Email: "other@example.com"}}, []*mail.Address{{Email: "alias@example.com"}}},
			{"To before Cc", "", "second@example.com", []*mail.Address{{Email: "second@example.com"}}, []*mail.Address{{Email: "alias@example.com"}}},
			{"message order", "", "second@example.com", []*mail.Address{{Email: "second@example.com"}, {Email: "alias@example.com"}}, nil},
			{"no match", "", "user@fastmail.com", []*mail.Address{{Email: "other@example.com"}}, nil},
			{"override", "second@example.com", "second@example.com", []*mail.Address{{Email: "alias@example.com"}}, nil},
		} {
			t.Run(string(mode)+"/"+tt.name, func(t *testing.T) {
				orig := &email.Email{From: []*mail.Address{{Email: "sender@example.com"}}, To: tt.to, CC: tt.cc, Subject: "Synthetic"}
				identities := []*identity.Identity{
					{Email: "alias@example.com", Name: "Alias"},
					{Email: "second@example.com", Name: "Second"},
				}
				var created *email.Email
				identityCalls := 0
				c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
					switch r := req.Calls[0].Args.(type) {
					case *email.Get:
						return &jmap.Response{Responses: []*jmap.Invocation{{Name: "Email/get", Args: &email.GetResponse{List: []*email.Email{orig}}}}}, nil
					case *identity.Get:
						identityCalls++
						return mockIdentityGetSuccess(identities)(req)
					case *email.Set:
						created = r.Create["draft-0"]
						return mockDraftCreateSuccess("synthetic")(req)
					default:
						t.Fatalf("unexpected synthetic request: %T", r)
						return nil, nil
					}
				})
				c.identityCache = nil
				result, err := c.CreateDraft(DraftOptions{Mode: mode, OriginalID: "original", From: tt.from, Body: "Synthetic"})
				if err != nil {
					t.Fatal(err)
				}
				if identityCalls != 1 || created == nil || len(created.From) != 1 || created.From[0].Email != tt.want {
					t.Fatalf("identity calls=%d, created=%+v; want From %s", identityCalls, created, tt.want)
				}
				if !reflect.DeepEqual(result.From, convertAddresses(created.From)) || len(result.Warnings) != 0 {
					t.Fatalf("result From/warnings=%+v/%v; want draft From and no warnings", result.From, result.Warnings)
				}
				if tt.want == "alias@example.com" && created.From[0].Name != "Alias" {
					t.Fatalf("From name=%q, want identity name Alias", created.From[0].Name)
				}
				if mode == DraftModeReplyAll {
					for _, addr := range created.CC {
						if strings.EqualFold(addr.Email, tt.want) {
							t.Fatalf("selected identity remains in Cc: %+v", created.CC)
						}
					}
				}
			})
		}
	}
}

func TestCreateDraft_ReplyIdentityMissingCapability(t *testing.T) {
	for _, mode := range []DraftMode{DraftModeReply, DraftModeReplyAll} {
		for _, explicit := range []bool{false, true} {
			t.Run(string(mode)+"/explicit="+strconv.FormatBool(explicit), func(t *testing.T) {
				sets := 0
				c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
					switch r := req.Calls[0].Args.(type) {
					case *email.Get:
						return &jmap.Response{Responses: []*jmap.Invocation{{Name: "Email/get", Args: &email.GetResponse{List: []*email.Email{{From: []*mail.Address{{Email: "sender@example.com"}}, To: []*mail.Address{{Email: "alias@example.com"}}}}}}}}, nil
					case *email.Set:
						sets++
						if len(r.Create["draft-0"].From) != 1 || r.Create["draft-0"].From[0].Email != "user@fastmail.com" {
							t.Fatal("missing capability did not retain default sender")
						}
						return mockDraftCreateSuccess("synthetic")(req)
					default:
						t.Fatalf("Identity/get or unexpected request without capability: %T", r)
						return nil, nil
					}
				})
				c.identityCache = nil
				c.jmap.Session.RawCapabilities = nil
				opts := DraftOptions{Mode: mode, OriginalID: "original", Body: "Synthetic"}
				if explicit {
					opts.From = "alias@example.com"
				}
				result, err := c.CreateDraft(opts)
				if explicit {
					if !errors.Is(err, ErrMissingSubmissionCapability) || sets != 0 {
						t.Fatalf("explicit From: error=%v sets=%d; want capability error and no create", err, sets)
					}
					return
				}
				if err != nil || sets != 1 || len(result.From) != 1 || result.From[0].Email != "user@fastmail.com" {
					t.Fatalf("fallback: result=%+v error=%v sets=%d", result, err, sets)
				}
				if len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0], "using the current default sender") || !strings.Contains(result.Warnings[0], ErrMissingSubmissionCapability.Error()) {
					t.Fatalf("fallback warnings=%v; want default sender and capability explanation", result.Warnings)
				}
			})
		}
	}
}

func TestCreateDraft_NonReplyNoIdentityLookup(t *testing.T) {
	for _, mode := range []DraftMode{DraftModeNew, DraftModeForward} {
		t.Run(string(mode), func(t *testing.T) {
			c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
				switch r := req.Calls[0].Args.(type) {
				case *email.Get:
					return &jmap.Response{Responses: []*jmap.Invocation{{Name: "Email/get", Args: &email.GetResponse{List: []*email.Email{{}}}}}}, nil
				case *email.Set:
					if len(r.Create["draft-0"].From) != 1 || r.Create["draft-0"].From[0].Email != "user@fastmail.com" {
						t.Fatal("non-reply sender changed")
					}
					return mockDraftCreateSuccess("synthetic")(req)
				default:
					t.Fatalf("non-reply identity lookup: %T", r)
					return nil, nil
				}
			})
			c.identityCache = nil
			c.jmap.Session.RawCapabilities = nil
			result, err := c.CreateDraft(DraftOptions{Mode: mode, OriginalID: "original", Body: "Synthetic"})
			if err != nil || len(result.Warnings) != 0 {
				t.Fatalf("non-reply result=%+v error=%v; want no capability requirement or warning", result, err)
			}
		})
	}
}

func TestCreateDraft_ReplyIdentityLookupError(t *testing.T) {
	lookupErr := errors.New("synthetic identity failure")
	c := testClientForDraft(func(req *jmap.Request) (*jmap.Response, error) {
		switch r := req.Calls[0].Args.(type) {
		case *email.Get:
			return &jmap.Response{Responses: []*jmap.Invocation{{Name: "Email/get", Args: &email.GetResponse{List: []*email.Email{{}}}}}}, nil
		case *identity.Get:
			return nil, lookupErr
		default:
			t.Fatalf("draft created after identity lookup error: %T", r)
			return nil, nil
		}
	})
	c.identityCache = nil
	_, err := c.CreateDraft(DraftOptions{Mode: DraftModeReply, OriginalID: "original", Body: "Synthetic"})
	if !errors.Is(err, lookupErr) {
		t.Fatalf("error=%v, want original lookup error", err)
	}
}
