# jm CLI Reference

Complete reference for all `jm` commands, flags, output schemas, and error codes.

## Global Flags

| Flag            | Env Var            | Default                                 | Description                     |
| --------------- | ------------------ | --------------------------------------- | ------------------------------- |
| `--token`       | `JMAP_TOKEN`       | (none)                                  | Bearer token for authentication |
| `--session-url` | `JMAP_SESSION_URL` | `https://api.fastmail.com/jmap/session` | JMAP session endpoint           |
| `--format`      | `JMAP_FORMAT`      | `json`                                  | Output format: `json` or `text` |
| `--account-id`  | `JMAP_ACCOUNT_ID`  | (auto-detected)                         | JMAP account ID override        |
| `--config`      | --                 | `~/.config/jm/config.yaml`              | Config file path                |

Configuration sources are resolved in priority order: flags > environment variables > config file.

---

## Commands

### session

Display JMAP session info. Useful for verifying connectivity, checking capabilities, and discovering account IDs.

```bash
jm session
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
jm mailboxes
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
jm list [flags]
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
jm list --flagged                # only flagged emails
jm list --unflagged              # only unflagged emails
jm list --unread --unflagged     # unread and unflagged emails
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
jm read <email-id> [flags]
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
jm search [query] [flags]
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
jm search --flagged                       # only flagged emails
jm search --unread --unflagged            # unread and unflagged emails
jm search "invoice" --flagged --from acme
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

### archive

Move one or more emails to the Archive mailbox.

```bash
jm archive <email-id> [email-id...]
```

1 or more arguments required. No command-specific flags.

**JSON output:**

```json
{
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
Archived: M-email-id-1, M-email-id-2
Destination: Archive (mb-archive-id)
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### spam

Move one or more emails to the Junk/Spam mailbox.

```bash
jm spam <email-id> [email-id...]
```

1 or more arguments required. No command-specific flags.

**JSON output:**

```json
{
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
Marked as spam: M-email-id-1
Destination: Junk Mail (mb-junk-id)
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### mark-read

Mark one or more emails as read by setting the `$seen` keyword.

```bash
jm mark-read <email-id> [email-id...]
```

1 or more arguments required. No command-specific flags.

**JSON output:**

```json
{
  "marked_as_read": ["M-email-id-1", "M-email-id-2"],
  "errors": []
}
```

**Text output:**

```text
Marked as read: M-email-id-1, M-email-id-2
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### flag

Flag one or more emails by setting the `$flagged` keyword.

```bash
jm flag <email-id> [email-id...]
```

1 or more arguments required. No command-specific flags.

**JSON output:**

```json
{
  "flagged": ["M-email-id-1", "M-email-id-2"],
  "errors": []
}
```

**Text output:**

```text
Flagged: M-email-id-1, M-email-id-2
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### unflag

Unflag one or more emails by removing the `$flagged` keyword.

```bash
jm unflag <email-id> [email-id...]
```

1 or more arguments required. No command-specific flags.

**JSON output:**

```json
{
  "unflagged": ["M-email-id-1", "M-email-id-2"],
  "errors": []
}
```

**Text output:**

```text
Unflagged: M-email-id-1, M-email-id-2
```

If some emails fail, the successful ones are still listed and errors appear in the `errors` array. A `partial_failure` error is also written to stderr.

---

### move

Move one or more emails to a specified mailbox by name or ID.

```bash
jm move <email-id> [email-id...] --to <mailbox>
```

1 or more arguments required.

| Flag   | Required | Default | Description               |
| ------ | -------- | ------- | ------------------------- |
| `--to` | yes      | (none)  | Target mailbox name or ID |

**Safety:** The `move` command refuses to target Trash, Deleted Items, or Deleted Messages (by role or name, case-insensitive). Attempting this returns a `forbidden_operation` error.

**JSON output:**

```json
{
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

| Field            | Type            | Notes                                |
| ---------------- | --------------- | ------------------------------------ |
| `moved`          | string[]        | Omitted unless `move` command        |
| `archived`       | string[]        | Omitted unless `archive` command     |
| `marked_as_spam` | string[]        | Omitted unless `spam` command        |
| `marked_as_read` | string[]        | Omitted unless `mark-read` command   |
| `flagged`        | string[]        | Omitted unless `flag` command        |
| `unflagged`      | string[]        | Omitted unless `unflag` command      |
| `destination`    | DestinationInfo | Omitted on total failure             |
| `errors`         | string[]        | Empty array on full success          |

### DestinationInfo

| Field  | Type   | Notes        |
| ------ | ------ | ------------ |
| `id`   | string | Mailbox ID   |
| `name` | string | Mailbox name |

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
Hint: Check your token in JMAP_TOKEN or config file
```

### Error Codes

| Code                    | Description                                         | Example hint                                               |
| ----------------------- | --------------------------------------------------- | ---------------------------------------------------------- |
| `authentication_failed` | Token is missing, invalid, or expired               | Check your token in JMAP_TOKEN or config file              |
| `not_found`             | Email ID or mailbox not found                       | (varies)                                                   |
| `forbidden_operation`   | Attempted a disallowed action (e.g., move to Trash) | Deletion is not permitted by this tool                     |
| `jmap_error`            | Server-side JMAP method error                       | (varies)                                                   |
| `network_error`         | Connection or timeout failure                       | (varies)                                                   |
| `general_error`         | Invalid flag values or other client-side errors     | (varies)                                                   |
| `config_error`          | Malformed config file                               | Fix the syntax in ~/.config/jm/config.yaml or use --config |
| `partial_failure`       | Some IDs in a batch operation failed                | (none)                                                     |

### Cobra Validation Errors

Cobra (the CLI framework) handles argument and flag validation before `jm` commands run. These errors are printed as **plain text to stderr** and do not use the structured JSON/text error format.

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
