# Email Drafting: Implementation Plan (Option C)

## Context

`jm` is a safety-first JMAP email CLI for Claude Code. It currently reads, searches, and triages email but cannot create or send. The goal is to add **draft creation** -- allowing Claude Code to compose email drafts that appear in the user's Fastmail Drafts folder for human review and sending -- while completely ruling out email sending.

**Key JMAP insight:** Draft creation (`Email/set` with `create`) and sending (`EmailSubmission/set`) are separate JMAP operations. `jm` can create server-side drafts using only the existing `urn:ietf:params:jmap:mail` scope -- no new permissions, and no `urn:ietf:params:jmap:submission`.

**Primary use case:** Claude Code reads an email, drafts a contextual reply, and the human reviews/edits/sends it from Fastmail web or mobile.

---

## Safety Design

**Sending is structurally impossible:**
- The `emailsubmission` package is never imported.
- The `urn:ietf:params:jmap:submission` scope is never requested.
- No `EmailSubmission/set` call is ever constructed.

**Draft creation is validated with defense-in-depth:**
- New `ValidateSetForDraft` function in `safety.go` validates the entire `Email/set` request before execution:
  - `Create` contains exactly one email.
  - That email's `MailboxIDs` targets only the Drafts mailbox (verified by JMAP role, not mailbox name).
  - That email has the `$draft` keyword set to `true`.
  - `Update` is nil/empty.
  - `Destroy` is nil/empty.

**What remains unchanged:**
- No `Destroy` is used.
- Existing triage commands still use `Update` only.
- No trash moves (existing `ValidateTargetMailbox` remains unchanged).

---

## Command Design

### `jm draft` -- Create a draft email

Four composition modes via mutually exclusive flags:

```
# New composition
jm draft --to alice@example.com --subject "Meeting" --body "Let's meet Thursday."

# Reply (auto-fills To, subject, threading headers)
jm draft --reply-to <email-id> --body "Thanks, that works for me."

# Reply-all (auto-computes recipients, excludes self)
jm draft --reply-all <email-id> --body "Agreed, let's proceed."

# Forward (quotes original body, requires --to)
jm draft --forward <email-id> --to bob@example.com --body "FYI see below."

# Body from stdin (works with any mode)
echo "body" | jm draft --reply-to <email-id> --body-stdin
```

### Flags

| Flag | Description |
|------|-------------|
| `--to` | Recipient address list. Required for new/forward. Optional for reply/reply-all and appended to auto-derived recipients. |
| `--cc` | CC address list. Optional. Appended to auto-derived CC for reply-all. |
| `--bcc` | BCC address list. Optional. |
| `--subject` | Subject line. Required for new composition. Auto-derived for reply/reply-all/forward unless explicitly provided. |
| `--body` | Message body. Mutually exclusive with `--body-stdin`. |
| `--body-stdin` | Read body from stdin. Mutually exclusive with `--body`. |
| `--reply-to` | Email ID to reply to. Mutually exclusive with `--reply-all`, `--forward`. |
| `--reply-all` | Email ID to reply-all to. Mutually exclusive with `--reply-to`, `--forward`. |
| `--forward` | Email ID to forward. Mutually exclusive with `--reply-to`, `--reply-all`. |
| `--html` | Treat body as HTML instead of plain text. |

Address flags (`--to`, `--cc`, `--bcc`) are parsed using RFC 5322 parsing (`net/mail` parsing), not naive comma splitting.

### Validation Rules

- Exactly one mode:
  - New: no reply/forward flag set, with `--to` and `--subject` required.
  - Reply: `--reply-to`.
  - Reply-all: `--reply-all`.
  - Forward: `--forward` with `--to` required.
- Exactly one of `--body` or `--body-stdin` must be provided.
- `--reply-to`, `--reply-all`, `--forward` are mutually exclusive.
- In reply and reply-all modes, user `--to/--cc/--bcc` are additive (appended, then deduplicated).

### Recipient Computation Rules

- Deduplication key is normalized email address (case-insensitive).
- **Reply:**
  - Base `To`: original `ReplyTo` if present, otherwise original `From`.
  - Append user `--to`, dedupe, preserve first-seen order.
  - `CC`: user-provided `--cc` only.
- **Reply-all:**
  - Base `To`: original `ReplyTo` if present, otherwise original `From`.
  - Append user `--to`, dedupe.
  - Base `CC`: original `To + CC` minus self, minus anything already in final `To`.
  - Append user `--cc`, dedupe again against final `To` and `CC`.
- **Forward/new:** recipients come from user flags only.

### Threading Header Rules

- Use RFC message IDs (`messageId` / `references`) from the original email, never JMAP email IDs.
- If original has one or more `messageId` values:
  - `InReplyTo` = original `messageId` list (as-is).
  - `References` = original `references` + original `messageId` list, order-preserving with dedupe.
- If original has no `messageId`, omit `InReplyTo`/`References` rather than inventing values.

### Output

```json
{
  "id": "M-new-draft-id",
  "mode": "reply",
  "mailbox": {"id": "mb-drafts-id", "name": "Drafts"},
  "from": [{"name": "", "email": "me@fastmail.com"}],
  "to": [{"name": "Alice", "email": "alice@example.com"}],
  "subject": "Re: Meeting",
  "in_reply_to": "<CAExample1234@example.com>"
}
```

---

## Implementation Details

### 1. New file: `cmd/draft.go`

Pattern: follow `cmd/move.go` and other mutating command patterns.

```
- Parse flags and determine composition mode (new / reply / reply-all / forward)
- Read body from --body flag or stdin (--body-stdin)
- Parse --to/--cc/--bcc with RFC-compliant parser
- Build DraftOptions struct (including additive user recipients)
- Call newClient()
- Call c.CreateDraft(opts)
- Format result via formatter().Format()
- Return exitError on failures
```

Register in `init()` with `rootCmd.AddCommand(draftCmd)`.

### 2. New file: `internal/client/draft.go`

Add types and helpers:
- `type DraftMode string`
- `type DraftOptions struct { ... }`
- `func CreateDraft(opts DraftOptions) (types.DraftResult, error)`
- private helpers for address normalization, dedupe, and subject prefixing.

**`CreateDraft(opts DraftOptions) (types.DraftResult, error)`**

Steps:
1. Resolve the Drafts mailbox via `c.GetMailboxByRole(mailbox.RoleDrafts)`.
2. Determine `From` behavior:
   - If `c.Session().Username` parses as a valid single email address, set it as `From`.
   - If it does not parse cleanly, omit explicit `From` and let server defaults apply.
3. For reply/reply-all/forward modes, fetch original email via dedicated `Email/get` with properties:
   - `id`, `messageId`, `inReplyTo`, `references`, `from`, `to`, `cc`, `replyTo`, `subject`, `textBody`, `bodyValues`
   - `FetchTextBodyValues: true` for forward quoting.
4. Build recipients + subject + threading data:
   - **Reply:** base `To` from original `ReplyTo` else `From`; subject `Re:` prefix if needed.
   - **Reply-all:** base `To` same as reply; compute `CC` from original `To+CC` excluding self and final `To`.
   - **Forward:** subject `Fwd:` prefix if needed; append quoted original plain-text body beneath user body.
5. Construct `email.Email`:
   ```go
   &email.Email{
       MailboxIDs: map[jmap.ID]bool{draftsMailboxID: true},
       Keywords:   map[string]bool{"$draft": true, "$seen": true},
       From:       fromAddrs, // optional; nil when username is not a valid address
       To:         toAddrs,
       CC:         ccAddrs,
       BCC:        bccAddrs,
       Subject:    subject,
       InReplyTo:  inReplyTo,  // []string RFC message IDs
       References: references, // []string RFC message IDs
       TextBody:   []*email.BodyPart{{PartID: "body", Type: "text/plain"}},
       BodyValues: map[string]*email.BodyValue{"body": {Value: body}},
   }
   ```
   (Or `HTMLBody` + `Type: "text/html"` when `--html` is set.)
6. Construct `email.Set` request with exactly one `Create` key (e.g., `"draft"`).
7. Call `ValidateSetForDraft(&set, draftsMailboxID)` before execution.
8. Call `c.Do(req)` and process `SetResponse`:
   - success from `Created["draft"]`
   - failure from `NotCreated["draft"]` or method error.
9. Return `types.DraftResult` with new ID, mailbox info, recipients, subject, and first `InReplyTo` value (if present).

**`fetchOriginalForReply(emailID string) (*email.Email, error)`**

Dedicated `Email/get` helper with properties and body fetch described above. Returns `ErrNotFound` when appropriate.

### 3. Modify: `internal/types/types.go`

Add:

```go
// DraftResult reports the outcome of draft creation.
type DraftResult struct {
    ID        string           `json:"id"`
    Mode      string           `json:"mode"`
    Mailbox   *DestinationInfo `json:"mailbox"`
    From      []Address        `json:"from,omitempty"`
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
    // 4. Created email MailboxIDs must contain exactly draftsMailboxID:true
    // 5. Created email Keywords must include "$draft":true
}
```

Decision: accept `*email.Set` directly.

### 5. Modify: `internal/output/text.go`

Add `types.DraftResult` handling in the formatter switch and a `formatDraftResult` method.

Expected text output:

```go
Draft created: M-new-draft-id
Mode: reply
To: Alice <alice@example.com>
Subject: Re: Meeting
Mailbox: Drafts
In-Reply-To: <CAExample1234@example.com>
```

### 6. New file: `internal/client/draft_test.go`

Use `Client.doFunc` to mock JMAP responses (same style as `internal/client/email_test.go`).

Coverage:
- **Mode validation:** new/reply/reply-all/forward exclusivity and required flags.
- **Address parsing:** RFC-compliant parsing including quoted display names with commas.
- **New draft:** basic fields and HTML mode path.
- **Reply:** recipient derivation (`ReplyTo` fallback to `From`), subject prefixing, threading headers from `messageId`/`references`.
- **Reply-all:** `To/CC` composition, self-exclusion, dedupe across `To+CC`, additive user recipients.
- **Forward:** subject prefixing and quoted original body assembly.
- **From behavior:** valid username sets `From`; invalid username omits `From`.
- **Safety guard integration:** draft set rejected on wrong mailbox, missing `$draft`, update/destroy present, wrong create cardinality.
- **Errors:** drafts mailbox missing, original email missing, server `NotCreated`, method errors.

### 7. Modify: `internal/client/safety_test.go`

Add focused tests for `ValidateSetForDraft`:
- Valid draft set passes.
- Reject non-empty `Destroy`.
- Reject non-empty `Update`.
- Reject missing/empty `Create`.
- Reject multiple `Create` entries.
- Reject wrong mailbox target.
- Reject missing `$draft` or `$draft=false`.

### 8. CLI tests and command wiring

Add/expand CLI tests (in existing CLI test location) to verify:
- Flag validation errors for invalid combinations.
- `--body-stdin` behavior.
- JSON output includes expected draft fields.
- Text output includes key lines and does not regress existing command output behavior.

### 9. Documentation updates (required)

Update docs in the same implementation PR so behavior and safety claims stay aligned:
- `README.md`: add `jm draft` examples and safety description.
- `docs/CLI-REFERENCE.md`: add full command syntax, flags, validation rules, and output examples.
- `docs/plans/done/PLAN.md`: update hard-constraint wording from broad `Email/set` create prohibition to:
  - no `EmailSubmission/set`, and
  - no `Email/set create` except validated Drafts-only `$draft` creation.
- Any command list/help snapshots used by tests.

---

## Files Summary

| File | Action | ~Lines |
|------|--------|--------|
| `cmd/draft.go` | Create | 140 |
| `internal/client/draft.go` | Create | 260 |
| `internal/client/draft_test.go` | Create | 280 |
| `internal/types/types.go` | Modify | +12 |
| `internal/client/safety.go` | Modify | +35 |
| `internal/client/safety_test.go` | Modify | +90 |
| `internal/output/text.go` | Modify | +25 |
| CLI test file(s) | Modify/Create | +80 |
| `README.md` | Modify | +25 |
| `docs/CLI-REFERENCE.md` | Modify | +60 |
| `docs/plans/done/PLAN.md` | Modify | +20 |

---

## Verification

1. `make build` compiles.
2. `make test` passes including new draft and safety unit tests.
3. `make test-cli` passes including new `jm draft` command validation and output tests.
4. `make vet && make fmt` pass cleanly.
5. Manual live test (with valid token):
   - `jm draft --to test@example.com --subject "Test" --body "Hello"`
   - verify draft appears in Fastmail Drafts.
6. Manual reply test:
   - `jm draft --reply-to <email-id> --body "Thanks"`
   - verify recipient and threading headers look correct in Fastmail draft view.
