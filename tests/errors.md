# fm error handling

Verify structured error output when things go wrong.

## Missing token produces structured JSON error

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Missing token with text format

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session --format text 2>&1
Error [authentication_failed]: no token configured; set FM_TOKEN, --token, or token in config file
Hint: Check your token in FM_TOKEN or config file
[1]
```

## List without token

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm list 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Read without token

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm read some-email-id 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Search without token

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search "test query" 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Archive without token

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive M123 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Spam without token

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm spam M123 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Mark-read without token

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mark-read M123 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Move without token

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm move M123 --to Archive 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Mailboxes without token

```scrut
$ env -u FM_TOKEN -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mailboxes 2>&1
{
  "error": "authentication_failed",
  "message": "no token configured; set FM_TOKEN, --token, or token in config file",
  "hint": "Check your token in FM_TOKEN or config file"
}
[1]
```

## Invalid token produces connection error

```scrut
$ env -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session --token "invalid-token-xxx" 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```
