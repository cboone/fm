# jm

A safe, read-oriented CLI for JMAP email (Fastmail). `jm` provides reading, searching, archiving, spam marking, and mailbox moves -- and nothing more. It is designed for use by Claude Code on the command line, with JSON output by default and structured errors.

## Safety

- **No sending email** -- `EmailSubmission` is never called
- **No deleting email** -- `Email/set` destroy is never used, and move to Trash is refused
- **The `move` command refuses** Trash, Deleted Items, and Deleted Messages as targets

Sending and deleting are structurally disallowed, not merely hidden behind flags.

## Install

```bash
go install github.com/cboone/jm@latest
```

## Configuration

### API Token

Create a Fastmail API token at **Settings > Privacy & Security > Integrations > API tokens** with these scopes:

- `urn:ietf:params:jmap:core`
- `urn:ietf:params:jmap:mail`

No other scopes are needed (specifically not `urn:ietf:params:jmap:submission`).

### Configuration Sources

Configuration is resolved in priority order:

1. **Command-line flags** (`--token`, `--format`, etc.)
2. **Environment variables** (`JMAP_TOKEN`, `JMAP_SESSION_URL`, etc.)
3. **Config file** (`~/.config/jm/config.yaml`)

### Config File

```yaml
# ~/.config/jm/config.yaml
token: "fmu1-..."
session_url: "https://api.fastmail.com/jmap/session"
format: "json"
account_id: ""
```

All fields are optional. `session_url` defaults to Fastmail's endpoint. `format` defaults to `json`. `account_id` is auto-detected from the session if blank.

### Environment Variables

| Variable           | Description                     | Default                                 |
| ------------------ | ------------------------------- | --------------------------------------- |
| `JMAP_TOKEN`       | Bearer token for authentication | (none)                                  |
| `JMAP_SESSION_URL` | JMAP session endpoint           | `https://api.fastmail.com/jmap/session` |
| `JMAP_FORMAT`      | Output format: `json` or `text` | `json`                                  |
| `JMAP_ACCOUNT_ID`  | JMAP account ID override        | (auto-detected)                         |

## Quick Start

```bash
# Verify connectivity and authentication
jm session

# List all mailboxes
jm mailboxes

# List the 10 most recent emails in your inbox
jm list --limit 10

# List only unread emails
jm list --unread

# Read a specific email
jm read <email-id>

# Read an email with its full thread
jm read <email-id> --thread

# Search for emails
jm search "meeting agenda"
jm search --from alice@example.com --after 2026-01-01T00:00:00Z

# Archive emails
jm archive <email-id>

# Move emails to a named mailbox
jm move <email-id-1> <email-id-2> --to Receipts
```

## Commands

### Read Commands

| Command             | Description                                              |
| ------------------- | -------------------------------------------------------- |
| `jm session`        | Display JMAP session info (verify connectivity and auth) |
| `jm mailboxes`      | List all mailboxes in the account                        |
| `jm list`           | List emails in a mailbox                                 |
| `jm read <id>`      | Read the full content of an email                        |
| `jm search [query]` | Search emails by text and/or filters                     |

### Triage Commands

| Command                               | Description                          |
| ------------------------------------- | ------------------------------------ |
| `jm archive <id> [id...]`             | Move emails to the Archive mailbox   |
| `jm spam <id> [id...]`                | Move emails to the Junk/Spam mailbox |
| `jm move <id> [id...] --to <mailbox>` | Move emails to a specified mailbox   |

See [docs/CLI-REFERENCE.md](docs/CLI-REFERENCE.md) for full details on all flags, output schemas, and examples.

## Output Formats

`jm` supports two output formats, selected with `--format` or `JMAP_FORMAT`:

**JSON (default):**

```json
{
  "total": 42,
  "offset": 0,
  "emails": [
    {
      "id": "M-email-id",
      "subject": "Meeting tomorrow",
      "is_unread": true,
      "preview": "Hi, just wanted to confirm..."
    }
  ]
}
```

**Text** (`--format text`):

```text
Total: 42 (showing 25 from offset 0)

* Alice <alice@example.com>          Meeting tomorrow                                    2026-02-04 10:30
  ID: M-email-id
```

See [docs/CLI-REFERENCE.md#output-schemas](docs/CLI-REFERENCE.md#output-schemas) for complete schema documentation.

## Error Handling

Errors are written to stderr as structured JSON (or text) with exit code 1:

```json
{
  "error": "authentication_failed",
  "message": "no token configured; set JMAP_TOKEN, --token, or token in config file",
  "hint": "Check your token in JMAP_TOKEN or config file"
}
```

Error codes: `authentication_failed`, `not_found`, `forbidden_operation`, `jmap_error`, `network_error`, `general_error`, `config_error`, `partial_failure`.

See [docs/CLI-REFERENCE.md#error-reference](docs/CLI-REFERENCE.md#error-reference) for the full error reference.

## Using with Claude Code

`jm` is designed as a tool for Claude Code: JSON by default, structured errors, and safe operations only. Claude Code calls `jm` directly via shell commands with no additional configuration.

See [docs/CLAUDE-CODE-GUIDE.md](docs/CLAUDE-CODE-GUIDE.md) for setup instructions, a CLAUDE.md snippet, and example workflows.

## License

[MIT](LICENSE). Have fun.
