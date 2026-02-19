package sieve

import "git.sr.ht/~rockorager/go-jmap"

// Get retrieves sieve scripts by ID, or all scripts when IDs is nil.
// https://datatracker.ietf.org/doc/html/rfc9661#section-3
type Get struct {
	Account    jmap.ID   `json:"accountId,omitempty"`
	IDs        []jmap.ID `json:"ids,omitempty"`
	Properties []string  `json:"properties,omitempty"`
}

// Name returns the JMAP method name.
func (m *Get) Name() string { return "SieveScript/get" }

// Requires returns the capability URIs this method depends on.
func (m *Get) Requires() []jmap.URI { return []jmap.URI{URI} }

// GetResponse is the server response to a SieveScript/get request.
type GetResponse struct {
	Account  jmap.ID        `json:"accountId,omitempty"`
	State    string         `json:"state,omitempty"`
	List     []*SieveScript `json:"list,omitempty"`
	NotFound []jmap.ID      `json:"notFound,omitempty"`
}

func newGetResponse() jmap.MethodResponse { return &GetResponse{} }
