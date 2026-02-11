# Email Drafting: Implementation Plan (Option C)

## Context

`jm` is a safety-first JMAP email CLI for Claude Code. It currently reads, searches, and triages email but cannot create or send. The goal is to add **draft creation** -- allowing Claude Code to compose email drafts that appear in the user's Fastmail Drafts folder for human review and sending -- while completely ruling out email sending.

**Key JMAP insight:** Draft creation (`Email/set` with `Create`) and sending (`EmailSubmission/set`) are completely separate JMAP operations. `jm` can create server-side drafts using only its existing `urn:ietf:params:jmap:mail` scope -- no new permissions, no `urn:ietf:params:jmap:submission`.

**Primary use case:** Claude Code reads an email, drafts a contextual reply, and the human reviews/edits/sends it from Fastmail's web or mobile UI.

---

## Safety Design

**Sending is structurally impossible:**
- The `emailsubmission` package is never imported
- The `urn:ietf:params:jmap:submission` scope is never requested
- No `EmailSubmission/set` call is ever constructed

**Draft creation is validated with defense-in-depth:**
- New `ValidateSetForDraft` function in `safety.go` validates the entire `Email/set` request before execution:
  - `Create` contains exactly one email
  - That email's `MailboxIDs` targets only the Drafts mailbox (verified by JMAP role, not name)
  - That email has the `$draft` keyword set
  - `Update` is nil/empty
  - `Destroy` is nil/empty

**What remains unchanged:**
- No `Destroy` (never used)
- No `Update` in draft Set calls (triage commands continue using `Update` separately)
- No trash moves (existing `ValidateTargetMailbox` unchanged)

---

## Command Design

### `jm draft` -- Create a draft email

Four composition modes via mutually exclusive flags:

```
# New composition
jm draft --to alice@example.com --subject "Meeting" --body "Let's meet Thursday."

# Reply (auto-fills to, subject, threading headers)
jm draft --reply-to <email-id> --body "Thanks, that works for me."

# Reply-all (includes all original recipients, excludes self)
jm draft --reply-all <email-id> --body "Agreed, let's proceed."

# Forward (quotes original body, requires --to)
jm draft --forward <email-id> --to bob@example.com --body "FYI see below."

# Body from stdin (works with any mode)
echo "body" | jm draft --reply-to <email-id> --body-stdin
```

### Flags

| Flag | Description |
|------|-------------|
| `--to` | Recipient address(es), comma-separated. Required for new/forward. |
| `--cc` | CC address(es), comma-separated. |
| `--bcc` | BCC address(es), comma-separated. |
| `--subject` | Subject line. Required for new composition. Auto-filled for reply/forward. |
| `--body` | Plain text body. Mutually exclusive with `--body-stdin`. |
| `--body-stdin` | Read body from stdin. Mutually exclusive with `--body`. |
| `--reply-to` | Email ID to reply to. Mutually exclusive with `--reply-all`, `--forward`. |
| `--reply-all` | Email ID to reply-all to. Mutually exclusive with `--reply-to`, `--forward`. |
| `--forward` | Email ID to forward. Mutually exclusive with `--reply-to`, `--reply-all`. |
| `--html` | Treat body as HTML instead of plain text. |

### Validation Rules

- Exactly one of: (a) `--to` + `--subject` (new), (b) `--reply-to`, (c) `--reply-all`, (d) `--forward` + `--to`
- Exactly one of `--body` or `--body-stdin` must be provided
- `--reply-to`, `--reply-all`, `--forward` are mutually exclusive
- For `--forward`, `--to` is required (must specify who to forward to)
- For `--reply-to` and `--reply-all`, `--to` is optional (auto-derived from original email)

### Output

```json
{
  "id": "M-new-draft-id",
  "mode": "reply",
  "mailbox": {"id": "mb-drafts-id", "name": "Drafts"},
  "from": [{"name": "Me", "email": "me@fastmail.com"}],
  "to": [{"name": "Alice", "email": "alice@example.com"}],
  "subject": "Re: Meeting",
  "in_reply_to": "M-original-email-id"
}
```

---

## Implementation Details

### 1. New file: `cmd/draft.go`

Pattern: Follow `cmd/move.go` structure.

```
- Parse flags and determine composition mode (new / reply / reply-all / forward)
- Read body from --body flag or stdin (--body-stdin)
- Parse address flags into []Address (comma-split, trim)
- Call newClient()
- Build DraftOptions struct
- Call c.CreateDraft(opts)
- Format result via formatter().Format()
- Return exitError on failures
```

Register command in `init()` with `rootCmd.AddCommand(draftCmd)`.

### 2. New file: `internal/client/draft.go`

**`CreateDraft(opts DraftOptions) (types.DraftResult, error)`**

Steps:
1. Resolve the Drafts mailbox via `c.GetMailboxByRole(mailbox.RoleDrafts)`
2. Determine the From address: use `c.Session().Username` as the email address
3. For reply/reply-all/forward modes:
   - Fetch the original email with properties: `id`, `messageId`, `inReplyTo`, `references`, `from`, `to`, `cc`, `replyTo`, `subject`, `textBody`, `bodyValues` (need `messageId`, `references` for threading -- these are NOT in the existing `detailProperties` so we make a dedicated `Email/get` call)
   - Fetch with `FetchTextBodyValues: true` for forward quoting
   - **Reply:** Set `To` = original's `ReplyTo` (if present) or original's `From`. Set `Subject` = "Re: " + original subject (if not already prefixed). Set `InReplyTo` = original's `MessageID`. Set `References` = original's `References` + original's `MessageID`.
   - **Reply-all:** Same as reply, but also set `CC` = original's `To` + original's `CC`, minus self (the From address). Deduplicate.
   - **Forward:** Set `Subject` = "Fwd: " + original subject. Append quoted original body below the new body text.
4. Construct the `email.Email` struct:
   ```go
   &email.Email{
       MailboxIDs: map[jmap.ID]bool{draftsMailboxID: true},
       Keywords:   map[string]bool{"$draft": true, "$seen": true},
       From:       []*mail.Address{{Email: fromAddr}},
       To:         opts.To,
       CC:         opts.CC,
       BCC:        opts.BCC,
       Subject:    subject,
       InReplyTo:  inReplyTo,   // []string
       References: references,  // []string
       TextBody:   []*email.BodyPart{{PartID: "body", Type: "text/plain"}},
       BodyValues: map[string]*email.BodyValue{"body": {Value: body}},
   }
   ```
   (Or `HTMLBody` with `Type: "text/html"` if `--html` flag is set.)
5. Construct the `email.Set` request with `Create: map[jmap.ID]*email.Email{"draft": emailObj}`
6. Call `ValidateSetForDraft(&set, draftsMailboxID)` -- defense-in-depth check before execution
7. Call `c.Do(req)` and process `SetResponse`:
   - Check `Created["draft"]` for the server-assigned ID
   - Check `NotCreated["draft"]` for errors
8. Return `types.DraftResult` with the new ID, mailbox info, recipients, subject

**`fetchOriginalForReply(emailID string) (*email.Email, error)`** (private helper)

Fetches the original email with the specific properties needed for reply threading. Makes an `Email/get` call requesting: `id`, `messageId`, `inReplyTo`, `references`, `from`, `to`, `cc`, `replyTo`, `subject`, `textBody`, `bodyValues` with `FetchTextBodyValues: true`.

### 3. Modify: `internal/types/types.go`

Add:

```go
// DraftResult reports the outcome of a draft creation.
type DraftResult struct {
    ID        string           `json:"id"`
    Mode      string           `json:"mode"`
    Mailbox   *DestinationInfo `json:"mailbox"`
    From      []Address        `json:"from"`
    To        []Address        `json:"to"`
    CC        []Address        `json:"cc,omitempty"`
    Subject   string           `json:"subject"`
    InReplyTo string           `json:"in_reply_to,omitempty"`
}
```

### 4. Modify: `internal/client/safety.go`

Add:

```go
// ValidateSetForDraft checks that an Email/set request is safe for draft creation.
func ValidateSetForDraft(set *email.Set, draftsMailboxID jmap.ID) error {
    // 1. Destroy must be empty
    // 2. Update must be empty
    // 3. Create must have exactly one entry
    // 4. That entry's MailboxIDs must contain only draftsMailboxID
    // 5. That entry's Keywords must include "$draft"
}
```

This imports `git.sr.ht/~rockorager/go-jmap/mail/email` -- the first import of this package in `safety.go`. To avoid a circular import or tight coupling, the function can accept the concrete fields rather than the `email.Set` struct. Design decision: accept `(mailboxIDs map[jmap.ID]bool, keywords map[string]bool, draftsMailboxID jmap.ID, hasUpdate bool, hasDestroy bool)` to keep `safety.go` decoupled from the email package. Or just accept `*email.Set` and add the import -- the rest of the client package already imports it.

Decision: Accept `*email.Set` directly. The `safety.go` file is in the `client` package which already imports `go-jmap/mail/email` in `email.go`. No new coupling.

### 5. Modify: `internal/output/text.go`

Add `types.DraftResult` case to the `Format` type switch:

```go
case types.DraftResult:
    return f.formatDraftResult(w, val)
```

```go
func (f *TextFormatter) formatDraftResult(w io.Writer, r types.DraftResult) error {
    fmt.Fprintf(w, "Draft created: %s\n", r.ID)
    fmt.Fprintf(w, "Mode: %s\n", r.Mode)
    fmt.Fprintf(w, "To: %s\n", formatAddrs(r.To))
    if len(r.CC) > 0 {
        fmt.Fprintf(w, "CC: %s\n", formatAddrs(r.CC))
    }
    fmt.Fprintf(w, "Subject: %s\n", r.Subject)
    if r.Mailbox != nil {
        fmt.Fprintf(w, "Mailbox: %s\n", r.Mailbox.Name)
    }
    if r.InReplyTo != "" {
        fmt.Fprintf(w, "In-Reply-To: %s\n", r.InReplyTo)
    }
    return nil
}
```

### 6. New file: `internal/client/draft_test.go`

Test cases:
- New composition: basic fields, HTML mode
- Reply: correct `To` (from `ReplyTo` or `From`), `Subject` prefixed, `InReplyTo`/`References` set
- Reply-all: correct recipient computation (original To + CC minus self), deduplication
- Forward: subject prefixed with "Fwd:", original body quoted
- Safety validation: `ValidateSetForDraft` rejects empty create, wrong mailbox, missing `$draft` keyword, non-empty destroy, non-empty update
- Error cases: drafts mailbox not found, server rejects creation

Uses the existing `doFunc` test hook on `Client` to mock JMAP responses (same pattern as `email_test.go`).

### 7. Modify: `internal/client/safety_test.go`

Add tests for `ValidateSetForDraft`:
- Valid draft Set passes
- Reject Set with Destroy populated
- Reject Set with Update populated
- Reject Set targeting wrong mailbox
- Reject Set missing `$draft` keyword
- Reject Set with multiple Create entries

---

## Files Summary

| File | Action | ~Lines |
|------|--------|--------|
| `cmd/draft.go` | Create | 120 |
| `internal/client/draft.go` | Create | 200 |
| `internal/client/draft_test.go` | Create | 200 |
| `internal/types/types.go` | Modify | +12 |
| `internal/client/safety.go` | Modify | +30 |
| `internal/client/safety_test.go` | Modify | +60 |
| `internal/output/text.go` | Modify | +20 |

---

## Open Question: Safety Philosophy

The original `PLAN.md` lists "No `Email/set` with creation of outbound messages" as a hard constraint. Adding draft creation is the first use of `Email/set Create`.

**Arguments for:** Drafts are inert (sit in Drafts folder, not sent). JMAP cleanly separates creation from submission. No new scopes needed. `ValidateSetForDraft` provides defense-in-depth.

**Arguments for caution:** Expands write surface from "modify metadata" to "create new objects." A draft with recipients is one click from sending in the Fastmail UI. The safety claim becomes more nuanced.

This question remains open for the project owner to decide.

---

## Verification

1. `make build` -- compiles
2. `make test` -- all unit tests pass (existing + new draft/safety tests)
3. `make test-cli` -- CLI integration tests pass
4. `make vet && make fmt` -- code quality
5. Manual/live test: `jm draft --to test@example.com --subject "Test" --body "Hello"` → verify draft appears in Fastmail Drafts folder
