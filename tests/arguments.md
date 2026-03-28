# fm argument validation

Verify that commands correctly validate their arguments.

## Read requires exactly one argument

```scrut
$ $TESTDIR/../fm read 2>&1
Error: accepts 1 arg(s), received 0
[1]
```

## Read rejects multiple arguments

```scrut
$ $TESTDIR/../fm read M1 M2 2>&1
Error: accepts 1 arg(s), received 2
[1]
```

## Archive requires email IDs or filter flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive 2>&1
{
  "error": "general_error",
  "message": "no emails specified",
  "hint": "Provide email IDs as arguments or use filter flags (e.g. --mailbox inbox --unread)"
}
[1]
```

## Spam requires email IDs or filter flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm spam 2>&1
{
  "error": "general_error",
  "message": "no emails specified",
  "hint": "Provide email IDs as arguments or use filter flags (e.g. --mailbox inbox --unread)"
}
[1]
```

## Mark-read requires email IDs or filter flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mark-read 2>&1
{
  "error": "general_error",
  "message": "no emails specified",
  "hint": "Provide email IDs as arguments or use filter flags (e.g. --mailbox inbox --unread)"
}
[1]
```

## Flag requires email IDs or filter flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm flag 2>&1
{
  "error": "general_error",
  "message": "no emails specified",
  "hint": "Provide email IDs as arguments or use filter flags (e.g. --mailbox inbox --unread)"
}
[1]
```

## Unflag requires email IDs or filter flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm unflag 2>&1
{
  "error": "general_error",
  "message": "no emails specified",
  "hint": "Provide email IDs as arguments or use filter flags (e.g. --mailbox inbox --unread)"
}
[1]
```

## Move requires email IDs or filter flags

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm move 2>&1
{
  "error": "general_error",
  "message": "no emails specified",
  "hint": "Provide email IDs as arguments or use filter flags (e.g. --mailbox inbox --unread)"
}
[1]
```

## Move requires --to flag

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm move M123 2>&1
{
  "error": "general_error",
  "message": "required flag \"to\" not set",
  "hint": "Specify the destination mailbox with --to <mailbox>"
}
[1]
```

## Unknown command shows error

```scrut
$ $TESTDIR/../fm delete 2>&1
Error: unknown command "delete" for "fm"
[1]
```

## Unknown flag shows error

```scrut
$ $TESTDIR/../fm list --nonexistent 2>&1
Error: unknown flag: --nonexistent
[1]
```

## Search accepts zero arguments (filter-only search)

Search without a positional query argument should still work (it fails
due to no token, not due to argument validation).

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search --from alice@test.com 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search accepts one argument

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search "test query" --from alice@test.com 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search rejects multiple positional arguments

```scrut
$ $TESTDIR/../fm search "query one" "query two" 2>&1
Error: accepts at most 1 arg(s), received 2
[1]
```
