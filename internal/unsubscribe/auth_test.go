package unsubscribe

import "testing"

func authenticatedHeaders() []Header {
	return []Header{
		{Name: "Received", Value: "from compute1.internal (compute1.internal [10.0.0.1]) by mail.internal.messagingengine.com with LMTP; date"},
		{Name: "Authentication-Results", Value: "mx1.messagingengine.com; x-me-sender=none"},
		{Name: "Authentication-Results", Value: "mx1.messagingengine.com; dkim=pass (2048-bit rsa (nested)) header.d=sender.example header.b=abc"},
		{Name: "Received", Value: "from smtp.sender.example (smtp.sender.example [93.184.216.34]) by mx1.messagingengine.com with ESMTP; date"},
		{Name: "DKIM-Signature", Value: "v=1; d=sender.example; b=abcdef; h=from:List-Unsubscribe:List-Unsubscribe-Post"},
	}
}

func TestAuthenticated(t *testing.T) {
	base := authenticatedHeaders()
	for _, tc := range []struct {
		name   string
		change func([]Header) []Header
		valid  bool
	}{
		{"pass", func(h []Header) []Header { return h }, true},
		{"documented internal LMTP hop", func(h []Header) []Header {
			h[0].Value = "from compute4.internal (compute4.nyi.internal [10.202.2.44]) by sloti2d1t19 (Cyrus) with LMTPA; date"
			return h
		}, true},
		{"internal SMTP HELO is a boundary", func(h []Header) []Header {
			return append([]Header{{Name: "Received", Value: "from remote.internal by mx1.messagingengine.com with ESMTP; date"}}, h[1:]...)
		}, false},
		{"provider SMTP HELO is a boundary", func(h []Header) []Header {
			return append([]Header{{Name: "Received", Value: "from mx2.messagingengine.com by mx1.messagingengine.com with ESMTP; date"}}, h[1:]...)
		}, false},
		{"no prefix signature id", func(h []Header) []Header {
			h[2].Value = "mx1.messagingengine.com; dkim=pass header.d=sender.example"
			return h
		}, true},
		{"missing trace", func(h []Header) []Header { return h[1:3] }, false},
		{"foreign provider", func(h []Header) []Header {
			h[2].Value = "attacker.example; dkim=pass header.d=sender.example"
			return h
		}, false},
		{"lookalike provider", func(h []Header) []Header {
			h[2].Value = "mx1.messagingengine.com.attacker.example; dkim=pass header.d=sender.example"
			return h
		}, false},
		{"failed", func(h []Header) []Header {
			h[2].Value = "mx1.messagingengine.com; dkim=fail header.d=sender.example"
			return h
		}, false},
		{"unknown", func(h []Header) []Header { h[2].Value = "mx1.messagingengine.com; dkim=none"; return h }, false},
		{"different domain", func(h []Header) []Header {
			h[2].Value = "mx1.messagingengine.com; dkim=pass header.d=sub.sender.example"
			return h
		}, false},
		{"comment injection", func(h []Header) []Header {
			h[2].Value = "mx1.messagingengine.com; dkim=fail (dkim=pass header.d=sender.example)"
			return h
		}, false},
		{"malformed comment", func(h []Header) []Header { h[2].Value += " ("; return h }, false},
		{"quoted injection", func(h []Header) []Header {
			h[2].Value = `mx1.messagingengine.com; reason=";dkim=pass header.d=sender.example"`
			return h
		}, false},
		{"sender supplied result", func(h []Header) []Header { h[2], h[4] = h[4], h[2]; return h }, false},
		{"uncovered", func(h []Header) []Header {
			h[4].Value = "d=sender.example; b=abcdef; h=from:list-unsubscribe"
			return h
		}, false},
		{"destination uncovered", func(h []Header) []Header {
			h[4].Value = "d=sender.example; b=abcdef; h=from:list-unsubscribe-post"
			return h
		}, false},
		{"wrong signature", func(h []Header) []Header {
			h[4].Value = "d=sender.example; b=xyz; h=from:list-unsubscribe:list-unsubscribe-post"
			return h
		}, false},
		{"duplicate signature tag", func(h []Header) []Header { h[4].Value += "; d=sender.example"; return h }, false},
		{"ambiguous signatures", func(h []Header) []Header { h = append(h, h[4]); return h }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			headers := tc.change(append([]Header(nil), base...))
			err := Authenticated(headers, "sender.example")
			if (err == nil) != tc.valid {
				t.Fatalf("Authenticated() = %v; valid=%v", err, tc.valid)
			}
		})
	}
}
