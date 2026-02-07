# jm flag handling

Verify that flags and configuration work correctly.

## Default format is JSON

The error output uses JSON format by default.

```scrut
$ $TESTDIR/../jm session 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set JMAP_TOKEN, --token, or token in config file",
  "hint": "Check your token in JMAP_TOKEN or config file"
}
[1]
```

## Format flag switches to text

```scrut
$ $TESTDIR/../jm session --format text 2>&1
Error [authentication_failed]: no token configured; set JMAP_TOKEN, --token, or token in config file
Hint: Check your token in JMAP_TOKEN or config file
[1]
```

## JMAP_FORMAT env var switches to text

```scrut
$ JMAP_FORMAT=text $TESTDIR/../jm session 2>&1
Error [authentication_failed]: no token configured; set JMAP_TOKEN, --token, or token in config file
Hint: Check your token in JMAP_TOKEN or config file
[1]
```

## JMAP_TOKEN env var is read

When a token is set via env var, the error changes from "no token" to
a connection/auth error.

```scrut
$ JMAP_TOKEN=test-token $TESTDIR/../jm session 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Token flag overrides env var

```scrut
$ JMAP_TOKEN=env-token $TESTDIR/../jm session --token flag-token 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Custom session URL via flag

```scrut
$ $TESTDIR/../jm session --token test --session-url http://localhost:1/jmap 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## JMAP_SESSION_URL env var

```scrut
$ JMAP_TOKEN=test JMAP_SESSION_URL=http://localhost:1/jmap $TESTDIR/../jm session 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## List default flags

List with default flags fails on auth, not on flag parsing.

```scrut
$ $TESTDIR/../jm list 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## List with all flags

```scrut
$ $TESTDIR/../jm list -m inbox -l 10 -o 5 -u -s "sentAt asc" 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search with all filter flags

```scrut
$ $TESTDIR/../jm search "test" --from alice --to bob --subject meeting --before 2026-01-15T00:00:00Z --after 2025-12-01T00:00:00Z --has-attachment -l 10 -m inbox 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search with invalid date format

```scrut
$ $TESTDIR/../jm search --before "not-a-date" --token test --session-url http://localhost:1/jmap 2>&1
{
  "error": "general_error",
  "message": "invalid --before date: *", (glob)
  "hint": "Use RFC 3339 format, e.g. 2026-01-15T00:00:00Z"
}
[1]
```

## List with invalid limit

```scrut
$ $TESTDIR/../jm list --limit 0 2>&1
{
  "error": "general_error",
  "message": "--limit must be at least 1"
}
[1]
```

## Read with all flags

```scrut
$ $TESTDIR/../jm read M123 --html --raw-headers --thread 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Mailboxes with roles-only flag

```scrut
$ $TESTDIR/../jm mailboxes --roles-only 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Archive with multiple IDs

```scrut
$ $TESTDIR/../jm archive M1 M2 M3 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Spam with multiple IDs

```scrut
$ $TESTDIR/../jm spam M1 M2 M3 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Move with multiple IDs

```scrut
$ $TESTDIR/../jm move M1 M2 --to Receipts 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```
