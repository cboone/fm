package unsubscribe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTarget(t *testing.T) {
	for _, tc := range []struct {
		name, links, post string
		valid             bool
	}{
		{"https", "<https://list.example/opaque%2Fpath?q=secret>", "List-Unsubscribe=One-Click", true},
		{"literal comma", "<https://list.example/path,a?q=a,b>", "List-Unsubscribe=One-Click", true},
		{"encoded comma", "<https://list.example/path%2Ca?q=a%2Cb>", "List-Unsubscribe=One-Click", true},
		{"comma alternative", "<mailto:bye@example.com?subject=a,b>, <https://list.example/path,a?q=a,b>", "List-Unsubscribe=One-Click", true},
		{"trailing comma", "<https://list.example/>,", "List-Unsubscribe=One-Click", false},
		{"missing separator", "<https://list.example/><mailto:bye@example.com>", "List-Unsubscribe=One-Click", false},
		{"nested bracket", "<https://list.example/<secret>", "List-Unsubscribe=One-Click", false},
		{"alternative", "<mailto:bye@example.com>, <https://list.example/>", "List-Unsubscribe=One-Click", true},
		{"substring", "<https://list.example/>", "xList-Unsubscribe=One-Click", false},
		{"extra marker", "<https://list.example/>", "List-Unsubscribe=One-Click; x=y", false},
		{"two targets", "<https://a.example/>, <https://b.example/>", "List-Unsubscribe=One-Click", false},
		{"http", "<http://list.example/>, <https://list.example/>", "List-Unsubscribe=One-Click", false},
		{"userinfo", "<https://a:b@list.example/>", "List-Unsubscribe=One-Click", false},
		{"fragment", "<https://list.example/#token>", "List-Unsubscribe=One-Click", false},
		{"invalid", "<https:///secret>", "List-Unsubscribe=One-Click", false},
		{"unbracketed", "https://list.example/", "List-Unsubscribe=One-Click", false},
		{"junk", "junk<https://list.example/>", "List-Unsubscribe=One-Click", false},
		{"loopback", "<https://127.0.0.1/>", "List-Unsubscribe=One-Click", false},
		{"private", "<https://10.2.3.4/>", "List-Unsubscribe=One-Click", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			headers := []Header{{Name: "List-Unsubscribe", Value: tc.links}, {Name: "List-Unsubscribe-Post", Value: tc.post}}
			target, err := Target(headers)
			if (err == nil) != tc.valid {
				t.Fatalf("Target() = %v, %v; valid=%v", target, err, tc.valid)
			}
			if tc.valid {
				if !strings.Contains(tc.links, "<"+target.String()+">") {
					t.Fatalf("target changed opaque URL: %s", target)
				}
				headers = append(headers, headers[0])
				if _, err := Target(headers); err == nil {
					t.Error("accepted duplicate header")
				}
			}
		})
	}
}

func TestTargetHeaderMultiplicity(t *testing.T) {
	link := Header{Name: "List-Unsubscribe", Value: "<https://list.example/>"}
	post := Header{Name: "List-Unsubscribe-Post", Value: "List-Unsubscribe=One-Click"}
	for _, headers := range [][]Header{nil, {link}, {post}, {link, post, post}, {link, link, post}} {
		if _, err := Target(headers); err == nil {
			t.Errorf("accepted headers %v", headers)
		}
	}
}

func TestPublicDial(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "10.0.0.1", "172.16.1.1", "192.168.1.1", "169.254.169.254", "::1", "fc00::1", "fe80::1", "::ffff:127.0.0.1", "100.64.0.1", "0.0.0.0", "224.0.0.1", "2001:db8::1"} {
		t.Run(address, func(t *testing.T) {
			lookup := func(context.Context, string, string) ([]netip.Addr, error) {
				return []netip.Addr{netip.MustParseAddr("93.184.216.34"), netip.MustParseAddr(address)}, nil
			}
			dial := func(context.Context, string, string) (net.Conn, error) {
				t.Fatal("unsafe resolution reached dial")
				return nil, nil
			}
			if _, err := publicDial(lookup, dial)(context.Background(), "tcp", "list.example:443"); err == nil {
				t.Error("accepted unsafe answer")
			}
		})
	}
	lookup := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("93.184.216.34")}, nil
	}
	called := false
	dial := func(_ context.Context, network, address string) (net.Conn, error) {
		called = true
		if network != "tcp" || address != "93.184.216.34:443" {
			t.Fatalf("unvalidated dial %s %s", network, address)
		}
		return nil, errors.New("synthetic dial only")
	}
	_, _ = publicDial(lookup, dial)(context.Background(), "tcp", "list.example:443")
	if !called {
		t.Error("public destination did not reach pinned dial")
	}
}

func TestPostIsolation(t *testing.T) {
	for _, status := range []int{200, 204, 302, 400, 429, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			c := NewHTTPClient()
			if c.Timeout == 0 || c.Jar != nil {
				t.Fatal("missing timeout or unexpected cookies")
			}
			transport := c.Transport.(*http.Transport)
			if transport.Proxy != nil || !transport.DisableKeepAlives {
				t.Fatal("ambient proxy or connection reuse enabled")
			}
			c.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls > 1 {
					t.Fatal("request retried")
				}
				if r.URL.Host != "list.example" {
					t.Fatal("test refuses any other host")
				}
				if r.URL.EscapedPath() != "/opaque%2Fpath" || r.URL.RawQuery != "q=secret" {
					t.Fatal("outgoing capability URL changed")
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				if r.Method != "POST" || string(body) != "List-Unsubscribe=One-Click" || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
					t.Fatalf("bad POST: %s %q %v", r.Method, body, r.Header)
				}
				for _, h := range []string{"Cookie", "Authorization", "Referer"} {
					if r.Header.Get(h) != "" {
						t.Errorf("unexpected %s", h)
					}
				}
				if r.GetBody != nil {
					t.Error("request is replayable")
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Location": {"https://other.example/secret"}}, Body: io.NopCloser(strings.NewReader("secret")), Request: r}, nil
			})
			target, err := url.Parse("https://list.example/opaque%2Fpath?q=secret")
			if err != nil {
				t.Fatal(err)
			}
			got, err := Post(context.Background(), c, target)
			if err != nil || got != status || calls != 1 {
				t.Fatalf("Post() = %d, %v, calls=%d", got, err, calls)
			}
		})
	}
}

func TestPublicDialFallback(t *testing.T) {
	lookup := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("93.184.216.34"), netip.MustParseAddr("93.184.216.35")}, nil
	}
	left, right := net.Pipe()
	t.Cleanup(func() {
		if err := left.Close(); err != nil {
			t.Error(err)
		}
		if err := right.Close(); err != nil {
			t.Error(err)
		}
	})
	var addresses []string
	dial := func(_ context.Context, _, address string) (net.Conn, error) {
		addresses = append(addresses, address)
		if address == "93.184.216.34:443" {
			return nil, errors.New("synthetic unavailable address")
		}
		if address != "93.184.216.35:443" {
			t.Fatal("unexpected address")
		}
		return left, nil
	}
	conn, err := publicDial(lookup, dial)(context.Background(), "tcp", "list.example:443")
	if err != nil || conn != left || len(addresses) != 2 {
		t.Fatalf("fallback: conn=%v err=%v addresses=%v", conn, err, addresses)
	}
}

func TestProductionDestinationGuard(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	target, err := url.Parse(server.URL + "/fixture")
	if err != nil {
		t.Fatal(err)
	}
	c := NewHTTPClient()
	transport := c.Transport.(*http.Transport)
	transport.TLSClientConfig = server.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	c.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != target.Host || r.URL.Scheme != "https" {
			t.Fatal("test refuses unexpected host")
		}
		return transport.RoundTrip(r)
	})
	_, err = Post(context.Background(), c, target)
	if err == nil || !strings.Contains(err.Error(), "blocked; no request sent") || requests.Load() != 0 {
		t.Fatalf("production guard: err=%v requests=%d", err, requests.Load())
	}
}

func TestPostSafeErrorCategories(t *testing.T) {
	for _, failure := range []error{ErrUnsafeDestination, ErrResolution, ErrConnection} {
		c := NewHTTPClient()
		c.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Host != "list.example" {
				t.Fatal("unexpected host")
			}
			return nil, fmt.Errorf("synthetic secret: %w", failure)
		})
		target, _ := url.Parse("https://list.example/secret?token=secret")
		_, err := Post(context.Background(), c, target)
		if err == nil || !strings.Contains(err.Error(), "no request sent") || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "unknown") {
			t.Fatalf("safe category: %v", err)
		}
	}
}

func TestPostErrorRedaction(t *testing.T) {
	calls := 0
	c := NewHTTPClient()
	c.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "list.example" {
			t.Fatal("unexpected host")
		}
		return nil, errors.New("https://list.example/secret?token=secret")
	})
	target, _ := url.Parse("https://list.example/secret?token=secret")
	_, err := Post(context.Background(), c, target)
	if err == nil || strings.Contains(err.Error(), "secret") || calls != 1 {
		t.Fatalf("Post error = %v; calls=%d", err, calls)
	}
}

func TestPostCancellationAndTimeout(t *testing.T) {
	for _, cancelFirst := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		c := NewHTTPClient()
		c.Timeout = time.Millisecond
		calls := 0
		c.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if r.URL.Host != "list.example" {
				t.Fatal("unexpected host")
			}
			<-r.Context().Done()
			return nil, r.Context().Err()
		})
		if cancelFirst {
			cancel()
		}
		target, _ := url.Parse("https://list.example/secret?token=secret")
		_, err := Post(ctx, c, target)
		cancel()
		if err == nil || strings.Contains(err.Error(), "secret") || calls > 1 {
			t.Fatalf("timeout/cancellation: err=%v calls=%d", err, calls)
		}
	}
}
