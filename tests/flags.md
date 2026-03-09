# fm flag handling

Verify that flags and configuration work correctly.

## Version flag

```scrut
$ $TESTDIR/../fm --version
fm version * (glob)
```

## Default format is JSON

The error output uses JSON format by default.

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Format flag switches to text

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session --format text 2>&1
Error [authentication_failed]: credential command failed: * (glob)
Hint: Check your credential command or the token it returns
[1]
```

## FM_FORMAT env var switches to text

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_ACCOUNT_ID HOME=/nonexistent FM_FORMAT=text $TESTDIR/../fm session 2>&1
Error [authentication_failed]: credential command failed: * (glob)
Hint: Check your credential command or the token it returns
[1]
```

## FM_CREDENTIAL_COMMAND env var is read

When a credential command is set via env var, the error changes from
"credential command failed" to a connection/auth error.

```scrut
$ env -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent FM_CREDENTIAL_COMMAND="echo test-token" $TESTDIR/../fm session 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Credential command flag overrides env var

```scrut
$ env -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent FM_CREDENTIAL_COMMAND="echo env-token" $TESTDIR/../fm session --credential-command "echo flag-token" 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Custom session URL via flag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session --credential-command "echo test" --session-url http://localhost:1/jmap 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## FM_SESSION_URL env var

```scrut
$ env -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent FM_CREDENTIAL_COMMAND="echo test" FM_SESSION_URL=http://localhost:1/jmap $TESTDIR/../fm session 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## List default flags

List with default flags fails on auth, not on flag parsing.

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm list 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## List with all flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm list -m inbox -l 10 -o 5 -u -s "sentAt asc" 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search with all filter flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search "test" --from alice --to bob --subject meeting --before 2026-01-15T00:00:00Z --after 2025-12-01T00:00:00Z --has-attachment -l 10 -m inbox 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search with bare date in --before flag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search --before 2026-02-01 --credential-command "echo test" --session-url http://localhost:1/jmap 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search with bare date in --after flag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search --after 2026-02-01 --credential-command "echo test" --session-url http://localhost:1/jmap 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search with invalid date format

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search --before "not-a-date" --credential-command "echo test" --session-url http://localhost:1/jmap 2>&1
{
  "error": "general_error",
  "message": "invalid --before date: *", (glob)
  "hint": "Use RFC 3339 format (e.g. 2026-01-15T00:00:00Z) or a bare date (e.g. 2026-01-15)"
}
[1]
```

## List with invalid limit

```scrut
$ $TESTDIR/../fm list --limit 0 2>&1
{
  "error": "general_error",
  "message": "--limit must be at least 1"
}
[1]
```

## Read with all flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm read M123 --html --raw-headers --thread 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Mailboxes with roles-only flag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mailboxes --roles-only 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run flag on archive

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive --dry-run M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run short flag on archive

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive -n M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run flag on spam

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm spam --dry-run M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run short flag on spam

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm spam -n M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run flag on mark-read

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mark-read --dry-run M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run short flag on mark-read

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mark-read -n M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run flag on flag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm flag --dry-run M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run short flag on flag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm flag -n M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run flag on unflag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm unflag --dry-run M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run short flag on unflag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm unflag -n M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Flag with --color accepts valid color

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm flag --color orange M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Flag with -c short flag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm flag -c blue M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Flag with invalid color

```scrut
$ $TESTDIR/../fm flag --color magenta M1 2>&1
{
  "error": "general_error",
  "message": "invalid color \"magenta\"; valid colors: red, orange, yellow, green, blue, purple, gray",
  "hint": "Valid colors: red, orange, yellow, green, blue, purple, gray"
}
[1]
```

## Unflag with --color accepts valid color

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm unflag --color M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Unflag with --color and extra argument

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm unflag --color orange M1 2>&1
{
  "error": "authentication_failed",
* (glob+)
}
[1]
```

## Dry-run flag on move

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm move --dry-run M1 --to Receipts 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Dry-run short flag on move

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm move -n M1 --to Receipts 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Archive with multiple IDs

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive M1 M2 M3 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Spam with multiple IDs

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm spam M1 M2 M3 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Mark-read with multiple IDs

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mark-read M1 M2 M3 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Move with multiple IDs

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm move M1 M2 --to Receipts 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Archive with filter flags produces auth error (not flag error)

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive --mailbox inbox --unread 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Spam with filter flags produces auth error

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm spam --mailbox inbox --from notifications@github.com 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Mark-read with filter flags produces auth error

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mark-read --mailbox inbox --unread 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## IDs and filter flags are mutually exclusive

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive M1 --from alice@test.com 2>&1
{
  "error": "general_error",
  "message": "cannot combine email IDs with filter flags",
  "hint": "Use either email IDs or filter flags, not both"
}
[1]
```

## Flagged and unflagged are mutually exclusive on action commands

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive --flagged --unflagged 2>&1
{
  "error": "general_error",
  "message": "--flagged and --unflagged are mutually exclusive"
}
[1]
```
