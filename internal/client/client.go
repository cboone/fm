package client

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"

	"github.com/cboone/jm/internal/types"
)

const maxRetries = 3

// Client wraps the go-jmap client with convenience methods and safety guardrails.
type Client struct {
	jmap      *jmap.Client
	accountID jmap.ID
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
	return c.jmap.Do(req)
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
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable {
			return resp, nil
		}
		if attempt == maxRetries {
			resp.Body.Close()
			return nil, fmt.Errorf("max retries exceeded: received status %d", resp.StatusCode)
		}

		wait := retryDelay(resp, attempt)
		resp.Body.Close()
		time.Sleep(wait)
	}

	return resp, err
}

func retryDelay(resp *http.Response, attempt int) time.Duration {
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if seconds, err := strconv.Atoi(ra); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	// Exponential backoff: 1s, 2s, 4s.
	return time.Duration(1<<uint(attempt)) * time.Second
}
