package unsubscribe

import (
	"strings"
	"testing"
	"unicode"
)

func TestEqualFoldASCII(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"List-Unsubscribe", "LIST-UNSUBSCRIBE", true},
		{"", "", true},
		{"s", "k", false},
		{"ſ", "s", false},
		{"s", "ſ", false},
		{"K", "k", false},
		{"k", "K", false},
		{"é", "é", false},
	} {
		if got := EqualFoldASCII(tc.a, tc.b); got != tc.want {
			t.Errorf("EqualFoldASCII(%q,%q)=%v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
	for a := 'A'; a <= 'Z'; a++ {
		for r := unicode.SimpleFold(a); r != a; r = unicode.SimpleFold(r) {
			if r > unicode.MaxASCII && EqualFoldASCII(string(r), string(a)) {
				t.Errorf("accepted non-ASCII folding rune U+%04X for %c", r, a)
			}
		}
	}
}

func TestProtocolMarkerASCII(t *testing.T) {
	const marker = "List-Unsubscribe=One-Click"
	for _, bad := range []string{"Liſt-Unsubscribe=One-Click", "List-Unsubscribe=One-ClicK", "Lİst-Unsubscribe=One-Click"} {
		if Parse("https://example.invalid/unsub", bad).OneClick {
			t.Errorf("accepted non-ASCII marker %q", bad)
		}
		if _, err := Target([]Header{{Name: "List-Unsubscribe", Value: "<https://example.invalid/unsub>"}, {Name: "List-Unsubscribe-Post", Value: bad}}); err == nil {
			t.Errorf("Target accepted non-ASCII marker %q", bad)
		}
	}
	if !Parse("https://example.invalid/unsub", " \t"+strings.ToLower(marker)+"\r\n").OneClick {
		t.Error("rejected ASCII case and whitespace marker")
	}
	if _, err := Target([]Header{{Name: "LIST-UNSUBSCRIBE", Value: "<https://example.invalid/unsub>"}, {Name: "LIST-UNSUBSCRIBE-POST", Value: marker}}); err != nil {
		t.Errorf("rejected ASCII header names: %v", err)
	}
	if _, err := Target([]Header{{Name: "List-Unsubscribe", Value: "<https://example.invalid/unsub>"}, {Name: "List-Unsubscribe-Post", Value: strings.ToLower(marker)}}); err == nil {
		t.Error("execution accepted noncanonical ASCII marker")
	}
	for _, name := range []string{"Liſt-Unsubscribe", "Liſt-Unsubscribe-Post", "Lİst-Unsubscribe", "Lİst-Unsubscribe-Post"} {
		headers := []Header{{Name: "List-Unsubscribe", Value: "<https://example.invalid/unsub>"}, {Name: "List-Unsubscribe-Post", Value: marker}}
		index := 0
		if strings.HasSuffix(name, "Post") {
			index = 1
		}
		headers[index].Name = name
		if _, err := Target(headers); err == nil {
			t.Errorf("accepted non-ASCII header %q", name)
		}
	}
}

func TestAuthenticatedASCII(t *testing.T) {
	for _, tc := range []struct {
		name             string
		index            int
		header, old, new string
	}{
		{"result header", 2, "Authentication-Reſults", "", ""},
		{"signature header", 4, "DKIM-Signature", "", ""},
		{"long s signature", 4, "DKIM-ſignature", "", ""},
		{"method Kelvin", 2, "", "dkim=pass", "dKim=pass"},
		{"method long s", 2, "", "dkim=pass", "dkim=paſs"},
		{"covered header long s", 4, "", "List-Unsubscribe:", "Liſt-Unsubscribe:"},
		{"covered header dotted I", 4, "", "List-Unsubscribe:", "Lİst-Unsubscribe:"},
		{"internal host dotted I", 0, "", "compute1.internal", "compute1.İnternal"},
		{"authserv id", 2, "", "mx1.messagingengine.com; dkim=pass", "mx1.meſſagingengine.com; dkim=pass"},
		{"result domain", 2, "", "header.d=sender.example", "header.d=ſender.example"},
		{"signature domain", 4, "", "d=sender.example", "d=ſender.example"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			headers := authenticatedHeaders()
			if tc.header != "" {
				headers[tc.index].Name = tc.header
			} else {
				headers[tc.index].Value = strings.Replace(headers[tc.index].Value, tc.old, tc.new, 1)
			}
			if err := Authenticated(headers, "sender.example"); err == nil {
				t.Errorf("accepted non-ASCII authentication token in %s", tc.name)
			}
		})
	}
}

func TestAuthenticatedProviderASCII(t *testing.T) {
	for _, host := range []string{"mx1.meſſagingengine.com", "mK.messagingengine.com"} {
		headers := authenticatedHeaders()
		headers[3].Value = strings.Replace(headers[3].Value, "mx1.messagingengine.com", host, 1)
		err := Authenticated(headers, "sender.example")
		if err == nil || !strings.Contains(err.Error(), "receiving boundary unavailable") {
			t.Errorf("provider %q: want unavailable boundary, got %v", host, err)
		}
	}
	if err := Authenticated(authenticatedHeaders(), "ſender.example"); err == nil {
		t.Error("accepted non-ASCII sender domain")
	}
}

func TestEqualSenderAddressASCII(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"A@LIST.EXAMPLE", "a@list.example", true},
		{"ſ@list.example", "s@list.example", true},
		{"a@liſt.example", "a@list.example", false},
		{"a@list.example", "a@liſt.example", false},
		{"a@K.example", "a@k.example", false},
		{"a@k.example", "a@K.example", false},
		{"a@ſ.example", "a@ſ.example", false},
		{"address", "address", false},
	} {
		if got := EqualSenderAddress(tc.a, tc.b); got != tc.want {
			t.Errorf("EqualSenderAddress(%q,%q)=%v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
