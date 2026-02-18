# fm CLI Reference

Complete reference for all `fm` commands, flags, output schemas, and error codes.

## Global Flags

| Flag            | Env Var            | Default                                 | Description                     |
| --------------- | ------------------ | --------------------------------------- | ------------------------------- |
| `--token`       | `FM_TOKEN`       | (none)                                  | Bearer token for authentication   |
| `--session-url` | `FM_SESSION_URL` | `https://api.fastmail.com/jmap/session` | Fastmail session endpoint         |
| `--format`      | `FM_FORMAT`      | `json`                                  | Output format: `json` or `text`   |
| `--account-id`  | `FM_ACCOUNT_ID`  | (auto-detected)                         | Fastmail account ID override      |
| `--config`      | --               | `~/.config/fm/config.yaml`              | Config file path                  |
| `--version`     | --               | --                                      | Print version and exit              |

Configuration sources are resolved in priority order: flags > environment variables > config file.

---

## Commands

### session

Display JMAP session info. Useful for verifying connectivity, checking capabilities, and discovering account IDs.

```bash
fm session
```

No arguments. No command-specific flags.

**JSON output:**

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

**Text output:**

```text
Username: user@fastmail.com
Capabilities: urn:ietf:params:jmap:core, urn:ietf:params:jmap:mail
Account: abc123 - user@fastmail.com (personal)
```

---

### mailboxes

List all mailboxes (folders/labels) in the account.

```bash
fm mailboxes
```

No arguments.

| Flag           | Default | Description                                                          |
| -------------- | ------- | -------------------------------------------------------------------- |
| `--roles-only` | `false` | Only show mailboxes with a defined role (inbox, archive, junk, etc.) |

**JSON output:**

```json
[
  {
    "id": "mb-inbox-id",
    "name": "Inbox",
    "role": "inbox",
    "total_emails": 1542,
    "unread_emails": 12
  },
  {
    "id": "mb-archive-id",
    "name": "Archive",
    "role": "archive",
    "total_emails": 48210,
    "unread_emails": 0
  }
]
```

**Text output:**

```text
Inbox                                    mb-inbox-id  total:1542   unread:12     [inbox]
Archive                                  mb-archive-id  total:48210  unread:0      [archive]
```

Fields with no role omit the `role` field in JSON and the `[role]` tag in text.

---

### list

List emails in a mailbox. Returns a summary of each email (not the full body).

```bash
fm list [flags]
```

No arguments.

| Flag           | Short | Default           | Description                           |
| -------------- | ----- | ----------------- | ------------------------------------- |
| `--mailbox`    | `-m`  | `inbox`           | Mailbox name or ID                    |
| `--limit`      | `-l`  | `25`              | Maximum number of results (minimum 1) |
| `--offset`     | `-o`  | `0`               | Pagination offset (non-negative)      |
| `--unread`     | `-u`  | `false`           | Only show unread messages             |
| `--flagged`    | `-f`  | `false`           | Only show flagged messages            |
| `--unflagged`  |       | `false`           | Only show unflagged messages          |
| `--sort`       | `-s`  | `receivedAt desc` | Sort order: field + direction         |

`--flagged` and `--unflagged` are mutually exclusive.

**Sort fields:** `receivedAt`, `sentAt`, `from`, `subject` (case-insensitive).
**Sort direction:** `asc` or `desc` (default: `desc`). Append after the field name, separated by a space or colon.

Examples: `"receivedAt desc"`, `"subject asc"`, `"from:asc"`.

**Filtering examples:**

```bash
fm list --flagged                # only flagged emails
fm list --unflagged              # only unflagged emails
fm list --unread --unflagged     # unread and unflagged emails
```

**JSON output:**

```json
{
  "total": 1542,
  "offset": 0,
  "emails": [
    {
      "id": "M-email-id",
      "thread_id": "T-thread-id",
      "from": [{ "name": "Alice", "email": "alice@example.com" }],
      "to": [{ "name": "Me", "email": "me@fastmail.com" }],
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

**Text output:**

```text
Total: 1542 (showing 25 from offset 0)

* Alice <alice@example.com>          Meeting tomorrow                                    2026-02-04 10:30
  ID: M-email-id
```

Unread emails are marked with `*` in text output.

---

### read

Read the full content of a specific email by ID.

```bash
fm read <email-id> [flags]
```

Exactly 1 argument required: the email ID.

| Flag            | Default | Description                                            |
| --------------- | ------- | ------------------------------------------------------ |
| `--html`        | `false` | Prefer HTML body (default: plain text)                 |
| `--raw-headers` | `false` | Include all raw email headers                          |
| `--thread`      | `false` | Show all emails in the same thread (conversation view) |

**JSON output (basic read):**

```json
{
  "id": "M-email-id",
  "thread_id": "T-thread-id",
  "from": [{ "name": "Alice", "email": "alice@example.com" }],
  "to": [{ "name": "Me", "email": "me@fastmail.com" }],
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

**JSON output (with --thread):**

```json
{
  "email": {
    "id": "M-email-id",
    "thread_id": "T-thread-id",
    "from": [{ "name": "Alice", "email": "alice@example.com" }],
    "to": [{ "name": "Me", "email": "me@fastmail.com" }],
    "cc": [],
    "subject": "Meeting tomorrow",
    "received_at": "2026-02-04T10:30:00Z",
    "is_unread": true,
    "is_flagged": false,
    "body": "Hi,\n\nJust wanted to confirm our meeting tomorrow at 3pm.\n\nBest,\nAlice",
    "attachments": []
  },
  "thread": [
    {
      "id": "M-earlier-email",
      "from": [{ "name": "Me", "email": "me@fastmail.com" }],
      "to": [{ "name": "Alice", "email": "alice@example.com" }],
      "subject": "Meeting tomorrow",
      "received_at": "2026-02-04T09:00:00Z",
      "preview": "Can we meet tomorrow at 3pm?",
      "is_unread": false
    },
    {
      "id": "M-email-id",
      "from": [{ "name": "Alice", "email": "alice@example.com" }],
      "to": [{ "name": "Me", "email": "me@fastmail.com" }],
      "subject": "Meeting tomorrow",
      "received_at": "2026-02-04T10:30:00Z",
      "preview": "Hi, just wanted to confirm our meeting...",
      "is_unread": true
    }
  ]
}
```

**Text output (basic read):**

```text
Subject: Meeting tomorrow
From: Alice <alice@example.com>
To: Me <me@fastmail.com>
Date: 2026-02-04 10:30:00 +0000
ID: M-email-id
------------------------------------------------------------------------
Hi,

Just wanted to confirm our meeting tomorrow at 3pm.

Best,
Alice
------------------------------------------------------------------------
Attachments (1):
  - agenda.pdf (application/pdf, 24000 bytes)
```

**Text output (with --thread):**

```text
Thread (2 messages):

  [1] Me <me@fastmail.com> - Meeting tomorrow (2026-02-04 09:00)
      Can we meet tomorrow at 3pm?
> [2] Alice <alice@example.com> - Meeting tomorrow (2026-02-04 10:30)

Subject: Meeting tomorrow
From: Alice <alice@example.com>
To: Me <me@fastmail.com>
Date: 2026-02-04 10:30:00 +0000
ID: M-email-id
------------------------------------------------------------------------
Hi,

Just wanted to confirm our meeting tomorrow at 3pm.

Best,
Alice
```

The `>` marker indicates the target email. Thread emails other than the target show a preview line.

**Note:** Attachments are listed as metadata only. Downloading attachment content is not supported.

---

### search

Search emails by full-text query and/or structured filters.

```bash
fm search [query] [flags]
```

0 or 1 argument. The optional `[query]` searches across subject, from, to, and body. If omitted, only the provided flags are used for filtering (filter-only search).

| Flag               | Short | Default           | Description                                 |
| ------------------ | ----- | ----------------- | ------------------------------------------- |
| `--mailbox`        | `-m`  | (all mailboxes)   | Restrict search to a specific mailbox       |
| `--limit`          | `-l`  | `25`              | Maximum results (minimum 1)                 |
| `--offset`         | `-o`  | `0`               | Pagination offset (non-negative)            |
| `--unread`         | `-u`  | `false`           | Only show unread messages                   |
| `--flagged`        | `-f`  | `false`           | Only show flagged messages                  |
| `--unflagged`      |       | `false`           | Only show unflagged messages                |
| `--sort`           | `-s`  | `receivedAt desc` | Sort order: field + direction               |
| `--from`           |       | (none)            | Filter by sender address or name            |
| `--to`             |       | (none)            | Filter by recipient address or name         |
| `--subject`        |       | (none)            | Filter by subject text                      |
| `--before`         |       | (none)            | Emails received before this date (RFC 3339 or YYYY-MM-DD) |
| `--after`          |       | (none)            | Emails received after this date (RFC 3339 or YYYY-MM-DD)  |
| `--has-attachment` |       | `false`           | Only emails with attachments                |

`--flagged` and `--unflagged` are mutually exclusive.

**Date format:** RFC 3339 (e.g. `2026-01-15T00:00:00Z`) or a bare date (e.g. `2026-01-15`). Bare dates are treated as midnight UTC.

**Sort fields:** `receivedAt`, `sentAt`, `from`, `subject` (case-insensitive).
**Sort direction:** `asc` or `desc` (default: `desc`). Append after the field name, separated by a space or colon.

Examples: `"receivedAt desc"`, `"subject asc"`, `"from:asc"`.

**Filtering examples:**

```bash
fm search --flagged                       # only flagged emails
fm search --unread --unflagged            # unread and unflagged emails
fm search "invoice" --flagged --from acme
```

All filters are combined with AND logic.

**JSON output:**

Same shape as `list` output (`EmailListResult`), with an additional `snippet` field on each email when a text query is provided:

```json
{
  "total": 3,
  "offset": 0,
  "emails": [
    {
      "id": "M-email-id",
      "thread_id": "T-thread-id",
      "from": [{ "name": "Alice", "email": "alice@example.com" }],
      "to": [{ "name": "Me", "email": "me@fastmail.com" }],
      "subject": "Meeting tomorrow",
      "received_at": "2026-02-04T10:30:00Z",
      "size": 4521,
      "is_unread": true,
      "is_flagged": false,
      "preview": "Hi, just wanted to confirm our meeting...",
      "snippet": "...confirm our <mark>meeting</mark> tomorrow at 3pm..."
    }
  ]
}
```

The `snippet` field contains HTML `<mark>` tags highlighting matched terms. It is omitted when no text query is provided.

**Text output:** Same format as `list` text output, with snippet lines shown below each email ID.

---

### stats

Aggregate emails by sender address and display per-sender counts. Queries all matching emails in the mailbox and groups them by sender, sorted by volume descending.

```bash
fm stats [flags]
```

No arguments.

| Flag          | Short | Default | Description                      |
| ------------- | ----- | ------- | -------------------------------- |
| `--mailbox`   | `-m`  | `inbox` | Mailbox name or ID               |
| `--unread`    | `-u`  | `false` | Only count unread messages       |
| `--flagged`   | `-f`  | `false` | Only count flagged messages      |
| `--unflagged` |       | `false` | Only count unflagged messages    |
| `--subjects`  |       | `false` | Include subject lines per sender |

`--flagged` and `--unflagged` are mutually exclusive.

**Usage examples:**

```bash
fm stats --mailbox Inbox --unread          # unread sender distribution
fm stats --unread --subjects               # with subject lines
fm stats --flagged                         # flagged sender distribution
```

**JSON output:**

```json
{
  "total": 142,
  "senders": [
    {
      "email": "newsletter@example.com",
      "name": "Example Newsletter",
      "count": 15,
      "subjects": ["Weekly Digest #42", "Weekly Digest #41"]
    },
    {
      "email": "alice@example.com",
      "name": "Alice Smith",
      "count": 8
    }
  ]
}
```

The `subjects` field is omitted when `--subjects` is not set.

**Text output:**

```text
Total: 142 emails from 23 senders

15  newsletter@example.com  Example Newsletter
      Weekly Digest #42
      Weekly Digest #41
 8  alice@example.com       Alice Smith
 3  bob@example.com         Bob Jones
```

Right-aligned counts, left-aligned emails, optional display name, indented subjects.

---

### archive

Move emails to the Archive mailbox. Specify emails by ID or by filter flags.

```bash
fm archive [email-id...]
fm archive --mailbox inbox --unread
fm archive --mailbox inbox --from notifications@github.com
```

Email IDs and filter flags are mutually exclusive.

| Flag               | Short | Default         | Description                                                |
| ------------------ | ----- | --------------- | ---------------------------------------------------------- |
| `--dry-run`        | `-n`  | false           | Preview affected emails without making changes             |
| `--mailbox`        | `-m`  | (all mailboxes) | Restrict to a specific mailbox                             |
| `--from`           |       | (none)          | Filter by sender address or name                           |
| `--to`             |       | (none)          | Filter by recipient address or name                        |
| `--subject`        |       | (none)          | Filter by subject text                                     |
| `--before`         |       | (none)          | Emails received before this date (RFC 3339 or YYYY-MM-DD)  |
| `--after`          |       | (none)          | Emails received after this date (RFC 3339 or YYYY-MM-DD)   |
| `--has-attachment`  |       | false           | Only emails with attachments                               |
| `--unread`         | `-u`  | false           | Only unread messages                                       |
| `--flagged`        | `-f`  | false           | Only flagged messages                                      |
| `--unflagged`      |       | false           | Only unflagged messages                                    |

`--flagged` and `--unflagged` are mutually exclusive.

**JSON output:**

```json
{
  "matched": 2,
  "processed": 2,
  "failed": 0,
  "archived": ["M-email-id-1", "M-email-id-2"],
  "destination": {
    "id": "mb-archive-id",
    "name": "Archive"
  },
  "errors": []
}
```

**Text output:**

```text
Matched: 2, Processed: 2, Failed: 0
Archived: M-email-id-1, M-email-id-2
Destination: Archive (mb-archive-id)
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### spam

Move emails to the Junk/Spam mailbox. Specify emails by ID or by filter flags.

```bash
fm spam [email-id...]
fm spam --mailbox inbox --from spammer@example.com
```

Email IDs and filter flags are mutually exclusive.

| Flag               | Short | Default         | Description                                                |
| ------------------ | ----- | --------------- | ---------------------------------------------------------- |
| `--dry-run`        | `-n`  | false           | Preview affected emails without making changes             |
| `--mailbox`        | `-m`  | (all mailboxes) | Restrict to a specific mailbox                             |
| `--from`           |       | (none)          | Filter by sender address or name                           |
| `--to`             |       | (none)          | Filter by recipient address or name                        |
| `--subject`        |       | (none)          | Filter by subject text                                     |
| `--before`         |       | (none)          | Emails received before this date (RFC 3339 or YYYY-MM-DD)  |
| `--after`          |       | (none)          | Emails received after this date (RFC 3339 or YYYY-MM-DD)   |
| `--has-attachment`  |       | false           | Only emails with attachments                               |
| `--unread`         | `-u`  | false           | Only unread messages                                       |
| `--flagged`        | `-f`  | false           | Only flagged messages                                      |
| `--unflagged`      |       | false           | Only unflagged messages                                    |

`--flagged` and `--unflagged` are mutually exclusive.

**JSON output:**

```json
{
  "matched": 1,
  "processed": 1,
  "failed": 0,
  "marked_as_spam": ["M-email-id-1"],
  "destination": {
    "id": "mb-junk-id",
    "name": "Junk Mail"
  },
  "errors": []
}
```

**Text output:**

```text
Matched: 1, Processed: 1, Failed: 0
Marked as spam: M-email-id-1
Destination: Junk Mail (mb-junk-id)
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### mark-read

Mark emails as read by setting the `$seen` keyword. Specify emails by ID or by filter flags.

```bash
fm mark-read [email-id...]
fm mark-read --mailbox inbox --unread
fm mark-read --mailbox inbox --from notifications@github.com --unread
```

Email IDs and filter flags are mutually exclusive.

| Flag               | Short | Default         | Description                                                |
| ------------------ | ----- | --------------- | ---------------------------------------------------------- |
| `--dry-run`        | `-n`  | false           | Preview affected emails without making changes             |
| `--mailbox`        | `-m`  | (all mailboxes) | Restrict to a specific mailbox                             |
| `--from`           |       | (none)          | Filter by sender address or name                           |
| `--to`             |       | (none)          | Filter by recipient address or name                        |
| `--subject`        |       | (none)          | Filter by subject text                                     |
| `--before`         |       | (none)          | Emails received before this date (RFC 3339 or YYYY-MM-DD)  |
| `--after`          |       | (none)          | Emails received after this date (RFC 3339 or YYYY-MM-DD)   |
| `--has-attachment`  |       | false           | Only emails with attachments                               |
| `--unread`         | `-u`  | false           | Only unread messages                                       |
| `--flagged`        | `-f`  | false           | Only flagged messages                                      |
| `--unflagged`      |       | false           | Only unflagged messages                                    |

`--flagged` and `--unflagged` are mutually exclusive.

**JSON output:**

```json
{
  "matched": 2,
  "processed": 2,
  "failed": 0,
  "marked_as_read": ["M-email-id-1", "M-email-id-2"],
  "errors": []
}
```

**Text output:**

```text
Matched: 2, Processed: 2, Failed: 0
Marked as read: M-email-id-1, M-email-id-2
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### flag

Flag emails by setting the `$flagged` keyword. Optionally set a flag color. Specify emails by ID or by filter flags.

```bash
fm flag [email-id...]
fm flag --color orange [email-id...]
fm flag --mailbox inbox --from boss@company.com --unread
```

Email IDs and filter flags are mutually exclusive.

| Flag               | Short | Default         | Description                                                              |
| ------------------ | ----- | --------------- | ------------------------------------------------------------------------ |
| `--color`          | `-c`  | (none)          | Flag color: `red`, `orange`, `yellow`, `green`, `blue`, `purple`, `gray` |
| `--dry-run`        | `-n`  | false           | Preview affected emails without making changes                           |
| `--mailbox`        | `-m`  | (all mailboxes) | Restrict to a specific mailbox                                           |
| `--from`           |       | (none)          | Filter by sender address or name                                         |
| `--to`             |       | (none)          | Filter by recipient address or name                                      |
| `--subject`        |       | (none)          | Filter by subject text                                                   |
| `--before`         |       | (none)          | Emails received before this date (RFC 3339 or YYYY-MM-DD)                |
| `--after`          |       | (none)          | Emails received after this date (RFC 3339 or YYYY-MM-DD)                 |
| `--has-attachment`  |       | false           | Only emails with attachments                                             |
| `--unread`         | `-u`  | false           | Only unread messages                                                     |
| `--flagged`        | `-f`  | false           | Only flagged messages                                                    |
| `--unflagged`      |       | false           | Only unflagged messages                                                  |

`--flagged` and `--unflagged` are mutually exclusive.

When `--color` is provided, the command sets both `$flagged` and the appropriate `$MailFlagBit` keywords per the [IETF MailFlagBit spec](https://www.ietf.org/archive/id/draft-eggert-mailflagcolors-00.html). These colors are displayed in Apple Mail and Fastmail. Without `--color`, only `$flagged` is set (backward compatible).

**Color examples:**

```bash
fm flag --color orange M-email-id   # flag with orange color
fm flag --color red M-email-id      # flag with red (default color, clears other color bits)
fm flag M-email-id                  # flag without color (existing behavior)
```

**JSON output:**

```json
{
  "matched": 2,
  "processed": 2,
  "failed": 0,
  "flagged": ["M-email-id-1", "M-email-id-2"],
  "errors": []
}
```

**Text output:**

```text
Matched: 2, Processed: 2, Failed: 0
Flagged: M-email-id-1, M-email-id-2
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### unflag

Unflag emails by removing the `$flagged` keyword and clearing all color bits. With `--color`, only the color bits are removed (the email stays flagged). Specify emails by ID or by filter flags.

```bash
fm unflag [email-id...]
fm unflag --color <email-id> [email-id...]
fm unflag --mailbox inbox --flagged --before 2025-01-01
```

Email IDs and filter flags are mutually exclusive.

| Flag               | Short | Default         | Description                                                |
| ------------------ | ----- | --------------- | ---------------------------------------------------------- |
| `--color`          | `-c`  | false           | Remove only the flag color (keep the email flagged)        |
| `--dry-run`        | `-n`  | false           | Preview affected emails without making changes             |
| `--mailbox`        | `-m`  | (all mailboxes) | Restrict to a specific mailbox                             |
| `--from`           |       | (none)          | Filter by sender address or name                           |
| `--to`             |       | (none)          | Filter by recipient address or name                        |
| `--subject`        |       | (none)          | Filter by subject text                                     |
| `--before`         |       | (none)          | Emails received before this date (RFC 3339 or YYYY-MM-DD)  |
| `--after`          |       | (none)          | Emails received after this date (RFC 3339 or YYYY-MM-DD)   |
| `--has-attachment`  |       | false           | Only emails with attachments                               |
| `--unread`         | `-u`  | false           | Only unread messages                                       |
| `--flagged`        | `-f`  | false           | Only flagged messages                                      |
| `--unflagged`      |       | false           | Only unflagged messages                                    |

`--flagged` and `--unflagged` are mutually exclusive.

Without `--color`, the command removes `$flagged` and clears all `$MailFlagBit` color keywords (per the [IETF MailFlagBit spec](https://www.ietf.org/archive/id/draft-eggert-mailflagcolors-00.html) recommendation). With `--color`, only the color bits are cleared, leaving the email flagged with the default red color.

**Examples:**

```bash
fm unflag M-email-id                 # fully unflag (remove flag + color)
fm unflag --color M-email-id         # remove color only, keep flagged
```

**JSON output:**

```json
{
  "matched": 2,
  "processed": 2,
  "failed": 0,
  "unflagged": ["M-email-id-1", "M-email-id-2"],
  "errors": []
}
```

**Text output:**

```text
Matched: 2, Processed: 2, Failed: 0
Unflagged: M-email-id-1, M-email-id-2
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### move

Move emails to a specified mailbox by name or ID. Specify emails by ID or by filter flags.

```bash
fm move [email-id...] --to <mailbox>
fm move --mailbox inbox --from notifications@github.com --to Archive
```

Email IDs and filter flags are mutually exclusive. The `--to` flag is always required as the destination mailbox.

| Flag               | Short | Required | Default         | Description                                                |
| ------------------ | ----- | -------- | --------------- | ---------------------------------------------------------- |
| `--to`             |       | yes      | (none)          | Target mailbox name or ID                                  |
| `--dry-run`        | `-n`  | no       | false           | Preview affected emails without making changes             |
| `--mailbox`        | `-m`  | no       | (all mailboxes) | Restrict to a specific mailbox                             |
| `--from`           |       | no       | (none)          | Filter by sender address or name                           |
| `--subject`        |       | no       | (none)          | Filter by subject text                                     |
| `--before`         |       | no       | (none)          | Emails received before this date (RFC 3339 or YYYY-MM-DD)  |
| `--after`          |       | no       | (none)          | Emails received after this date (RFC 3339 or YYYY-MM-DD)   |
| `--has-attachment`  |       | no       | false           | Only emails with attachments                               |
| `--unread`         | `-u`  | no       | false           | Only unread messages                                       |
| `--flagged`        | `-f`  | no       | false           | Only flagged messages                                      |
| `--unflagged`      |       | no       | false           | Only unflagged messages                                    |

`--flagged` and `--unflagged` are mutually exclusive.

The `--to` flag on `move` is the destination mailbox, not a recipient filter. To filter by recipient, use the `search` command first and pass the resulting IDs.

**Safety:** The `move` command refuses to target Trash, Deleted Items, or Deleted Messages (by role or name, case-insensitive). Attempting this returns a `forbidden_operation` error.

**JSON output:**

```json
{
  "matched": 1,
  "processed": 1,
  "failed": 0,
  "moved": ["M-email-id-1"],
  "destination": {
    "id": "mb-receipts-id",
    "name": "Receipts"
  },
  "errors": []
}
```

**Text output:**

```text
Matched: 1, Processed: 1, Failed: 0
Moved: M-email-id-1
Destination: Receipts (mb-receipts-id)
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

## Output Schemas

All JSON output is pretty-printed (2-space indent). These schemas are derived from the Go types in `internal/types/types.go`.

### Address

```json
{
  "name": "Alice",
  "email": "alice@example.com"
}
```

### Attachment

```json
{
  "name": "document.pdf",
  "type": "application/pdf",
  "size": 24000
}
```

### MailboxInfo

Returned by the `mailboxes` command (as an array).

| Field           | Type   | Notes            |
| --------------- | ------ | ---------------- |
| `id`            | string |                  |
| `name`          | string |                  |
| `role`          | string | Omitted if empty |
| `total_emails`  | number |                  |
| `unread_emails` | number |                  |
| `parent_id`     | string | Omitted if empty |

### EmailSummary

Returned within `EmailListResult` by the `list` and `search` commands.

| Field         | Type      | Notes                              |
| ------------- | --------- | ---------------------------------- |
| `id`          | string    |                                    |
| `thread_id`   | string    |                                    |
| `from`        | Address[] |                                    |
| `to`          | Address[] |                                    |
| `subject`     | string    |                                    |
| `received_at` | string    | RFC 3339 timestamp                 |
| `size`        | number    | Bytes                              |
| `is_unread`   | boolean   |                                    |
| `is_flagged`  | boolean   |                                    |
| `preview`     | string    | Server-generated preview           |
| `snippet`     | string    | Omitted unless text search is used |

### EmailListResult

Top-level response from `list` and `search` commands.

| Field    | Type           | Notes                     |
| -------- | -------------- | ------------------------- |
| `total`  | number         | Total matching emails     |
| `offset` | number         | Current pagination offset |
| `emails` | EmailSummary[] |                           |

### SenderStat

Aggregated count for a single sender address, returned within `StatsResult`.

| Field      | Type     | Notes                                    |
| ---------- | -------- | ---------------------------------------- |
| `email`    | string   | Lowercased sender address (grouping key) |
| `name`     | string   | Most recent display name for the address |
| `count`    | number   | Number of matching emails from sender    |
| `subjects` | string[] | Omitted unless `--subjects` is used      |

### StatsResult

Top-level response from the `stats` command.

| Field     | Type         | Notes                             |
| --------- | ------------ | --------------------------------- |
| `total`   | number       | Total matching emails             |
| `senders` | SenderStat[] | Sorted by count descending        |

### EmailDetail

Returned by the `read` command (without `--thread`).

| Field         | Type         | Notes                                      |
| ------------- | ------------ | ------------------------------------------ |
| `id`          | string       |                                            |
| `thread_id`   | string       |                                            |
| `from`        | Address[]    |                                            |
| `to`          | Address[]    |                                            |
| `cc`          | Address[]    |                                            |
| `bcc`         | Address[]    | Omitted if empty                           |
| `reply_to`    | Address[]    | Omitted if empty                           |
| `subject`     | string       |                                            |
| `sent_at`     | string       | RFC 3339 timestamp; omitted if unavailable |
| `received_at` | string       | RFC 3339 timestamp                         |
| `is_unread`   | boolean      |                                            |
| `is_flagged`  | boolean      |                                            |
| `body`        | string       | Plain text by default; HTML with `--html`  |
| `attachments` | Attachment[] |                                            |
| `headers`     | Header[]     | Omitted unless `--raw-headers` is used     |

### Header

```json
{
  "name": "X-Mailer",
  "value": "Fastmail"
}
```

### ThreadEmail

Condensed view of an email within a thread listing.

| Field         | Type      | Notes              |
| ------------- | --------- | ------------------ |
| `id`          | string    |                    |
| `from`        | Address[] |                    |
| `to`          | Address[] |                    |
| `subject`     | string    |                    |
| `received_at` | string    | RFC 3339 timestamp |
| `preview`     | string    |                    |
| `is_unread`   | boolean   |                    |

### ThreadView

Returned by `read --thread`. Wraps a full email with surrounding thread context.

| Field    | Type          | Notes                                             |
| -------- | ------------- | ------------------------------------------------- |
| `email`  | EmailDetail   | The requested email in full                       |
| `thread` | ThreadEmail[] | All emails in the thread, sorted by `received_at` |

### SessionInfo

Returned by the `session` command.

| Field          | Type                   | Notes               |
| -------------- | ---------------------- | ------------------- |
| `username`     | string                 |                     |
| `accounts`     | map[string]AccountInfo | Keyed by account ID |
| `capabilities` | string[]               |                     |

### AccountInfo

| Field         | Type    | Notes |
| ------------- | ------- | ----- |
| `name`        | string  |       |
| `is_personal` | boolean |       |

### MoveResult

Returned by `archive`, `spam`, `mark-read`, `flag`, `unflag`, and `move` commands. Only the relevant action field is populated.

| Field            | Type            | Notes                                                     |
| ---------------- | --------------- | --------------------------------------------------------- |
| `matched`        | number          | Number of input IDs                                       |
| `processed`      | number          | Number of IDs attempted (succeeded + failed)              |
| `failed`         | number          | Number of IDs that failed                                 |
| `moved`          | string[]        | Omitted unless `move` command                             |
| `archived`       | string[]        | Omitted unless `archive` command                          |
| `marked_as_spam` | string[]        | Omitted unless `spam` command                             |
| `marked_as_read` | string[]        | Omitted unless `mark-read` command                        |
| `flagged`        | string[]        | Omitted unless `flag` command                             |
| `unflagged`      | string[]        | Omitted unless `unflag` command                           |
| `destination`    | DestinationInfo | Omitted on total failure                                  |
| `errors`         | string[]        | Empty array on full success                               |

### DestinationInfo

| Field  | Type   | Notes        |
| ------ | ------ | ------------ |
| `id`   | string | Mailbox ID   |
| `name` | string | Mailbox name |

### DryRunResult

Returned by any mutating command when `--dry-run` / `-n` is passed. Previews the emails that would be affected without making changes.

| Field         | Type            | Notes                                             |
| ------------- | --------------- | ------------------------------------------------- |
| `operation`   | string          | One of: `archive`, `move`, `spam`, `mark-read`, `flag`, `unflag` |
| `count`       | number          | Number of emails that would be mutated            |
| `emails`      | EmailSummary[]  | Summaries of found emails                         |
| `not_found`   | string[]        | Omitted if empty; IDs that failed `Email/get`     |
| `destination` | DestinationInfo | Omitted for mark-read/flag/unflag                 |

**JSON example:**

```json
{
  "operation": "archive",
  "count": 2,
  "emails": [
    {
      "id": "M-email-id-1",
      "thread_id": "T-thread-id",
      "from": [{ "name": "Alice", "email": "alice@example.com" }],
      "to": [{ "name": "Me", "email": "me@fastmail.com" }],
      "subject": "Meeting tomorrow",
      "received_at": "2026-02-14T10:30:00Z",
      "size": 4521,
      "is_unread": true,
      "is_flagged": false,
      "preview": "Hi, just wanted to confirm..."
    }
  ],
  "destination": {
    "id": "mb-archive-id",
    "name": "Archive"
  }
}
```

**Text example:**

```text
Dry run: would archive 2 email(s)

  M-email-id-1  Alice <alice@example.com>  Meeting tomorrow  2026-02-14 10:30
  M-email-id-2  Bob <bob@example.com>      Invoice #1234     2026-02-13 09:15

Destination: Archive (mb-archive-id)
```

---

## Error Reference

### Error Formats

Errors are written to **stderr**. The format depends on the `--format` setting.

**JSON (default):**

```json
{
  "error": "not_found",
  "message": "email M-nonexistent not found"
}
```

The `hint` field is omitted when empty.

**Text:**

```text
Error [not_found]: email M-nonexistent not found
```

When a hint is present:

```text
Error [authentication_failed]: JMAP session request returned 401: invalid bearer token
Hint: Check your token in FM_TOKEN or config file
```

### Error Codes

| Code                    | Description                                         | Example hint                                               |
| ----------------------- | --------------------------------------------------- | ---------------------------------------------------------- |
| `authentication_failed` | Token is missing, invalid, or expired               | Check your token in FM_TOKEN or config file                |
| `not_found`             | Email ID or mailbox not found                       | (varies)                                                   |
| `forbidden_operation`   | Attempted a disallowed action (e.g., move to Trash) | Deletion is not permitted by this tool                     |
| `jmap_error`            | Server-side JMAP method error                       | (varies)                                                   |
| `network_error`         | Connection or timeout failure                       | (varies)                                                   |
| `general_error`         | Invalid flag values or other client-side errors     | (varies)                                                   |
| `config_error`          | Malformed config file                               | Fix the syntax in ~/.config/fm/config.yaml or use --config |
| `partial_failure`       | Some IDs in a batch operation failed                | (none)                                                     |

### Cobra Validation Errors

Cobra (the CLI framework) handles argument and flag validation before `fm` commands run. These errors are printed as **plain text to stderr** and do not use the structured JSON/text error format.

Examples:

```text
Error: accepts 1 arg(s), received 0
Error: required flag(s) "to" not set
Error: unknown flag: --bogus
```

### Exit Codes

| Code | Meaning                                                                          |
| ---- | -------------------------------------------------------------------------------- |
| `0`  | Success                                                                          |
| `1`  | Any error (authentication, not found, forbidden, JMAP, network, config, general) |
