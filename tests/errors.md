# fm error handling

Verify structured error output when things go wrong.

## Missing credential command produces structured JSON error

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Missing credential command with text format

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session --format text 2>&1
Error [authentication_failed]: credential command failed: * (glob)
Hint: Check your credential command or the token it returns
[1]
```

## List without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm list 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Read without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm read some-email-id 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Search without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm search "test query" 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Archive without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm archive M123 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Spam without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm spam M123 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Mark-read without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mark-read M123 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Move without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm move M123 --to Archive 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Mailboxes without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm mailboxes 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Invalid token produces connection error

```scrut
$ env -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm session --credential-command "echo invalid-token-xxx" 2>&1
{
  "error": "authentication_failed",
* (glob+)
[1]
```
