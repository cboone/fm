# fm

A safe, read-oriented CLI for Fastmail email via JMAP. `fm` provides reading, searching, archiving, spam marking, and mailbox moves -- and nothing more. It is designed for use by Claude Code on the command line, with JSON output by default and structured errors.

## Safety

- **No sending email** -- `EmailSubmission` is never called
- **No deleting email** -- `Email/set` destroy is never used, and move to Trash is refused
- **The `move` command refuses** Trash, Deleted Items, and Deleted Messages as targets

Sending and deleting are structurally disallowed, not merely hidden behind flags.

## Install

### Homebrew

```bash
brew install cboone/tap/fm
```

### Go install

```bash
go install github.com/cboone/fm@latest
```

## Versioning

`fm` follows [Semantic Versioning](https://semver.org/). The `--version` flag prints the current version:

```bash
fm --version
# fm version 0.2.0
```

Development builds from source show `fm version dev`. Release binaries have the version injected at build time. Tagged releases (`v*`) on GitHub automatically publish binaries and update the [Homebrew formula](https://github.com/cboone/homebrew-tap).

## Configuration

### API Token

Create a Fastmail API token at **Settings > Privacy & Security > Integrations > API tokens** with these scopes:

- `urn:ietf:params:jmap:core`
- `urn:ietf:params:jmap:mail`

No other scopes are needed (specifically not `urn:ietf:params:jmap:submission`).

### Configuration Sources

Configuration is resolved in priority order:

1. **Command-line flags** (`--token`, `--format`, etc.)
2. **Environment variables** (`FM_TOKEN`, `FM_SESSION_URL`, etc.)
3. **Config file** (`~/.config/fm/config.yaml`)

### Config File

```yaml
# ~/.config/fm/config.yaml
token: "fmu1-..."
session_url: "https://api.fastmail.com/jmap/session"
format: "json"
account_id: ""
```

All fields are optional. `session_url` defaults to Fastmail's endpoint. `format` defaults to `json`. `account_id` is auto-detected from the session if blank.

### Environment Variables

| Variable         | Description                     | Default                                 |
| ---------------- | ------------------------------- | --------------------------------------- |
| `FM_TOKEN`       | Bearer token for authentication | (none)                                  |
| `FM_SESSION_URL` | JMAP session endpoint           | `https://api.fastmail.com/jmap/session` |
| `FM_FORMAT`      | Output format: `json` or `text` | `json`                                  |
| `FM_ACCOUNT_ID`  | JMAP account ID override        | (auto-detected)                         |

## Quick Start

```bash
# Verify connectivity and authentication
fm session

# List all mailboxes
fm mailboxes

# List the 10 most recent emails in your inbox
fm list --limit 10

# List only unread emails
fm list --unread

# Read a specific email
fm read <email-id>

# Read an email with its full thread
fm read <email-id> --thread

# Search for emails
fm search "meeting agenda"
fm search --from alice@example.com --after 2026-01-01T00:00:00Z
fm search --from alice@example.com --after 2026-01-01

# Mark emails as read
fm mark-read <email-id>

# Flag or unflag emails
fm flag <email-id>
fm unflag <email-id>

# Preview what archive would do (dry run)
fm archive --dry-run <email-id-1> <email-id-2>

# Then actually archive
fm archive <email-id-1> <email-id-2>

# Move emails to a named mailbox
fm move <email-id-1> <email-id-2> --to Receipts
```

All triage commands support `--dry-run` (`-n`) to preview affected emails without making changes.

## Commands

### Read Commands

| Command             | Description                                              |
| ------------------- | -------------------------------------------------------- |
| `fm session`        | Display JMAP session info (verify connectivity and auth) |
| `fm mailboxes`      | List all mailboxes in the account                        |
| `fm list`           | List emails in a mailbox                                 |
| `fm read <id>`      | Read the full content of an email                        |
| `fm search [query]` | Search emails by text and/or filters                     |

### Triage Commands

| Command                               | Description                          |
| ------------------------------------- | ------------------------------------ |
| `fm archive <id> [id...]`             | Move emails to the Archive mailbox   |
| `fm spam <id> [id...]`                | Move emails to the Junk/Spam mailbox |
| `fm mark-read <id> [id...]`           | Mark emails as read                  |
| `fm flag <id> [id...]`                | Flag emails (set $flagged keyword)   |
| `fm unflag <id> [id...]`              | Unflag emails (remove $flagged)      |
| `fm move <id> [id...] --to <mailbox>` | Move emails to a specified mailbox   |

See [docs/CLI-REFERENCE.md](docs/CLI-REFERENCE.md) for full details on all flags, output schemas, and examples.

## Output Formats

`fm` supports two output formats, selected with `--format` or `FM_FORMAT`:

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

**Bulk operation JSON** (`archive`, `spam`, `mark-read`, `flag`, `unflag`, `move`):

```json
{
  "matched": 3,
  "processed": 3,
  "failed": 1,
  "archived": ["M1", "M2"],
  "errors": ["M3: not found"]
}
```

**Bulk operation text** (`--format text`):

```text
Matched: 3, Processed: 3, Failed: 1
Archived: M1, M2
Errors:
  - M3: not found
```

See [docs/CLI-REFERENCE.md#output-schemas](docs/CLI-REFERENCE.md#output-schemas) for complete schema documentation.

## Error Handling

Errors are written to stderr as structured JSON (or text) with exit code 1:

```json
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
```

Error codes: `authentication_failed`, `not_found`, `forbidden_operation`, `jmap_error`, `network_error`, `general_error`, `config_error`, `partial_failure`.

See [docs/CLI-REFERENCE.md#error-reference](docs/CLI-REFERENCE.md#error-reference) for the full error reference.

## Using with Claude Code

`fm` is designed as a tool for Claude Code: JSON by default, structured errors, and safe operations only. Claude Code calls `fm` directly via shell commands with no additional configuration.

See [docs/CLAUDE-CODE-GUIDE.md](docs/CLAUDE-CODE-GUIDE.md) for setup instructions, a CLAUDE.md snippet, and example workflows.

## License

[MIT License](./LICENSE). TL;DR: Do whatever you want with this software, just keep the copyright notice included. The authors aren't liable if something goes wrong.
