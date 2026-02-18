# Using fm with Claude Code

`fm` is a command-line tool for reading, searching, and triaging JMAP email. It is designed as a tool for Claude Code: JSON output by default, structured errors on stderr, and safety constraints that prevent sending or deleting email.

## Setup

### 1. Install

```bash
go install github.com/cboone/fm@latest
```

### 2. Configure your JMAP token

Create a Fastmail API token at **Settings > Privacy & Security > Integrations > API tokens** with these scopes:

- `urn:ietf:params:jmap:core`
- `urn:ietf:params:jmap:mail`

Set it as an environment variable:

```bash
export FM_TOKEN="fmu1-..."
```

Or add it to `~/.config/fm/config.yaml`:

```yaml
token: "fmu1-..."
```

### 3. Verify

```bash
fm session
```

This should return JSON with your username, account info, and capabilities.

## Integration

### Shell Commands (Primary Pattern)

Claude Code calls `fm` directly via shell. No MCP server configuration is needed. All commands return JSON to stdout by default, and errors return structured JSON to stderr with exit code 1.

### CLAUDE.md Snippet

Add this to your project's `CLAUDE.md` to give Claude Code context about `fm`:

```markdown
## Email (fm)

`fm` is a CLI for reading and triaging JMAP email (Fastmail). All commands output JSON by default.

### Commands

**Read commands:**

- `fm session` -- verify connectivity and auth
- `fm mailboxes` -- list all mailboxes (add `--roles-only` for just system mailboxes)
- `fm list` -- list emails in inbox (flags: `--mailbox`, `--limit`, `--offset`, `--unread`, `--sort`)
- `fm read <id>` -- read full email (flags: `--html`, `--raw-headers`, `--thread`)
- `fm search [query]` -- search by text and/or filters (flags: `--mailbox`, `--limit`, `--from`, `--to`, `--subject`, `--before`, `--after`, `--has-attachment`)
- `fm stats` -- aggregate emails by sender (flags: `--mailbox`, `--unread`, `--flagged`, `--unflagged`, `--subjects`)

**Triage commands (by ID or filter flags):**

- `fm archive [id...] [--mailbox inbox --unread]` -- move to Archive
- `fm spam [id...] [--mailbox inbox --from spammer]` -- move to Junk
- `fm mark-read [id...] [--mailbox inbox --unread]` -- mark as read
- `fm flag [id...] [--mailbox inbox --from boss]` -- flag emails
- `fm unflag [id...] [--flagged --before 2025-01-01]` -- unflag emails
- `fm move [id...] --to <mailbox> [--mailbox inbox --from sender]` -- move to a named mailbox

### Notes

- Output is JSON by default; errors are JSON on stderr with exit code 1
- Email IDs from `list` and `search` chain directly into `read`, `archive`, `spam`, `mark-read`, `flag`, `unflag`, and `move`
- Triage commands accept email IDs or filter flags (not both): `fm archive M1 M2` or `fm archive --mailbox inbox --unread`
- Sending and deleting email are structurally disallowed
- Date filters accept RFC 3339 (e.g., `2026-01-15T00:00:00Z`) or bare dates (e.g., `2026-01-15`)
- The `move` command's `--to` flag is the destination, not a recipient filter
```

## Workflows

### Email Triage

**Prompt:** "Check my inbox for unread emails and summarize them"

```bash
# List unread emails
fm list --unread

# Read each email for full content
fm read <email-id>
```

### Search and Organize

**Prompt:** "Find emails from alice in the last week, archive the project updates"

```bash
# Search with filters
fm search --from alice --after 2026-02-01T00:00:00Z

# Read specific emails to check content
fm read <email-id>

# Archive the ones that are project updates
fm archive <id-1> <id-2> <id-3>
```

### Bulk Inbox Cleanup with Filters

**Prompt:** "Archive all GitHub notifications in my inbox"

```bash
# Preview what would be affected
fm archive --mailbox inbox --from notifications@github.com --unread --dry-run

# Execute the archive
fm archive --mailbox inbox --from notifications@github.com --unread
```

**Prompt:** "Mark all unread emails in inbox as read"

```bash
fm mark-read --mailbox inbox --unread
```

**Prompt:** "Unflag old flagged emails"

```bash
fm unflag --mailbox inbox --flagged --before 2025-01-01
```

### Conversation Context

**Prompt:** "Read the latest email from Bob with full thread"

```bash
# Find Bob's latest email
fm search --from bob --limit 1

# Read it with thread context
fm read <email-id> --thread
```

### Mailbox Organization

**Prompt:** "Move receipt emails from inbox to Receipts"

```bash
# Option 1: Search then move by ID
fm search "receipt" --mailbox inbox
fm move <id-1> <id-2> --to Receipts

# Option 2: Move with filter flags (matches all emails with "receipt" in subject)
fm move --mailbox inbox --subject receipt --to Receipts
```

### Filter-Only Search

**Prompt:** "Find emails with attachments from last month"

```bash
# No text query needed -- just use flags
fm search --has-attachment --after 2026-01-01T00:00:00Z --before 2026-02-01T00:00:00Z
```

### Sender Triage

**Prompt:** "Show me who's sending the most unread email and help me triage"

```bash
# Get sender distribution for unread inbox
fm stats --mailbox Inbox --unread

# Include subject lines to see what each sender is sending
fm stats --mailbox Inbox --unread --subjects

# Search for a high-volume sender to review
fm search --from newsletter@example.com --mailbox Inbox --unread

# Archive or mark-read in bulk
fm archive <id-1> <id-2> <id-3>
```

## Tips

- **Chaining IDs:** Email IDs from `list` and `search` results chain directly into `read`, `archive`, `spam`, `mark-read`, `flag`, `unflag`, and `move`.
- **Batch operations:** Triage commands accept multiple email IDs or filter flags (e.g. `--mailbox inbox --unread`) for bulk operations.
- **Filter-only search:** Omit the query argument and use only flags to search by sender, date range, attachments, etc.
- **Dry-run preview:** All triage commands support `--dry-run` / `-n` to preview affected emails before mutating.
- **Date format:** All date flags (`--before`, `--after`) accept RFC 3339 format (e.g. `2026-01-15T00:00:00Z`) or bare dates (e.g. `2026-01-15`).
- **Thread view:** Use `fm read <id> --thread` to see the full conversation context for a single email.
- **Mailbox names:** Both `--mailbox` and `--to` accept mailbox names (e.g., "Inbox", "Receipts") or mailbox IDs.
- **Sort order:** `fm list --sort "subject asc"` sorts by subject ascending. Fields: `receivedAt`, `sentAt`, `from`, `subject`.

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
| `authentication_failed` | Token is missing or invalid | `FM_TOKEN` not set or expired           |
| `not_found`             | Email or mailbox not found  | Stale email ID or typo in mailbox name |
| `forbidden_operation`   | Safety constraint triggered | Attempted to move to Trash             |

For the full error code list and schema details, see [CLI Reference: Error Reference](CLI-REFERENCE.md#error-reference).

## Limitations

- **Attachment metadata only:** Attachment names, types, and sizes are included in email output, but downloading attachment content is not supported.
- **No sending:** There is no command for composing or sending email.
- **No deleting:** There is no command for deleting email. The `move` command refuses Trash, Deleted Items, and Deleted Messages as targets.
- **No session caching:** Each command makes a session request to the JMAP server (~100-300ms overhead). This is negligible relative to LLM API call latency.
