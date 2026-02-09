# jm argument validation

Verify that commands correctly validate their arguments.

## Read requires exactly one argument

```scrut
$ $TESTDIR/../jm read 2>&1
Error: accepts 1 arg(s), received 0
[1]
```

## Read rejects multiple arguments

```scrut
$ $TESTDIR/../jm read M1 M2 2>&1
Error: accepts 1 arg(s), received 2
[1]
```

## Archive requires at least one argument

```scrut
$ $TESTDIR/../jm archive 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Spam requires at least one argument

```scrut
$ $TESTDIR/../jm spam 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Move requires at least one argument

```scrut
$ $TESTDIR/../jm move 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Move requires --to flag

```scrut
$ $TESTDIR/../jm move M123 2>&1
Error: required flag(s) "to" not set
[1]
```

## Unknown command shows error

```scrut
$ $TESTDIR/../jm delete 2>&1
Error: unknown command "delete" for "jm"
[1]
```

## Unknown flag shows error

```scrut
$ $TESTDIR/../jm list --nonexistent 2>&1
Error: unknown flag: --nonexistent
[1]
```

## Search accepts zero arguments (filter-only search)

Search without a positional query argument should still work (it fails
due to no token, not due to argument validation).

```scrut
$ env -u JMAP_TOKEN -u JMAP_SESSION_URL -u JMAP_FORMAT -u JMAP_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../jm search --from alice@test.com 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search accepts one argument

```scrut
$ env -u JMAP_TOKEN -u JMAP_SESSION_URL -u JMAP_FORMAT -u JMAP_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../jm search "test query" --from alice@test.com 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search rejects multiple positional arguments

```scrut
$ $TESTDIR/../jm search "query one" "query two" 2>&1
Error: accepts at most 1 arg(s), received 2
[1]
```
