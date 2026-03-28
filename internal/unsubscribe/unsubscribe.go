// Package unsubscribe parses List-Unsubscribe and List-Unsubscribe-Post
// email headers (RFC 2369) into structured data.
package unsubscribe

import (
	"mime"
	"net/url"
	"strings"
)

// Mechanism constants describe the type of unsubscribe method available.
const (
	MechanismMailto = "mailto"
	MechanismURL    = "url"
	MechanismBoth   = "both"
	MechanismNone   = "none"
)

// MailtoParams holds parsed components of a mailto: unsubscribe URI.
type MailtoParams struct {
	Address string
	Subject string
	Body    string
}

// ParsedHeader holds the result of parsing List-Unsubscribe headers.
type ParsedHeader struct {
	Mechanism string
	Mailto    *MailtoParams
	URL       string
	OneClick  bool
	Raw       string
}

// Parse decodes and parses List-Unsubscribe and List-Unsubscribe-Post header
// values. The listUnsubscribe value may contain MIME encoded-words (RFC 2047)
// and angle-bracket-delimited URIs separated by commas (RFC 2369).
func Parse(listUnsubscribe, listUnsubscribePost string) ParsedHeader {
	if strings.TrimSpace(listUnsubscribe) == "" {
		return ParsedHeader{Mechanism: MechanismNone}
	}

	decoded := decodeHeader(listUnsubscribe)
	uris := extractBracketedURIs(decoded)

	var mailto *MailtoParams
	var httpURL string

	for _, uri := range uris {
		u, err := url.Parse(uri)
		if err != nil {
			continue
		}

		switch strings.ToLower(u.Scheme) {
		case "mailto":
			if mailto == nil {
				mailto = parseMailtoURI(u)
			}
		case "http", "https":
			if httpURL == "" {
				httpURL = uri
			}
		}
	}

	result := ParsedHeader{
		Mailto:   mailto,
		URL:      httpURL,
		OneClick: isOneClick(listUnsubscribePost),
		Raw:      decoded,
	}

	switch {
	case mailto != nil && httpURL != "":
		result.Mechanism = MechanismBoth
	case mailto != nil:
		result.Mechanism = MechanismMailto
	case httpURL != "":
		result.Mechanism = MechanismURL
	default:
		result.Mechanism = MechanismNone
	}

	return result
}

// decodeHeader attempts MIME encoded-word decoding (RFC 2047). If decoding
// fails, the original value is returned unchanged.
func decodeHeader(raw string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(raw)
	if err != nil {
		return raw
	}
	return decoded
}

// extractBracketedURIs extracts all URIs enclosed in angle brackets from s.
// Per RFC 2369, List-Unsubscribe URIs must be wrapped in < >.
func extractBracketedURIs(s string) []string {
	var uris []string
	remaining := s
	for {
		start := strings.IndexByte(remaining, '<')
		if start == -1 {
			break
		}
		end := strings.IndexByte(remaining[start+1:], '>')
		if end == -1 {
			break
		}
		uri := strings.TrimSpace(remaining[start+1 : start+1+end])
		if uri != "" {
			uris = append(uris, uri)
		}
		remaining = remaining[start+1+end+1:]
	}
	return uris
}

// parseMailtoURI extracts the address, subject, and body from a parsed mailto URI.
func parseMailtoURI(u *url.URL) *MailtoParams {
	// For mailto: URIs, the address is in the Opaque field.
	address := u.Opaque
	if address == "" {
		return nil
	}

	// The address may contain a query string (e.g., "addr?subject=foo").
	// Split on '?' to separate the address from query parameters.
	queryStr := u.RawQuery
	if idx := strings.IndexByte(address, '?'); idx != -1 {
		queryStr = address[idx+1:]
		address = address[:idx]
	}

	// Percent-decode the address.
	if decoded, err := url.PathUnescape(address); err == nil {
		address = decoded
	}

	params := MailtoParams{Address: address}
	if queryStr != "" {
		if q, err := url.ParseQuery(queryStr); err == nil {
			params.Subject = q.Get("subject")
			params.Body = q.Get("body")
		}
	}
	return &params
}

// isOneClick checks whether the List-Unsubscribe-Post header indicates
// one-click unsubscribe support (RFC 8058).
func isOneClick(listUnsubscribePost string) bool {
	return strings.Contains(
		strings.ToLower(listUnsubscribePost),
		"list-unsubscribe=one-click",
	)
}
