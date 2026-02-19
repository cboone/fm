package sieve

import "git.sr.ht/~rockorager/go-jmap"

// Query searches for sieve script IDs matching the given criteria.
// https://datatracker.ietf.org/doc/html/rfc9661#section-5
type Query struct {
	Account jmap.ID `json:"accountId,omitempty"`
}

// Name returns the JMAP method name.
func (m *Query) Name() string { return "SieveScript/query" }

// Requires returns the capability URIs this method depends on.
func (m *Query) Requires() []jmap.URI { return []jmap.URI{URI} }

// QueryResponse is the server response to a SieveScript/query request.
type QueryResponse struct {
	Account    jmap.ID   `json:"accountId,omitempty"`
	QueryState string    `json:"queryState,omitempty"`
	IDs        []jmap.ID `json:"ids,omitempty"`
	Total      uint64    `json:"total,omitempty"`
}

func newQueryResponse() jmap.MethodResponse { return &QueryResponse{} }
