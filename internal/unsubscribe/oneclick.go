package unsubscribe

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/cboone/fm/internal/types"
)

// Header preserves field order and multiplicity for execution validation.
type Header = types.Header

// Candidate is the newest message with an exact matching From address.
type Candidate struct {
	ID         string
	ReceivedAt time.Time
	From       []string
	Headers    []Header
}

// Target validates the complete unsubscribe fields without exposing URL errors.
func Target(headers []Header) (*url.URL, error) {
	var links, posts []string
	for _, h := range headers {
		switch strings.ToLower(h.Name) {
		case "list-unsubscribe":
			links = append(links, h.Value)
		case "list-unsubscribe-post":
			posts = append(posts, h.Value)
		}
	}
	if len(links) != 1 || len(posts) != 1 || strings.TrimSpace(posts[0]) != "List-Unsubscribe=One-Click" {
		return nil, errors.New("missing or ambiguous one-click headers")
	}
	var target *url.URL
	// RFC 2369 allows mailto alternatives alongside the single HTTPS URI.
	remaining := strings.TrimSpace(links[0])
	for remaining != "" {
		end := strings.IndexByte(remaining, '>')
		if !strings.HasPrefix(remaining, "<") || end < 2 {
			return nil, errors.New("malformed unsubscribe target")
		}
		part := remaining[1:end]
		remaining = strings.TrimSpace(remaining[end+1:])
		if remaining != "" {
			if !strings.HasPrefix(remaining, ",") || strings.TrimSpace(remaining[1:]) == "" {
				return nil, errors.New("malformed unsubscribe target")
			}
			remaining = strings.TrimSpace(remaining[1:])
		}
		u, err := url.Parse(part)
		if err != nil || u.User != nil || u.Fragment != "" || strings.ContainsAny(part, "<>\r\n\t ") {
			return nil, errors.New("malformed unsubscribe target")
		}
		if u.Scheme == "mailto" && u.Opaque != "" {
			continue
		}
		if u.Scheme != "https" || u.Hostname() == "" || u.Opaque != "" || target != nil {
			return nil, errors.New("expected exactly one HTTPS unsubscribe target")
		}
		if ip, err := netip.ParseAddr(u.Hostname()); err == nil && !publicAddress(ip) {
			return nil, errors.New("unsafe unsubscribe destination")
		}
		target = u
	}
	if target == nil {
		return nil, errors.New("expected exactly one HTTPS unsubscribe target")
	}
	return target, nil
}

var deniedNetworks = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
}

var globalIPv6 = netip.MustParsePrefix("2000::/3")

func publicAddress(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	for _, block := range deniedNetworks {
		if block.Contains(ip) {
			return false
		}
	}
	return ip.Is4() || globalIPv6.Contains(ip)
}

type lookupFunc func(context.Context, string, string) ([]netip.Addr, error)
type dialFunc func(context.Context, string, string) (net.Conn, error)

// ErrUnsafeDestination reports a policy refusal before any connection is made.
var ErrUnsafeDestination = errors.New("unsafe destination")

// ErrResolution reports a DNS failure before any connection is made.
var ErrResolution = errors.New("destination resolution failed")

// ErrConnection reports that all checked addresses failed before HTTP execution.
var ErrConnection = errors.New("destination connection failed")

// publicDial checks every DNS answer, then dials a checked IP without resolving again.
func publicDial(lookup lookupFunc, dial dialFunc) dialFunc {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, errors.New("invalid destination")
		}
		ips, err := lookup(ctx, "ip", host)
		if err != nil || len(ips) == 0 {
			return nil, ErrResolution
		}
		for _, ip := range ips {
			if !publicAddress(ip) {
				return nil, ErrUnsafeDestination
			}
		}
		for _, ip := range ips {
			conn, err := dial(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
		}
		return nil, ErrConnection
	}
}

// NewHTTPClient returns an isolated, single-attempt HTTPS client. It has no proxy,
// cookies or credentials, and does not follow redirects or reuse connections.
func NewHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport: &http.Transport{
			DialContext:         publicDial(net.DefaultResolver.LookupNetIP, dialer.DialContext),
			DisableKeepAlives:   true,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

// Post sends the RFC 8058 form once. Errors omit the capability URL and response
// body; a transport error may mean the request was received, so callers must not retry.
func Post(ctx context.Context, c *http.Client, target *url.URL) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), strings.NewReader("List-Unsubscribe=One-Click"))
	if err != nil {
		return 0, errors.New("cannot construct unsubscribe request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.GetBody = nil
	resp, err := c.Do(req)
	if err != nil {
		if errors.Is(err, ErrUnsafeDestination) {
			return 0, errors.New("unsubscribe destination blocked; no request sent")
		}
		if errors.Is(err, ErrResolution) {
			return 0, errors.New("unsubscribe DNS resolution failed; no request sent")
		}
		if errors.Is(err, ErrConnection) {
			return 0, errors.New("unsubscribe connection failed; no request sent")
		}
		var certificateErr *tls.CertificateVerificationError
		if errors.As(err, &certificateErr) {
			return 0, errors.New("unsubscribe TLS certificate verification failed; no request sent")
		}
		return 0, errors.New("unsubscribe request failed; outcome unknown, do not retry automatically")
	}
	if err := resp.Body.Close(); err != nil {
		return resp.StatusCode, fmt.Errorf("HTTP %d; response close failed", resp.StatusCode)
	}
	return resp.StatusCode, nil
}
