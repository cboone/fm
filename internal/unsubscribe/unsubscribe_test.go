package unsubscribe

import (
	"os"
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
			wantMechanism: MechanismMailto,
			wantAddress:   "unsub@example.com",
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

func TestParseBareAndOneClick(t *testing.T) {
	const marker = "List-Unsubscribe=One-Click"
	tests := []struct {
		name, header, post, mechanism, url, address string
		oneClick                                    bool
		subject, body                               string
	}{
		{"bare HTTPS", "https://example.invalid/unsub", marker, MechanismURL, "https://example.invalid/unsub", "", true, "", ""},
		{"bare HTTP", "http://example.invalid/unsub", marker, MechanismURL, "http://example.invalid/unsub", "", false, "", ""},
		{"bare mailto", "mailto:unsub@example.invalid?subject=Unsubscribe&body=Remove%20me", marker, MechanismMailto, "", "unsub@example.invalid", false, "Unsubscribe", "Remove me"},
		{"surrounding whitespace", " \thttps://example.invalid/unsub\r\n", marker, MechanismURL, "https://example.invalid/unsub", "", true, "", ""},
		{"case insensitive scheme", "HTTPS://example.invalid/unsub", marker, MechanismURL, "HTTPS://example.invalid/unsub", "", true, "", ""},
		{"missing marker", "https://example.invalid/unsub", "", MechanismURL, "https://example.invalid/unsub", "", false, "", ""},
		{"bracketed HTTP", "<http://example.invalid/unsub>", marker, MechanismURL, "http://example.invalid/unsub", "", false, "", ""},
		{"bracketed mailto", "<mailto:unsub@example.invalid>", marker, MechanismMailto, "", "unsub@example.invalid", false, "", ""},
		{"bracketed alternatives", "<mailto:unsub@example.invalid>, <https://example.invalid/u,a>", marker, MechanismBoth, "https://example.invalid/u,a", "unsub@example.invalid", true, "", ""},
		{"bare payload comma", "https://example.invalid/u,a?key=a,b", marker, MechanismURL, "https://example.invalid/u,a?key=a,b", "", true, "", ""},
		{"encoded bare HTTPS", "=?us-ascii?Q?https=3A=2F=2Fexample.invalid=2Funsub?=", marker, MechanismURL, "https://example.invalid/unsub", "", true, "", ""},
		{"empty", "", marker, MechanismNone, "", "", false, "", ""},
		{"unsupported scheme", "ftp://example.invalid/unsub", marker, MechanismNone, "", "", false, "", ""},
		{"relative URL", "/unsubscribe", marker, MechanismNone, "", "", false, "", ""},
		{"missing host", "https:///unsubscribe", marker, MechanismNone, "", "", false, "", ""},
		{"opaque HTTPS", "https:unsubscribe", marker, MechanismNone, "", "", false, "", ""},
		{"invalid escape", "https://example.invalid/%zz", marker, MechanismNone, "", "", false, "", ""},
		{"embedded whitespace", "https://example.invalid/un subscribe", marker, MechanismNone, "", "", false, "", ""},
		{"multiple bare values", "mailto:unsub@example.invalid, https://example.invalid/unsub", marker, MechanismNone, "", "", false, "", ""},
		{"unclosed bracket", "<https://example.invalid/unsub", marker, MechanismNone, "", "", false, "", ""},
		{"unopened bracket", "https://example.invalid/unsub>", marker, MechanismNone, "", "", false, "", ""},
		{"malformed bracketed URL", "<https:///unsubscribe>", marker, MechanismNone, "", "", false, "", ""},
		{"unrecognized value", "not a URL", marker, MechanismNone, "", "", false, "", ""},
		{"prefixed marker", "https://example.invalid/unsub", "xList-Unsubscribe=One-Click", MechanismURL, "https://example.invalid/unsub", "", false, "", ""},
		{"suffixed marker", "<https://example.invalid/unsub>", "List-Unsubscribe=One-Click-disabled", MechanismURL, "https://example.invalid/unsub", "", false, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "bare HTTPS" && os.Getenv("FM_PARSER_TEST_WRONG_URL") == "1" {
				tt.url = "https://wrong.invalid/judge"
			}
			got := Parse(tt.header, tt.post)
			if got.Mechanism != tt.mechanism || got.URL != tt.url || got.OneClick != tt.oneClick {
				t.Errorf("Parse(%q, %q) = %+v; want mechanism=%q URL=%q oneClick=%v", tt.header, tt.post, got, tt.mechanism, tt.url, tt.oneClick)
			}
			if tt.address == "" {
				if got.Mailto != nil {
					t.Errorf("Mailto = %+v, want nil", got.Mailto)
				}
			} else if got.Mailto == nil || got.Mailto.Address != tt.address {
				t.Errorf("Mailto = %+v, want address %q", got.Mailto, tt.address)
			}
			if got.Mailto != nil && (got.Mailto.Subject != tt.subject || got.Mailto.Body != tt.body) {
				t.Errorf("Mailto = %+v, want decoded subject and body", got.Mailto)
			}
			if tt.header != "" && got.Raw != decodeHeader(tt.header) {
				t.Errorf("Raw = %q, want decoded original header", got.Raw)
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
