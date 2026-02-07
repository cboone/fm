# Using jm with Claude Code

`jm` is a command-line tool for reading, searching, and triaging JMAP email. It is designed as a tool for Claude Code: JSON output by default, structured errors on stderr, and safety constraints that prevent sending or deleting email.

## Setup

### 1. Install

```bash
go install github.com/cboone/jm@latest
```

### 2. Configure your JMAP token

Create a Fastmail API token at **Settings > Privacy & Security > Integrations > API tokens** with these scopes:

- `urn:ietf:params:jmap:core`
- `urn:ietf:params:jmap:mail`

Set it as an environment variable:

```bash
export JMAP_TOKEN="fmu1-..."
```

Or add it to `~/.config/jm/config.yaml`:

```yaml
token: "fmu1-..."
```

### 3. Verify

```bash
jm session
```

This should return JSON with your username, account info, and capabilities.

## Integration

### Shell Commands (Primary Pattern)

Claude Code calls `jm` directly via shell. No MCP server configuration is needed. All commands return JSON to stdout by default, and errors return structured JSON to stderr with exit code 1.

### CLAUDE.md Snippet

Add this to your project's `CLAUDE.md` to give Claude Code context about `jm`:

```markdown
## Email (jm)

`jm` is a CLI for reading and triaging JMAP email (Fastmail). All commands output JSON by default.

### Commands

**Read commands:**

- `jm session` -- verify connectivity and auth
- `jm mailboxes` -- list all mailboxes (add `--roles-only` for just system mailboxes)
- `jm list` -- list emails in inbox (flags: `--mailbox`, `--limit`, `--offset`, `--unread`, `--sort`)
- `jm read <id>` -- read full email (flags: `--html`, `--raw-headers`, `--thread`)
- `jm search [query]` -- search by text and/or filters (flags: `--mailbox`, `--limit`, `--from`, `--to`, `--subject`, `--before`, `--after`, `--has-attachment`)

**Triage commands:**

- `jm archive <id> [id...]` -- move to Archive
- `jm spam <id> [id...]` -- move to Junk
- `jm move <id> [id...] --to <mailbox>` -- move to a named mailbox

### Notes

- Output is JSON by default; errors are JSON on stderr with exit code 1
- Email IDs from `list` and `search` chain directly into `read`, `archive`, `spam`, and `move`
- Sending and deleting email are structurally disallowed
- Date filters use RFC 3339 format (e.g., `2026-01-15T00:00:00Z`)
- Batch operations (`archive`, `spam`, `move`) accept multiple IDs
```

## Workflows

### Email Triage

**Prompt:** "Check my inbox for unread emails and summarize them"

```bash
# List unread emails
jm list --unread

# Read each email for full content
jm read <email-id>
```

### Search and Organize

**Prompt:** "Find emails from alice in the last week, archive the project updates"

```bash
# Search with filters
jm search --from alice --after 2026-02-01T00:00:00Z

# Read specific emails to check content
jm read <email-id>

# Archive the ones that are project updates
jm archive <id-1> <id-2> <id-3>
```

### Conversation Context

**Prompt:** "Read the latest email from Bob with full thread"

```bash
# Find Bob's latest email
jm search --from bob --limit 1

# Read it with thread context
jm read <email-id> --thread
```

### Mailbox Organization

**Prompt:** "Move receipt emails from inbox to Receipts"

```bash
# Search for receipts in inbox
jm search "receipt" --mailbox inbox

# Move matching emails
jm move <id-1> <id-2> --to Receipts
```

### Filter-Only Search

**Prompt:** "Find emails with attachments from last month"

```bash
# No text query needed -- just use flags
jm search --has-attachment --after 2026-01-01T00:00:00Z --before 2026-02-01T00:00:00Z
```

## Tips

- **Chaining IDs:** Email IDs from `list` and `search` results chain directly into `read`, `archive`, `spam`, and `move`.
- **Batch operations:** `archive`, `spam`, and `move` accept multiple email IDs in a single call.
- **Filter-only search:** Omit the query argument and use only flags to search by sender, date range, attachments, etc.
- **Date format:** All date flags (`--before`, `--after`) use RFC 3339 format, e.g. `2026-01-15T00:00:00Z`.
- **Thread view:** Use `jm read <id> --thread` to see the full conversation context for a single email.
- **Mailbox names:** Both `--mailbox` and `--to` accept mailbox names (e.g., "Inbox", "Receipts") or mailbox IDs.
- **Sort order:** `jm list --sort "subject asc"` sorts by subject ascending. Fields: `receivedAt`, `sentAt`, `from`, `subject`.

## Error Handling

All errors produce exit code 1. Structured error details are written to **stderr** as JSON (default) or text.

**JSON error format:**

```json
{
  "error": "not_found",
  "message": "email M-nonexistent not found"
}
```

**Common errors:**

| Error code              | Meaning                     | Typical cause                          |
| ----------------------- | --------------------------- | -------------------------------------- |
| `authentication_failed` | Token is missing or invalid | `JMAP_TOKEN` not set or expired        |
| `not_found`             | Email or mailbox not found  | Stale email ID or typo in mailbox name |
| `forbidden_operation`   | Safety constraint triggered | Attempted to move to Trash             |

For the full error code list and schema details, see [CLI Reference: Error Reference](CLI-REFERENCE.md#error-reference).

## Limitations

- **Attachment metadata only:** Attachment names, types, and sizes are included in email output, but downloading attachment content is not supported.
- **No sending:** There is no command for composing or sending email.
- **No deleting:** There is no command for deleting email. The `move` command refuses Trash, Deleted Items, and Deleted Messages as targets.
- **No session caching:** Each command makes a session request to the JMAP server (~100-300ms overhead). This is negligible relative to LLM API call latency.
