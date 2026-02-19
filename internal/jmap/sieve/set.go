package sieve

import "git.sr.ht/~rockorager/go-jmap"

// Set creates, updates, or destroys sieve scripts.
// https://datatracker.ietf.org/doc/html/rfc9661#section-4
type Set struct {
	Account jmap.ID `json:"accountId,omitempty"`

	IfInState string `json:"ifInState,omitempty"`

	Create map[jmap.ID]*SieveScript `json:"create,omitempty"`

	Update map[jmap.ID]jmap.Patch `json:"update,omitempty"`

	Destroy []jmap.ID `json:"destroy,omitempty"`

	// OnSuccessActivateScript is the ID of the script to activate after a
	// successful set operation. Use a creation reference (e.g. "#create0")
	// to activate a newly created script. When omitted, activation state
	// is unchanged.
	OnSuccessActivateScript *jmap.ID `json:"onSuccessActivateScript,omitempty"`

	// OnSuccessDeactivateScript, when true, deactivates the currently active
	// script after a successful set operation. Deactivation is processed
	// before activation when both are present.
	OnSuccessDeactivateScript *bool `json:"onSuccessDeactivateScript,omitempty"`
}

// Name returns the JMAP method name.
func (m *Set) Name() string { return "SieveScript/set" }

// Requires returns the capability URIs this method depends on.
func (m *Set) Requires() []jmap.URI { return []jmap.URI{URI} }

// SetResponse is the server response to a SieveScript/set request.
type SetResponse struct {
	Account  jmap.ID `json:"accountId,omitempty"`
	OldState string  `json:"oldState,omitempty"`
	NewState string  `json:"newState,omitempty"`

	Created map[jmap.ID]*SieveScript `json:"created,omitempty"`
	Updated map[jmap.ID]*SieveScript `json:"updated,omitempty"`

	Destroyed []jmap.ID `json:"destroyed,omitempty"`

	NotCreated   map[jmap.ID]*jmap.SetError `json:"notCreated,omitempty"`
	NotUpdated   map[jmap.ID]*jmap.SetError `json:"notUpdated,omitempty"`
	NotDestroyed map[jmap.ID]*jmap.SetError `json:"notDestroyed,omitempty"`
}

func newSetResponse() jmap.MethodResponse { return &SetResponse{} }
