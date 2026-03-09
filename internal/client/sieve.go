package client

import (
	"fmt"
	"io"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"

	"github.com/cboone/fm/internal/jmap/sieve"
	"github.com/cboone/fm/internal/types"
)

// Blank import triggers sieve capability and method registration.
var _ = sieve.URI

// hasSieveCapability checks whether the server advertises sieve support.
func (c *Client) hasSieveCapability() bool {
	if c.jmap == nil || c.jmap.Session == nil {
		return false
	}
	_, ok := c.jmap.Session.RawCapabilities[sieve.URI]
	return ok
}

// requireSieve returns an error if the server does not support sieve.
func (c *Client) requireSieve() error {
	if !c.hasSieveCapability() {
		return fmt.Errorf("server does not support sieve scripts (missing %s capability)", sieve.URI)
	}
	return nil
}

// ListSieveScripts returns all sieve scripts in the account.
func (c *Client) ListSieveScripts() (types.SieveScriptListResult, error) {
	if err := c.requireSieve(); err != nil {
		return types.SieveScriptListResult{}, err
	}

	req := &jmap.Request{}
	req.Invoke(&sieve.Get{Account: c.accountID})

	resp, err := c.Do(req)
	if err != nil {
		return types.SieveScriptListResult{}, fmt.Errorf("listing sieve scripts: %w", err)
	}

	var result types.SieveScriptListResult
	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *sieve.GetResponse:
			for _, s := range r.List {
				result.Scripts = append(result.Scripts, types.SieveScriptInfo{
					ID:       string(s.ID),
					Name:     s.Name,
					IsActive: s.IsActive,
				})
			}
			result.Total = len(result.Scripts)
		case *jmap.MethodError:
			return types.SieveScriptListResult{}, fmt.Errorf("listing sieve scripts: %s", r.Error())
		}
	}

	return result, nil
}

// GetSieveScript returns a sieve script's metadata and content.
func (c *Client) GetSieveScript(id string) (types.SieveScriptDetail, error) {
	if err := c.requireSieve(); err != nil {
		return types.SieveScriptDetail{}, err
	}

	req := &jmap.Request{}
	req.Invoke(&sieve.Get{
		Account: c.accountID,
		IDs:     []jmap.ID{jmap.ID(id)},
	})

	resp, err := c.Do(req)
	if err != nil {
		return types.SieveScriptDetail{}, fmt.Errorf("getting sieve script: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *sieve.GetResponse:
			if len(r.NotFound) > 0 {
				return types.SieveScriptDetail{}, fmt.Errorf("sieve script %s: %w", id, ErrNotFound)
			}
			if len(r.List) == 0 {
				return types.SieveScriptDetail{}, fmt.Errorf("sieve script %s: %w", id, ErrNotFound)
			}
			s := r.List[0]
			content, err := c.getSieveScriptContent(s.BlobID)
			if err != nil {
				return types.SieveScriptDetail{}, err
			}
			return types.SieveScriptDetail{
				ID:       string(s.ID),
				Name:     s.Name,
				BlobID:   string(s.BlobID),
				IsActive: s.IsActive,
				Content:  content,
			}, nil
		case *jmap.MethodError:
			return types.SieveScriptDetail{}, fmt.Errorf("getting sieve script: %s", r.Error())
		}
	}

	return types.SieveScriptDetail{}, fmt.Errorf("getting sieve script: unexpected response")
}

// getSieveScriptContent downloads a sieve script's content by blob ID.
func (c *Client) getSieveScriptContent(blobID jmap.ID) (string, error) {
	body, err := c.Download(c.accountID, blobID)
	if err != nil {
		return "", fmt.Errorf("downloading sieve script content: %w", err)
	}
	defer func() { _ = body.Close() }()

	data, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("reading sieve script content: %w", err)
	}
	return string(data), nil
}

// CreateSieveScript uploads script content as a blob and creates a SieveScript.
// When activate is true, the script is activated upon creation.
func (c *Client) CreateSieveScript(name, content string, activate bool) (types.SieveCreateResult, error) {
	if err := c.requireSieve(); err != nil {
		return types.SieveCreateResult{}, err
	}

	upload, err := c.Upload(c.accountID, strings.NewReader(content))
	if err != nil {
		return types.SieveCreateResult{}, fmt.Errorf("uploading sieve script: %w", err)
	}

	createID := jmap.ID("create0")
	set := &sieve.Set{
		Account: c.accountID,
		Create: map[jmap.ID]*sieve.SieveScript{
			createID: {Name: name, BlobID: upload.ID},
		},
	}

	if activate {
		ref := jmap.ID("#create0")
		set.OnSuccessActivateScript = &ref
	}

	req := &jmap.Request{}
	req.Invoke(set)

	resp, err := c.Do(req)
	if err != nil {
		return types.SieveCreateResult{}, fmt.Errorf("creating sieve script: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *sieve.SetResponse:
			if created, ok := r.Created[createID]; ok {
				return types.SieveCreateResult{
					ID:       string(created.ID),
					Name:     name,
					BlobID:   string(upload.ID),
					IsActive: activate,
					Content:  content,
				}, nil
			}
			if setErr, ok := r.NotCreated[createID]; ok {
				desc := "unknown error"
				if setErr.Description != nil {
					desc = *setErr.Description
				}
				return types.SieveCreateResult{}, fmt.Errorf("creating sieve script: %s", desc)
			}
		case *jmap.MethodError:
			return types.SieveCreateResult{}, fmt.Errorf("creating sieve script: %s", r.Error())
		}
	}

	return types.SieveCreateResult{}, fmt.Errorf("creating sieve script: unexpected response")
}

// ValidateSieveScript uploads content as a blob and validates it server-side.
func (c *Client) ValidateSieveScript(content string) (types.SieveValidateResult, error) {
	if err := c.requireSieve(); err != nil {
		return types.SieveValidateResult{}, err
	}

	upload, err := c.Upload(c.accountID, strings.NewReader(content))
	if err != nil {
		return types.SieveValidateResult{}, fmt.Errorf("uploading sieve script for validation: %w", err)
	}

	req := &jmap.Request{}
	req.Invoke(&sieve.Validate{
		Account: c.accountID,
		BlobID:  upload.ID,
	})

	resp, err := c.Do(req)
	if err != nil {
		return types.SieveValidateResult{}, fmt.Errorf("validating sieve script: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *sieve.ValidateResponse:
			result := types.SieveValidateResult{
				Valid:   r.Error == nil,
				Content: content,
			}
			if r.Error != nil && r.Error.Description != nil {
				result.Error = *r.Error.Description
			}
			return result, nil
		case *jmap.MethodError:
			return types.SieveValidateResult{}, fmt.Errorf("validating sieve script: %s", r.Error())
		}
	}

	return types.SieveValidateResult{}, fmt.Errorf("validating sieve script: unexpected response")
}

// ActivateSieveScript activates a script by ID, deactivating any currently active one.
func (c *Client) ActivateSieveScript(id string) (types.SieveActivateResult, error) {
	if err := c.requireSieve(); err != nil {
		return types.SieveActivateResult{}, err
	}

	scriptID := jmap.ID(id)
	set := &sieve.Set{
		Account:                 c.accountID,
		OnSuccessActivateScript: &scriptID,
	}

	req := &jmap.Request{}
	req.Invoke(set)

	resp, err := c.Do(req)
	if err != nil {
		return types.SieveActivateResult{}, fmt.Errorf("activating sieve script: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *sieve.SetResponse:
			return types.SieveActivateResult{
				ID:       id,
				IsActive: true,
			}, nil
		case *jmap.MethodError:
			return types.SieveActivateResult{}, fmt.Errorf("activating sieve script: %s", r.Error())
		}
	}

	return types.SieveActivateResult{}, fmt.Errorf("activating sieve script: unexpected response")
}

// DeactivateSieveScript deactivates the currently active script.
func (c *Client) DeactivateSieveScript() (types.SieveActivateResult, error) {
	if err := c.requireSieve(); err != nil {
		return types.SieveActivateResult{}, err
	}

	deactivate := true
	set := &sieve.Set{
		Account:                   c.accountID,
		OnSuccessDeactivateScript: &deactivate,
	}

	req := &jmap.Request{}
	req.Invoke(set)

	resp, err := c.Do(req)
	if err != nil {
		return types.SieveActivateResult{}, fmt.Errorf("deactivating sieve script: %w", err)
	}

	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *sieve.SetResponse:
			return types.SieveActivateResult{
				IsActive: false,
			}, nil
		case *jmap.MethodError:
			return types.SieveActivateResult{}, fmt.Errorf("deactivating sieve script: %s", r.Error())
		}
	}

	return types.SieveActivateResult{}, fmt.Errorf("deactivating sieve script: unexpected response")
}

// DeleteSieveScript deletes a script by ID. Returns an error if the script is
// currently active (it must be deactivated first).
func (c *Client) DeleteSieveScript(id string) (types.SieveDeleteResult, error) {
	if err := c.requireSieve(); err != nil {
		return types.SieveDeleteResult{}, err
	}

	req := &jmap.Request{}
	req.Invoke(&sieve.Get{
		Account: c.accountID,
		IDs:     []jmap.ID{jmap.ID(id)},
	})
	req.Invoke(&sieve.Set{
		Account: c.accountID,
		Destroy: []jmap.ID{jmap.ID(id)},
	})

	resp, err := c.Do(req)
	if err != nil {
		return types.SieveDeleteResult{}, fmt.Errorf("deleting sieve script: %w", err)
	}

	var name string
	for _, inv := range resp.Responses {
		switch r := inv.Args.(type) {
		case *sieve.GetResponse:
			if len(r.List) > 0 {
				name = r.List[0].Name
			}
		case *sieve.SetResponse:
			if len(r.Destroyed) > 0 {
				return types.SieveDeleteResult{ID: id, Name: name}, nil
			}
			if setErr, ok := r.NotDestroyed[jmap.ID(id)]; ok {
				desc := "unknown error"
				if setErr.Description != nil {
					desc = *setErr.Description
				}
				if setErr.Type == "sieveIsActive" {
					return types.SieveDeleteResult{}, fmt.Errorf("cannot delete active sieve script: deactivate it first")
				}
				return types.SieveDeleteResult{}, fmt.Errorf("deleting sieve script: %s", desc)
			}
		case *jmap.MethodError:
			return types.SieveDeleteResult{}, fmt.Errorf("deleting sieve script: %s", r.Error())
		}
	}

	return types.SieveDeleteResult{}, fmt.Errorf("deleting sieve script: unexpected response")
}
