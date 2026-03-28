package client

import (
	"fmt"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/identity"
)

// GetAllIdentities retrieves all identities in the account.
// Results are cached for the lifetime of the Client instance.
//
// This method invokes Identity/get, which requires the JMAP submission
// capability. Importing the identity package transitively registers
// emailsubmission methods via init(). This is safe: fm never constructs
// EmailSubmission/set calls, so no email can be sent.
func (c *Client) GetAllIdentities() ([]*identity.Identity, error) {
	if c.identityCache != nil {
		return c.identityCache, nil
	}

	req := &jmap.Request{}
	req.Invoke(&identity.Get{
		Account: c.accountID,
	})

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("identity/get: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *identity.GetResponse:
			c.identityCache = r.List
			return r.List, nil
		case *jmap.MethodError:
			return nil, fmt.Errorf("identity/get: %s", r.Error())
		}
	}

	return nil, fmt.Errorf("identity/get: unexpected response")
}

// ResolveIdentityByEmail finds an identity by email address (case-insensitive).
// If no match is found, the error lists all available identity emails.
func (c *Client) ResolveIdentityByEmail(addr string) (*identity.Identity, error) {
	identities, err := c.GetAllIdentities()
	if err != nil {
		return nil, err
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no identities configured for this account")
	}

	lower := strings.ToLower(addr)
	for _, id := range identities {
		if strings.ToLower(id.Email) == lower {
			return id, nil
		}
	}

	available := make([]string, len(identities))
	for i, id := range identities {
		available[i] = id.Email
	}
	return nil, fmt.Errorf("no identity found for %q; available identities: %s",
		addr, strings.Join(available, ", "))
}
