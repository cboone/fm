package client

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/core"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestRetryTransport_RetriesWithFreshRequestBody(t *testing.T) {
	var calls int
	var bodies []string

	rt := &retryTransport{base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		payload, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		bodies = append(bodies, string(payload))

		if calls == 1 {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header:     http.Header{"Retry-After": []string{"0"}},
				Body:       io.NopCloser(strings.NewReader("busy")),
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})}

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBufferString("payload"))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if calls != 2 {
		t.Fatalf("expected 2 transport calls, got %d", calls)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 captured bodies, got %d", len(bodies))
	}
	if bodies[0] != "payload" || bodies[1] != "payload" {
		t.Fatalf("expected both attempts to send full payload, got %q", bodies)
	}
}

func TestRetryDelay_IntegerSeconds(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"5"}}}
	d := retryDelay(resp, 0)
	if d != 5*time.Second {
		t.Fatalf("expected 5s, got %v", d)
	}
}

func TestRetryDelay_HTTPDate(t *testing.T) {
	future := time.Now().Add(10 * time.Second)
	resp := &http.Response{Header: http.Header{
		"Retry-After": []string{future.UTC().Format(http.TimeFormat)},
	}}
	d := retryDelay(resp, 0)
	if d < 9*time.Second || d > 11*time.Second {
		t.Fatalf("expected ~10s, got %v", d)
	}
}

func TestRetryDelay_HTTPDateInPast(t *testing.T) {
	past := time.Now().Add(-10 * time.Second)
	resp := &http.Response{Header: http.Header{
		"Retry-After": []string{past.UTC().Format(http.TimeFormat)},
	}}
	d := retryDelay(resp, 1)
	// Past date should fall through to exponential backoff (2s for attempt 1).
	if d != 2*time.Second {
		t.Fatalf("expected 2s exponential backoff, got %v", d)
	}
}

func TestRetryDelay_ExponentialBackoff(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	for attempt, want := range []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second} {
		d := retryDelay(resp, attempt)
		if d != want {
			t.Fatalf("attempt %d: expected %v, got %v", attempt, want, d)
		}
	}
}

type oneShotReader struct {
	data []byte
	off  int
}

func (r *oneShotReader) Read(p []byte) (int, error) {
	if r.off >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.off:])
	r.off += n
	return n, nil
}

func TestRetryTransport_FailsForNonRewindableBody(t *testing.T) {
	var calls int

	rt := &retryTransport{base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		_, _ = io.ReadAll(req.Body)
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{"Retry-After": []string{"0"}},
			Body:       io.NopCloser(strings.NewReader("retry later")),
		}, nil
	})}

	req, err := http.NewRequest(http.MethodPost, "http://example.com", &oneShotReader{data: []byte("payload")})
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	_, err = rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected retry to fail for non-rewindable body")
	}
	if !strings.Contains(err.Error(), "non-rewindable body") {
		t.Fatalf("expected non-rewindable body error, got: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 transport call, got %d", calls)
	}
}

// --- maxBatchSize tests ---

func TestMaxBatchSize_NilClient(t *testing.T) {
	var c *Client
	if got := c.maxBatchSize(); got != defaultBatchSize {
		t.Errorf("expected %d, got %d", defaultBatchSize, got)
	}
}

func TestMaxBatchSize_NilJmapClient(t *testing.T) {
	c := &Client{}
	if got := c.maxBatchSize(); got != defaultBatchSize {
		t.Errorf("expected %d, got %d", defaultBatchSize, got)
	}
}

func TestMaxBatchSize_NilSession(t *testing.T) {
	c := &Client{jmap: &jmap.Client{}}
	if got := c.maxBatchSize(); got != defaultBatchSize {
		t.Errorf("expected %d, got %d", defaultBatchSize, got)
	}
}

func TestMaxBatchSize_MissingCapability(t *testing.T) {
	c := &Client{jmap: &jmap.Client{
		Session: &jmap.Session{
			Capabilities: map[jmap.URI]jmap.Capability{},
		},
	}}
	if got := c.maxBatchSize(); got != defaultBatchSize {
		t.Errorf("expected %d, got %d", defaultBatchSize, got)
	}
}

func TestMaxBatchSize_ZeroValue(t *testing.T) {
	c := &Client{jmap: &jmap.Client{
		Session: &jmap.Session{
			Capabilities: map[jmap.URI]jmap.Capability{
				jmap.CoreURI: &core.Core{MaxObjectsInSet: 0},
			},
		},
	}}
	if got := c.maxBatchSize(); got != defaultBatchSize {
		t.Errorf("expected %d, got %d", defaultBatchSize, got)
	}
}

func TestMaxBatchSize_ValidValue(t *testing.T) {
	c := &Client{jmap: &jmap.Client{
		Session: &jmap.Session{
			Capabilities: map[jmap.URI]jmap.Capability{
				jmap.CoreURI: &core.Core{MaxObjectsInSet: 100},
			},
		},
	}}
	if got := c.maxBatchSize(); got != 100 {
		t.Errorf("expected 100, got %d", got)
	}
}

func TestDefaultBatchSizeConstant(t *testing.T) {
	if defaultBatchSize != 50 {
		t.Errorf("expected defaultBatchSize=50, got %d", defaultBatchSize)
	}
}
