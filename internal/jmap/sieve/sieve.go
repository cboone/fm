// Package sieve implements the JMAP SieveScript extension (RFC 9661).
package sieve

import "git.sr.ht/~rockorager/go-jmap"

// URI is the capability identifier for JMAP Sieve script management.
const URI jmap.URI = "urn:ietf:params:jmap:sieve"

func init() {
	jmap.RegisterCapability(&Capability{})
	jmap.RegisterMethod("SieveScript/get", newGetResponse)
	jmap.RegisterMethod("SieveScript/set", newSetResponse)
	jmap.RegisterMethod("SieveScript/query", newQueryResponse)
	jmap.RegisterMethod("SieveScript/validate", newValidateResponse)
}

// Capability is the JMAP capability object for urn:ietf:params:jmap:sieve.
type Capability struct{}

// URI returns the sieve capability URI.
func (c *Capability) URI() jmap.URI { return URI }

// New returns a new empty Capability instance.
func (c *Capability) New() jmap.Capability { return &Capability{} }

// SieveScript represents a server-side sieve filtering script (RFC 9661, Section 2).
type SieveScript struct {
	ID       jmap.ID `json:"id,omitempty"`
	Name     string  `json:"name,omitempty"`
	BlobID   jmap.ID `json:"blobId,omitempty"`
	IsActive bool    `json:"isActive,omitempty"`
}
