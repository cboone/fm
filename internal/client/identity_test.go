package client

import (
	"fmt"
	"strings"
	"testing"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/identity"
)

func testIdentities() []*identity.Identity {
	return []*identity.Identity{
		{ID: "id-primary", Name: "Chris", Email: "chris@fastmail.com"},
		{ID: "id-alias", Name: "Chris B", Email: "chris@sent.com"},
		{ID: "id-other", Name: "", Email: "cboone@fea.st"},
	}
}

func mockIdentityGetSuccess(identities []*identity.Identity) func(*jmap.Request) (*jmap.Response, error) {
	return func(req *jmap.Request) (*jmap.Response, error) {
		return &jmap.Response{Responses: []*jmap.Invocation{
			{Name: "Identity/get", CallID: "0", Args: &identity.GetResponse{
				List: identities,
			}},
		}}, nil
	}
}

func TestGetAllIdentities_Success(t *testing.T) {
	ids := testIdentities()
	c := &Client{
		accountID: "test-account",
		doFunc:    mockIdentityGetSuccess(ids),
	}

	result, err := c.GetAllIdentities()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("GetAllIdentities() returned %d identities, want 3", len(result))
	}
	if result[0].Email != "chris@fastmail.com" {
		t.Errorf("GetAllIdentities()[0].Email = %q, want %q", result[0].Email, "chris@fastmail.com")
	}
}

func TestGetAllIdentities_Cached(t *testing.T) {
	calls := 0
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			calls++
			return mockIdentityGetSuccess(testIdentities())(req)
		},
	}

	if _, err := c.GetAllIdentities(); err != nil {
		t.Fatalf("first call: %v", err)
	}
	if _, err := c.GetAllIdentities(); err != nil {
		t.Fatalf("second call: %v", err)
	}
	if calls != 1 {
		t.Errorf("doFunc called %d times, want 1 (cached)", calls)
	}
}

func TestGetAllIdentities_DoError(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return nil, fmt.Errorf("network error")
		},
	}

	_, err := c.GetAllIdentities()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "network error") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "network error")
	}
}

func TestGetAllIdentities_MethodError(t *testing.T) {
	c := &Client{
		accountID: "test-account",
		doFunc: func(req *jmap.Request) (*jmap.Response, error) {
			return &jmap.Response{Responses: []*jmap.Invocation{
				{Name: "Identity/get", CallID: "0", Args: &jmap.MethodError{
					Type: "accountNotFound",
				}},
			}}, nil
		},
	}

	_, err := c.GetAllIdentities()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "identity/get") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "identity/get")
	}
}

func TestResolveIdentityByEmail_ExactMatch(t *testing.T) {
	c := &Client{identityCache: testIdentities()}

	id, err := c.ResolveIdentityByEmail("chris@sent.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.ID != "id-alias" {
		t.Errorf("ID = %q, want %q", id.ID, "id-alias")
	}
	if id.Name != "Chris B" {
		t.Errorf("Name = %q, want %q", id.Name, "Chris B")
	}
}

func TestResolveIdentityByEmail_CaseInsensitive(t *testing.T) {
	c := &Client{identityCache: testIdentities()}

	id, err := c.ResolveIdentityByEmail("Chris@Sent.COM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.ID != "id-alias" {
		t.Errorf("ID = %q, want %q", id.ID, "id-alias")
	}
}

func TestResolveIdentityByEmail_NoMatch(t *testing.T) {
	c := &Client{identityCache: testIdentities()}

	_, err := c.ResolveIdentityByEmail("unknown@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no identity found") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "no identity found")
	}
	if !strings.Contains(err.Error(), "chris@fastmail.com") {
		t.Errorf("error = %q, want it to list available identities", err.Error())
	}
	if !strings.Contains(err.Error(), "chris@sent.com") {
		t.Errorf("error = %q, want it to list available identities", err.Error())
	}
}

func TestResolveIdentityByEmail_EmptyList(t *testing.T) {
	c := &Client{identityCache: []*identity.Identity{}}

	_, err := c.ResolveIdentityByEmail("any@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no identities configured") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "no identities configured")
	}
}
