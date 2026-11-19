package unsubscribe

import (
	"errors"
	"regexp"
	"strings"
)

var (
	fastmailHost     = regexp.MustCompile(`(?i)^(?:[a-z0-9-]+\.)*messagingengine\.com$`)
	receivedHosts    = regexp.MustCompile(`(?i)^from\s+(\S+)\s+.*?by\s+(\S+)\s`)
	internalDelivery = regexp.MustCompile(`(?i)\swith\s+LMTPA?\s*;`)
)

// Authenticated requires a provider-prefix DKIM pass and a matching signature
// covering both unsubscribe fields. It assumes headers from Fastmail delivery;
// raw message imports cannot be authenticated by their trace headers alone.
func Authenticated(headers []Header, domain string) error {
	var results []string
	var receiver string
	for _, h := range headers {
		if strings.EqualFold(h.Name, "Received") {
			value, ok := withoutComments(h.Value)
			if !ok {
				break
			}
			parts := receivedHosts.FindStringSubmatch(strings.Join(strings.Fields(value), " ") + " ")
			if len(parts) != 3 {
				break
			}
			// Local LMTP delivery precedes the provider's SMTP trace and results.
			if len(results) == 0 && strings.HasSuffix(strings.ToLower(parts[1]), ".internal") && internalDelivery.MatchString(value) {
				continue
			}
			if fastmailHost.MatchString(parts[2]) {
				receiver = parts[2]
			}
			break
		}
		if strings.EqualFold(h.Name, "Authentication-Results") {
			results = append(results, h.Value)
		}
	}
	if receiver == "" {
		return errors.New("unverified: Fastmail receiving boundary unavailable")
	}
	var passes []string
	for _, raw := range results {
		value, ok := withoutComments(raw)
		if !ok {
			return errors.New("unverified: malformed authentication result")
		}
		parts := strings.Split(value, ";")
		if !strings.EqualFold(strings.TrimSpace(parts[0]), receiver) {
			continue
		}
		for _, method := range parts[1:] {
			fields := strings.Fields(method)
			if len(fields) == 0 || !strings.EqualFold(fields[0], "dkim=pass") {
				continue
			}
			props := make(map[string]string)
			for _, field := range fields[1:] {
				key, val, ok := strings.Cut(field, "=")
				if !ok || props[strings.ToLower(key)] != "" {
					return errors.New("unverified: ambiguous DKIM result")
				}
				props[strings.ToLower(key)] = val
			}
			if strings.EqualFold(props["header.d"], domain) {
				passes = append(passes, props["header.b"])
			}
		}
	}
	if len(passes) == 0 {
		return errors.New("unverified: no Fastmail DKIM pass for sender domain")
	}
	for _, signaturePrefix := range passes {
		var matching []map[string]string
		for _, h := range headers {
			if !strings.EqualFold(h.Name, "DKIM-Signature") {
				continue
			}
			tags := signatureTags(h.Value)
			if !strings.EqualFold(tags["d"], domain) || tags["b"] == "" {
				continue
			}
			if signaturePrefix == "" || strings.HasPrefix(tags["b"], signaturePrefix) {
				matching = append(matching, tags)
			}
		}
		if len(matching) != 1 {
			continue
		}
		covered := make(map[string]bool)
		for _, name := range strings.Split(strings.ToLower(matching[0]["h"]), ":") {
			covered[strings.TrimSpace(name)] = true
		}
		if covered["list-unsubscribe"] && covered["list-unsubscribe-post"] {
			return nil
		}
	}
	return errors.New("unverified: passing DKIM signature or unsubscribe coverage unknown")
}

func signatureTags(raw string) map[string]string {
	tags := make(map[string]string)
	for _, part := range strings.Split(raw, ";") {
		if strings.TrimSpace(part) == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		if _, exists := tags[key]; !ok || exists {
			return nil
		}
		tags[key] = strings.Join(strings.Fields(value), "")
	}
	return tags
}

// withoutComments rejects quoted strings and malformed comments rather than
// accepting authentication tokens hidden inside them.
func withoutComments(raw string) (string, bool) {
	var out strings.Builder
	depth := 0
	escaped := false
	for _, r := range raw {
		if escaped {
			escaped = false
			continue
		}
		if depth > 0 && r == '\\' {
			escaped = true
			continue
		}
		switch r {
		case '(':
			depth++
			out.WriteByte(' ')
		case ')':
			depth--
			if depth < 0 {
				return "", false
			}
		case '"', '\\':
			if depth == 0 {
				return "", false
			}
		default:
			if depth == 0 {
				out.WriteRune(r)
			}
		}
	}
	return out.String(), depth == 0 && !escaped
}
