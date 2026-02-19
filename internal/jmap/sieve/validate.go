package sieve

import "git.sr.ht/~rockorager/go-jmap"

// Validate checks sieve script syntax without storing the script.
// https://datatracker.ietf.org/doc/html/rfc9661#section-6
type Validate struct {
	Account jmap.ID `json:"accountId,omitempty"`
	BlobID  jmap.ID `json:"blobId,omitempty"`
}

// Name returns the JMAP method name.
func (m *Validate) Name() string { return "SieveScript/validate" }

// Requires returns the capability URIs this method depends on.
func (m *Validate) Requires() []jmap.URI { return []jmap.URI{URI} }

// ValidateResponse is the server response to a SieveScript/validate request.
type ValidateResponse struct {
	Account jmap.ID        `json:"accountId,omitempty"`
	Error   *jmap.SetError `json:"error,omitempty"`
}

func newValidateResponse() jmap.MethodResponse { return &ValidateResponse{} }
