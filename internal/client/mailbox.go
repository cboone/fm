package client

import (
	"fmt"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"

	"github.com/cboone/jm/internal/types"
)

// GetAllMailboxes retrieves all mailboxes in the account.
// Results are cached for the lifetime of the Client instance.
func (c *Client) GetAllMailboxes() ([]*mailbox.Mailbox, error) {
	if c.mailboxCache != nil {
		return c.mailboxCache, nil
	}

	req := &jmap.Request{}
	req.Invoke(&mailbox.Get{
		Account: c.accountID,
	})

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("mailbox/get: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *mailbox.GetResponse:
			c.mailboxCache = r.List
			return r.List, nil
		case *jmap.MethodError:
			return nil, fmt.Errorf("mailbox/get: %s", r.Error())
		}
	}

	return nil, fmt.Errorf("mailbox/get: unexpected response")
}

// GetMailboxByRole finds a mailbox by its JMAP role.
func (c *Client) GetMailboxByRole(role mailbox.Role) (*mailbox.Mailbox, error) {
	mailboxes, err := c.GetAllMailboxes()
	if err != nil {
		return nil, err
	}
	for _, mb := range mailboxes {
		if mb.Role == role {
			return mb, nil
		}
	}
	return nil, fmt.Errorf("no mailbox found with role %q", role)
}

// GetMailboxByNameOrID finds a mailbox by name (case-insensitive) or by ID.
func (c *Client) GetMailboxByNameOrID(nameOrID string) (*mailbox.Mailbox, error) {
	mailboxes, err := c.GetAllMailboxes()
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(nameOrID)
	for _, mb := range mailboxes {
		if string(mb.ID) == nameOrID {
			return mb, nil
		}
		if strings.ToLower(mb.Name) == lower {
			return mb, nil
		}
	}
	return nil, fmt.Errorf("mailbox not found: %q", nameOrID)
}

// ResolveMailboxID resolves "inbox", other role names, a mailbox name, or a
// raw mailbox ID to a JMAP mailbox ID.
func (c *Client) ResolveMailboxID(nameOrID string) (jmap.ID, error) {
	// Check if it's a well-known role name first.
	role := mailbox.Role(strings.ToLower(nameOrID))
	switch role {
	case mailbox.RoleInbox, mailbox.RoleArchive, mailbox.RoleJunk,
		mailbox.RoleDrafts, mailbox.RoleSent, mailbox.RoleTrash:
		mb, err := c.GetMailboxByRole(role)
		if err == nil {
			return mb.ID, nil
		}
	}

	mb, err := c.GetMailboxByNameOrID(nameOrID)
	if err != nil {
		return "", err
	}
	return mb.ID, nil
}

// ListMailboxes returns simplified mailbox info for output.
func (c *Client) ListMailboxes(rolesOnly bool) ([]types.MailboxInfo, error) {
	mailboxes, err := c.GetAllMailboxes()
	if err != nil {
		return nil, err
	}

	var result []types.MailboxInfo
	for _, mb := range mailboxes {
		if rolesOnly && mb.Role == "" {
			continue
		}
		result = append(result, types.MailboxInfo{
			ID:           string(mb.ID),
			Name:         mb.Name,
			Role:         string(mb.Role),
			TotalEmails:  mb.TotalEmails,
			UnreadEmails: mb.UnreadEmails,
			ParentID:     string(mb.ParentID),
		})
	}
	return result, nil
}
