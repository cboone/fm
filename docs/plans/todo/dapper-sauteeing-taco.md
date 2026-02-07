# Plan: Comprehensive Documentation for jm

## Context

The `jm` CLI has a minimal 48-line README with just install, basic config, and brief command examples. Detailed design information lives in `docs/plans/done/PLAN.md`, which is an implementation plan rather than user-facing documentation. There is no CLI reference and no guide for using `jm` with Claude Code (its primary consumer). This plan adds three documentation deliverables to make the project approachable and fully documented.

## Deliverables

1. **Expanded `README.md`** -- Landing page with safety, config, quick start, and links to detailed docs
2. **`docs/cli-reference.md`** -- Exhaustive command/flag/schema/error reference
3. **`docs/claude-code-guide.md`** -- Integration guide with setup, CLAUDE.md snippet, and workflows

## Implementation Order

Write `docs/cli-reference.md` first (canonical reference), then `docs/claude-code-guide.md` (links to cli-reference), then `README.md` last (links to both).

---

## 1. `docs/cli-reference.md` (~400 lines)

Source data from: `cmd/*.go` (flags), `internal/types/types.go` (schemas), `internal/output/text.go` (text format), `docs/plans/done/PLAN.md` (example outputs).

### Structure

```
# jm CLI Reference

## Global Flags
Table: Flag | Short | Env Var | Default | Description
(--token, --session-url, --format, --account-id, --config)

## Commands

### session
Synopsis, description, no args, no command-specific flags.
JSON + text output examples.

### mailboxes
Synopsis, description, no args.
Flag: --roles-only.
JSON + text output examples.

### list
Synopsis, description, no args.
Flags: --mailbox/-m, --limit/-l, --offset/-o, --unread/-u, --sort/-s.
JSON + text output examples. Sort field/direction explanation.

### read
Synopsis: jm read <email-id>
Exactly 1 arg required.
Flags: --html, --raw-headers, --thread.
JSON examples for basic read + thread view. Text example.

### search
Synopsis: jm search [query]
0 or 1 arg. Filter-only search supported.
Flags: --mailbox/-m, --limit/-l, --from, --to, --subject, --before, --after, --has-attachment.
JSON examples with snippet. Date format note (RFC 3339).

### archive
Synopsis: jm archive <email-id> [email-id...]
1+ args. No command-specific flags.
JSON + text output examples.

### spam
Synopsis: jm spam <email-id> [email-id...]
1+ args. No command-specific flags.
JSON + text output examples.

### move
Synopsis: jm move <email-id> [email-id...] --to <mailbox>
1+ args. Flag: --to (required).
JSON + text output examples. Safety note about refused targets.

## Output Schemas
Derived directly from internal/types/types.go:
- Address {name, email}
- Attachment {name, type, size}
- MailboxInfo {id, name, role?, total_emails, unread_emails, parent_id?}
- EmailSummary {id, thread_id, from, to, subject, received_at, size, is_unread, is_flagged, preview, snippet?}
- EmailListResult {total, offset, emails}
- EmailDetail {id, thread_id, from, to, cc, bcc?, reply_to?, subject, sent_at?, received_at, is_unread, is_flagged, body, attachments, headers?}
- Header {name, value}
- ThreadEmail {id, from, to, subject, received_at, preview, is_unread}
- ThreadView {email, thread}
- SessionInfo {username, accounts, capabilities}
- AccountInfo {name, is_personal}
- MoveResult {moved?, archived?, marked_as_spam?, destination?, errors}
- DestinationInfo {id, name}

## Error Reference

### Error Formats
JSON: {"error": "...", "message": "...", "hint": "..."}
Text: Error [...]: ...\nHint: ...

### Error Codes
Table: Code | Description | Example hint
- authentication_failed
- not_found
- forbidden_operation
- jmap_error
- network_error
- general_error
- config_error

### Cobra Validation Errors
Plain text (not structured JSON). Examples from tests.

### Exit Codes
0 = success, 1 = all errors.
```

---

## 2. `docs/claude-code-guide.md` (~200 lines)

### Structure

```
# Using jm with Claude Code

## Overview
jm is designed for Claude Code: JSON by default, structured errors, safe operations only.

## Setup
Install, configure JMAP_TOKEN, verify with jm session.

## Integration

### Shell Commands (Primary Pattern)
Claude Code calls jm directly via shell. No MCP config needed.

### CLAUDE.md Snippet
Copy-pasteable block to add to a project's CLAUDE.md:
- Lists all commands with brief descriptions
- Notes JSON default output
- Notes safety constraints

## Workflows

### Email Triage
Prompt: "Check my inbox for unread emails and summarize them"
Steps: jm list --unread -> jm read <id> for each

### Search and Organize
Prompt: "Find emails from alice in the last week, archive project updates"
Steps: jm search --from alice --after <date> -> jm read <id> -> jm archive <ids>

### Conversation Context
Prompt: "Read the latest email from Bob with full thread"
Steps: jm search --from bob --limit 1 -> jm read <id> --thread

### Mailbox Organization
Prompt: "Move receipt emails from inbox to Receipts"
Steps: jm search "receipt" --mailbox inbox -> jm move <ids> --to Receipts

## Tips
- Email IDs from list/search chain into read/archive/spam/move
- Batch operations accept multiple IDs
- Filter-only search (no query, just flags) is supported
- Date filters use RFC 3339 format
- Thread view shows conversation context for a single email

## Error Handling
How to interpret exit code 1 + structured JSON on stderr.
Common errors: auth, not_found, forbidden_operation.
Link to cli-reference.md#error-reference.

## Limitations
- Attachment metadata only (no download)
- No sending or deleting
- No session caching (~100-300ms per command)
```

---

## 3. `README.md` (~150 lines)

Rewrite the existing README, keeping its content but expanding and restructuring.

### Structure

```
# jm

One-paragraph description.

## Safety
Move UP from bottom. Three bullet points (no send, no delete, move refuses trash).

## Install
go install (unchanged).

## Configuration

### API Token
How to get a Fastmail token. Required scopes.

### Configuration Sources
Priority: flags > env vars > config file.

### Config File
Full ~/.config/jm/config.yaml example (all 4 fields).

### Environment Variables
Table: JMAP_TOKEN, JMAP_SESSION_URL, JMAP_FORMAT, JMAP_ACCOUNT_ID.

## Quick Start
Narrative walkthrough: session -> mailboxes -> list -> read -> search -> archive.

## Commands
Brief table of all 8 commands.
Group into Read (session, mailboxes, list, read, search) and Triage (archive, spam, move).
Link: "See docs/cli-reference.md for full details."

## Output Formats
JSON (default) vs text (--format text). Brief example of both.
Link to cli-reference.md#output-schemas.

## Error Handling
Brief overview: structured JSON/text, exit code 1, error code list.
Link to cli-reference.md#error-reference.

## Using with Claude Code
Two-sentence pitch.
Link to docs/claude-code-guide.md.

## License
MIT.
```

---

## Files Modified

| File | Action |
|------|--------|
| `README.md` | Rewrite (expand from ~48 to ~150 lines) |
| `docs/cli-reference.md` | Create new |
| `docs/claude-code-guide.md` | Create new |

## Source Files Referenced (read-only)

| File | Data extracted |
|------|---------------|
| `cmd/root.go` | Global flags, config loading, error formatting |
| `cmd/list.go` | List flags with short forms (-m, -l, -o, -u, -s) |
| `cmd/search.go` | Search flags with short forms (-m, -l) |
| `cmd/read.go` | Read flags |
| `cmd/archive.go` | Archive usage |
| `cmd/spam.go` | Spam usage |
| `cmd/move.go` | Move flags, safety check |
| `internal/types/types.go` | All output type definitions with JSON tags |
| `internal/output/text.go` | Text format patterns |
| `docs/plans/done/PLAN.md` | Example JSON outputs for all commands |

## Verification

1. Read through each new doc file to confirm accuracy against source code
2. Verify all cross-reference links point to valid sections
3. Run `scrut test tests/` to ensure no existing tests break (docs-only change, should pass)
4. Spot-check JSON examples against `internal/types/types.go` field names and `omitempty` tags
