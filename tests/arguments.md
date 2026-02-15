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

## Archive requires at least one argument

```scrut
$ $TESTDIR/../fm archive 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Spam requires at least one argument

```scrut
$ $TESTDIR/../fm spam 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Mark-read requires at least one argument

```scrut
$ $TESTDIR/../fm mark-read 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Flag requires at least one argument

```scrut
$ $TESTDIR/../fm flag 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Unflag requires at least one argument

```scrut
$ $TESTDIR/../fm unflag 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Move requires at least one argument

```scrut
$ $TESTDIR/../fm move 2>&1
Error: requires at least 1 arg(s), only received 0
[1]
```

## Move requires --to flag

```scrut
$ $TESTDIR/../fm move M123 2>&1
Error: required flag(s) "to" not set
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
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search --from alice@test.com 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```

## Search accepts one argument

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search "test query" --from alice@test.com 2>&1
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
