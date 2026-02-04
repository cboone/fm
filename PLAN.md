# simple-jmap-cli: Implementation Plan

A CLI tool written in Go (Cobra + Viper) that provides a limited, safe interface to
a JMAP mail server (Fastmail). Designed to be used by Claude Code on the command line
to read, search, and triage email.

## Design Principles

1. **Safety first** -- Destructive or outbound actions are structurally prevented
2. **Machine-readable output** -- JSON by default, since the primary consumer is Claude Code
3. **Minimal surface area** -- Only the operations needed; no kitchen sink
4. **Standard Go tooling** -- Cobra for CLI, Viper for config, a proven JMAP library for protocol

---

## Disallowed Operations (Hard Constraints)

These are **never** implemented, not behind a flag, not behind confirmation:

- **Sending email** -- No `EmailSubmission` calls, no `Email/set` with creation of
  outbound messages
- **Deleting email** -- No `Email/set` with `destroy`, no moving to Trash
- **Purging / expunging** -- No mailbox deletion, no bulk destroy
- The `move` command will **refuse** to target Trash, Deleted Items, or Deleted Messages

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework (commands, flags, help generation) |
| `github.com/spf13/viper` | Configuration (file, env vars, flag binding) |
| `git.sr.ht/~rockorager/go-jmap` | JMAP client library (session, core protocol, mail types) |
| Standard library (`net/http`, `encoding/json`, `fmt`, `os`) | HTTP, JSON, I/O |

### Why `rockorager/go-jmap`?

It is the most actively maintained Go JMAP library. It implements RFC 8620 (Core) and
RFC 8621 (Mail) completely, provides typed Go structs for all JMAP mail objects
(Email, Mailbox, Thread, SearchSnippet), and handles session discovery and
request/response marshaling. Using it avoids reimplementing the JMAP protocol from
scratch (session negotiation, method invocation IDs, back-references, error handling).

---

## Authentication & Configuration

### Fastmail API Token

Authentication uses a Bearer token generated in Fastmail under
**Settings > Privacy & Security > Integrations > API tokens**.
The token requires these JMAP scopes:

- `urn:ietf:params:jmap:core`
- `urn:ietf:params:jmap:mail`

No other scopes are needed (specifically not `urn:ietf:params:jmap:submission`).

### Configuration Sources (Viper Priority Order)

1. **Command-line flags** (`--token`, `--format`, etc.)
2. **Environment variables** (`JMAP_TOKEN`, `JMAP_SESSION_URL`, `JMAP_FORMAT`)
3. **Config file** at `~/.config/simple-jmap-cli/config.yaml`

### Config File Schema

```yaml
# ~/.config/simple-jmap-cli/config.yaml
session_url: "https://api.fastmail.com/jmap/session"  # default
token: "fmu1-..."  # Fastmail API token (or use JMAP_TOKEN env var)
format: "json"     # json | text (default: json)
account_id: ""     # optional override; auto-detected from session if blank
```

### Env Var Prefix

All environment variables are prefixed with `JMAP_`:
`JMAP_TOKEN`, `JMAP_SESSION_URL`, `JMAP_FORMAT`, `JMAP_ACCOUNT_ID`.

---

## CLI Commands

### Global Flags

| Flag | Env Var | Default | Description |
|---|---|---|---|
| `--token` | `JMAP_TOKEN` | (none) | Bearer token for authentication |
| `--session-url` | `JMAP_SESSION_URL` | `https://api.fastmail.com/jmap/session` | JMAP session endpoint |
| `--format` | `JMAP_FORMAT` | `json` | Output format: `json` or `text` |
| `--account-id` | `JMAP_ACCOUNT_ID` | (auto) | JMAP account ID override |
| `--config` | -- | `~/.config/simple-jmap-cli/config.yaml` | Config file path |

---

### `jmap session`

Fetch and display the JMAP session resource. Useful for verifying connectivity,
checking capabilities, and discovering account IDs.

**JMAP calls:** `GET {session_url}`

**Output (JSON):**
```json
{
  "username": "user@fastmail.com",
  "accounts": {
    "abc123": {
      "name": "user@fastmail.com",
      "is_personal": true
    }
  },
  "capabilities": ["urn:ietf:params:jmap:core", "urn:ietf:params:jmap:mail"]
}
```

---

### `jmap mailboxes`

List all mailboxes (folders/labels) in the account.

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--roles-only` | `false` | Only show mailboxes with a defined role (inbox, archive, junk, etc.) |

**JMAP calls:** `Mailbox/get` (all mailboxes)

**Output (JSON):**
```json
[
  {
    "id": "mb-inbox-id",
    "name": "Inbox",
    "role": "inbox",
    "total_emails": 1542,
    "unread_emails": 12,
    "parent_id": null
  },
  {
    "id": "mb-archive-id",
    "name": "Archive",
    "role": "archive",
    "total_emails": 48210,
    "unread_emails": 0,
    "parent_id": null
  }
]
```

---

### `jmap list`

List emails in a mailbox. Returns a summary of each email (not full body).

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--mailbox` | `inbox` | Mailbox name or ID to list |
| `--limit` | `25` | Maximum number of results |
| `--offset` | `0` | Pagination offset |
| `--unread` | `false` | Only show unread messages |
| `--sort` | `receivedAt desc` | Sort order (`receivedAt`, `sentAt`, `from`, `subject`) |

**JMAP calls:** `Email/query` + `Email/get` (using back-reference)

**Email/query filter:** `inMailbox: {mailbox_id}`, optionally `hasKeyword: $seen` (inverted for unread)

**Email/get properties:** `id`, `threadId`, `mailboxIds`, `from`, `to`, `subject`,
`receivedAt`, `size`, `keywords`, `preview`

**Output (JSON):**
```json
{
  "total": 1542,
  "offset": 0,
  "emails": [
    {
      "id": "M-email-id",
      "thread_id": "T-thread-id",
      "from": [{"name": "Alice", "email": "alice@example.com"}],
      "to": [{"name": "Me", "email": "me@fastmail.com"}],
      "subject": "Meeting tomorrow",
      "received_at": "2026-02-04T10:30:00Z",
      "size": 4521,
      "is_unread": true,
      "is_flagged": false,
      "preview": "Hi, just wanted to confirm our meeting..."
    }
  ]
}
```

---

### `jmap read <email-id>`

Read the full content of a specific email by ID.

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--raw-headers` | `false` | Include all raw headers |
| `--html` | `false` | Prefer HTML body (default: plain text) |

**JMAP calls:** `Email/get` with `bodyProperties` and `fetchTextBodyValues` / `fetchHTMLBodyValues`

**Email/get properties:** `id`, `threadId`, `mailboxIds`, `from`, `to`, `cc`, `bcc`,
`replyTo`, `subject`, `sentAt`, `receivedAt`, `size`, `keywords`, `headers` (if --raw-headers),
`bodyValues`, `textBody`, `htmlBody`, `attachments`

**Output (JSON):**
```json
{
  "id": "M-email-id",
  "thread_id": "T-thread-id",
  "from": [{"name": "Alice", "email": "alice@example.com"}],
  "to": [{"name": "Me", "email": "me@fastmail.com"}],
  "cc": [],
  "subject": "Meeting tomorrow",
  "sent_at": "2026-02-04T10:29:00Z",
  "received_at": "2026-02-04T10:30:00Z",
  "is_unread": true,
  "is_flagged": false,
  "body": "Hi,\n\nJust wanted to confirm our meeting tomorrow at 3pm.\n\nBest,\nAlice",
  "attachments": [
    {
      "name": "agenda.pdf",
      "type": "application/pdf",
      "size": 24000
    }
  ]
}
```

**Note:** Attachments are listed by metadata only. Downloading attachment content is
out of scope for this tool.

---

### `jmap search <query>`

Full-text search across emails. Uses JMAP's `text` filter which matches against
subject, from, to, and body content.

**Flags:**

| Flag | Default | Description |
|---|---|---|
| `--mailbox` | (all) | Restrict search to a specific mailbox |
| `--limit` | `25` | Maximum results |
| `--from` | (none) | Filter by sender address/name |
| `--to` | (none) | Filter by recipient address/name |
| `--subject` | (none) | Filter by subject text |
| `--before` | (none) | Emails received before this date (RFC 3339) |
| `--after` | (none) | Emails received after this date (RFC 3339) |
| `--has-attachment` | `false` | Only emails with attachments |

**JMAP calls:** `Email/query` + `Email/get` (back-reference) + optionally `SearchSnippet/get`

**Email/query filter:** Combines `text`, `from`, `to`, `subject`, `before`, `after`,
`hasAttachment`, `inMailbox` as an AND filter.

**Output:** Same shape as `jmap list` output, with an additional `snippet` field when
text search is used (highlighted matching text from `SearchSnippet/get`).

---

### `jmap archive <email-id> [email-id...]`

Move one or more emails to the Archive mailbox.

**JMAP calls:**
1. `Mailbox/get` -- resolve the Archive mailbox ID (by `role: archive`)
2. `Email/set` -- update each email's `mailboxIds`: remove current mailbox, add Archive

**Output (JSON):**
```json
{
  "archived": ["M-email-id-1", "M-email-id-2"],
  "errors": []
}
```

---

### `jmap spam <email-id> [email-id...]`

Move one or more emails to the Junk/Spam mailbox.

**JMAP calls:**
1. `Mailbox/get` -- resolve the Junk mailbox ID (by `role: junk`)
2. `Email/set` -- update each email's `mailboxIds`: remove current mailbox, add Junk;
   also set keyword `$junk` if Fastmail expects it

**Output (JSON):**
```json
{
  "marked_as_spam": ["M-email-id-1"],
  "errors": []
}
```

---

### `jmap move <email-id> [email-id...] --to <mailbox>`

Move one or more emails to a specified mailbox (by name or ID).

**Safety check:** Before executing, the command resolves the target mailbox and
**refuses** if the target role is `trash` or if the mailbox name matches
`Trash`, `Deleted Items`, or `Deleted Messages` (case-insensitive). Returns an
error explaining that deletion is not permitted.

**Flags:**

| Flag | Required | Description |
|---|---|---|
| `--to` | yes | Target mailbox name or ID |

**JMAP calls:**
1. `Mailbox/get` -- resolve target mailbox ID and verify it's not trash
2. `Email/set` -- update `mailboxIds`

**Output (JSON):**
```json
{
  "moved": ["M-email-id-1"],
  "destination": {"id": "mb-id", "name": "Receipts"},
  "errors": []
}
```

---

## Project Structure

```
simple-jmap-cli/
├── main.go                          # Entry point: calls cmd.Execute()
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go                      # Root command, global persistent flags, Viper init
│   ├── session.go                   # jmap session
│   ├── mailboxes.go                 # jmap mailboxes
│   ├── list.go                      # jmap list
│   ├── read.go                      # jmap read
│   ├── search.go                    # jmap search
│   ├── archive.go                   # jmap archive
│   ├── spam.go                      # jmap spam
│   └── move.go                      # jmap move
├── internal/
│   ├── client/
│   │   ├── client.go                # JMAP client wrapper (session init, Do())
│   │   ├── mailbox.go               # Mailbox resolution helpers (by name, by role)
│   │   ├── email.go                 # Email query/get/set helpers
│   │   └── safety.go                # Safety guardrails (blocked operations)
│   ├── output/
│   │   ├── formatter.go             # Output interface (JSON / text formatting)
│   │   ├── json.go                  # JSON output implementation
│   │   └── text.go                  # Human-readable text output implementation
│   └── types/
│       └── types.go                 # App-level types (simplified email, mailbox structs)
├── .env.example                     # Example environment variables
├── .goreleaser.yaml                 # Optional: release automation
├── LICENSE
├── README.md
└── PLAN.md                          # This file
```

### Package Responsibilities

**`cmd/`** -- One file per command. Each file defines a `cobra.Command`, registers
flags, calls into `internal/client` for JMAP operations, and uses `internal/output`
to render results. No JMAP protocol logic lives here.

**`internal/client/`** -- Wraps `go-jmap` library. Manages session lifecycle,
provides high-level methods like `ListEmails(mailbox, limit, offset)`,
`ReadEmail(id)`, `MoveEmails(ids, targetMailbox)`. Contains all safety checks
in `safety.go`.

**`internal/output/`** -- Handles formatting. The `Formatter` interface has two
implementations: JSON (default) and text. Selected via `--format` flag.

**`internal/types/`** -- Simplified structs that map from the `go-jmap` library's
types to our output shapes. Keeps the command layer decoupled from the JMAP library's
internal representations.

---

## Safety Implementation (`internal/client/safety.go`)

Centralized safety checks, called before any `Email/set` invocation:

```go
// Blocked operations -- these functions always return an error
func ValidateNoDestroy(ids []string) error       // prevents Email/set destroy
func ValidateNoSubmission() error                // prevents EmailSubmission calls
func ValidateTargetMailbox(mailbox Mailbox) error // rejects trash-role targets
```

The `Email/set` wrapper in `client.go` will:
1. Never populate the `Destroy` field
2. Never create emails (no `Create` field)
3. Only use `Update` to change `mailboxIds` and `keywords`

These are **compile-time structural guarantees**, not just runtime checks -- the
wrapper functions simply don't accept parameters for forbidden operations.

---

## Error Handling

All errors are returned as structured JSON to stderr:

```json
{
  "error": "authentication_failed",
  "message": "JMAP session request returned 401: invalid bearer token",
  "hint": "Check your token in JMAP_TOKEN or config file"
}
```

Categories:
- **authentication_failed** -- 401 from session endpoint
- **not_found** -- email ID or mailbox not found
- **forbidden_operation** -- attempt to delete, send, or move to trash
- **jmap_error** -- JMAP method-level error from server
- **network_error** -- connection/timeout issues

Exit codes:
- `0` -- success
- `1` -- general error
- `2` -- authentication error
- `3` -- forbidden operation (safety guardrail triggered)

---

## Build & Install

```bash
go build -o jmap-cli .
# or
go install .
```

The binary name will be `jmap-cli` (derived from the module name, but overridable
via `-o` flag or goreleaser config).

---

## Testing Strategy

- **Unit tests** for `internal/client/safety.go` -- verify all guardrails
- **Unit tests** for `internal/output/` -- verify JSON and text formatting
- **Unit tests** for `internal/client/` -- mock the `go-jmap` client, verify correct
  JMAP method calls are constructed
- **Integration test** (optional, requires real token) -- tagged with `//go:build integration`,
  exercises the full flow against Fastmail's API

---

## Implementation Order

Suggested sequence for building this out:

1. **Project scaffolding** -- `go mod init`, install dependencies, create directory
   structure, `main.go`, `cmd/root.go` with Viper config loading
2. **`internal/client/client.go`** -- JMAP session init, authentication, basic `Do()` wrapper
3. **`jmap session`** -- verify connectivity end-to-end
4. **`internal/client/mailbox.go`** + **`jmap mailboxes`** -- mailbox listing and
   resolution by name/role (needed by most other commands)
5. **`internal/client/email.go`** + **`jmap list`** -- email query + get
6. **`jmap read`** -- full email content retrieval
7. **`jmap search`** -- query with filters and search snippets
8. **`internal/client/safety.go`** -- safety guardrails
9. **`jmap archive`** -- move to archive (first write operation, exercises Email/set)
10. **`jmap spam`** -- move to junk
11. **`jmap move`** -- general move with safety checks
12. **`internal/output/`** -- text formatter (JSON is handled by `encoding/json`
    throughout, text formatter added last)
13. **Tests and polish**

---

## Open Questions

- **Session caching:** Should the session response be cached to disk to avoid an
  extra HTTP round-trip on every invocation? The JMAP spec supports this via the
  session `state` field. Could store in `~/.cache/simple-jmap-cli/session.json`.
- **Thread view:** Should `jmap read` optionally show the full thread (all emails
  in the thread)? Or is a separate `jmap thread <thread-id>` command warranted?
- **Batch size limits:** Fastmail may have limits on how many IDs can be passed to
  a single `Email/set` call. Need to test and potentially chunk large batches.
- **Rate limiting:** Should the client handle HTTP 429 responses with automatic retry?
- **Binary name:** `jmap-cli`, `jmap`, or `simple-jmap-cli`? The shorter `jmap` is
  convenient but might conflict with other tools.
