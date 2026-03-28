# Plan: Add `fm unsubscribe` command (Issue #70)

## Context

During email triage, unsubscribing from promotional senders is a multi-step manual process: reading the email as JSON, decoding MIME/QP-encoded `List-Unsubscribe` headers with external Python scripts, extracting mailto parameters, and manually calling `fm draft`. This had to be repeated for 30+ senders in a single session.

The `unsubscribe` command automates this into a single step: extract, decode, parse, and optionally draft.

## Proposed behavior

```bash
fm unsubscribe <email-id>                        # show unsubscribe mechanism
fm unsubscribe <email-id> --draft                # create draft for mailto unsubscribe
fm unsubscribe --from "sender@example.com"       # use filter to find email, show mechanism
```

## Implementation

### 1. Create `internal/unsubscribe/unsubscribe.go` -- header parsing

Pure-logic package with no JMAP dependency. Uses only Go stdlib (`mime`, `net/url`, `strings`).

**Types:**

```go
type MailtoParams struct {
    Address string
    Subject string
    Body    string
}

type ParsedHeader struct {
    Mechanism string        // "mailto", "url", "both", "none"
    Mailto    *MailtoParams // nil if no mailto URI
    URL       string        // empty if no http(s) URI
    OneClick  bool          // List-Unsubscribe-Post contains "List-Unsubscribe=One-Click"
    Raw       string        // header value after MIME decoding
}
```

**Core function:**

```go
func Parse(listUnsubscribe, listUnsubscribePost string) ParsedHeader
```

Logic:
1. MIME-decode via `mime.WordDecoder{}.DecodeHeader()` (handles `=?us-ascii?Q?...?=`)
2. Extract URIs between `<` and `>` (RFC 2369 format)
3. Classify each as `mailto:` or `http(s):`
4. For mailto: parse with `net/url.Parse()`, extract Opaque as address, query params for subject/body
5. Determine mechanism: "mailto", "url", "both", or "none"
6. Check post header for "List-Unsubscribe=One-Click" (case-insensitive)

**Edge cases:** empty header, MIME-encoded wrapping, multiple URIs (take first of each type), malformed URIs (skip), bare URIs without angle brackets (skip per RFC 2369).

### 2. Create `internal/unsubscribe/unsubscribe_test.go`

Table-driven tests covering:
- Simple mailto, mailto with subject+body, simple HTTPS URL
- Both mailto and URL in same header
- MIME encoded-word wrapping (QP-encoded angle brackets, at-signs, slashes)
- One-click detection (case-insensitive)
- Empty header, no angle brackets, malformed URIs
- Multiple mailtos (first wins), whitespace variations
- Percent-encoded mailto addresses

### 3. Add `UnsubscribeResult` to `internal/types/types.go`

```go
type UnsubscribeResult struct {
    EmailID   string `json:"email_id"`
    Mechanism string `json:"mechanism"`
    Mailto    string `json:"mailto,omitempty"`
    Subject   string `json:"subject,omitempty"`
    Body      string `json:"body,omitempty"`
    URL       string `json:"url,omitempty"`
    OneClick  bool   `json:"one_click"`
    DraftID   string `json:"draft_id,omitempty"`
}
```

### 4. Add text formatter to `internal/output/text.go`

Add `types.UnsubscribeResult` case to `Format()` switch and `formatUnsubscribeResult` method.

Text output:
```
Unsubscribe: mailto
Email: M12345
Address: unsub@example.com
Subject: Unsubscribe
```

When draft created, append: `Draft created: M-draft-id`

When none: `Unsubscribe: none` / `Email: M12345`

### 5. Create `cmd/unsubscribe.go`

```go
var unsubscribeCmd = &cobra.Command{
    Use:   "unsubscribe [email-id]",
    Short: "Show or act on the List-Unsubscribe header of an email",
    Args:  cobra.MaximumNArgs(1),
    RunE:  ...,
}
```

Flags: `--draft` (bool), plus filter flags via `addFilterFlags()`.

**Command flow:**
1. `validateIDsOrFilters(cmd, args)`
2. `newClient()`
3. `resolveEmailIDs(cmd, args, c)` then take `ids[0]`
4. `c.ReadEmail(emailID, false, false)` to get `EmailDetail` with headers
5. `unsubscribe.Parse(detail.ListUnsubscribe, detail.ListUnsubscribePost)`
6. Build `types.UnsubscribeResult`
7. If `--draft` and mailto available: call `c.CreateDraft(DraftOptions{Mode: DraftModeNew, To: [...], Subject: ..., Body: ...})`
   - Default subject: `"Unsubscribe"` if not in header
   - Default body: empty string if not in header
8. If `--draft` and no mailto: error with hint
9. `formatter().Format(os.Stdout, result)`

Reuses: `validateIDsOrFilters`, `resolveEmailIDs`, `addFilterFlags`, `ReadEmail`, `CreateDraft`, `parseAddressFlag` patterns, `exitError`.

### 6. Update scrut tests

**`tests/help.md`**: Add `unsubscribe` line between `unflag` and the blank line in root help output.

**`tests/arguments.md`**: Add test for `fm unsubscribe` with no args/filters (expects "no emails specified" error).

### 7. Update `docs/CLI-REFERENCE.md`

Add `### unsubscribe` section with usage, flags, and output examples.

## Files summary

| File | Action |
|------|--------|
| `internal/unsubscribe/unsubscribe.go` | Create |
| `internal/unsubscribe/unsubscribe_test.go` | Create |
| `internal/types/types.go` | Modify (add UnsubscribeResult) |
| `internal/output/text.go` | Modify (add formatter case) |
| `internal/output/text_test.go` | Modify (add formatter tests) |
| `cmd/unsubscribe.go` | Create |
| `tests/help.md` | Modify |
| `tests/arguments.md` | Modify |
| `docs/CLI-REFERENCE.md` | Modify |

## Deferred (out of scope)

- **`--from` on draft** (setting the draft's From to match the original email's To address): depends on issue #69 landing first.
- **URL opening**: the issue mentions "or open it, if that's in scope" for URL-only headers. Defer for now; just display the URL.

## Verification

1. `go build ./...`
2. `go test ./...` (all new + existing)
3. `go vet ./...`
4. `make lint` / `make test-ci`
5. Manual: `fm unsubscribe <id>` on a real newsletter email
6. Manual: `fm unsubscribe <id> --draft` to verify draft creation
7. Manual: `fm unsubscribe --from "newsletter@example.com" --mailbox inbox`
