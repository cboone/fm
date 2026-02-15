package client

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/core"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"

	"github.com/cboone/fm/internal/types"
)

// ErrNotFound indicates that a requested resource was not found.
var ErrNotFound = fmt.Errorf("not found")

const maxRetries = 3
const defaultBatchSize = 50

// Client wraps the go-jmap client with convenience methods and safety guardrails.
type Client struct {
	jmap         *jmap.Client
	accountID    jmap.ID
	mailboxCache []*mailbox.Mailbox
	doFunc       func(*jmap.Request) (*jmap.Response, error)
}

// New creates a Client, authenticates, and discovers the session.
func New(sessionURL, token, accountID string) (*Client, error) {
	httpClient := &http.Client{
		Transport: &retryTransport{base: http.DefaultTransport},
		Timeout:   30 * time.Second,
	}

	jc := &jmap.Client{
		SessionEndpoint: sessionURL,
		HttpClient:      httpClient,
	}
	jc.WithAccessToken(token)

	if err := jc.Authenticate(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	c := &Client{jmap: jc}

	if accountID != "" {
		c.accountID = jmap.ID(accountID)
	} else {
		primary, ok := jc.Session.PrimaryAccounts[mail.URI]
		if !ok {
			return nil, fmt.Errorf("no primary mail account found in session")
		}
		c.accountID = primary
	}

	return c, nil
}

// AccountID returns the active account ID.
func (c *Client) AccountID() jmap.ID {
	return c.accountID
}

// Session returns the underlying JMAP session.
func (c *Client) Session() *jmap.Session {
	return c.jmap.Session
}

// Do executes a JMAP request.
func (c *Client) Do(req *jmap.Request) (*jmap.Response, error) {
	if c.doFunc != nil {
		return c.doFunc(req)
	}
	return c.jmap.Do(req)
}

// maxBatchSize returns the server's MaxObjectsInSet from the JMAP session
// capabilities, falling back to defaultBatchSize when unavailable.
func (c *Client) maxBatchSize() int {
	if c == nil || c.jmap == nil || c.jmap.Session == nil {
		return defaultBatchSize
	}
	if capability, ok := c.jmap.Session.Capabilities[jmap.CoreURI]; ok {
		if coreCap, ok := capability.(*core.Core); ok && coreCap != nil && coreCap.MaxObjectsInSet > 0 {
			return int(coreCap.MaxObjectsInSet)
		}
	}
	return defaultBatchSize
}

// SessionInfo returns a simplified view of the current session.
func (c *Client) SessionInfo() types.SessionInfo {
	s := c.jmap.Session
	accounts := make(map[string]types.AccountInfo)
	for id, acct := range s.Accounts {
		accounts[string(id)] = types.AccountInfo{
			Name:       acct.Name,
			IsPersonal: acct.IsPersonal,
		}
	}

	caps := make([]string, 0, len(s.Capabilities))
	for uri := range s.Capabilities {
		caps = append(caps, string(uri))
	}
	sort.Strings(caps)

	return types.SessionInfo{
		Username:     s.Username,
		Accounts:     accounts,
		Capabilities: caps,
	}
}

// retryTransport wraps an http.RoundTripper to retry on 429 and 503.
type retryTransport struct {
	base http.RoundTripper
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		attemptReq := req
		if attempt > 0 {
			attemptReq, err = cloneRequestForRetry(req)
			if err != nil {
				return nil, err
			}
		}

		resp, err = t.base.RoundTrip(attemptReq)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable {
			return resp, nil
		}
		if attempt == maxRetries {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("max retries exceeded: received status %d", resp.StatusCode)
		}

		wait := retryDelay(resp, attempt)
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		}
	}

	return resp, err
}

func cloneRequestForRetry(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.Body == nil {
		return clone, nil
	}
	if req.GetBody == nil {
		return nil, fmt.Errorf("cannot retry request with non-rewindable body")
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, fmt.Errorf("failed to clone request body for retry: %w", err)
	}
	clone.Body = body
	return clone, nil
}

func retryDelay(resp *http.Response, attempt int) time.Duration {
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if seconds, err := strconv.Atoi(ra); err == nil {
			return time.Duration(seconds) * time.Second
		}
		if t, err := http.ParseTime(ra); err == nil {
			if d := time.Until(t); d > 0 {
				return d
			}
		}
	}
	// Exponential backoff: 1s, 2s, 4s.
	return time.Duration(1<<uint(attempt)) * time.Second
}
