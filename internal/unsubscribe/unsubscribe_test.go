package unsubscribe

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name          string
		header        string
		post          string
		wantMechanism string
		wantAddress   string
		wantSubject   string
		wantBody      string
		wantURL       string
		wantOneClick  bool
	}{
		{
			name:          "simple mailto",
			header:        "<mailto:unsub@example.com>",
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@example.com",
		},
		{
			name:          "mailto with subject",
			header:        "<mailto:unsub@example.com?subject=Unsubscribe>",
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@example.com",
			wantSubject:   "Unsubscribe",
		},
		{
			name:          "mailto with subject and body",
			header:        "<mailto:unsub@example.com?subject=Unsub&body=Please%20remove%20me>",
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@example.com",
			wantSubject:   "Unsub",
			wantBody:      "Please remove me",
		},
		{
			name:          "simple https url",
			header:        "<https://example.com/unsubscribe?token=abc123>",
			wantMechanism: MechanismURL,
			wantURL:       "https://example.com/unsubscribe?token=abc123",
		},
		{
			name:          "http url",
			header:        "<http://example.com/unsub>",
			wantMechanism: MechanismURL,
			wantURL:       "http://example.com/unsub",
		},
		{
			name:          "both mailto and url",
			header:        "<mailto:unsub@example.com>, <https://example.com/unsub>",
			wantMechanism: MechanismBoth,
			wantAddress:   "unsub@example.com",
			wantURL:       "https://example.com/unsub",
		},
		{
			name:          "url before mailto",
			header:        "<https://example.com/unsub>, <mailto:unsub@example.com>",
			wantMechanism: MechanismBoth,
			wantAddress:   "unsub@example.com",
			wantURL:       "https://example.com/unsub",
		},
		{
			name:          "empty header",
			header:        "",
			wantMechanism: MechanismNone,
		},
		{
			name:          "whitespace only",
			header:        "   ",
			wantMechanism: MechanismNone,
		},
		{
			name:          "bare uri without brackets",
			header:        "mailto:unsub@example.com",
			wantMechanism: MechanismNone,
		},
		{
			name:          "one-click post header",
			header:        "<https://example.com/unsub>",
			post:          "List-Unsubscribe=One-Click",
			wantMechanism: MechanismURL,
			wantURL:       "https://example.com/unsub",
			wantOneClick:  true,
		},
		{
			name:          "one-click case insensitive",
			header:        "<https://example.com/unsub>",
			post:          "list-unsubscribe=one-click",
			wantMechanism: MechanismURL,
			wantURL:       "https://example.com/unsub",
			wantOneClick:  true,
		},
		{
			name:          "no one-click without post header",
			header:        "<https://example.com/unsub>",
			post:          "",
			wantMechanism: MechanismURL,
			wantURL:       "https://example.com/unsub",
			wantOneClick:  false,
		},
		{
			name:          "multiple mailtos takes first",
			header:        "<mailto:first@example.com>, <mailto:second@example.com>",
			wantMechanism: MechanismMailto,
			wantAddress:   "first@example.com",
		},
		{
			name:          "whitespace variations",
			header:        "  <mailto:unsub@example.com> ,  <https://example.com/unsub>  ",
			wantMechanism: MechanismBoth,
			wantAddress:   "unsub@example.com",
			wantURL:       "https://example.com/unsub",
		},
		{
			name:          "percent-encoded mailto address",
			header:        "<mailto:unsub%40token@example.com>",
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@token@example.com",
		},
		{
			name:          "mime encoded-word mailto",
			header:        "=?us-ascii?Q?=3Cmailto=3Aunsub=40example=2Ecom=3E?=",
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@example.com",
		},
		{
			name:          "mime encoded-word url",
			header:        "=?us-ascii?Q?=3Chttps=3A=2F=2Fexample=2Ecom=2Funsub=3E?=",
			wantMechanism: MechanismURL,
			wantURL:       "https://example.com/unsub",
		},
		{
			name:          "mime base64 encoded",
			header:        "=?us-ascii?B?PG1haWx0bzp1bnN1YkBleGFtcGxlLmNvbT4=?=",
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@example.com",
		},
		{
			name:          "malformed uri in brackets skipped",
			header:        "<not valid>, <mailto:unsub@example.com>",
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@example.com",
		},
		{
			name:          "empty brackets skipped",
			header:        "<>, <mailto:unsub@example.com>",
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.header, tt.post)

			if got.Mechanism != tt.wantMechanism {
				t.Errorf("Mechanism = %q, want %q", got.Mechanism, tt.wantMechanism)
			}

			if tt.wantAddress != "" {
				if got.Mailto == nil {
					t.Fatalf("Mailto is nil, want address %q", tt.wantAddress)
				}
				if got.Mailto.Address != tt.wantAddress {
					t.Errorf("Mailto.Address = %q, want %q", got.Mailto.Address, tt.wantAddress)
				}
			}

			if tt.wantSubject != "" && got.Mailto != nil && got.Mailto.Subject != tt.wantSubject {
				t.Errorf("Mailto.Subject = %q, want %q", got.Mailto.Subject, tt.wantSubject)
			}

			if tt.wantBody != "" && got.Mailto != nil && got.Mailto.Body != tt.wantBody {
				t.Errorf("Mailto.Body = %q, want %q", got.Mailto.Body, tt.wantBody)
			}

			if got.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tt.wantURL)
			}

			if got.OneClick != tt.wantOneClick {
				t.Errorf("OneClick = %v, want %v", got.OneClick, tt.wantOneClick)
			}
		})
	}
}

func TestExtractBracketedURIs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single", "<foo>", []string{"foo"}},
		{"multiple", "<a>, <b>, <c>", []string{"a", "b", "c"}},
		{"no brackets", "foo", nil},
		{"empty brackets", "<>", nil},
		{"nested ignored", "<<foo>>", []string{"<foo"}},
		{"unclosed", "<foo", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBracketedURIs(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("extractBracketedURIs(%q) returned %d URIs, want %d", tt.input, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractBracketedURIs(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}
